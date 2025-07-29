package models

import (
	"time"
)

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
