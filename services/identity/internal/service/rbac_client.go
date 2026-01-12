package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/shared/clients"
)

// RBACClient handles communication with the RBAC service.
type RBACClient struct {
	*clients.BaseClient
}

// NewRBACClient creates a new RBAC service client.
func NewRBACClient(baseURL string) *RBACClient {
	return &RBACClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.ShortTimeout),
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
	var result UserPermissionsResponse
	path := fmt.Sprintf("/internal/v1/users/%s/permissions", userID)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AssignRoleToUser assigns a role to a user.
func (c *RBACClient) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	payload := map[string]string{
		"role_id": roleID,
	}
	path := fmt.Sprintf("/api/v1/users/%s/roles", userID)
	if err := c.Post(ctx, path, payload, nil); err != nil {
		return err
	}
	return nil
}

// AssignDefaultRole assigns the default "user" role to a newly registered user.
func (c *RBACClient) AssignDefaultRole(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/internal/v1/users/%s/assign-default-role", userID)
	if err := c.Post(ctx, path, nil, nil); err != nil {
		return err
	}
	return nil
}

// AssignUserAdminRole assigns the "user_admin" role to a User-Admin account.
func (c *RBACClient) AssignUserAdminRole(ctx context.Context, userID string) error {
	payload := map[string]string{
		"role_name": "user_admin",
	}
	path := fmt.Sprintf("/internal/v1/users/%s/assign-role", userID)
	if err := c.Post(ctx, path, payload, nil); err != nil {
		return err
	}
	return nil
}
