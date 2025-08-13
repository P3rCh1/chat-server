package models

import "time"

type Profile struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
}
