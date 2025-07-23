package models

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
)

type Tools struct {
	DB            *sql.DB
	TokenProvider tokens.TokenProvider
	Log           *slog.Logger
}

type Profile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created-at"`
}

type Room struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	CreatorID int    `json:"creator_id"`
}

type Message struct {
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Text      *string   `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
