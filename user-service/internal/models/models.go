package models

import "time"

type Profile struct {
	ID        int
	Username  string
	Email     string
	CreatedAt time.Time
}
