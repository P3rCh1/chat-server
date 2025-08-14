package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const RequestIDKey = "requestID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()[:8]
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LogRequests(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Context().Value(RequestIDKey).(string)
			next.ServeHTTP(w, r)
			logger.Info(
				"HTTP request",
				slog.String("id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Duration("duration", time.Duration(time.Since(start))),
			)
		})

	}
}

func LogInternalErrors(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := NewResponseWrapper(w)
			requestID := r.Context().Value(RequestIDKey).(string)
			next.ServeHTTP(rw, r)
			if rw.Status() == http.StatusInternalServerError {
				logger.Info(
					"internal error",
					slog.String("id", requestID),
					slog.Int("status", rw.Status()),
					slog.String("path", r.URL.Path),
				)
			}
		})
	}
}

type ResponseWrapper struct {
	http.ResponseWriter
	status int
}

func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{w, http.StatusOK}
}

func (rw *ResponseWrapper) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *ResponseWrapper) Status() int {
	return rw.status
}
