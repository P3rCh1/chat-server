package models

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/P3rCh1/chat-server/internal/config"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
	"github.com/gorilla/websocket"
)

type Tools struct {
	DB            *sql.DB
	TokenProvider tokens.TokenProvider
	Log           *slog.Logger
	WSUpgrader    *websocket.Upgrader
	Cfg           *config.Config
	Pkg           *Package
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
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	UserID   int
	Username string
	RoomID   int
}

type Package struct {
	SystemUserID int
	ErrorUserID  int
}
