package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vnykmshr/nivo/shared/logger"
)

func TestLogging(t *testing.T) {
	t.Run("logs request start and completion", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		middleware := Logging(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "test-123")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		output := buf.String()

		// Verify request started log
		if !strings.Contains(output, "request started") {
			t.Error("expected 'request started' log message")
		}

		// Verify request completed log
		if !strings.Contains(output, "request completed") {
			t.Error("expected 'request completed' log message")
		}

		// Verify request ID is logged
		if !strings.Contains(output, "test-123") {
			t.Error("expected request ID to be logged")
		}

		// Verify method is logged
		if !strings.Contains(output, "GET") {
			t.Error("expected method to be logged")
		}

		// Verify path is logged
		if !strings.Contains(output, "/test") {
			t.Error("expected path to be logged")
		}
	})

	t.Run("logs status code and bytes written", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		})

		middleware := Logging(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodPost, "/users", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		output := buf.String()

		// Verify status code is logged
		if !strings.Contains(output, "201") {
			t.Error("expected status code 201 to be logged")
		}

		// Verify bytes written (length of "created" = 7)
		if !strings.Contains(output, "\"bytes\":7") {
			t.Error("expected bytes written to be logged")
		}

		// Verify duration is logged
		if !strings.Contains(output, "duration_ms") {
			t.Error("expected duration to be logged")
		}
	})

	t.Run("handles missing request ID", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := Logging(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		output := buf.String()

		// Should use "unknown" when no request ID
		if !strings.Contains(output, "unknown") {
			t.Error("expected 'unknown' request ID to be logged")
		}
	})

	t.Run("logs all HTTP methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		}

		for _, method := range methods {
			var buf bytes.Buffer
			log := logger.New(logger.Config{
				Level:  "info",
				Format: "json",
				Output: &buf,
			})

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := Logging(log)
			wrapped := middleware(handler)

			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			output := buf.String()
			if !strings.Contains(output, method) {
				t.Errorf("expected method %s to be logged", method)
			}
		}
	})

	t.Run("captures errors from handler", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		})

		middleware := Logging(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		output := buf.String()

		// Verify 500 status is logged
		if !strings.Contains(output, "500") {
			t.Error("expected status 500 to be logged")
		}
	})
}
