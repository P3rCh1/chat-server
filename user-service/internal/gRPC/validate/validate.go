package validate

import (
	"regexp"
	"unicode"

	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
)

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func Email(email string) bool {
	return emailRegexp.MatchString(email)
}

func Name(username string) bool {
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return false
		}
	}
	return true
}

func Password(password string) bool {
	if len(password) < 8 || len(password) > 128 {
		return false
	}
	return true
}

func Register(username, email, password string) error {
	switch {
	case !Name(username):
		return status_error.InvalidUsername
	case !Email(email):
		return status_error.InvalidEmail
	case !Password(password):
		return status_error.InvalidPassword
	}
	return nil
}
