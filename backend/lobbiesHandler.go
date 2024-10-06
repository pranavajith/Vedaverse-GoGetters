package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

// Handle searching for lobbies
func (s *Server) searchLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	cursor, err := s.lobbiesCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve lobbies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var lobbies []Lobby
	if err = cursor.All(context.TODO(), &lobbies); err != nil {
		http.Error(w, "Failed to decode lobbies list", http.StatusInternalServerError)
		return
	}

	lobbiesJSON, err := json.Marshal(lobbies)
	if err != nil {
		http.Error(w, "Failed to marshal lobbies list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(lobbiesJSON)
}


// Handle creating a lobby
func (s *Server) createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lobby Lobby
	if err := json.NewDecoder(r.Body).Decode(&lobby); err != nil {
		log.Printf("Error decoding request payload: %v", err) // Log the error for debugging
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	lobby.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	lobby.CreatedAt = time.Now()
	lobby.Status = "waiting"
	lobby.Participants = append(lobby.Participants, lobby.Creator)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, err := s.lobbiesCollection.InsertOne(context.TODO(), lobby)
	if err != nil {
		http.Error(w, "Failed to create lobby", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lobby)
}

// Handle joining a lobby
func (s *Server) joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req JoinLobbyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var lobby Lobby
	err = s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": req.LobbyID}).Decode(&lobby)
	if err != nil {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	if lobby.Status != "waiting" || len(lobby.Participants) >= 2 {
		http.Error(w, "Lobby is either full or not active", http.StatusForbidden)
		return
	}

	// Add the user to the participants list
	lobby.Participants = append(lobby.Participants, req.Username)

	// Update the lobby in the database
	filter := bson.M{"_id": req.LobbyID}
	update := bson.M{
		"$set": bson.M{
			"participants": lobby.Participants,
		},
	}
	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update lobby", http.StatusInternalServerError)
		return
	}

	// If there are now 2 participants, establish WebSocket connections and start the game

	if len(lobby.Participants) == 2 {
		// Call a method to handle the WebSocket connection for the participants
		s.startGame(lobby)
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// New method to start the game
func (s *Server) startGame(lobby Lobby) {
	// Get the question at the current index
	if lobby.CurrentIndex >= len(lobby.Questions) {
		log.Println("No questions available in lobby")
		return
	}

	// question := lobby.Questions[lobby.CurrentIndex]

	// Set up WebSocket connections for both participants
	// for _, username := range lobby.Participants {
	// 	// Handle WebSocket connections for both participants
	// 	// wsConn := s.getConnectionForUser(username) // Implement this method to get the connection
	// 	go s.sendQuestion(wsConn, question, lobby) // Pass the lobby as a parameter
	// }
}

func (s *Server) sendQuestion(ws *websocket.Conn, question Question, lobby Lobby) {
	err := ws.WriteJSON(question)
	if err != nil {
		log.Println("Failed to send question:", err)
		return
	}

	// Create a channel to wait for the answer
	answerChannel := make(chan Answer, 1)
	go s.waitForAnswer(ws, answerChannel)

	select {
	case answer := <-answerChannel:
		// Process the answer
		s.submitAnswer(lobby.ID, answer.Username, answer.Answer)
		// Send the next question
		s.sendNextQuestion(ws, lobby)
	case <-time.After(30 * time.Second):
		log.Println("Timeout waiting for answer")
	}
}

func (s *Server) waitForAnswer(ws *websocket.Conn, answerChannel chan Answer) {
	var answer Answer
	err := ws.ReadJSON(&answer)
	if err != nil {
		log.Println("Failed to read answer:", err)
		return
	}
	answerChannel <- answer
}

func (s *Server) submitAnswer(lobbyID string, username string, answer string) {
	// Locking to prevent concurrent writes to lobby scores
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Retrieve the lobby
	var lobby Lobby
	err := s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": lobbyID}).Decode(&lobby)
	if err != nil {
		log.Println("Lobby not found:", err)
		return
	}

	if lobby.Status != "active" || lobby.CurrentIndex >= len(lobby.Questions) {
		log.Println("Lobby is not active")
		return
	}

	// Find the question in the lobby questions by questionID
	correctAnswer := lobby.Questions[lobby.CurrentIndex].CorrectAnswer

	// Check if the answer matches
	if correctAnswer == answer {
		// Answer is correct; add 10 points
		lobby.Scores[username] += 10
	} else {
		// Answer is incorrect; subtract 10 points
		lobby.Scores[username] -= 10
	}

	// Update the scores in the database
	_, err = s.lobbiesCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": lobbyID},
		bson.M{"$set": bson.M{"scores": lobby.Scores}},
	)
	if err != nil {
		log.Println("Failed to update scores:", err)
	}
}

// Sending the next question after the answer is processed
func (s *Server) sendNextQuestion(ws *websocket.Conn, lobby Lobby) {
	lobby.CurrentIndex++
	if lobby.CurrentIndex < len(lobby.Questions) {
		question := lobby.Questions[lobby.CurrentIndex]
		s.sendQuestion(ws, question, lobby)
	} else {
		// Game over logic
		s.endGame(lobby)
	}
}

func (s *Server) endGame(lobby Lobby) {
	// Update scores in the database
	for username, score := range lobby.Scores {
		_, err := s.usersCollection.UpdateOne(
			context.TODO(),
			bson.M{"username": username},
			bson.M{"$inc": bson.M{"multiPlayerScore": score}},
		)
		if err != nil {
			log.Println("Failed to update user scores:", err)
		}
	}

	// Update lobby status to ended
	lobby.Status = "ended"
	filter := bson.M{"_id": lobby.ID}
	_, err := s.lobbiesCollection.UpdateOne(context.TODO(), filter, bson.M{"$set": bson.M{"status": lobby.Status}})
	if err != nil {
		log.Println("Failed to update lobby status:", err)
	}
}