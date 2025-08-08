package database

import (
	"errors"
	"strings"

	"github.com/P3rCh1/chat-server/user-service/internal/gRPC/status_error"
	"github.com/lib/pq"
)

const (
	alreadyExistsCode = "23505"
)

func isAlreadyExists(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == alreadyExistsCode
	}
	return false
}

func AsUsernameOrEmailExistsErr(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == alreadyExistsCode {
		if strings.Contains(pqErr.Message, "username") {
			return status_error.NameExists
		}
		if strings.Contains(pqErr.Message, "email") {
			return status_error.EmailExists
		}
	}
	return nil
}
