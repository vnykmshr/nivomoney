package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// IdentityClient is a client for interacting with the Identity Service.
type IdentityClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewIdentityClient creates a new identity client.
func NewIdentityClient(baseURL string) *IdentityClient {
	return &IdentityClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LookupUserByPhone looks up a user by phone number.
func (c *IdentityClient) LookupUserByPhone(ctx context.Context, phone string) (*UserInfo, *errors.Error) {
	// Build URL with query parameter
	lookupURL, err := url.Parse(fmt.Sprintf("%s/api/v1/users/lookup", c.baseURL))
	if err != nil {
		return nil, errors.Internal("failed to parse identity service URL")
	}

	query := lookupURL.Query()
	query.Set("phone", phone)
	lookupURL.RawQuery = query.Encode()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL.String(), nil)
	if err != nil {
		return nil, errors.Internal("failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Extract JWT token from context and forward it (for service-to-service authenticated requests)
	if token := ctx.Value(middleware.JWTTokenKey); token != nil {
		if tokenStr, ok := token.(string); ok {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenStr))
		}
	}

	// Make HTTP request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to call identity service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal("failed to read response body")
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(fmt.Sprintf("identity service returned error: %d - %s", resp.StatusCode, string(body)))
	}

	// Parse response
	var apiResponse struct {
		Data UserInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, errors.Internal("failed to parse response")
	}

	return &apiResponse.Data, nil
}

// GetUserKYCStatus gets the KYC status for a user.
func (c *IdentityClient) GetUserKYCStatus(ctx context.Context, userID string) (string, *errors.Error) {
	// Build URL
	getUserURL := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getUserURL, nil)
	if err != nil {
		return "", errors.Internal("failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request (internal service-to-service, no auth required)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", errors.Internal(fmt.Sprintf("failed to call identity service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Internal("failed to read response body")
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", errors.NotFound("user not found")
		}
		return "", errors.Internal(fmt.Sprintf("identity service returned error: %d - %s", resp.StatusCode, string(body)))
	}

	// Parse response
	var apiResponse struct {
		Data struct {
			KYC struct {
				Status string `json:"status"`
			} `json:"kyc"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", errors.Internal("failed to parse response")
	}

	return apiResponse.Data.KYC.Status, nil
}

// GetUser retrieves user information by user ID.
func (c *IdentityClient) GetUser(ctx context.Context, userID string) (*UserInfo, *errors.Error) {
	// Build URL
	getUserURL := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getUserURL, nil)
	if err != nil {
		return nil, errors.Internal("failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request (internal service-to-service, no auth required)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to call identity service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal("failed to read response body")
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal(fmt.Sprintf("identity service returned error: %d - %s", resp.StatusCode, string(body)))
	}

	// Parse response
	var apiResponse struct {
		Data UserInfo `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, errors.Internal("failed to parse response")
	}

	// Handle both phone and phone_number fields
	if apiResponse.Data.PhoneNumber == "" && apiResponse.Data.Phone != "" {
		apiResponse.Data.PhoneNumber = apiResponse.Data.Phone
	} else if apiResponse.Data.Phone == "" && apiResponse.Data.PhoneNumber != "" {
		apiResponse.Data.Phone = apiResponse.Data.PhoneNumber
	}

	return &apiResponse.Data, nil
}
