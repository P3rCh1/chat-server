package models

import "time"

type Room struct {
	CreatorUID int32     `json:"creatorUID"`
	RoomID     int32     `json:"roomID"`
	Name       string    `json:"name"`
	IsPrivate  bool      `json:"isPrivate"`
	CreatedAt  time.Time `json:"createdAt"`
}
