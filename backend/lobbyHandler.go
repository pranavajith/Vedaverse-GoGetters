package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (s *Server) lobbyHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/lobby/create":
		s.createLobbyHandler(w, r)
	case "/lobby/search/all":
		s.searchLobbiesHandler(w, r)
	case "/lobby/join":
		s.joinLobbyHandler(w, r)
	case "/lobby/end":
		s.endGameHandler(w, r)
	}
}

// Create a new lobby
func (s *Server) createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newLobby Lobby
	if err := json.NewDecoder(r.Body).Decode(&newLobby); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Assign ID and status
	newLobby.ID = "lobby-" + time.Now().Format("20060102150405")
	newLobby.Status = "available"

	_, err := s.lobbiesCollection.InsertOne(context.TODO(), newLobby)
	if err != nil {
		http.Error(w, "Failed to create lobby", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newLobby)
}

// Search for available lobbies
func (s *Server) searchLobbiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cursor, err := s.lobbiesCollection.Find(context.TODO(), bson.M{"status": "available"})
	if err != nil {
		http.Error(w, "Failed to retrieve lobbies", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var lobbies []Lobby
	if err := cursor.All(context.TODO(), &lobbies); err != nil {
		http.Error(w, "Failed to decode lobbies", http.StatusInternalServerError)
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

// Join an existing lobby
func (s *Server) joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var joinRequest struct {
		LobbyID  string `json:"lobbyId"`  // This should match how clients send the ID
		Username string `json:"username"` // Username of the joining player
	}

	if err := json.NewDecoder(r.Body).Decode(&joinRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var lobby Lobby
	err := s.lobbiesCollection.FindOne(context.TODO(), bson.M{"_id": joinRequest.LobbyID}).Decode(&lobby)
	if err != nil || lobby.Status != "available" {
		http.Error(w, "Lobby not available", http.StatusNotFound)
		return
	}

	lobby.JoinedPlayer = joinRequest.Username
	lobby.Status = "active"

	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), bson.M{"_id": joinRequest.LobbyID}, bson.M{"$set": lobby})
	if err != nil {
		http.Error(w, "Failed to join lobby", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lobby)
}

// End game handler
func (s *Server) endGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var endRequest struct {
		LobbyID string `json:"lobbyId"`
		Winner  string `json:"winner"`
	}

	if err := json.NewDecoder(r.Body).Decode(&endRequest); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := s.lobbiesCollection.UpdateOne(context.TODO(), bson.M{"id": endRequest.LobbyID}, bson.M{"$set": bson.M{"status": "closed", "winner": endRequest.Winner}})
	if err != nil {
		http.Error(w, "Failed to end game", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Game ended successfully"))
}
