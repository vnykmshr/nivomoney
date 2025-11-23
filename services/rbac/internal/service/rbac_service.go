package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vnykmshr/nivo/services/rbac/internal/models"
	"github.com/vnykmshr/nivo/services/rbac/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// RBACService handles all RBAC business logic.
type RBACService struct {
	repo *repository.RBACRepository
}

// NewRBACService creates a new RBAC service.
func NewRBACService(repo *repository.RBACRepository) *RBACService {
	return &RBACService{repo: repo}
}

// ============================================================================
// Role Operations
// ============================================================================

// CreateRole creates a new role with validation.
func (s *RBACService) CreateRole(ctx context.Context, req *models.CreateRoleRequest, creatorID string) (*models.Role, *errors.Error) {
	// Validate parent role exists if specified
	if req.ParentRoleID != nil {
		parentRole, err := s.repo.GetRoleByID(ctx, *req.ParentRoleID)
		if err != nil {
			return nil, err
		}
		if !parentRole.IsActive {
			return nil, errors.BadRequest("parent role is not active")
		}
	}

	role := &models.Role{
		Name:         req.Name,
		Description:  req.Description,
		ParentRoleID: req.ParentRoleID,
		IsSystem:     false, // User-created roles are never system roles
		IsActive:     true,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// GetRole retrieves a role by ID with its permissions.
func (s *RBACService) GetRole(ctx context.Context, roleID string) (*models.Role, *errors.Error) {
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Load direct permissions
	permissions, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	role.Permissions = permissions

	// Load parent role if exists
	if role.ParentRoleID != nil {
		parentRole, err := s.repo.GetRoleByID(ctx, *role.ParentRoleID)
		if err == nil {
			role.ParentRole = parentRole
		}
	}

	return role, nil
}

// GetRoleByName retrieves a role by name.
func (s *RBACService) GetRoleByName(ctx context.Context, name string) (*models.Role, *errors.Error) {
	return s.repo.GetRoleByName(ctx, name)
}

// ListRoles retrieves all roles.
func (s *RBACService) ListRoles(ctx context.Context, activeOnly bool) ([]models.Role, *errors.Error) {
	return s.repo.ListRoles(ctx, activeOnly)
}

// UpdateRole updates a role.
func (s *RBACService) UpdateRole(ctx context.Context, roleID string, req *models.UpdateRoleRequest) (*models.Role, *errors.Error) {
	// Check role exists
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Prevent updating system roles
	if role.IsSystem {
		return nil, errors.Forbidden("cannot update system role")
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ParentRoleID != nil {
		// Validate parent role
		parentRole, err := s.repo.GetRoleByID(ctx, *req.ParentRoleID)
		if err != nil {
			return nil, err
		}
		if !parentRole.IsActive {
			return nil, errors.BadRequest("parent role is not active")
		}
		// Prevent circular hierarchy
		if *req.ParentRoleID == roleID {
			return nil, errors.BadRequest("role cannot be its own parent")
		}
		updates["parent_role_id"] = *req.ParentRoleID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := s.repo.UpdateRole(ctx, roleID, updates); err != nil {
		return nil, err
	}

	// Return updated role
	return s.GetRole(ctx, roleID)
}

// DeleteRole deletes a role (only non-system roles).
func (s *RBACService) DeleteRole(ctx context.Context, roleID string) *errors.Error {
	return s.repo.DeleteRole(ctx, roleID)
}

// GetRoleWithHierarchy retrieves a role with all its parent roles.
func (s *RBACService) GetRoleWithHierarchy(ctx context.Context, roleID string) ([]models.Role, *errors.Error) {
	return s.repo.GetRoleHierarchy(ctx, roleID)
}

// ============================================================================
// Permission Operations
// ============================================================================

// CreatePermission creates a new permission.
func (s *RBACService) CreatePermission(ctx context.Context, req *models.CreatePermissionRequest) (*models.Permission, *errors.Error) {
	// Validate permission name format
	expectedName := fmt.Sprintf("%s:%s:%s", req.Service, req.Resource, req.Action)
	if req.Name != expectedName {
		return nil, errors.BadRequest(fmt.Sprintf("permission name must match format: %s", expectedName))
	}

	perm := &models.Permission{
		Name:        req.Name,
		Service:     req.Service,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		IsSystem:    false, // User-created permissions are not system permissions
	}

	if err := s.repo.CreatePermission(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

// GetPermission retrieves a permission by ID.
func (s *RBACService) GetPermission(ctx context.Context, permissionID string) (*models.Permission, *errors.Error) {
	return s.repo.GetPermissionByID(ctx, permissionID)
}

// GetPermissionByName retrieves a permission by name.
func (s *RBACService) GetPermissionByName(ctx context.Context, name string) (*models.Permission, *errors.Error) {
	return s.repo.GetPermissionByName(ctx, name)
}

// ListPermissions retrieves all permissions, optionally filtered by service.
func (s *RBACService) ListPermissions(ctx context.Context, service string) ([]models.Permission, *errors.Error) {
	return s.repo.ListPermissions(ctx, service)
}

// ============================================================================
// Role-Permission Assignment
// ============================================================================

// AssignPermissionToRole assigns a permission to a role.
func (s *RBACService) AssignPermissionToRole(ctx context.Context, roleID, permissionID string, assignedBy *string) *errors.Error {
	// Validate role exists
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// Prevent modifying system roles
	if role.IsSystem {
		return errors.Forbidden("cannot modify permissions of system role")
	}

	// Validate permission exists
	if _, err := s.repo.GetPermissionByID(ctx, permissionID); err != nil {
		return err
	}

	return s.repo.AssignPermissionToRole(ctx, roleID, permissionID, assignedBy)
}

// RemovePermissionFromRole removes a permission from a role.
func (s *RBACService) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) *errors.Error {
	// Validate role exists and is not system role
	role, err := s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return errors.Forbidden("cannot modify permissions of system role")
	}

	return s.repo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

// GetRolePermissions retrieves all permissions for a role (direct only).
func (s *RBACService) GetRolePermissions(ctx context.Context, roleID string) ([]models.Permission, *errors.Error) {
	// Validate role exists
	if _, err := s.repo.GetRoleByID(ctx, roleID); err != nil {
		return nil, err
	}

	return s.repo.GetRolePermissions(ctx, roleID)
}

// GetRolePermissionsWithHierarchy retrieves all permissions for a role including inherited.
func (s *RBACService) GetRolePermissionsWithHierarchy(ctx context.Context, roleID string) ([]models.Permission, *errors.Error) {
	// Get role hierarchy (current role + all parents)
	roles, err := s.repo.GetRoleHierarchy(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// Collect all unique permissions from role hierarchy
	permissionMap := make(map[string]models.Permission)
	for _, role := range roles {
		permissions, err := s.repo.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, perm := range permissions {
			permissionMap[perm.ID] = perm
		}
	}

	// Convert map to slice
	var allPermissions []models.Permission
	for _, perm := range permissionMap {
		allPermissions = append(allPermissions, perm)
	}

	return allPermissions, nil
}

// ============================================================================
// User-Role Assignment
// ============================================================================

// AssignRoleToUser assigns a role to a user.
func (s *RBACService) AssignRoleToUser(ctx context.Context, req *models.AssignRoleToUserRequest, assignedBy *string) (*models.UserRole, *errors.Error) {
	// Validate role exists and is active
	role, err := s.repo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}

	if !role.IsActive {
		return nil, errors.BadRequest("cannot assign inactive role")
	}

	// Parse expiry time if provided
	var expiresAt *sharedModels.Timestamp
	if req.ExpiresAt != nil {
		parsedTime, parseErr := time.Parse(time.RFC3339, *req.ExpiresAt)
		if parseErr != nil {
			return nil, errors.BadRequest("invalid expires_at format, use ISO 8601")
		}
		if parsedTime.Before(time.Now()) {
			return nil, errors.BadRequest("expires_at must be in the future")
		}
		timestamp := sharedModels.NewTimestamp(parsedTime)
		expiresAt = &timestamp
	}

	userRole := &models.UserRole{
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		AssignedBy: assignedBy,
		ExpiresAt:  expiresAt,
		IsActive:   true,
	}

	if err := s.repo.AssignRoleToUser(ctx, userRole); err != nil {
		return nil, err
	}

	// Load role details
	userRole.Role = role

	return userRole, nil
}

// RemoveRoleFromUser removes a role from a user.
func (s *RBACService) RemoveRoleFromUser(ctx context.Context, userID, roleID string) *errors.Error {
	return s.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

// GetUserRoles retrieves all active roles for a user.
func (s *RBACService) GetUserRoles(ctx context.Context, userID string) ([]models.UserRole, *errors.Error) {
	userRoles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load role details for each assignment
	for i := range userRoles {
		role, roleErr := s.repo.GetRoleByID(ctx, userRoles[i].RoleID)
		if roleErr == nil {
			userRoles[i].Role = role
		}
	}

	return userRoles, nil
}

// GetUserPermissions retrieves all permissions for a user (includes hierarchy).
func (s *RBACService) GetUserPermissions(ctx context.Context, userID string) (*models.UserPermissionsResponse, *errors.Error) {
	// Get user roles
	userRoles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all permissions (with hierarchy)
	permissions, err := s.repo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build response with roles
	var roles []models.Role
	for _, ur := range userRoles {
		role, roleErr := s.repo.GetRoleByID(ctx, ur.RoleID)
		if roleErr == nil {
			roles = append(roles, *role)
		}
	}

	return &models.UserPermissionsResponse{
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// ============================================================================
// Permission Checking (Core RBAC Logic)
// ============================================================================

// CheckPermission checks if a user has a specific permission.
func (s *RBACService) CheckPermission(ctx context.Context, req *models.CheckPermissionRequest) (*models.CheckPermissionResponse, *errors.Error) {
	// Check if user has the permission (includes hierarchy)
	hasPermission, err := s.repo.HasPermission(ctx, req.UserID, req.Permission)
	if err != nil {
		return nil, err
	}

	response := &models.CheckPermissionResponse{
		Allowed: hasPermission,
	}

	// If allowed, get the roles that granted it
	if hasPermission {
		userRoles, err := s.repo.GetUserRoles(ctx, req.UserID)
		if err == nil {
			var roleNames []string
			for _, ur := range userRoles {
				role, roleErr := s.repo.GetRoleByID(ctx, ur.RoleID)
				if roleErr == nil {
					roleNames = append(roleNames, role.Name)
				}
			}
			response.Roles = roleNames
			response.Reason = "Permission granted via roles: " + fmt.Sprint(roleNames)
		}
	} else {
		response.Reason = "User does not have the required permission"
	}

	return response, nil
}

// CheckPermissions checks multiple permissions at once (batch check).
func (s *RBACService) CheckPermissions(ctx context.Context, req *models.CheckPermissionsRequest) (*models.CheckPermissionsResponse, *errors.Error) {
	results := make(map[string]bool)

	// Check each permission
	for _, permission := range req.Permissions {
		hasPermission, err := s.repo.HasPermission(ctx, req.UserID, permission)
		if err != nil {
			return nil, err
		}
		results[permission] = hasPermission
	}

	// Get user roles
	userRoles, err := s.repo.GetUserRoles(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	var roleNames []string
	for _, ur := range userRoles {
		role, roleErr := s.repo.GetRoleByID(ctx, ur.RoleID)
		if roleErr == nil {
			roleNames = append(roleNames, role.Name)
		}
	}

	return &models.CheckPermissionsResponse{
		Results: results,
		Roles:   roleNames,
	}, nil
}

// HasPermission is a convenience method for simple permission checks.
func (s *RBACService) HasPermission(ctx context.Context, userID, permission string) (bool, *errors.Error) {
	return s.repo.HasPermission(ctx, userID, permission)
}
