package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/middleware"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	userpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/user"
	"github.com/go-chi/chi/v5"
)

const URLParam = "UID"

func ChangeName(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var changeNameRequest userpb.ChangeNameRequest
		if err := json.NewDecoder(r.Body).Decode(&changeNameRequest); err != nil {
			http.Error(w, "invalid argument", http.StatusBadRequest)
			return
		}
		changeNameRequest.UID = r.Context().Value(middleware.UIDContextKey).(int64)
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
		defer cancel()
		_, err := s.User.ChangeName(ctx, &changeNameRequest)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, "user", err)
			return
		}
		responses.SendJSON(w, http.StatusOK, map[string]string{"status": "name changed"})
	}
}

func MyProfile(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		uid := r.Context().Value(middleware.UIDContextKey).(int64)
		profile(uid, s, w, r)
	}
}

func AnotherProfile(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		uid, err := strconv.ParseInt(chi.URLParam(r, URLParam), 10, 64)
		if err != nil {
			http.Error(w, "invalid UID", http.StatusBadRequest)
			return
		}
		profile(uid, s, w, r)
	}
}

func profile(uid int64, s *gateway.Services, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
	defer cancel()
	profileProto, err := s.User.Profile(ctx, &userpb.ProfileRequest{UID: uid})
	if err != nil {
		responses.GatewayGRPCErr(w, s.Log, "user", err)
		return
	}
	profile := struct {
		uid       int64     `json:"uid"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"createdAt"`
	}{
		uid:       profileProto.UID,
		Username:  profileProto.Username,
		Email:     profileProto.Email,
		CreatedAt: profileProto.CreatedAt.AsTime(),
	}
	responses.SendJSON(w, http.StatusOK, profile)
}
