// Package middleware provides common HTTP middleware for Nivo services.
package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware in order.
// Middleware are applied in the order they are provided.
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	// Apply in reverse order so they execute in the order provided
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// ResponseWriter wraps http.ResponseWriter to capture status code and size.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	BytesWritten int
	wroteHeader  bool
}

// NewResponseWriter creates a new ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code.
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.StatusCode = statusCode
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.BytesWritten += n
	return n, err
}
