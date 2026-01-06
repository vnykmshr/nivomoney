package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/rbac/internal/models"
	"github.com/vnykmshr/nivo/services/rbac/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// RBACHandler handles all RBAC HTTP requests.
type RBACHandler struct {
	service *service.RBACService
}

// NewRBACHandler creates a new RBAC handler.
func NewRBACHandler(service *service.RBACService) *RBACHandler {
	return &RBACHandler{service: service}
}

// ============================================================================
// Role Handlers
// ============================================================================

// CreateRole handles POST /api/v1/roles
func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateRoleRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// TODO: Get creator user ID from auth context (for now, use nil)
	var creatorID string

	// Create role
	role, createErr := h.service.CreateRole(r.Context(), &req, creatorID)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Success(w, http.StatusCreated, role)
}

// GetRole handles GET /api/v1/roles/{id}
func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	role, err := h.service.GetRole(r.Context(), roleID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, role)
}

// ListRoles handles GET /api/v1/roles
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	// Check if we should filter to active only
	activeOnly := r.URL.Query().Get("active") == "true"

	roles, err := h.service.ListRoles(r.Context(), activeOnly)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, roles)
}

// UpdateRole handles PUT /api/v1/roles/{id}
func (h *RBACHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.UpdateRoleRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Update role
	role, updateErr := h.service.UpdateRole(r.Context(), roleID, &req)
	if updateErr != nil {
		response.Error(w, updateErr)
		return
	}

	response.Success(w, http.StatusOK, role)
}

// DeleteRole handles DELETE /api/v1/roles/{id}
func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	if err := h.service.DeleteRole(r.Context(), roleID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "role deleted successfully"})
}

// GetRoleHierarchy handles GET /api/v1/roles/{id}/hierarchy
func (h *RBACHandler) GetRoleHierarchy(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	roles, err := h.service.GetRoleWithHierarchy(r.Context(), roleID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, roles)
}

// ============================================================================
// Permission Handlers
// ============================================================================

// CreatePermission handles POST /api/v1/permissions
func (h *RBACHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreatePermissionRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Create permission
	permission, createErr := h.service.CreatePermission(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Success(w, http.StatusCreated, permission)
}

// GetPermission handles GET /api/v1/permissions/{id}
func (h *RBACHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	permissionID := r.PathValue("id")
	if permissionID == "" {
		response.Error(w, errors.BadRequest("permission ID is required"))
		return
	}

	permission, err := h.service.GetPermission(r.Context(), permissionID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, permission)
}

// ListPermissions handles GET /api/v1/permissions
func (h *RBACHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	// Optional service filter
	service := r.URL.Query().Get("service")

	permissions, err := h.service.ListPermissions(r.Context(), service)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, permissions)
}

// ============================================================================
// Role-Permission Assignment Handlers
// ============================================================================

// AssignPermissionToRole handles POST /api/v1/roles/{id}/permissions
func (h *RBACHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.AssignPermissionToRoleRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// TODO: Get assignor user ID from auth context
	var assignedBy *string

	// Assign permission
	if assignErr := h.service.AssignPermissionToRole(r.Context(), roleID, req.PermissionID, assignedBy); assignErr != nil {
		response.Error(w, assignErr)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "permission assigned to role successfully"})
}

// RemovePermissionFromRole handles DELETE /api/v1/roles/{roleId}/permissions/{permissionId}
func (h *RBACHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("roleId")
	permissionID := r.PathValue("permissionId")

	if roleID == "" || permissionID == "" {
		response.Error(w, errors.BadRequest("role ID and permission ID are required"))
		return
	}

	if err := h.service.RemovePermissionFromRole(r.Context(), roleID, permissionID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "permission removed from role successfully"})
}

// GetRolePermissions handles GET /api/v1/roles/{id}/permissions
func (h *RBACHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		response.Error(w, errors.BadRequest("role ID is required"))
		return
	}

	// Check if we should include inherited permissions
	includeInherited := r.URL.Query().Get("inherited") == "true"

	var permissions []models.Permission
	var err *errors.Error

	if includeInherited {
		permissions, err = h.service.GetRolePermissionsWithHierarchy(r.Context(), roleID)
	} else {
		permissions, err = h.service.GetRolePermissions(r.Context(), roleID)
	}

	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, permissions)
}

// ============================================================================
// User-Role Assignment Handlers
// ============================================================================

// AssignRoleToUser handles POST /api/v1/users/{userId}/roles
func (h *RBACHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.AssignRoleToUserRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Set user ID from path
	req.UserID = userID

	// TODO: Get assignor user ID from auth context
	var assignedBy *string

	// Assign role
	userRole, assignErr := h.service.AssignRoleToUser(r.Context(), &req, assignedBy)
	if assignErr != nil {
		response.Error(w, assignErr)
		return
	}

	response.Success(w, http.StatusCreated, userRole)
}

// AssignDefaultRoleInternal handles POST /internal/v1/users/{userId}/assign-default-role
// This is an internal endpoint for service-to-service communication (no auth required).
func (h *RBACHandler) AssignDefaultRoleInternal(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	// Hardcoded default "user" role ID from seed data
	const defaultUserRoleID = "00000000-0000-0000-0000-000000000001"

	// Create assignment request
	req := models.AssignRoleToUserRequest{
		UserID: userID,
		RoleID: defaultUserRoleID,
	}

	// Assign role (no assignedBy for internal service calls)
	userRole, assignErr := h.service.AssignRoleToUser(r.Context(), &req, nil)
	if assignErr != nil {
		response.Error(w, assignErr)
		return
	}

	response.Success(w, http.StatusCreated, userRole)
}

// AssignRoleByNameInternal handles POST /internal/v1/users/{userId}/assign-role
// This is an internal endpoint for service-to-service communication (no auth required).
// Request body: {"role_name": "user_admin"}
func (h *RBACHandler) AssignRoleByNameInternal(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse role name from body
	var req struct {
		RoleName string `json:"role_name"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.BadRequest("invalid request body"))
		return
	}

	if req.RoleName == "" {
		response.Error(w, errors.BadRequest("role_name is required"))
		return
	}

	// Look up role by name
	role, roleErr := h.service.GetRoleByName(r.Context(), req.RoleName)
	if roleErr != nil {
		response.Error(w, roleErr)
		return
	}

	// Create assignment request
	assignReq := models.AssignRoleToUserRequest{
		UserID: userID,
		RoleID: role.ID,
	}

	// Assign role (no assignedBy for internal service calls)
	userRole, assignErr := h.service.AssignRoleToUser(r.Context(), &assignReq, nil)
	if assignErr != nil {
		response.Error(w, assignErr)
		return
	}

	response.Success(w, http.StatusCreated, userRole)
}

// RemoveRoleFromUser handles DELETE /api/v1/users/{userId}/roles/{roleId}
func (h *RBACHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	roleID := r.PathValue("roleId")

	if userID == "" || roleID == "" {
		response.Error(w, errors.BadRequest("user ID and role ID are required"))
		return
	}

	if err := h.service.RemoveRoleFromUser(r.Context(), userID, roleID); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "role removed from user successfully"})
}

// GetUserRoles handles GET /api/v1/users/{userId}/roles
func (h *RBACHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	userRoles, err := h.service.GetUserRoles(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, userRoles)
}

// GetUserPermissions handles GET /api/v1/users/{userId}/permissions
func (h *RBACHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	permissions, err := h.service.GetUserPermissions(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, permissions)
}

// GetUserPermissionsInternal handles GET /internal/v1/users/{userId}/permissions
// This is an internal endpoint for service-to-service communication (no authentication required).
func (h *RBACHandler) GetUserPermissionsInternal(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	permissions, err := h.service.GetUserPermissions(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, http.StatusOK, permissions)
}

// ============================================================================
// Permission Check Handlers
// ============================================================================

// CheckPermission handles POST /api/v1/check-permission
func (h *RBACHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CheckPermissionRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Check permission
	result, checkErr := h.service.CheckPermission(r.Context(), &req)
	if checkErr != nil {
		response.Error(w, checkErr)
		return
	}

	response.Success(w, http.StatusOK, result)
}

// CheckPermissions handles POST /api/v1/check-permissions (batch)
func (h *RBACHandler) CheckPermissions(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CheckPermissionsRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Check permissions
	result, checkErr := h.service.CheckPermissions(r.Context(), &req)
	if checkErr != nil {
		response.Error(w, checkErr)
		return
	}

	response.Success(w, http.StatusOK, result)
}
