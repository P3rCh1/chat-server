package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/gateway-service/internal/gateway"
	"github.com/P3rCh1/chat-server/gateway-service/internal/responses"
	userpb "github.com/P3rCh1/chat-server/gateway-service/shared/proto/gen/go/user"
	"github.com/go-chi/chi/v5"
)

func ChangeName(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var changeNameRequest userpb.ChangeNameRequest
		changeNameRequest.Id = r.Context().Value("uid").(int32)
		if err := json.NewDecoder(r.Body).Decode(&changeNameRequest); err != nil {
			http.Error(w, "invalid argument", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
		defer cancel()
		_, err := s.User.ChangeName(ctx, &changeNameRequest)
		if err != nil {
			responses.GatewayGRPCErr(w, s.Log, err)
			return
		}
		responses.SendJSON(w, http.StatusOK, struct{}{})
	}
}

func MyProfile(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		uid := r.Context().Value("uid").(int32)
		profile(uid, s, w, r)
	}
}

func AnotherProfile(s *gateway.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		uid, err := strconv.Atoi(chi.URLParam(r, "uid"))
		if err != nil {
			http.Error(w, "invalid uid", http.StatusBadRequest)
			return
		}
		profile(int32(uid), s, w, r)
	}
}

func profile(uid int32, s *gateway.Services, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.Timeouts.User)
	defer cancel()
	profile, err := s.User.Profile(ctx, &userpb.ProfileRequest{Id: uid})
	if err != nil {
		responses.GatewayGRPCErr(w, s.Log, err)
		return
	}
	responses.SendJSON(w, http.StatusOK, profile)
}
