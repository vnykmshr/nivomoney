package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/errors"
)

// Default timeouts for service clients
const (
	DefaultTimeout = 10 * time.Second
	ShortTimeout   = 5 * time.Second
	LongTimeout    = 30 * time.Second
)

// BaseClient provides common HTTP functionality for service-to-service communication.
// Embed this in service-specific clients to get consistent error handling,
// timeouts, and response envelope parsing.
type BaseClient struct {
	baseURL        string
	httpClient     *http.Client
	defaultHeaders map[string]string
}

// NewBaseClient creates a new base client with the specified timeout.
func NewBaseClient(baseURL string, timeout time.Duration) *BaseClient {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &BaseClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		defaultHeaders: make(map[string]string),
	}
}

// NewBaseClientWithHeaders creates a new base client with default headers.
// These headers will be applied to all requests made by this client.
func NewBaseClientWithHeaders(baseURL string, timeout time.Duration, headers map[string]string) *BaseClient {
	client := NewBaseClient(baseURL, timeout)
	for k, v := range headers {
		client.defaultHeaders[k] = v
	}
	return client
}

// SetAuthToken sets a Bearer token for all requests.
// This is a convenience method for the common auth pattern.
func (c *BaseClient) SetAuthToken(token string) {
	c.defaultHeaders["Authorization"] = "Bearer " + token
}

// SetInternalSecret sets the internal service secret for service-to-service calls.
// This header is validated by the InternalAuth middleware on receiving services.
func (c *BaseClient) SetInternalSecret(secret string) {
	if secret != "" {
		c.defaultHeaders["X-Internal-Secret"] = secret
	}
}

// NewInternalClient creates a base client configured for internal service-to-service calls.
// The secret is sent as X-Internal-Secret header and validated by InternalAuth middleware.
func NewInternalClient(baseURL string, timeout time.Duration, internalSecret string) *BaseClient {
	client := NewBaseClient(baseURL, timeout)
	if internalSecret != "" {
		client.defaultHeaders["X-Internal-Secret"] = internalSecret
	}
	return client
}

// BaseURL returns the base URL for building endpoint paths.
func (c *BaseClient) BaseURL() string {
	return c.baseURL
}

// Get performs a GET request and parses the envelope response into result.
// The result parameter should be a pointer to the expected data type.
func (c *BaseClient) Get(ctx context.Context, path string, result any) *errors.Error {
	return c.GetWithHeaders(ctx, path, result, nil)
}

// GetWithHeaders performs a GET request with additional headers.
func (c *BaseClient) GetWithHeaders(ctx context.Context, path string, result any, headers map[string]string) *errors.Error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.Internal(fmt.Sprintf("failed to create request: %v", err))
	}

	return c.doRequest(req, result, headers, http.StatusOK)
}

// Post performs a POST request with JSON body and parses the envelope response.
// The body parameter will be marshaled to JSON. Pass nil for empty body.
// The result parameter should be a pointer to the expected data type, or nil if no response data expected.
func (c *BaseClient) Post(ctx context.Context, path string, body, result any) *errors.Error {
	return c.PostWithHeaders(ctx, path, body, result, nil)
}

// PostWithHeaders performs a POST request with additional headers.
func (c *BaseClient) PostWithHeaders(ctx context.Context, path string, body, result any, headers map[string]string) *errors.Error {
	return c.doJSON(ctx, http.MethodPost, path, body, result, headers, http.StatusOK, http.StatusCreated)
}

// Put performs a PUT request with JSON body and parses the envelope response.
func (c *BaseClient) Put(ctx context.Context, path string, body, result any) *errors.Error {
	return c.PutWithHeaders(ctx, path, body, result, nil)
}

// PutWithHeaders performs a PUT request with additional headers.
func (c *BaseClient) PutWithHeaders(ctx context.Context, path string, body, result any, headers map[string]string) *errors.Error {
	return c.doJSON(ctx, http.MethodPut, path, body, result, headers, http.StatusOK)
}

// Delete performs a DELETE request and parses the envelope response.
func (c *BaseClient) Delete(ctx context.Context, path string, result any) *errors.Error {
	return c.DeleteWithHeaders(ctx, path, result, nil)
}

// DeleteWithHeaders performs a DELETE request with additional headers.
func (c *BaseClient) DeleteWithHeaders(ctx context.Context, path string, result any, headers map[string]string) *errors.Error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return errors.Internal(fmt.Sprintf("failed to create request: %v", err))
	}

	return c.doRequest(req, result, headers, http.StatusOK, http.StatusNoContent)
}

// doJSON handles JSON body requests (POST, PUT, PATCH).
func (c *BaseClient) doJSON(ctx context.Context, method, path string, body, result any, headers map[string]string, successCodes ...int) *errors.Error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return errors.Internal(fmt.Sprintf("failed to marshal request body: %v", err))
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return errors.Internal(fmt.Sprintf("failed to create request: %v", err))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(req, result, headers, successCodes...)
}

// doRequest executes the HTTP request and handles response parsing.
func (c *BaseClient) doRequest(req *http.Request, result any, headers map[string]string, successCodes ...int) *errors.Error {
	// Apply default headers first
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}
	// Apply per-request headers (override defaults if same key)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Internal(fmt.Sprintf("request failed: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	// Limit response body size to prevent OOM from large responses
	limitedReader := io.LimitReader(resp.Body, config.MaxResponseBodySize+1)
	respBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return errors.Internal(fmt.Sprintf("failed to read response: %v", err))
	}
	if len(respBody) > config.MaxResponseBodySize {
		return errors.Internal("response body too large")
	}

	// Check if status code is in success codes
	if !slices.Contains(successCodes, resp.StatusCode) {
		return c.parseErrorResponse(resp.StatusCode, respBody)
	}

	// If no result expected, we're done
	if result == nil {
		return nil
	}

	return parseEnvelope(respBody, result)
}

// parseEnvelope parses the standard API envelope {success, data, error}.
func parseEnvelope(respBody []byte, result any) *errors.Error {
	// Handle empty response
	if len(respBody) == 0 {
		return nil
	}

	// Parse envelope structure
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   *EnvelopeError  `json:"error"`
	}

	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return errors.Internal(fmt.Sprintf("failed to parse response: %v", err))
	}

	if !envelope.Success {
		if envelope.Error != nil {
			return errors.Internal(envelope.Error.Message)
		}
		return errors.Internal("request failed: unknown error")
	}

	// Parse data into result if present
	if len(envelope.Data) > 0 && result != nil {
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return errors.Internal(fmt.Sprintf("failed to parse response data: %v", err))
		}
	}

	return nil
}

// EnvelopeError represents the error field in API responses.
// Supports both string and object formats.
type EnvelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// UnmarshalJSON handles both string and object error formats.
func (e *EnvelopeError) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Message = s
		return nil
	}

	// Try object
	type errorObj struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	var obj errorObj
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	e.Code = obj.Code
	e.Message = obj.Message
	return nil
}

// parseErrorResponse extracts error details from non-success responses.
func (c *BaseClient) parseErrorResponse(statusCode int, respBody []byte) *errors.Error {
	// Try to parse as envelope error first
	var envelope struct {
		Success bool           `json:"success"`
		Error   *EnvelopeError `json:"error"`
	}

	msg := string(respBody)
	if err := json.Unmarshal(respBody, &envelope); err == nil && envelope.Error != nil {
		msg = envelope.Error.Message
		if envelope.Error.Code != "" {
			msg = fmt.Sprintf("%s: %s", envelope.Error.Code, envelope.Error.Message)
		}
	}

	// Fallback message if empty
	if msg == "" {
		msg = fmt.Sprintf("service returned status %d", statusCode)
	}

	return errorForStatusCode(statusCode, msg)
}

// errorForStatusCode maps HTTP status codes to appropriate error types.
func errorForStatusCode(statusCode int, msg string) *errors.Error {
	switch statusCode {
	case http.StatusNotFound:
		return errors.NotFound(msg)
	case http.StatusBadRequest:
		return errors.BadRequest(msg)
	case http.StatusUnauthorized:
		return errors.Unauthorized(msg)
	case http.StatusForbidden:
		return errors.Forbidden(msg)
	default:
		return errors.Internal(msg)
	}
}
