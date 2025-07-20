package msg

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func SendJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("json encode error: %w", err)
	}
	return nil
}

func SendJSONWithLog(w http.ResponseWriter, status int, msg any, log *slog.Logger, logInfo string) {
	log.Info(
		logInfo,
		"status", status,
		"message", msg,
	)
	if err := SendJSON(w, status, msg); err != nil {
		log.Info(err.Error())
	}
}
