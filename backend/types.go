package main

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Lobby struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	HostPlayer   string    `json:"hostPlayer"`
	Questions    []string  `json:"questions"`
	JoinedPlayer string    `json:"joinedPlayer"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	Winner       string    `json:"winner"`
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

type LobbyRequest struct {
	ID        string     `json:"id" bson:"_id,omitempty"`
	Host      string     `json:"host"`      // Username of the user who created the lobby
	Questions []Question `json:"questions"` // Questions for the game
}

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

type Server struct {
	serverAddress     string
	mongoClient       *mongo.Client
	usersCollection   *mongo.Collection
	lobbiesCollection *mongo.Collection // New collection for lobbies
	mutex             sync.Mutex
}
