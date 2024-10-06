package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"



	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

)

// Define types for the User and UserRequest structures




// Define types for the Lobby, Question, and GameState structures




type GameState struct {
	LobbyID         string         `json:"lobbyId"`
	CurrentQuestion int            `json:"currentQuestion"`
	Scores          map[string]int `json:"scores"`
}

// Removed duplicate Message struct definition

// Define the Message type


func NewServer(serverAddress string) *Server {
	return &Server{
		serverAddress: serverAddress,
		clients:       make(map[*websocket.Conn]bool),
		broadcast:     make(chan Message),
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	s.clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(s.clients, ws)
			break
		}
		s.broadcast <- msg
	}
}

func (s *Server) handleMessages() {
	for {
		msg := <-s.broadcast
		for client := range s.clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(s.clients, client)
			}
		}
	}
}

// MongoDB connection setup
func (s *Server) ConnectMongoDB() error {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = os.Getenv("MONGO_URI_LOCAL")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	// Test the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}

	s.mongoClient = client
	s.usersCollection = client.Database("game").Collection("users")
	s.lobbiesCollection = client.Database("game").Collection("lobbies")
	s.questionsCollection = client.Database("game").Collection("questions")
	return nil
}

// Handle searching for lobbies


// Handle joining a lobby


// Handle submitting an answer
func (s *Server) submitAnswerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		LobbyID    string `json:"lobbyId"`
		Username   string `json:"username"`
		Answer     string `json:"answer"`
		QuestionID string `json:"questionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var lobby Lobby
	err := s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": requestData.LobbyID}).Decode(&lobby)
	if err != nil {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	var question Question
	err = s.questionsCollection.FindOne(context.TODO(), bson.M{"_id": requestData.QuestionID}).Decode(&question)
	if err != nil {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	if question.CorrectAnswer == requestData.Answer {
		// Update the score
		if lobby.Scores == nil {
			lobby.Scores = make(map[string]int)
		}
		lobby.Scores[requestData.Username]++
	}

	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), bson.M{"_id": requestData.LobbyID}, bson.M{"$set": bson.M{"scores": lobby.Scores}})
	if err != nil {
		http.Error(w, "Failed to submit answer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lobby)
}

// Handle ending a game
func (s *Server) endGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		LobbyID string `json:"lobbyId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var lobby Lobby
	err := s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": requestData.LobbyID}).Decode(&lobby)
	if err != nil {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	lobby.Status = "ended"

	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), bson.M{"_id": requestData.LobbyID}, bson.M{"$set": bson.M{"status": lobby.Status}})
	if err != nil {
		http.Error(w, "Failed to end game", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lobby)
}

// Handle creating a lobby


// Start the server
func (s *Server) Run() {
	http.HandleFunc("/ws", s.corsMiddleware(s.handleConnections))
	go s.handleMessages()

	http.HandleFunc("/user/login", s.corsMiddleware(s.userLoginHandler))
	http.HandleFunc("/users", s.corsMiddleware(s.usersHandler))
	http.HandleFunc("/user/", s.corsMiddleware(s.userHandler))
	http.HandleFunc("/createLobby", s.corsMiddleware(s.createLobbyHandler))
	http.HandleFunc("/searchLobbies", s.corsMiddleware(s.searchLobbiesHandler))
	http.HandleFunc("/joinLobby", s.corsMiddleware(s.joinLobbyHandler))
	http.HandleFunc("/submitAnswer", s.corsMiddleware(s.submitAnswerHandler))
	http.HandleFunc("/endGame", s.corsMiddleware(s.endGameHandler))

	fmt.Println("Server running at", s.serverAddress)
	log.Fatal(http.ListenAndServe(s.serverAddress, nil))
}

// CORS middleware
func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

