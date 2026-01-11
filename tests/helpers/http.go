// Package helpers provides test utilities for HTTP handler testing.
package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRequest represents a test HTTP request configuration.
type TestRequest struct {
	Method      string
	Path        string
	Body        interface{}
	Headers     map[string]string
	QueryParams map[string]string
	AuthToken   string
}

// TestResponse represents the parsed test HTTP response.
type TestResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// APIResponse is the standard API response structure.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// APIError represents an error response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MakeRequest creates and executes a test HTTP request against a handler.
func MakeRequest(t *testing.T, handler http.Handler, req TestRequest) *TestResponse {
	t.Helper()

	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		require.NoError(t, err, "failed to marshal request body")
		bodyReader = bytes.NewBuffer(bodyBytes)
	}

	httpReq := httptest.NewRequest(req.Method, req.Path, bodyReader)

	// Set default content type for POST/PUT/PATCH
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set auth token if provided
	if req.AuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken)
	}

	// Add query parameters
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Record the response
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httpReq)

	return &TestResponse{
		StatusCode: rec.Code,
		Body:       rec.Body.Bytes(),
		Headers:    rec.Header(),
	}
}

// ParseJSON parses the response body as JSON into the target struct.
func (r *TestResponse) ParseJSON(t *testing.T, target interface{}) {
	t.Helper()
	err := json.Unmarshal(r.Body, target)
	require.NoError(t, err, "failed to parse response JSON: %s", string(r.Body))
}

// ParseAPIResponse parses the response as a standard API response.
func (r *TestResponse) ParseAPIResponse(t *testing.T) *APIResponse {
	t.Helper()
	var resp APIResponse
	r.ParseJSON(t, &resp)
	return &resp
}

// ParseData parses the Data field from an API response into the target struct.
func (resp *APIResponse) ParseData(t *testing.T, target interface{}) {
	t.Helper()
	err := json.Unmarshal(resp.Data, target)
	require.NoError(t, err, "failed to parse response data: %s", string(resp.Data))
}

// AssertSuccess asserts that the response indicates success.
func (r *TestResponse) AssertSuccess(t *testing.T) {
	t.Helper()
	resp := r.ParseAPIResponse(t)
	require.True(t, resp.Success, "expected success=true, got response: %s", string(r.Body))
}

// AssertError asserts that the response indicates an error with the expected code.
func (r *TestResponse) AssertError(t *testing.T, expectedCode string) {
	t.Helper()
	resp := r.ParseAPIResponse(t)
	require.False(t, resp.Success, "expected success=false")
	require.NotNil(t, resp.Error, "expected error response")
	require.Equal(t, expectedCode, resp.Error.Code, "error code mismatch")
}

// AssertStatusCode asserts the expected HTTP status code.
func (r *TestResponse) AssertStatusCode(t *testing.T, expected int) {
	t.Helper()
	require.Equal(t, expected, r.StatusCode, "status code mismatch, body: %s", string(r.Body))
}

// AssertHeader asserts a header value.
func (r *TestResponse) AssertHeader(t *testing.T, key, expected string) {
	t.Helper()
	actual := r.Headers.Get(key)
	require.Equal(t, expected, actual, "header %s mismatch", key)
}

// AssertJSONContains asserts that the response body contains a JSON field.
func (r *TestResponse) AssertJSONContains(t *testing.T, path string, expected interface{}) {
	t.Helper()
	var data map[string]interface{}
	r.ParseJSON(t, &data)

	// Simple single-level path for now
	actual, ok := data[path]
	require.True(t, ok, "JSON path %s not found in response", path)
	require.Equal(t, expected, actual, "JSON value at path %s mismatch", path)
}

// JSONBody is a helper to create a map[string]interface{} body.
func JSONBody(pairs ...interface{}) map[string]interface{} {
	if len(pairs)%2 != 0 {
		panic("JSONBody requires even number of arguments")
	}

	result := make(map[string]interface{})
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic("JSONBody keys must be strings")
		}
		result[key] = pairs[i+1]
	}
	return result
}

// ContextWithUserID creates a request context with user ID for auth testing.
// This is useful for testing handlers that expect authenticated users.
type contextKey string

const UserIDContextKey contextKey = "user_id"

// AddUserIDToRequest adds user ID to the request context.
func AddUserIDToRequest(r *http.Request, userID string) *http.Request {
	return r.WithContext(r.Context())
}
