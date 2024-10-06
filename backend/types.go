package main

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
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
	Questions    []Question     `json:"questions"`
	Participants []string       `json:"participants"`
	Status       string         `json:"status"`
	CreatedAt    time.Time      `json:"createdAt"`
	Scores       map[string]int `json:"scores"`       // Add Scores field
	CurrentIndex int            `json:"currentIndex"` // Add CurrentQuestion field
}

type Question struct {
	ID            string   `json:"id" bson:"_id,omitempty"`
	QuestionText  string   `json:"questionText"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
}

// Removed duplicate Message struct definition
type Server struct {
	serverAddress     string
	mongoClient       *mongo.Client
	usersCollection   *mongo.Collection
	lobbiesCollection *mongo.Collection
	questionsCollection *mongo.Collection
	mutex           sync.Mutex // Add a mutex for concurrency safety
	clients         map[*websocket.Conn]bool
	broadcast       chan Message
	userConnections map[string]*websocket.Conn
}

// Define the Message type
type Message struct {
	LobbyID    string `json:"lobbyId"`
	Username   string `json:"username"`
	Answer     string `json:"answer"`
	QuestionID string `json:"questionId"`
	Action     string `json:"action"`
}

type Answer struct {
	Username string `json:"username"`
	Answer   string `json:"answer"`
}

type JoinLobbyRequest struct {
	LobbyID  string `json:"lobby_id"`
	Username string `json:"username"`
}