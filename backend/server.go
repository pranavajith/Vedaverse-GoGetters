package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Define types for the User and UserRequest structures
type CompletedLevel struct {
	LevelID int `json:"levelId"`
	Score   int `json:"score"`
}

type ProfileImage struct {
	Format string `json:"format"`
	Path   string `json:"path"`
}

type StreakDataType struct {
	LatestPlayed          time.Time `json:"latestPlayed"`
	LatestStreakStartDate time.Time `json:"latestStreakStartDate"`
}

type StreakDataRequestType struct {
	LatestPlayed          string `json:"latestPlayed"`
	LatestStreakStartDate string `json:"latestStreakStartDate"`
}

type UserRequest struct {
	FirstName        string                `json:"firstName"`
	LastName         string                `json:"lastName"`
	Username         string                `json:"username"`
	Email            string                `json:"email"`
	DOB              string                `json:"dob"`
	CompletedLevels  []CompletedLevel      `json:"completedLevels"`
	MultiPlayerScore int                   `json:"multiPlayerScore"`
	Password         string                `json:"password"`
	StreakData       StreakDataRequestType `json:"streakData"`
	UserProfileImage ProfileImage          `json:"userProfileImage"`
	OngoingLevel     float64               `json:"ongoingLevel"`
}

type User struct {
	FirstName        string           `json:"firstName"`
	LastName         string           `json:"lastName"`
	Username         string           `json:"username"`
	Email            string           `json:"email"`
	DOB              time.Time        `json:"dob"`
	CompletedLevels  []CompletedLevel `json:"completedLevels"`
	OngoingLevel     float64          `json:"ongoingLevel"`
	MultiPlayerScore int              `json:"multiPlayerScore"`
	PasswordHash     string           `json:"-"` // password is excluded from JSON
	StreakData       StreakDataType   `json:"streakData"`
	UserProfileImage ProfileImage     `json:"userProfileImage"`
}

// Define types for the Lobby, Question, and GameState structures
type Lobby struct {
	ID           string         `json:"id" bson:"_id,omitempty"`
	Creator      string         `json:"creator"`
	Questions    []string       `json:"questions"`
	Participants []string       `json:"participants"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	Scores       map[string]int `json:"scores"` // Add Scores field
}

type Question struct {
	ID            string   `json:"id" bson:"_id,omitempty"`
	QuestionText  string   `json:"questionText"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
}

type GameState struct {
	LobbyID         string         `json:"lobbyId"`
	CurrentQuestion int            `json:"currentQuestion"`
	Scores          map[string]int `json:"scores"`
}

// Removed duplicate Message struct definition
type Server struct {
	serverAddress       string
	mongoClient         *mongo.Client
	usersCollection     *mongo.Collection
	lobbiesCollection   *mongo.Collection
	questionsCollection *mongo.Collection
	mutex               sync.Mutex // Add a mutex for concurrency safety
	clients             map[*websocket.Conn]bool
	broadcast           chan Message
}

// Define the Message type
type Message struct {
	LobbyID string `json:"lobbyId"`
	UserID  string `json:"userId"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

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

// Handle joining a lobby
func (s *Server) joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LobbyID  string `json:"lobby_id"`
		Username string `json:"username"`
	}
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

	if lobby.Status != "waiting" {
		http.Error(w, "Lobby full", http.StatusForbidden)
		return
	}

	// Add the user to the participants list
	lobby.Participants = append(lobby.Participants, req.Username)

	lobby.Status = "active"

	// Update the lobby in the database
	filter := bson.M{"_id": req.LobbyID}
	update := bson.M{
		"$set": bson.M{
			"participants": lobby.Participants,
			"status":       lobby.Status,
		},
	}
	_, err = s.lobbiesCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		http.Error(w, "Failed to update lobby", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

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
func (s *Server) createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var lobby Lobby
	if err := json.NewDecoder(r.Body).Decode(&lobby); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	lobby.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	lobby.CreatedAt = time.Now()
	lobby.Status = "waiting"

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

// Handle user login
func (s *Server) userLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var user User
	err := s.usersCollection.FindOne(context.TODO(), bson.M{"username": requestData.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if !CheckPasswordHash(requestData.Password, user.PasswordHash) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create a response struct with formatted DOB
	type UserResponse struct {
		FirstName        string           `json:"firstName"`
		LastName         string           `json:"lastName"`
		Username         string           `json:"username"`
		Email            string           `json:"email"`
		DOB              string           `json:"dob"`
		CompletedLevels  []CompletedLevel `json:"completedLevels"`
		MultiPlayerScore int              `json:"multiPlayerScore"`
		StreakData       StreakDataType   `json:"streakData"`
		UserProfileImage ProfileImage     `json:"userProfileImage"`
		OngoingLevel     float64          `json:"ongoingLevel"`
	}

	// Remove password hash before sending user info
	user.PasswordHash = ""

	// Format DOB to yyyy-mm-dd
	formattedDOB := user.DOB.Format("2006-01-02")

	userResponse := UserResponse{
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Username:         user.Username,
		Email:            user.Email,
		DOB:              formattedDOB, // Use the formatted DOB here
		CompletedLevels:  user.CompletedLevels,
		MultiPlayerScore: user.MultiPlayerScore,
		StreakData:       user.StreakData,
		UserProfileImage: user.UserProfileImage,
		OngoingLevel:     user.OngoingLevel,
	}

	userJSON, err := json.Marshal(userResponse)
	if err != nil {
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

// Handle getting and adding users
func (s *Server) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGetUsers(w)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Retrieve all users
func (s *Server) handleGetUsers(w http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	cursor, err := s.usersCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var users []User
	if err = cursor.All(context.TODO(), &users); err != nil {
		http.Error(w, "Failed to decode users list", http.StatusInternalServerError)
		return
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		http.Error(w, "Failed to marshal users list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(usersJSON)
}

// Update the main route handler to include password change
func (s *Server) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleGetUser(w, r)
	case "POST":
		path := r.URL.Path
		if path == "/user/add" {
			s.handleAddUser(w, r)
		} else if path == "/user/modify" {
			fmt.Println("userHandler has been touched")
			s.handleModifyUser(w, r)
		} else if path == "/user/change-password" {
			s.handleChangePassword(w, r)
		} else {
			http.Error(w, "Invalid POST path", http.StatusNotFound)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Retrieve a single user by username
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var user User
	err := s.usersCollection.FindOne(context.TODO(), bson.M{"username": requestData.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.PasswordHash = ""
	userData, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to encode user data", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userData)
}

// Hash password for storage
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check if a provided password matches the stored hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Add a new user
func (s *Server) handleAddUser(w http.ResponseWriter, r *http.Request) {
	var newUserReq UserRequest
	if err := json.NewDecoder(r.Body).Decode(&newUserReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	count, err := s.usersCollection.CountDocuments(context.TODO(), bson.M{"username": newUserReq.Username})
	if err != nil {
		http.Error(w, "Error checking username uniqueness", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	dob, err := time.Parse("2006-01-02", newUserReq.DOB)
	if err != nil {
		http.Error(w, "Invalid DOB format, should be YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	passwordHash, err := HashPassword(newUserReq.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Parse streak data dates
	latestPlayed, _ := time.Parse("2006-01-02", newUserReq.StreakData.LatestPlayed)
	latestStreakStartDate, _ := time.Parse("2006-01-02", newUserReq.StreakData.LatestStreakStartDate)

	newUser := User{
		FirstName:        newUserReq.FirstName,
		LastName:         newUserReq.LastName,
		Username:         newUserReq.Username,
		Email:            newUserReq.Email,
		DOB:              dob,
		CompletedLevels:  newUserReq.CompletedLevels,
		MultiPlayerScore: newUserReq.MultiPlayerScore,
		PasswordHash:     passwordHash,
		StreakData: StreakDataType{
			LatestPlayed:          latestPlayed,
			LatestStreakStartDate: latestStreakStartDate,
		},
		UserProfileImage: newUserReq.UserProfileImage,
		OngoingLevel:     newUserReq.OngoingLevel,
	}

	_, err = s.usersCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User added successfully"))
}

// Modify an existing user (for general updates without password change)
func (s *Server) handleModifyUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleModifyUser has been touched")
	var modifyUserReq UserRequest
	if err := json.NewDecoder(r.Body).Decode(&modifyUserReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var user User
	fmt.Println("searching for user")
	err := s.usersCollection.FindOne(context.TODO(), bson.M{"username": modifyUserReq.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	fmt.Println("user found")
	// Update the profile fields (excluding password)
	if modifyUserReq.FirstName != "" {
		user.FirstName = modifyUserReq.FirstName
	}
	if modifyUserReq.LastName != "" {
		user.LastName = modifyUserReq.LastName
	}
	if modifyUserReq.Email != "" {
		user.Email = modifyUserReq.Email
	}
	if modifyUserReq.OngoingLevel > user.OngoingLevel {
		user.OngoingLevel = modifyUserReq.OngoingLevel
	}
	if modifyUserReq.DOB != "" {
		dob, err := time.Parse("2006-01-02", modifyUserReq.DOB)
		if err != nil {
			http.Error(w, "Invalid DOB format, should be YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		user.DOB = dob
	}
	// fmt.Println("DOB issue")
	if len(modifyUserReq.CompletedLevels) > 0 {
		user.CompletedLevels = modifyUserReq.CompletedLevels
	}
	if modifyUserReq.MultiPlayerScore > 0 {
		user.MultiPlayerScore = modifyUserReq.MultiPlayerScore
	}

	// Parse streak data
	if modifyUserReq.StreakData.LatestPlayed != "" {
		latestPlayed, _ := time.Parse("2006-01-02", modifyUserReq.StreakData.LatestPlayed)
		user.StreakData.LatestPlayed = latestPlayed
	}
	if modifyUserReq.StreakData.LatestStreakStartDate != "" {
		latestStreakStartDate, _ := time.Parse("2006-01-02", modifyUserReq.StreakData.LatestStreakStartDate)
		user.StreakData.LatestStreakStartDate = latestStreakStartDate
	}

	if modifyUserReq.UserProfileImage.Path != "" {
		user.UserProfileImage.Path = modifyUserReq.UserProfileImage.Path
	}
	if modifyUserReq.UserProfileImage.Format != "" {
		user.UserProfileImage.Format = modifyUserReq.UserProfileImage.Format
	}

	// Update the user document in the database
	_, err = s.usersCollection.UpdateOne(context.TODO(), bson.M{"username": modifyUserReq.Username}, bson.M{"$set": user})
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// fmt.Println("Updation issue")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User modified successfully"))
}

// Change password handler
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var passwordChangeReq struct {
		Username        string `json:"username"`
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&passwordChangeReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var user User
	err := s.usersCollection.FindOne(context.TODO(), bson.M{"username": passwordChangeReq.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the current password is correct
	if !CheckPasswordHash(passwordChangeReq.CurrentPassword, user.PasswordHash) {
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Hash the new password
	newPasswordHash, err := HashPassword(passwordChangeReq.NewPassword)
	if err != nil {
		http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
		return
	}

	// Update the password in the database
	_, err = s.usersCollection.UpdateOne(context.TODO(), bson.M{"username": passwordChangeReq.Username}, bson.M{"$set": bson.M{"passwordHash": newPasswordHash}})
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Password changed successfully"))
}
