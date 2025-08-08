package responses

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func SendJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("json encode error: %w", err)
	}
	return nil
}

func GatewayGRPCErr(w http.ResponseWriter, log *slog.Logger, err error) {
	stat, ok := status.FromError(err)
	if !ok || stat.Code() == codes.Internal {
		http.Error(w, "internal error", http.StatusInternalServerError)
	} else {
		http.Error(w, stat.Message(), GRPCToHTTP(stat.Code()))
	}
}
