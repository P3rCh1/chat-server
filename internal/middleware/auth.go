package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/P3rCh1/chat-server/internal/utils"
)

func JWTAuth(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}
		token = strings.TrimPrefix(token, "Bearer ")
		userID, err := utils.VerifyJWT(token)
		if err != nil {
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
