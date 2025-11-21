package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	t.Run("chains middleware in correct order", func(t *testing.T) {
		var calls []string

		// Create middleware that tracks call order
		m1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, "m1-before")
				next.ServeHTTP(w, r)
				calls = append(calls, "m1-after")
			})
		}

		m2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, "m2-before")
				next.ServeHTTP(w, r)
				calls = append(calls, "m2-after")
			})
		}

		m3 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, "m3-before")
				next.ServeHTTP(w, r)
				calls = append(calls, "m3-after")
			})
		}

		// Final handler
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "handler")
			w.WriteHeader(http.StatusOK)
		})

		// Chain middleware
		chained := Chain(handler, m1, m2, m3)

		// Execute request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)

		// Verify call order
		expected := []string{
			"m1-before", "m2-before", "m3-before",
			"handler",
			"m3-after", "m2-after", "m1-after",
		}

		if len(calls) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(calls))
		}

		for i, call := range calls {
			if call != expected[i] {
				t.Errorf("call %d: expected %s, got %s", i, expected[i], call)
			}
		}
	})

	t.Run("works with no middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		chained := Chain(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		if rec.Body.String() != "ok" {
			t.Errorf("expected body 'ok', got %s", rec.Body.String())
		}
	})

	t.Run("middleware can modify request", func(t *testing.T) {
		middleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("X-Modified", "true")
				next.ServeHTTP(w, r)
			})
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Modified") != "true" {
				t.Error("expected X-Modified header to be set")
			}
			w.WriteHeader(http.StatusOK)
		})

		chained := Chain(handler, middleware)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		chained.ServeHTTP(rec, req)
	})
}

func TestResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.WriteHeader(http.StatusCreated)

		if rw.StatusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rw.StatusCode)
		}
	})

	t.Run("defaults to 200 when WriteHeader not called", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.Write([]byte("test"))

		if rw.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rw.StatusCode)
		}
	})

	t.Run("captures bytes written", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		data := []byte("hello world")
		n, err := rw.Write(data)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if n != len(data) {
			t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
		}

		if rw.BytesWritten != len(data) {
			t.Errorf("expected BytesWritten=%d, got %d", len(data), rw.BytesWritten)
		}
	})

	t.Run("accumulates bytes from multiple writes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.Write([]byte("hello"))
		rw.Write([]byte(" "))
		rw.Write([]byte("world"))

		expected := len("hello world")
		if rw.BytesWritten != expected {
			t.Errorf("expected BytesWritten=%d, got %d", expected, rw.BytesWritten)
		}
	})

	t.Run("proxies header operations", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.Header().Set("X-Test", "value")

		if rec.Header().Get("X-Test") != "value" {
			t.Error("expected header to be set on underlying writer")
		}
	})

	t.Run("only sets status code once", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rw := NewResponseWriter(rec)

		rw.WriteHeader(http.StatusCreated)
		rw.WriteHeader(http.StatusBadRequest) // Should be ignored

		if rw.StatusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rw.StatusCode)
		}
	})
}
