package validate

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func joinInvalidArgs(errs ...error) error {
	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			status, _ := status.FromError(err)
			msgs = append(msgs, status.Message())
		}
	}
	return status.Error(codes.InvalidArgument, strings.Join(msgs, "; "))
}

func Email(email string) error {
	if email == "" {
		return status_error.EmptyEmail
	}
	if emailRegexp.MatchString(email) {
		return nil
	}
	return status_error.InvalidEmail
}

func Name(username string) error {
	if username == "" {
		return status_error.EmptyUsername
	}
	var err error
	if len(username) < 3 || len(username) > 20 {
		err = status_error.InvalidUsernameLen
	}
	for _, char := range username {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			err = joinInvalidArgs(err, status_error.InvalidUsernameChr)
		}
	}
	return err
}

func Password(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return status_error.InvalidPassword
	}
	return nil
}

func Register(username, email, password string) error {
	errs := []error{Name(username), Email(email), Password(password)}
	return joinInvalidArgs(errs...)
}
