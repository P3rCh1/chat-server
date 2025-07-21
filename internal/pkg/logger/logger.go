package logger

import (
	"log/slog"
	"os"

	"github.com/P3rCh1/chat-server/internal/config"
)

func New(c *config.LogConfig) *slog.Logger {
	var logLevel slog.Level
	switch c.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	return slog.New(handler)
}
