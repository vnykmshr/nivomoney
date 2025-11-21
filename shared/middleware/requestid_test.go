package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vnykmshr/nivo/shared/logger"
)

func TestRequestID(t *testing.T) {
	t.Run("extracts request ID from header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request ID is in context
			requestID := r.Context().Value(logger.RequestIDKey)
			if requestID != "test-request-id-123" {
				t.Errorf("expected request ID 'test-request-id-123', got %v", requestID)
			}

			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "test-request-id-123")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify request ID is set in response header
		if rec.Header().Get("X-Request-ID") != "test-request-id-123" {
			t.Errorf("expected X-Request-ID header, got %s", rec.Header().Get("X-Request-ID"))
		}
	})

	t.Run("generates request ID when not provided", func(t *testing.T) {
		var capturedRequestID string

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logger.RequestIDKey)
			if requestID == nil {
				t.Error("expected request ID to be in context")
			}

			capturedRequestID = requestID.(string)
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No X-Request-ID header set
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify request ID was generated
		if capturedRequestID == "" {
			t.Error("expected request ID to be generated")
		}

		// Verify it's a valid hex string (32 characters for 16 bytes)
		if len(capturedRequestID) != 32 {
			t.Errorf("expected request ID length 32, got %d", len(capturedRequestID))
		}

		// Verify request ID is set in response header
		responseID := rec.Header().Get("X-Request-ID")
		if responseID != capturedRequestID {
			t.Error("response header should match generated request ID")
		}
	})

	t.Run("generates unique request IDs", func(t *testing.T) {
		seenIDs := make(map[string]bool)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logger.RequestIDKey)
			if requestID != nil {
				id := requestID.(string)
				if seenIDs[id] {
					t.Errorf("duplicate request ID generated: %s", id)
				}
				seenIDs[id] = true
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		// Generate 100 request IDs
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)
		}

		if len(seenIDs) != 100 {
			t.Errorf("expected 100 unique IDs, got %d", len(seenIDs))
		}
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logger.RequestIDKey)
			if requestID != "original-id" {
				t.Errorf("expected original request ID to be preserved, got %v", requestID)
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "original-id")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Header().Get("X-Request-ID") != "original-id" {
			t.Error("expected original request ID to be preserved")
		}
	})

	t.Run("handles empty request ID header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Context().Value(logger.RequestIDKey)
			if requestID == nil || requestID.(string) == "" {
				t.Error("expected request ID to be generated for empty header")
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Should generate new ID when header is empty
		responseID := rec.Header().Get("X-Request-ID")
		if responseID == "" {
			t.Error("expected request ID to be generated")
		}
	})

	t.Run("works with multiple requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := RequestID()
		wrapped := middleware(handler)

		// First request with ID
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.Header.Set("X-Request-ID", "req-1")
		rec1 := httptest.NewRecorder()
		wrapped.ServeHTTP(rec1, req1)

		// Second request without ID
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec2 := httptest.NewRecorder()
		wrapped.ServeHTTP(rec2, req2)

		// Third request with different ID
		req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req3.Header.Set("X-Request-ID", "req-3")
		rec3 := httptest.NewRecorder()
		wrapped.ServeHTTP(rec3, req3)

		// Verify each response has correct ID
		if rec1.Header().Get("X-Request-ID") != "req-1" {
			t.Error("first request ID not preserved")
		}

		id2 := rec2.Header().Get("X-Request-ID")
		if id2 == "" || id2 == "req-1" || id2 == "req-3" {
			t.Error("second request should have generated unique ID")
		}

		if rec3.Header().Get("X-Request-ID") != "req-3" {
			t.Error("third request ID not preserved")
		}
	})
}

func TestGenerateRequestID(t *testing.T) {
	t.Run("generates non-empty ID", func(t *testing.T) {
		id := generateRequestID()
		if id == "" {
			t.Error("expected non-empty request ID")
		}
	})

	t.Run("generates hex string", func(t *testing.T) {
		id := generateRequestID()

		// Should be 32 characters (16 bytes in hex)
		if len(id) != 32 {
			t.Errorf("expected length 32, got %d", len(id))
		}

		// Should only contain hex characters
		for _, c := range id {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("invalid hex character: %c", c)
			}
		}
	})

	t.Run("generates different IDs", func(t *testing.T) {
		id1 := generateRequestID()
		id2 := generateRequestID()

		if id1 == id2 {
			t.Error("expected different request IDs")
		}
	})
}
