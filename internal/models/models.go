package models

import "time"

type UserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"Password"`
}

type Profile struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created-at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Room struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	CreatorID int    `json:"creator_id"`
}

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"room_id"`
	UserID    int       `json:"user_id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
