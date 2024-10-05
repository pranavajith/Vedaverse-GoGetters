package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewServer(serverAddress string) *Server {
	return &Server{
		serverAddress:   serverAddress,
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan Message),
		userConnections: make(map[string]*websocket.Conn),
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	defer ws.Close()

	var username string
	username = r.URL.Query().Get("username")

	s.mutex.Lock() // Lock for thread safety
	s.clients[ws] = true
	s.userConnections[username] = ws // Store the connection
	s.mutex.Unlock()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(s.clients, ws)
			delete(s.userConnections, username)
			break
		}
		s.broadcast <- msg
	}
}

func (s *Server) getConnectionForUser(username string) *websocket.Conn {
	s.mutex.Lock() // Lock for thread safety
	defer s.mutex.Unlock()

	return s.userConnections[username] // Return the WebSocket connection for the user
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
	// s.questionsCollection = client.Database("game").Collection("questions")
	return nil
}

// Start the server
func (s *Server) Run() {
	http.HandleFunc("/ws", s.corsMiddleware(s.handleConnections))
	go s.handleMessages()

	http.HandleFunc("/user/login", s.corsMiddleware(s.userLoginHandler))
	http.HandleFunc("/users", s.corsMiddleware(s.usersHandler))
	http.HandleFunc("/user/", s.corsMiddleware(s.userHandler))
	http.HandleFunc("/createLobby", s.corsMiddleware(s.createLobbyHandler))
	http.HandleFunc("/searchLobbies", s.corsMiddleware(s.searchLobbiesHandler)) // only send waiting
	http.HandleFunc("/joinLobby", s.corsMiddleware(s.joinLobbyHandler))
	http.HandleFunc("/submitAnswer", s.corsMiddleware(s.submitAnswerHandler))
	// http.HandleFunc("/endGame", s.corsMiddleware(s.endGameHandler))

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
