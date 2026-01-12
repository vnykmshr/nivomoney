package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// writeJSON is a helper for tests to write JSON responses.
func writeJSON(w http.ResponseWriter, v any) {
	_ = json.NewEncoder(w).Encode(v)
}

// readJSON is a helper for tests to read JSON request bodies.
func readJSON(r *http.Request, v any) {
	_ = json.NewDecoder(r.Body).Decode(v)
}

func TestNewBaseClient(t *testing.T) {
	t.Run("uses default timeout when zero", func(t *testing.T) {
		client := NewBaseClient("http://example.com", 0)
		if client.httpClient.Timeout != DefaultTimeout {
			t.Errorf("expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
		}
	})

	t.Run("uses custom timeout", func(t *testing.T) {
		timeout := 5 * time.Second
		client := NewBaseClient("http://example.com", timeout)
		if client.httpClient.Timeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, client.httpClient.Timeout)
		}
	})

	t.Run("stores base URL", func(t *testing.T) {
		baseURL := "http://example.com"
		client := NewBaseClient(baseURL, 0)
		if client.BaseURL() != baseURL {
			t.Errorf("expected baseURL %s, got %s", baseURL, client.BaseURL())
		}
	})
}

func TestBaseClient_Get(t *testing.T) {
	t.Run("successful GET with envelope", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/api/test" {
				t.Errorf("expected path /api/test, got %s", r.URL.Path)
			}

			resp := map[string]any{
				"success": true,
				"data": map[string]string{
					"id":   "123",
					"name": "test",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)

		var result struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		err := client.Get(context.Background(), "/api/test", &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.ID != "123" {
			t.Errorf("expected ID '123', got '%s'", result.ID)
		}
		if result.Name != "test" {
			t.Errorf("expected Name 'test', got '%s'", result.Name)
		}
	})

	t.Run("GET with error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			resp := map[string]any{
				"success": false,
				"error": map[string]string{
					"code":    "NOT_FOUND",
					"message": "resource not found",
				},
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		var result any

		err := client.Get(context.Background(), "/api/missing", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.HTTPStatusCode() != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", err.HTTPStatusCode())
		}
	})

	t.Run("GET with string error format", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]any{
				"success": false,
				"error":   "invalid request",
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		var result any

		err := client.Get(context.Background(), "/api/bad", &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.HTTPStatusCode() != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", err.HTTPStatusCode())
		}
	})
}

func TestBaseClient_Post(t *testing.T) {
	t.Run("successful POST with body and response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			var body map[string]string
			readJSON(r, &body)
			if body["name"] != "test" {
				t.Errorf("expected body name 'test', got '%s'", body["name"])
			}

			w.WriteHeader(http.StatusCreated)
			resp := map[string]any{
				"success": true,
				"data": map[string]string{
					"id":   "456",
					"name": body["name"],
				},
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)

		reqBody := map[string]string{"name": "test"}
		var result struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		err := client.Post(context.Background(), "/api/create", reqBody, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.ID != "456" {
			t.Errorf("expected ID '456', got '%s'", result.ID)
		}
	})

	t.Run("POST with nil body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"success": true,
				"data":    nil,
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		err := client.Post(context.Background(), "/api/action", nil, nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestBaseClient_Put(t *testing.T) {
	t.Run("successful PUT", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}

			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"success": true,
				"data": map[string]string{
					"updated": "true",
				},
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		reqBody := map[string]string{"name": "updated"}
		var result map[string]string

		err := client.Put(context.Background(), "/api/update", reqBody, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result["updated"] != "true" {
			t.Errorf("expected updated 'true', got '%s'", result["updated"])
		}
	})
}

func TestBaseClient_Delete(t *testing.T) {
	t.Run("successful DELETE with 204", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		err := client.Delete(context.Background(), "/api/resource/123", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successful DELETE with 200 and response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := map[string]any{
				"success": true,
				"data": map[string]string{
					"deleted": "true",
				},
			}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		var result map[string]string

		err := client.Delete(context.Background(), "/api/resource/123", &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result["deleted"] != "true" {
			t.Errorf("expected deleted 'true', got '%s'", result["deleted"])
		}
	})
}

func TestEnvelopeError_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshals string error", func(t *testing.T) {
		data := []byte(`"something went wrong"`)
		var e EnvelopeError
		err := e.UnmarshalJSON(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if e.Message != "something went wrong" {
			t.Errorf("expected message 'something went wrong', got '%s'", e.Message)
		}
		if e.Code != "" {
			t.Errorf("expected empty code, got '%s'", e.Code)
		}
	})

	t.Run("unmarshals object error", func(t *testing.T) {
		data := []byte(`{"code": "ERR_001", "message": "detailed error"}`)
		var e EnvelopeError
		err := e.UnmarshalJSON(data)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if e.Code != "ERR_001" {
			t.Errorf("expected code 'ERR_001', got '%s'", e.Code)
		}
		if e.Message != "detailed error" {
			t.Errorf("expected message 'detailed error', got '%s'", e.Message)
		}
	})
}

func TestParseEnvelope(t *testing.T) {
	t.Run("handles empty response", func(t *testing.T) {
		var result any
		err := parseEnvelope([]byte{}, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("handles success false with error", func(t *testing.T) {
		data := []byte(`{"success": false, "error": "bad request"}`)
		var result any
		err := parseEnvelope(data, &result)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("parses data into result", func(t *testing.T) {
		data := []byte(`{"success": true, "data": {"id": "123"}}`)
		var result struct {
			ID string `json:"id"`
		}
		err := parseEnvelope(data, &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.ID != "123" {
			t.Errorf("expected ID '123', got '%s'", result.ID)
		}
	})
}

func TestBaseClient_ResponseBodyLimit(t *testing.T) {
	t.Run("rejects response larger than limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Write more than 1MB (MaxResponseBodySize)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Write 1MB + 1 byte of data
			largeData := make([]byte, 1<<20+1)
			for i := range largeData {
				largeData[i] = 'x'
			}
			_, _ = w.Write(largeData)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		var result any

		err := client.Get(context.Background(), "/api/large", &result)
		if err == nil {
			t.Fatal("expected error for large response, got nil")
		}
		if err.Message != "response body too large" {
			t.Errorf("expected 'response body too large', got '%s'", err.Message)
		}
	})
}

func TestBaseClient_ErrorStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode     int
		expectedStatus int
		errorMessage   string
	}{
		{http.StatusBadRequest, http.StatusBadRequest, "bad request"},
		{http.StatusUnauthorized, http.StatusUnauthorized, "unauthorized"},
		{http.StatusForbidden, http.StatusForbidden, "forbidden"},
		{http.StatusNotFound, http.StatusNotFound, "not found"},
		{http.StatusInternalServerError, http.StatusInternalServerError, "server error"},
	}

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				resp := map[string]any{
					"success": false,
					"error": map[string]string{
						"message": tc.errorMessage,
					},
				}
				writeJSON(w, resp)
			}))
			defer server.Close()

			client := NewBaseClient(server.URL, DefaultTimeout)
			var result any

			err := client.Get(context.Background(), "/api/test", &result)
			if err == nil {
				t.Error("expected error, got nil")
			}
			if err.HTTPStatusCode() != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, err.HTTPStatusCode())
			}
		})
	}
}

func TestBaseClient_Headers(t *testing.T) {
	t.Run("default headers are sent on all requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify default header is present
			if r.Header.Get("X-Custom-Header") != "custom-value" {
				t.Errorf("expected X-Custom-Header 'custom-value', got '%s'", r.Header.Get("X-Custom-Header"))
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization 'Bearer test-token', got '%s'", r.Header.Get("Authorization"))
			}

			w.WriteHeader(http.StatusOK)
			resp := map[string]any{"success": true, "data": nil}
			writeJSON(w, resp)
		}))
		defer server.Close()

		headers := map[string]string{
			"X-Custom-Header": "custom-value",
			"Authorization":   "Bearer test-token",
		}
		client := NewBaseClientWithHeaders(server.URL, DefaultTimeout, headers)

		err := client.Get(context.Background(), "/api/test", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SetAuthToken sets bearer token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer my-auth-token" {
				t.Errorf("expected Authorization 'Bearer my-auth-token', got '%s'", r.Header.Get("Authorization"))
			}

			w.WriteHeader(http.StatusOK)
			resp := map[string]any{"success": true, "data": nil}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		client.SetAuthToken("my-auth-token")

		err := client.Get(context.Background(), "/api/test", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("per-request headers override defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Default should be overridden
			if r.Header.Get("Authorization") != "Bearer override-token" {
				t.Errorf("expected Authorization 'Bearer override-token', got '%s'", r.Header.Get("Authorization"))
			}
			// Default should still be present
			if r.Header.Get("X-Default") != "default-value" {
				t.Errorf("expected X-Default 'default-value', got '%s'", r.Header.Get("X-Default"))
			}
			// Per-request should be present
			if r.Header.Get("X-Request") != "request-value" {
				t.Errorf("expected X-Request 'request-value', got '%s'", r.Header.Get("X-Request"))
			}

			w.WriteHeader(http.StatusOK)
			resp := map[string]any{"success": true, "data": nil}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClientWithHeaders(server.URL, DefaultTimeout, map[string]string{
			"Authorization": "Bearer default-token",
			"X-Default":     "default-value",
		})

		perRequestHeaders := map[string]string{
			"Authorization": "Bearer override-token",
			"X-Request":     "request-value",
		}

		err := client.GetWithHeaders(context.Background(), "/api/test", nil, perRequestHeaders)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("PostWithHeaders sends headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Custom") != "custom" {
				t.Errorf("expected X-Custom 'custom', got '%s'", r.Header.Get("X-Custom"))
			}

			w.WriteHeader(http.StatusCreated)
			resp := map[string]any{"success": true, "data": map[string]string{"id": "123"}}
			writeJSON(w, resp)
		}))
		defer server.Close()

		client := NewBaseClient(server.URL, DefaultTimeout)
		var result struct {
			ID string `json:"id"`
		}

		err := client.PostWithHeaders(context.Background(), "/api/test", nil, &result, map[string]string{"X-Custom": "custom"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
