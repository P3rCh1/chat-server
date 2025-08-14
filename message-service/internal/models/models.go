package models

import (
	"time"
)

type Message struct {
	ID        int64     `json:"ID"`
	RoomID    int64     `json:"RoomID"`
	UID       int64     `json:"UID"`
	Type      string    `json:"Type"`
	Text      string    `json:"Text"`
	Timestamp time.Time `json:"Timestamp"`
}
