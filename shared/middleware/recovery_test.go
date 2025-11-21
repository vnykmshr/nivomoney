package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vnykmshr/nivo/shared/logger"
)

func TestRecovery(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		middleware := Recovery(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		// Should not panic
		wrapped.ServeHTTP(rec, req)

		// Verify 500 status code
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		// Verify error response
		if !strings.Contains(rec.Body.String(), "internal server error") {
			t.Error("expected error message in response")
		}

		// Verify panic is logged
		output := buf.String()
		if !strings.Contains(output, "panic recovered") {
			t.Error("expected 'panic recovered' log message")
		}

		// Verify panic value is logged
		if !strings.Contains(output, "test panic") {
			t.Error("expected panic value to be logged")
		}

		// Verify stack trace is logged
		if !strings.Contains(output, "stack") {
			t.Error("expected stack trace to be logged")
		}
	})

	t.Run("does not interfere with normal execution", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		middleware := Recovery(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify normal response
		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Body.String() != "success" {
			t.Errorf("expected body 'success', got %s", rec.Body.String())
		}

		// Verify no panic logs
		output := buf.String()
		if strings.Contains(output, "panic recovered") {
			t.Error("unexpected panic log for normal execution")
		}
	})

	t.Run("logs request details on panic", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("error")
		})

		middleware := Recovery(log)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodPost, "/users/123", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		output := buf.String()

		// Verify method is logged
		if !strings.Contains(output, "POST") {
			t.Error("expected method to be logged")
		}

		// Verify path is logged
		if !strings.Contains(output, "/users/123") {
			t.Error("expected path to be logged")
		}
	})

	t.Run("handles different panic types", func(t *testing.T) {
		testCases := []struct {
			name       string
			panicValue interface{}
		}{
			{"string panic", "panic message"},
			{"error panic", http.ErrAbortHandler},
			{"int panic", 42},
			{"nil panic", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var buf bytes.Buffer
				log := logger.New(logger.Config{
					Level:  "info",
					Format: "json",
					Output: &buf,
				})

				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic(tc.panicValue)
				})

				middleware := Recovery(log)
				wrapped := middleware(handler)

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				rec := httptest.NewRecorder()

				// Should not panic
				wrapped.ServeHTTP(rec, req)

				// Verify recovery happened
				if rec.Code != http.StatusInternalServerError {
					t.Errorf("expected status 500, got %d", rec.Code)
				}
			})
		}
	})
}

func TestRecoveryWithHandler(t *testing.T) {
	t.Run("calls custom handler on panic", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		customHandlerCalled := false
		var capturedPanic interface{}

		customHandler := func(w http.ResponseWriter, r *http.Request, err interface{}) {
			customHandlerCalled = true
			capturedPanic = err
			w.WriteHeader(http.StatusTeapot) // Use unusual status to verify custom handler
			w.Write([]byte("custom recovery"))
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("custom panic")
		})

		middleware := RecoveryWithHandler(log, customHandler)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify custom handler was called
		if !customHandlerCalled {
			t.Error("expected custom handler to be called")
		}

		// Verify panic value was passed
		if capturedPanic != "custom panic" {
			t.Errorf("expected panic value 'custom panic', got %v", capturedPanic)
		}

		// Verify custom response
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected status 418, got %d", rec.Code)
		}

		if rec.Body.String() != "custom recovery" {
			t.Errorf("expected body 'custom recovery', got %s", rec.Body.String())
		}

		// Verify panic is still logged
		output := buf.String()
		if !strings.Contains(output, "panic recovered") {
			t.Error("expected panic to be logged")
		}
	})

	t.Run("does not call custom handler on normal execution", func(t *testing.T) {
		var buf bytes.Buffer
		log := logger.New(logger.Config{
			Level:  "info",
			Format: "json",
			Output: &buf,
		})

		customHandlerCalled := false
		customHandler := func(w http.ResponseWriter, r *http.Request, err interface{}) {
			customHandlerCalled = true
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RecoveryWithHandler(log, customHandler)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if customHandlerCalled {
			t.Error("custom handler should not be called on normal execution")
		}
	})
}

func TestDefaultPanicHandler(t *testing.T) {
	t.Run("returns JSON error response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		defaultPanicHandler(rec, req, "test error")

		// Verify content type
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type to be application/json")
		}

		// Verify status code
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		// Verify response includes error message
		body := rec.Body.String()
		if !strings.Contains(body, "internal server error") {
			t.Error("expected 'internal server error' in response")
		}

		if !strings.Contains(body, "test error") {
			t.Error("expected error message in response")
		}
	})
}
