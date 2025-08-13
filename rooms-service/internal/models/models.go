package models

import "time"

type Room struct {
	CreatorUID int64     `json:"creatorUID"`
	RoomID     int64     `json:"roomID"`
	Name       string    `json:"name"`
	IsPrivate  bool      `json:"isPrivate"`
	CreatedAt  time.Time `json:"createdAt"`
}
