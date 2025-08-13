package models

import (
	"time"
)

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomID"`
	UID       int64     `json:"uid"`
	Type      string    `json:"type"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}
