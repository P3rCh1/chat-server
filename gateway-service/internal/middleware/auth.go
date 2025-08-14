package middleware

import (
	"context"
	"net/http"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	sessionpb "github.com/P3rCh1/chat-server/gateway-service/pkg/proto/gen/go/session"
)

const UIDContextKey = "UID"

func Auth(s *gateway.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.Session)
			defer cancel()
			uid, err := s.Session.Verify(ctx, &sessionpb.VerifyRequest{Token: token})
			if err != nil {
				responses.GatewayGRPCErr(w, s.Log, "auth", err)
				return
			}
			ctx = context.WithValue(r.Context(), UIDContextKey, uid.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
