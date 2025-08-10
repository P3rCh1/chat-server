package user

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	userpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/user"
)

func Register(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user userpb.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "invalid argument", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
		defer cancel()
		id, err := s.User.Register(ctx, &user)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "user", err)
			return
		}
		responses.SendJSON(w, http.StatusCreated, id)
	}
}

func Login(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginRequest userpb.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
			http.Error(w, "invalid argument", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
		defer cancel()
		token, err := s.User.Login(ctx, &loginRequest)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "user", err)
			return
		}
		responses.SendJSON(w, http.StatusOK, token)
	}
}
