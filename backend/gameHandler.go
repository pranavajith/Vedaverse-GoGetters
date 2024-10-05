package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

type Player struct {
	Conn     *websocket.Conn
	Username string
}

type Game struct {
	Players map[string]*Player
	Mutex   sync.Mutex
}

// Global variable for managing the game
var game = Game{
	Players: make(map[string]*Player),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for testing, change in production)
	},
}

func (s *Server) gameHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	username := r.URL.Query().Get("username")
	lobbyID := r.URL.Query().Get("lobbyId") // Assume you're passing the lobby ID in the query string
	if username == "" || lobbyID == "" {
		http.Error(w, "Username and Lobby ID are required", http.StatusBadRequest)
		return
	}

	// Prepare the join request
	joinRequest := LobbyRequest{
		ID:   lobbyID,
		Host: username, // or another host-related username if needed
	}

	// Attempt to join the lobby
	if err := s.joinLobby(joinRequest); err != nil {
		http.Error(w, "Failed to join lobby: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add player to the game
	game.Mutex.Lock()
	game.Players[username] = &Player{Conn: conn, Username: username}
	game.Mutex.Unlock()

	// Notify everyone that a new player has joined
	s.broadcastMessage(fmt.Sprintf("%s has joined the game.", username))

	for {
		// Read messages from the player
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		messageString := string(msg)
		fmt.Println("Received message:", messageString, ", and of type: ", messageType)

		// Broadcast the received message to all players
		s.broadcastMessage(fmt.Sprintf("%s: %s", username, messageString))
	}

	// Remove the player when they leave
	s.removePlayer(username)
}

// New method to encapsulate joining logic
func (s *Server) joinLobby(joinRequest LobbyRequest) error {
	var lobby Lobby
	err := s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": joinRequest.ID}).Decode(&lobby)
	if err != nil || lobby.Status != "available" {
		return fmt.Errorf("Lobby not available")
	}

	lobby.JoinedPlayer = joinRequest.Host // Assuming the Host field should be the player joining
	lobby.Status = "active"

	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), bson.M{"_id": joinRequest.ID}, bson.M{"$set": lobby})
	return err
}

func (s *Server) removePlayer(username string) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	delete(game.Players, username)
	s.broadcastMessage(fmt.Sprintf("%s has left the game.", username))
}

// Broadcast a message to all connected players
func (s *Server) broadcastMessage(message string) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	for username, player := range game.Players {
		if err := player.Conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Printf("Error sending message to %s: %v", username, err)
			player.Conn.Close()            // Close connection if there's an error
			delete(game.Players, username) // Remove player
		}
	}
}
