package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout returns a middleware that sets a timeout for requests.
func Timeout(duration time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			// Create a channel to signal completion
			done := make(chan struct{})

			// Run handler in goroutine
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			// Wait for completion or timeout
			select {
			case <-done:
				// Handler completed successfully
				return
			case <-ctx.Done():
				// Timeout or cancellation
				w.WriteHeader(http.StatusGatewayTimeout)
				w.Write([]byte(`{"error": "request timeout"}`))
			}
		})
	}
}
