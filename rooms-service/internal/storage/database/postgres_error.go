package database

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

const (
	alreadyExistsCode   = "23505"
	foreignKeyViolation = "23503"
)

func ExpectedPGErr(err, notFound, exists error) error {
	var pqErr *pq.Error
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	if errors.As(err, &pqErr) {
		if pqErr.Code == alreadyExistsCode {
			return exists
		}
		if pqErr.Code == foreignKeyViolation {
			return notFound
		}
	}
	return nil
}
