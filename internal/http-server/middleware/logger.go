package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func LogRequests(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info(
				"HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Duration("duration", time.Duration(time.Since(start))),
			)
		})

	}
}

func LogErrors(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := NewResponseWrapper(w)
			next.ServeHTTP(rw, r)
			if rw.Status() >= 400 {
				logger.Info(
					"HTTP error",
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
