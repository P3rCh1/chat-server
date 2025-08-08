package middleware

import (
	"context"
	"net/http"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	sessionpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/session"
)

func Auth(s *gateway.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.Session)
			defer cancel()
			uid, err := s.Session.Verify(ctx, &sessionpb.VerifyRequest{Token: token})
			if err != nil {
				responses.GatewayGRPCErr(w, s.Log, err)
				return
			}
			ctx = context.WithValue(r.Context(), "uid", uid.UID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
