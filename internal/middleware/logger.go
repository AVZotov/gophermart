package middleware

import (
	"net/http"
	"time"

	"github.com/AVZotov/gophermart/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware returns HTTP middleware that logs each request's
// method, path, response status, and duration via l after the request
// completes.
func LoggingMiddleware(l logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				start := time.Now()
				lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

				next.ServeHTTP(lw, r)

				l.Info(
					"request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", lw.statusCode,
					"duration", time.Since(start),
				)
			},
		)
	}
}
