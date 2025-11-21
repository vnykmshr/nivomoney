package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/vnykmshr/nivo/shared/logger"
)

// RequestID returns a middleware that generates or extracts request IDs.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get request ID from header
			requestID := r.Header.Get("X-Request-ID")

			// Generate new request ID if not present
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Set request ID in response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// generateRequestID generates a random request ID.
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter if random fails
		return "unknown"
	}
	return hex.EncodeToString(bytes)
}
