package responses

import (
	"encoding/json"
	"fmt"
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

func SendOk(w http.ResponseWriter, msg string) {
	SendJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{msg})
}
