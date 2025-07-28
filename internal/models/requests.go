package models

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"Password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type WSRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type CreateRoomRequest struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

type InviteRequest struct {
	UserID int `json:"user_id"`
	RoomID int `json:"room_id"`
}
