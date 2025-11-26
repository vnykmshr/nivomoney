package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RBACClient handles communication with the RBAC service.
type RBACClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRBACClient creates a new RBAC service client.
func NewRBACClient(baseURL string) *RBACClient {
	return &RBACClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// UserPermissionsResponse represents the response from RBAC service.
type UserPermissionsResponse struct {
	UserID      string       `json:"user_id"`
	Roles       []RoleInfo   `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

// RoleInfo represents basic role information.
type RoleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Permission represents a permission.
type Permission struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetUserPermissions fetches all roles and permissions for a user.
// Uses internal endpoint for service-to-service communication (no auth required).
func (c *RBACClient) GetUserPermissions(ctx context.Context, userID string) (*UserPermissionsResponse, error) {
	url := fmt.Sprintf("%s/internal/v1/users/%s/permissions", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call RBAC service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RBAC service returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response envelope
	var envelope struct {
		Success bool                     `json:"success"`
		Data    *UserPermissionsResponse `json:"data"`
		Error   *string                  `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, fmt.Errorf("RBAC request failed: %s", errMsg)
	}

	return envelope.Data, nil
}

// AssignRoleToUser assigns a role to a user.
func (c *RBACClient) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	url := fmt.Sprintf("%s/api/v1/users/%s/roles", c.baseURL, userID)

	payload := map[string]string{
		"role_id": roleID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call RBAC service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RBAC service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// AssignDefaultRole assigns the default "user" role to a newly registered user.
func (c *RBACClient) AssignDefaultRole(ctx context.Context, userID string) error {
	// Use internal endpoint (no authentication required) for service-to-service communication
	url := fmt.Sprintf("%s/internal/v1/users/%s/assign-default-role", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call RBAC service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RBAC service returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
