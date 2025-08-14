package models

import "time"

type Room struct {
	CreatorUID int64     `json:"CreatorUID"`
	RoomID     int64     `json:"RoomID"`
	Name       string    `json:"Name"`
	IsPrivate  bool      `json:"IsPrivate"`
	CreatedAt  time.Time `json:"CreatedAt"`
}

type Message struct {
	ID        int64     `json:"ID"`
	RoomID    int64     `json:"RoomID"`
	UID       int64     `json:"UID"`
	Type      string    `json:"Type"`
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
}
