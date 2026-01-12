package service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// IdentityClient is a client for interacting with the Identity Service.
type IdentityClient struct {
	*clients.BaseClient
}

// NewIdentityClient creates a new identity client.
func NewIdentityClient(baseURL string) *IdentityClient {
	return &IdentityClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.DefaultTimeout),
	}
}

// getAuthHeaders extracts JWT token from context for service-to-service authenticated requests.
func getAuthHeaders(ctx context.Context) map[string]string {
	if token := ctx.Value(middleware.JWTTokenKey); token != nil {
		if tokenStr, ok := token.(string); ok {
			return map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", tokenStr),
			}
		}
	}
	return nil
}

// LookupUserByPhone looks up a user by phone number.
func (c *IdentityClient) LookupUserByPhone(ctx context.Context, phone string) (*UserInfo, *errors.Error) {
	// Build path with query parameter
	path := fmt.Sprintf("/api/v1/users/lookup?phone=%s", url.QueryEscape(phone))

	var result UserInfo
	// Forward JWT from context for authenticated lookup
	if err := c.GetWithHeaders(ctx, path, &result, getAuthHeaders(ctx)); err != nil {
		return nil, err
	}
	return &result, nil
}

// userKYCResponse is used to parse the nested KYC status response.
type userKYCResponse struct {
	KYC struct {
		Status string `json:"status"`
	} `json:"kyc"`
}

// GetUserKYCStatus gets the KYC status for a user.
func (c *IdentityClient) GetUserKYCStatus(ctx context.Context, userID string) (string, *errors.Error) {
	path := fmt.Sprintf("/api/v1/users/%s", userID)

	var result userKYCResponse
	if err := c.Get(ctx, path, &result); err != nil {
		return "", err
	}
	return result.KYC.Status, nil
}

// GetUser retrieves user information by user ID.
func (c *IdentityClient) GetUser(ctx context.Context, userID string) (*UserInfo, *errors.Error) {
	path := fmt.Sprintf("/api/v1/users/%s", userID)

	var result UserInfo
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}

	// Handle both phone and phone_number fields
	if result.PhoneNumber == "" && result.Phone != "" {
		result.PhoneNumber = result.Phone
	} else if result.Phone == "" && result.PhoneNumber != "" {
		result.Phone = result.PhoneNumber
	}

	return &result, nil
}
