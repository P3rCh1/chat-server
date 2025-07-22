package msg

import (
	"net/http"
)

var (
	InvalidData             = New(http.StatusBadRequest, "invalid data")
	BadEmail                = New(http.StatusBadRequest, "invalid email")
	BadUsername             = New(http.StatusBadRequest, "Username must be 3-20 characters long and contain only letters, numbers or underscores")
	BadPassword             = New(http.StatusBadRequest, "Password must be 8-128 characters long")
	EmptyFields             = New(http.StatusBadRequest, "missing fields")
	ServerError             = New(http.StatusInternalServerError, "server error")
	UserNotFound            = New(http.StatusUnauthorized, "user not found")
	UserAlreadyExist        = New(http.StatusConflict, "user with same name already exists")
	UserOrEmailAlreadyExist = New(http.StatusConflict, "user with same name or email already exists")
	InvalidPassword         = New(http.StatusUnauthorized, "invalid password")
)

type ErrorHTTP struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e ErrorHTTP) Error() string {
	return e.Message
}

func New(code int, message string) ErrorHTTP {
	return ErrorHTTP{
		Code:    code,
		Message: message,
	}
}

func (e ErrorHTTP) Drop(w http.ResponseWriter) {
	http.Error(w, e.Message, e.Code)
}
