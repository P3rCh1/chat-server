package users

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/pkg/tools"
	"github.com/P3rCh1/chat-server/internal/pkg/validate"
	"github.com/go-chi/chi/v5"
)

func MyProfile(tools *tools.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		profile(userID, tools, w, r)
	}
}

func AnotherProfile(tools *tools.Tools) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
		if err != nil {
			responses.InvalidURL.Drop(w)
		}
		profile(userID, tools, w, r)
	}
}

func profile(userID int, tools *tools.Tools, w http.ResponseWriter, r *http.Request) {
	const op = "internal.http-server.handlers.users.user.profile"
	profile, err := tools.Repository.Profile(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responses.UserNotFound.Drop(w)
		} else {
			responses.ServerError.Drop(w)
			logger.LogError(tools.Log, op, err)
		}
		return
	}
	responses.SendJSON(w, http.StatusOK, profile)
}

func ChangeName(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.users.user.Profile"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		id := r.Context().Value("userID").(int)
		var request struct {
			NewName string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		if request.NewName == "" {
			responses.EmptyFields.Drop(w)
			return
		}
		if !validate.Name(request.NewName) {
			responses.BadName.Drop(w)
			return
		}
		curName, err := tools.Repository.GetUsername(id)
		if err != nil {
			responses.ServerError.Drop(w)
			logger.LogError(tools.Log, op, err)
			return
		}
		if request.NewName == curName {
			responses.NewNameMatchesCur.Drop(w)
			return
		}
		err = tools.Repository.ChangeName(id, request.NewName)
		if err != nil {
			respErr := responses.MyErrorHTTP{}
			if errors.As(err, &respErr) {
				respErr.Drop(w)
			} else {
				responses.ServerError.Drop(w)
				logger.LogError(tools.Log, op, err)
			}
			return
		}
		responses.SendOk(w, "name change succeed")
	}
}
