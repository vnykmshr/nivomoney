package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	t.Run("allows fast requests to complete", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Fast handler that completes immediately
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		middleware := Timeout(1 * time.Second)
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
	})

	t.Run("times out slow requests", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Slow handler that takes longer than timeout
			select {
			case <-time.After(200 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("too late"))
			case <-r.Context().Done():
				// Context cancelled, handler should stop
				return
			}
		})

		middleware := Timeout(50 * time.Millisecond)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify timeout response
		if rec.Code != http.StatusGatewayTimeout {
			t.Errorf("expected status 504, got %d", rec.Code)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "request timeout") {
			t.Errorf("expected timeout message, got %s", body)
		}
	})

	t.Run("provides context with timeout to handler", func(t *testing.T) {
		var contextHadDeadline bool

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if context has a deadline
			_, hasDeadline := r.Context().Deadline()
			contextHadDeadline = hasDeadline
			w.WriteHeader(http.StatusOK)
		})

		middleware := Timeout(1 * time.Second)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if !contextHadDeadline {
			t.Error("expected context to have deadline")
		}
	})

	t.Run("handler can check for cancellation", func(t *testing.T) {
		contextCancelled := make(chan bool, 1)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate work with context checking
			for i := 0; i < 100; i++ {
				select {
				case <-r.Context().Done():
					contextCancelled <- true
					return
				case <-time.After(5 * time.Millisecond):
					// Continue working
				}
			}
			w.WriteHeader(http.StatusOK)
		})

		middleware := Timeout(30 * time.Millisecond)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify timeout response
		if rec.Code != http.StatusGatewayTimeout {
			t.Errorf("expected status 504, got %d", rec.Code)
		}

		// Wait for handler to detect cancellation
		select {
		case <-contextCancelled:
			// Good, handler detected cancellation
		case <-time.After(100 * time.Millisecond):
			t.Error("expected handler to detect context cancellation")
		}
	})

	t.Run("different timeout durations", func(t *testing.T) {
		testCases := []struct {
			name           string
			timeout        time.Duration
			handlerDelay   time.Duration
			expectedStatus int
		}{
			{
				name:           "very short timeout",
				timeout:        10 * time.Millisecond,
				handlerDelay:   50 * time.Millisecond,
				expectedStatus: http.StatusGatewayTimeout,
			},
			{
				name:           "generous timeout",
				timeout:        500 * time.Millisecond,
				handlerDelay:   10 * time.Millisecond,
				expectedStatus: http.StatusOK,
			},
			{
				name:           "exact boundary",
				timeout:        100 * time.Millisecond,
				handlerDelay:   50 * time.Millisecond,
				expectedStatus: http.StatusOK,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					select {
					case <-time.After(tc.handlerDelay):
						w.WriteHeader(http.StatusOK)
					case <-r.Context().Done():
						return
					}
				})

				middleware := Timeout(tc.timeout)
				wrapped := middleware(handler)

				req := httptest.NewRequest(http.MethodGet, "/test", nil)
				rec := httptest.NewRecorder()

				wrapped.ServeHTTP(rec, req)

				if rec.Code != tc.expectedStatus {
					t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
				}
			})
		}
	})

	t.Run("returns JSON error on timeout", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-time.After(200 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
			case <-r.Context().Done():
				return
			}
		})

		middleware := Timeout(50 * time.Millisecond)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify JSON response
		body := rec.Body.String()
		if !strings.Contains(body, "{") || !strings.Contains(body, "}") {
			t.Error("expected JSON response")
		}

		if !strings.Contains(body, "error") {
			t.Error("expected error field in JSON")
		}

		if !strings.Contains(body, "request timeout") {
			t.Error("expected timeout message in JSON")
		}
	})

	t.Run("preserves handler response when no timeout", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "custom-value")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("custom response"))
		})

		middleware := Timeout(1 * time.Second)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify custom response is preserved
		if rec.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rec.Code)
		}

		if rec.Header().Get("X-Custom-Header") != "custom-value" {
			t.Error("expected custom header to be preserved")
		}

		if rec.Body.String() != "custom response" {
			t.Errorf("expected custom body, got %s", rec.Body.String())
		}
	})

	t.Run("works with other HTTP methods", func(t *testing.T) {
		methods := []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != method {
						t.Errorf("expected method %s, got %s", method, r.Method)
					}
					w.WriteHeader(http.StatusOK)
				})

				middleware := Timeout(1 * time.Second)
				wrapped := middleware(handler)

				req := httptest.NewRequest(method, "/test", nil)
				rec := httptest.NewRecorder()

				wrapped.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("expected status 200, got %d", rec.Code)
				}
			})
		}
	})

	t.Run("concurrent requests with timeouts", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Variable delay based on query param
			delay := 10 * time.Millisecond
			if r.URL.Query().Get("slow") == "true" {
				delay = 200 * time.Millisecond
			}

			select {
			case <-time.After(delay):
				w.WriteHeader(http.StatusOK)
			case <-r.Context().Done():
				return
			}
		})

		middleware := Timeout(50 * time.Millisecond)
		wrapped := middleware(handler)

		// Fast request
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec1 := httptest.NewRecorder()

		// Slow request
		req2 := httptest.NewRequest(http.MethodGet, "/test?slow=true", nil)
		rec2 := httptest.NewRecorder()

		// Execute concurrently
		done1 := make(chan bool)
		done2 := make(chan bool)

		go func() {
			wrapped.ServeHTTP(rec1, req1)
			done1 <- true
		}()

		go func() {
			wrapped.ServeHTTP(rec2, req2)
			done2 <- true
		}()

		<-done1
		<-done2

		// Fast request should succeed
		if rec1.Code != http.StatusOK {
			t.Errorf("expected fast request to succeed with status 200, got %d", rec1.Code)
		}

		// Slow request should timeout
		if rec2.Code != http.StatusGatewayTimeout {
			t.Errorf("expected slow request to timeout with status 504, got %d", rec2.Code)
		}
	})
}
