package responses

import (
	"net/http"
)

var (
	InvalidData             = New(http.StatusBadRequest, "invalid data")
	BadEmail                = New(http.StatusBadRequest, "invalid email")
	BadName                 = New(http.StatusBadRequest, "name must be 3-20 characters long and contain only letters, numbers or underscores")
	BadPassword             = New(http.StatusBadRequest, "password must be 8-128 characters long")
	EmptyFields             = New(http.StatusBadRequest, "missing fields")
	ServerError             = New(http.StatusInternalServerError, "server error")
	UserNotFound            = New(http.StatusUnauthorized, "user not found")
	InvalidToken            = New(http.StatusUnauthorized, "invalid token")
	NewNameMatchesCur       = New(http.StatusBadRequest, "new name matches the current")
	UserAlreadyExist        = New(http.StatusConflict, "user with same name already exists")
	UserOrEmailAlreadyExist = New(http.StatusConflict, "user with same name or email already exists")
	InvalidPassword         = New(http.StatusUnauthorized, "invalid password")
	RoomNotFound            = New(http.StatusUnauthorized, "room not found")
	RoomIsPrivate           = New(http.StatusUnauthorized, "room is private")
	RoomAlreadyExist        = New(http.StatusConflict, "room with same name already exists")
	NoAccessToRoom          = New(http.StatusForbidden, "only creator can invite to room")
	AlreadyInRoom           = New(http.StatusConflict, "user is already a member of room")
	AuthFail                = New(http.StatusUnauthorized, "authorization failed")
	InvalidURL              = New(http.StatusNotFound, "invalid URL")
	NotRoomMember           = New(http.StatusNotFound, "user not a member of room")
)

type ErrorHTTP interface {
	Error() string
	Code() int
	Drop(w http.ResponseWriter)
}

type MyErrorHTTP struct {
	HTTPCode int    `json:"code"`
	Message  string `json:"message"`
}

func (e MyErrorHTTP) Error() string {
	return e.Message
}

func New(code int, message string) MyErrorHTTP {
	return MyErrorHTTP{
		HTTPCode: code,
		Message:  message,
	}
}

func (e MyErrorHTTP) Code() int {
	return e.HTTPCode
}

func (e MyErrorHTTP) Drop(w http.ResponseWriter) {
	http.Error(w, e.Message, e.HTTPCode)
}
