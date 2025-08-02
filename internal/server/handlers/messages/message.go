package messages

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/P3rCh1/chat-server/internal/models"
	"github.com/P3rCh1/chat-server/internal/pkg/logger"
	"github.com/P3rCh1/chat-server/internal/pkg/responses"
	"github.com/P3rCh1/chat-server/internal/pkg/tools"
	"github.com/go-chi/chi/v5"
)

func Get(tools *tools.Tools) http.HandlerFunc {
	const op = "internal.http-server.handlers.messages.message.Get"
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID := r.Context().Value("userID").(int)
		roomID, err := strconv.Atoi(chi.URLParam(r, "roomID"))
		if err != nil {
			responses.InvalidURL.Drop(w)
			return
		}
		req := models.MsgRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.InvalidData.Drop(w)
			return
		}
		isMember, err := tools.Repository.IsRoomMember(userID, roomID)
		if err != nil {
			logger.LogError(tools.Log, op, err)
			responses.ServerError.Drop(w)
			return
		}
		if !isMember {
			responses.NotRoomMember.Drop(w)
			return
		}
		msgs, err := tools.Repository.GetMsgs(roomID, req.LastID, tools.Cfg.PKG.GetMessagesLimit)
		if err != nil {
			logger.LogError(tools.Log, op, err)
			responses.ServerError.Drop(w)
			return
		}
		responses.SendJSON(w, http.StatusOK, msgs)
	}
}
