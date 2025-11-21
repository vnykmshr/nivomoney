package middleware

import (
	"net/http"
	"time"

	"github.com/vnykmshr/nivo/shared/logger"
)

// Logging returns a middleware that logs HTTP requests and responses.
func Logging(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context if available
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = "unknown"
			}

			// Create context logger
			contextLog := log.WithField("request_id", requestID).
				WithField("method", r.Method).
				WithField("path", r.URL.Path).
				WithField("remote_addr", r.RemoteAddr)

			// Log request
			contextLog.Info("request started")

			// Wrap response writer to capture status code
			rw := NewResponseWriter(w)

			// Call next handler
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			contextLog.WithField("status", rw.StatusCode).
				WithField("bytes", rw.BytesWritten).
				WithField("duration_ms", duration.Milliseconds()).
				Info("request completed")
		})
	}
}
