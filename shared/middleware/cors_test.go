package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	if len(config.AllowedOrigins) != 1 || config.AllowedOrigins[0] != "*" {
		t.Error("expected AllowedOrigins to be [\"*\"]")
	}

	if len(config.AllowedMethods) == 0 {
		t.Error("expected AllowedMethods to be populated")
	}

	if config.MaxAge != 3600 {
		t.Errorf("expected MaxAge=3600, got %d", config.MaxAge)
	}
}

func TestCORS(t *testing.T) {
	t.Run("allows all origins with wildcard", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify wildcard is set
		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("expected wildcard origin, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("allows specific origin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://example.com", "http://test.com"},
			AllowedMethods: []string{"GET"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify origin is echoed back
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
			t.Errorf("expected origin http://example.com, got %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("rejects non-allowed origin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://example.com"},
			AllowedMethods: []string{"GET"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://malicious.com")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify origin header is not set
		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("expected no Access-Control-Allow-Origin header for non-allowed origin")
		}
	})

	t.Run("sets allowed methods", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		methods := rec.Header().Get("Access-Control-Allow-Methods")
		if !strings.Contains(methods, "GET") || !strings.Contains(methods, "POST") || !strings.Contains(methods, "PUT") {
			t.Errorf("expected methods to contain GET, POST, PUT, got %s", methods)
		}
	})

	t.Run("sets allowed headers", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"Content-Type", "Authorization", "X-Custom"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		headers := rec.Header().Get("Access-Control-Allow-Headers")
		if !strings.Contains(headers, "Content-Type") || !strings.Contains(headers, "Authorization") {
			t.Errorf("expected headers to contain Content-Type and Authorization, got %s", headers)
		}
	})

	t.Run("sets exposed headers", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			ExposedHeaders: []string{"X-Request-ID", "X-Rate-Limit"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		exposed := rec.Header().Get("Access-Control-Expose-Headers")
		if !strings.Contains(exposed, "X-Request-ID") {
			t.Errorf("expected exposed headers to contain X-Request-ID, got %s", exposed)
		}
	})

	t.Run("sets credentials flag", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"http://example.com"},
			AllowCredentials: true,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("expected Access-Control-Allow-Credentials to be true")
		}
	})

	t.Run("sets max age", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			MaxAge:         7200,
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		if rec.Header().Get("Access-Control-Max-Age") != "7200" {
			t.Errorf("expected MaxAge=7200, got %s", rec.Header().Get("Access-Control-Max-Age"))
		}
	})

	t.Run("handles preflight OPTIONS request", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
		}

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify handler was not called
		if handlerCalled {
			t.Error("handler should not be called for OPTIONS request")
		}

		// Verify 204 No Content response
		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}

		// Verify CORS headers are set
		if rec.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("expected Access-Control-Allow-Origin header")
		}
	})

	t.Run("passes through non-OPTIONS request", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
		}

		handlerCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created"))
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Verify handler was called
		if !handlerCalled {
			t.Error("expected handler to be called")
		}

		// Verify normal response
		if rec.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rec.Code)
		}

		if rec.Body.String() != "created" {
			t.Errorf("expected body 'created', got %s", rec.Body.String())
		}
	})

	t.Run("handles request without origin header", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"http://example.com"},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := CORS(config)
		wrapped := middleware(handler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// No Origin header set
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, req)

		// Should not set CORS headers when no origin
		if rec.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Error("should not set CORS headers without Origin header")
		}
	})
}

func TestIsOriginAllowed(t *testing.T) {
	testCases := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "wildcard allows any origin",
			origin:         "http://example.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "exact match",
			origin:         "http://example.com",
			allowedOrigins: []string{"http://example.com"},
			expected:       true,
		},
		{
			name:           "not in list",
			origin:         "http://malicious.com",
			allowedOrigins: []string{"http://example.com", "http://test.com"},
			expected:       false,
		},
		{
			name:           "case sensitive",
			origin:         "http://Example.com",
			allowedOrigins: []string{"http://example.com"},
			expected:       false,
		},
		{
			name:           "empty list",
			origin:         "http://example.com",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isOriginAllowed(tc.origin, tc.allowedOrigins)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
