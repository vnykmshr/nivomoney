package models

import (
	"github.com/vnykmshr/nivo/shared/models"
)

// Role represents a system role with hierarchical support.
type Role struct {
	ID           string           `json:"id" db:"id"`
	Name         string           `json:"name" db:"name"`
	Description  string           `json:"description" db:"description"`
	ParentRoleID *string          `json:"parent_role_id,omitempty" db:"parent_role_id"`
	IsSystem     bool             `json:"is_system" db:"is_system"`
	IsActive     bool             `json:"is_active" db:"is_active"`
	CreatedAt    models.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    models.Timestamp `json:"updated_at" db:"updated_at"`

	// Relationships (populated via joins, not in DB)
	ParentRole  *Role        `json:"parent_role,omitempty" db:"-"`
	Permissions []Permission `json:"permissions,omitempty" db:"-"`
}

// Permission represents a granular permission.
type Permission struct {
	ID          string           `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"` // Format: service:resource:action
	Service     string           `json:"service" db:"service"`
	Resource    string           `json:"resource" db:"resource"`
	Action      string           `json:"action" db:"action"`
	Description string           `json:"description" db:"description"`
	IsSystem    bool             `json:"is_system" db:"is_system"`
	CreatedAt   models.Timestamp `json:"created_at" db:"created_at"`
}

// RolePermission represents the mapping between roles and permissions.
type RolePermission struct {
	RoleID       string           `json:"role_id" db:"role_id"`
	PermissionID string           `json:"permission_id" db:"permission_id"`
	GrantedBy    *string          `json:"granted_by,omitempty" db:"granted_by"`
	GrantedAt    models.Timestamp `json:"granted_at" db:"granted_at"`
}

// UserRole represents a user's role assignment.
type UserRole struct {
	UserID     string            `json:"user_id" db:"user_id"`
	RoleID     string            `json:"role_id" db:"role_id"`
	AssignedBy *string           `json:"assigned_by,omitempty" db:"assigned_by"`
	AssignedAt models.Timestamp  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt  *models.Timestamp `json:"expires_at,omitempty" db:"expires_at"`
	IsActive   bool              `json:"is_active" db:"is_active"`

	// Relationships (populated via joins)
	Role *Role `json:"role,omitempty" db:"-"`
}

// CreateRoleRequest represents the request to create a new role.
type CreateRoleRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=50"`
	Description  string  `json:"description" validate:"required,min=5,max=500"`
	ParentRoleID *string `json:"parent_role_id,omitempty" validate:"omitempty,uuid"`
}

// UpdateRoleRequest represents the request to update a role.
type UpdateRoleRequest struct {
	Name         *string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Description  *string `json:"description,omitempty" validate:"omitempty,min=5,max=500"`
	ParentRoleID *string `json:"parent_role_id,omitempty" validate:"omitempty,uuid"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// CreatePermissionRequest represents the request to create a new permission.
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,permission_format"`
	Service     string `json:"service" validate:"required,min=2,max=50"`
	Resource    string `json:"resource" validate:"required,min=2,max=50"`
	Action      string `json:"action" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"required,min=5,max=500"`
}

// AssignRoleToUserRequest represents the request to assign a role to a user.
type AssignRoleToUserRequest struct {
	UserID    string  `json:"user_id" validate:"omitempty,uuid"` // Populated from path, not body
	RoleID    string  `json:"role_id" validate:"required,uuid"`
	ExpiresAt *string `json:"expires_at,omitempty"` // ISO 8601 format
}

// AssignPermissionToRoleRequest represents the request to assign a permission to a role.
type AssignPermissionToRoleRequest struct {
	PermissionID string `json:"permission_id" validate:"required,uuid"`
}

// CheckPermissionRequest represents a permission check request.
type CheckPermissionRequest struct {
	UserID     string `json:"user_id" validate:"required,uuid"`
	Permission string `json:"permission" validate:"required,permission_format"`
}

// CheckPermissionsRequest represents a batch permission check request.
type CheckPermissionsRequest struct {
	UserID      string   `json:"user_id" validate:"required,uuid"`
	Permissions []string `json:"permissions" validate:"required,min=1,dive,permission_format"`
}

// CheckPermissionResponse represents a permission check response.
type CheckPermissionResponse struct {
	Allowed bool     `json:"allowed"`
	Roles   []string `json:"roles,omitempty"`  // Roles that granted the permission
	Reason  string   `json:"reason,omitempty"` // Why permission was allowed/denied
}

// CheckPermissionsResponse represents a batch permission check response.
type CheckPermissionsResponse struct {
	Results map[string]bool `json:"results"` // permission -> allowed
	Roles   []string        `json:"roles"`   // All user's active roles
}

// UserPermissionsResponse represents all permissions for a user.
type UserPermissionsResponse struct {
	UserID      string       `json:"user_id"`
	Roles       []Role       `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

// IsExpired checks if a user role assignment has expired.
func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return ur.ExpiresAt.Before(models.Now())
}

// IsEffectivelyActive checks if user role is active and not expired.
func (ur *UserRole) IsEffectivelyActive() bool {
	return ur.IsActive && !ur.IsExpired()
}
