package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/P3rCh1/chat-server/internal/pkg/msg"
	"github.com/P3rCh1/chat-server/internal/utils"
)

func JWTAuth(log *slog.Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http-server.handlers.middleware.auth.JWTAuth"
		token := r.Header.Get("Authorization")
		userID, err := utils.VerifyJWT(token)
		if err != nil {
			msg.UserNotFound.DropWithLog(w, log, op)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
