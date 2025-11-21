package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/vnykmshr/nivo/shared/logger"
)

// Recovery returns a middleware that recovers from panics and logs them.
func Recovery(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					log.WithField("panic", err).
						WithField("stack", string(debug.Stack())).
						WithField("path", r.URL.Path).
						WithField("method", r.Method).
						Error("panic recovered")

					// Return 500 Internal Server Error
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal server error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryWithHandler returns a middleware that recovers from panics and calls a custom handler.
func RecoveryWithHandler(log *logger.Logger, handler func(http.ResponseWriter, *http.Request, interface{})) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.WithField("panic", err).
						WithField("stack", string(debug.Stack())).
						Error("panic recovered")

					// Call custom handler
					handler(w, r, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// defaultPanicHandler is a default handler for panics.
func defaultPanicHandler(w http.ResponseWriter, r *http.Request, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, `{"error": "internal server error", "message": "%v"}`, err)
}
