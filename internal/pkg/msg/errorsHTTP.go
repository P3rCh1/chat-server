package msg

import (
	"log/slog"
	"net/http"
)

var (
	InvalidData             = New(http.StatusBadRequest, "invalid data")
	BadEmail                = New(http.StatusBadRequest, "invalid email")
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

func (e ErrorHTTP) DropWithLog(w http.ResponseWriter, log *slog.Logger, logInfo string) {
	log.Info(
		logInfo,
		"code", e.Code,
		"message", e.Message,
	)
	e.Drop(w)
}
