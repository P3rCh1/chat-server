package middleware

import (
	"context"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/pkg/tokens"
)

func Auth(jwt tokens.TokenProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			userID, err := jwt.Verify(token)
			if err != nil {
				responses.AuthFail.Drop(w)
				return
			}
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
