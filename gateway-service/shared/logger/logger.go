package logger

import (
	"log/slog"
	"os"
)

const (
	DebugLVL  = "debug"
	InfoLVL = "info"
	WarnLVL  = "warn"
	ErrorLVL = "error"
)

func New(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
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
