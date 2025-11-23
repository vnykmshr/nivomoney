package service

import (
	"context"
	"testing"
	"time"

	"github.com/vnykmshr/nivo/services/rbac/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================================
// Mock Repository
// ============================================================================

type mockRBACRepository struct {
	// Role storage
	roles       map[string]*models.Role
	rolesByName map[string]*models.Role

	// Permission storage
	permissions map[string]*models.Permission
	permsByName map[string]*models.Permission

	// Role-Permission assignments
	rolePermissions map[string][]string // roleID -> []permissionID

	// User-Role assignments
	userRoles map[string][]models.UserRole // userID -> []UserRole

	// Function hooks for error injection
	getRoleByIDFunc      func(ctx context.Context, id string) (*models.Role, *errors.Error)
	createRoleFunc       func(ctx context.Context, role *models.Role) *errors.Error
	updateRoleFunc       func(ctx context.Context, id string, updates map[string]interface{}) *errors.Error
	deleteRoleFunc       func(ctx context.Context, id string) *errors.Error
	createPermissionFunc func(ctx context.Context, perm *models.Permission) *errors.Error
	hasPermissionFunc    func(ctx context.Context, userID, permission string) (bool, *errors.Error)
}

func newMockRBACRepository() *mockRBACRepository {
	return &mockRBACRepository{
		roles:           make(map[string]*models.Role),
		rolesByName:     make(map[string]*models.Role),
		permissions:     make(map[string]*models.Permission),
		permsByName:     make(map[string]*models.Permission),
		rolePermissions: make(map[string][]string),
		userRoles:       make(map[string][]models.UserRole),
	}
}

// Role methods
func (m *mockRBACRepository) CreateRole(ctx context.Context, role *models.Role) *errors.Error {
	if m.createRoleFunc != nil {
		return m.createRoleFunc(ctx, role)
	}

	// Check for duplicate name
	if _, exists := m.rolesByName[role.Name]; exists {
		return errors.Conflict("role with this name already exists")
	}

	// Generate ID
	role.ID = "role_" + role.Name
	role.CreatedAt = sharedModels.NewTimestamp(time.Now())
	role.UpdatedAt = sharedModels.NewTimestamp(time.Now())

	m.roles[role.ID] = role
	m.rolesByName[role.Name] = role

	return nil
}

func (m *mockRBACRepository) GetRoleByID(ctx context.Context, id string) (*models.Role, *errors.Error) {
	if m.getRoleByIDFunc != nil {
		return m.getRoleByIDFunc(ctx, id)
	}

	role, exists := m.roles[id]
	if !exists {
		return nil, errors.NotFound("role not found")
	}

	// Return a copy
	roleCopy := *role
	return &roleCopy, nil
}

func (m *mockRBACRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, *errors.Error) {
	role, exists := m.rolesByName[name]
	if !exists {
		return nil, errors.NotFound("role not found")
	}

	roleCopy := *role
	return &roleCopy, nil
}

func (m *mockRBACRepository) ListRoles(ctx context.Context, activeOnly bool) ([]models.Role, *errors.Error) {
	var roleList []models.Role

	for _, role := range m.roles {
		if !activeOnly || role.IsActive {
			roleList = append(roleList, *role)
		}
	}

	return roleList, nil
}

func (m *mockRBACRepository) UpdateRole(ctx context.Context, id string, updates map[string]interface{}) *errors.Error {
	if m.updateRoleFunc != nil {
		return m.updateRoleFunc(ctx, id, updates)
	}

	role, exists := m.roles[id]
	if !exists {
		return errors.NotFound("role not found")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		// Remove old name index
		delete(m.rolesByName, role.Name)
		role.Name = name
		m.rolesByName[name] = role
	}
	if desc, ok := updates["description"].(string); ok {
		role.Description = desc
	}
	if parentID, ok := updates["parent_role_id"].(string); ok {
		role.ParentRoleID = &parentID
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		role.IsActive = isActive
	}

	role.UpdatedAt = sharedModels.NewTimestamp(time.Now())

	return nil
}

func (m *mockRBACRepository) DeleteRole(ctx context.Context, id string) *errors.Error {
	if m.deleteRoleFunc != nil {
		return m.deleteRoleFunc(ctx, id)
	}

	role, exists := m.roles[id]
	if !exists {
		return errors.NotFound("role not found")
	}

	if role.IsSystem {
		return errors.Forbidden("cannot delete system role")
	}

	delete(m.roles, id)
	delete(m.rolesByName, role.Name)

	return nil
}

func (m *mockRBACRepository) GetRoleHierarchy(ctx context.Context, roleID string) ([]models.Role, *errors.Error) {
	var hierarchy []models.Role

	currentRole, err := m.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	hierarchy = append(hierarchy, *currentRole)

	// Traverse parent hierarchy
	for currentRole.ParentRoleID != nil {
		parentRole, err := m.GetRoleByID(ctx, *currentRole.ParentRoleID)
		if err != nil {
			break
		}
		hierarchy = append(hierarchy, *parentRole)
		currentRole = parentRole
	}

	return hierarchy, nil
}

// Permission methods
func (m *mockRBACRepository) CreatePermission(ctx context.Context, perm *models.Permission) *errors.Error {
	if m.createPermissionFunc != nil {
		return m.createPermissionFunc(ctx, perm)
	}

	// Check for duplicate name
	if _, exists := m.permsByName[perm.Name]; exists {
		return errors.Conflict("permission with this name already exists")
	}

	// Generate ID
	perm.ID = "perm_" + perm.Name
	perm.CreatedAt = sharedModels.NewTimestamp(time.Now())

	m.permissions[perm.ID] = perm
	m.permsByName[perm.Name] = perm

	return nil
}

func (m *mockRBACRepository) GetPermissionByID(ctx context.Context, id string) (*models.Permission, *errors.Error) {
	perm, exists := m.permissions[id]
	if !exists {
		return nil, errors.NotFound("permission not found")
	}

	permCopy := *perm
	return &permCopy, nil
}

func (m *mockRBACRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, *errors.Error) {
	perm, exists := m.permsByName[name]
	if !exists {
		return nil, errors.NotFound("permission not found")
	}

	permCopy := *perm
	return &permCopy, nil
}

func (m *mockRBACRepository) ListPermissions(ctx context.Context, service string) ([]models.Permission, *errors.Error) {
	var permList []models.Permission

	for _, perm := range m.permissions {
		if service == "" || perm.Service == service {
			permList = append(permList, *perm)
		}
	}

	return permList, nil
}

// Role-Permission assignment methods
func (m *mockRBACRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string, grantedBy *string) *errors.Error {
	// Check role exists
	if _, exists := m.roles[roleID]; !exists {
		return errors.NotFound("role not found")
	}

	// Check permission exists
	if _, exists := m.permissions[permissionID]; !exists {
		return errors.NotFound("permission not found")
	}

	// Check if already assigned
	perms, _ := m.rolePermissions[roleID]
	for _, pid := range perms {
		if pid == permissionID {
			return errors.Conflict("permission already assigned to role")
		}
	}

	m.rolePermissions[roleID] = append(perms, permissionID)

	return nil
}

func (m *mockRBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) *errors.Error {
	perms, exists := m.rolePermissions[roleID]
	if !exists {
		return nil
	}

	// Remove permission
	var updated []string
	for _, pid := range perms {
		if pid != permissionID {
			updated = append(updated, pid)
		}
	}

	m.rolePermissions[roleID] = updated

	return nil
}

func (m *mockRBACRepository) GetRolePermissions(ctx context.Context, roleID string) ([]models.Permission, *errors.Error) {
	permIDs, exists := m.rolePermissions[roleID]
	if !exists {
		return []models.Permission{}, nil
	}

	var perms []models.Permission
	for _, pid := range permIDs {
		if perm, exists := m.permissions[pid]; exists {
			perms = append(perms, *perm)
		}
	}

	return perms, nil
}

// User-Role assignment methods
func (m *mockRBACRepository) AssignRoleToUser(ctx context.Context, userRole *models.UserRole) *errors.Error {
	// Check role exists
	if _, exists := m.roles[userRole.RoleID]; !exists {
		return errors.NotFound("role not found")
	}

	// Set assignment timestamp
	userRole.AssignedAt = sharedModels.NewTimestamp(time.Now())

	// Check if already assigned
	userRoles, _ := m.userRoles[userRole.UserID]
	for _, ur := range userRoles {
		if ur.RoleID == userRole.RoleID {
			return errors.Conflict("role already assigned to user")
		}
	}

	m.userRoles[userRole.UserID] = append(userRoles, *userRole)

	return nil
}

func (m *mockRBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) *errors.Error {
	userRoles, exists := m.userRoles[userID]
	if !exists {
		return nil
	}

	// Remove role
	var updated []models.UserRole
	for _, ur := range userRoles {
		if ur.RoleID != roleID {
			updated = append(updated, ur)
		}
	}

	m.userRoles[userID] = updated

	return nil
}

func (m *mockRBACRepository) GetUserRoles(ctx context.Context, userID string) ([]models.UserRole, *errors.Error) {
	userRoles, exists := m.userRoles[userID]
	if !exists {
		return []models.UserRole{}, nil
	}

	// Filter out expired roles
	var activeRoles []models.UserRole
	now := time.Now()

	for _, ur := range userRoles {
		if !ur.IsActive {
			continue
		}
		if ur.ExpiresAt != nil && ur.ExpiresAt.Time.Before(now) {
			continue
		}
		activeRoles = append(activeRoles, ur)
	}

	return activeRoles, nil
}

func (m *mockRBACRepository) GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, *errors.Error) {
	// Get user roles
	userRoles, err := m.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Collect unique permissions from all roles (including hierarchy)
	permMap := make(map[string]models.Permission)

	for _, ur := range userRoles {
		// Get role hierarchy
		hierarchy, err := m.GetRoleHierarchy(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		// Get permissions from each role in hierarchy
		for _, role := range hierarchy {
			perms, _ := m.GetRolePermissions(ctx, role.ID)
			for _, perm := range perms {
				permMap[perm.ID] = perm
			}
		}
	}

	// Convert to slice
	var allPerms []models.Permission
	for _, perm := range permMap {
		allPerms = append(allPerms, perm)
	}

	return allPerms, nil
}

func (m *mockRBACRepository) HasPermission(ctx context.Context, userID, permissionName string) (bool, *errors.Error) {
	if m.hasPermissionFunc != nil {
		return m.hasPermissionFunc(ctx, userID, permissionName)
	}

	// Get all user permissions
	perms, err := m.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if permission exists
	for _, perm := range perms {
		if perm.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

// ============================================================================
// Tests: Role Operations
// ============================================================================

func TestCreateRole_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	req := &models.CreateRoleRequest{
		Name:        "manager",
		Description: "Manager role",
	}

	role, err := service.CreateRole(ctx, req, "admin_123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if role.Name != "manager" {
		t.Errorf("expected role name 'manager', got %s", role.Name)
	}

	if role.IsSystem {
		t.Error("user-created role should not be system role")
	}

	if !role.IsActive {
		t.Error("new role should be active")
	}
}

func TestCreateRole_WithParent_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create parent role
	parentReq := &models.CreateRoleRequest{
		Name:        "admin",
		Description: "Admin role",
	}
	parent, _ := service.CreateRole(ctx, parentReq, "system")

	// Create child role
	childReq := &models.CreateRoleRequest{
		Name:         "super_admin",
		Description:  "Super Admin role",
		ParentRoleID: &parent.ID,
	}

	child, err := service.CreateRole(ctx, childReq, "admin_123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if child.ParentRoleID == nil || *child.ParentRoleID != parent.ID {
		t.Errorf("expected parent role ID %s, got %v", parent.ID, child.ParentRoleID)
	}
}

func TestCreateRole_Error_InactiveParent(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create inactive parent role
	parentReq := &models.CreateRoleRequest{
		Name:        "inactive_admin",
		Description: "Inactive Admin role",
	}
	parent, _ := service.CreateRole(ctx, parentReq, "system")

	// Deactivate parent
	isActive := false
	service.UpdateRole(ctx, parent.ID, &models.UpdateRoleRequest{
		IsActive: &isActive,
	})

	// Try to create child with inactive parent
	childReq := &models.CreateRoleRequest{
		Name:         "child",
		Description:  "Child role",
		ParentRoleID: &parent.ID,
	}

	_, err := service.CreateRole(ctx, childReq, "admin_123")

	if err == nil {
		t.Fatal("expected error for inactive parent role")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestUpdateRole_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	createReq := &models.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
	}
	role, _ := service.CreateRole(ctx, createReq, "admin_123")

	// Update role
	newName := "senior_editor"
	newDesc := "Senior Editor role"
	updateReq := &models.UpdateRoleRequest{
		Name:        &newName,
		Description: &newDesc,
	}

	updated, err := service.UpdateRole(ctx, role.ID, updateReq)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "senior_editor" {
		t.Errorf("expected updated name 'senior_editor', got %s", updated.Name)
	}

	if updated.Description != "Senior Editor role" {
		t.Errorf("expected updated description, got %s", updated.Description)
	}
}

func TestUpdateRole_Error_SystemRole(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create system role manually
	systemRole := &models.Role{
		Name:     "system_admin",
		IsSystem: true,
		IsActive: true,
	}
	repo.CreateRole(ctx, systemRole)

	// Try to update system role
	newName := "hacked_admin"
	updateReq := &models.UpdateRoleRequest{
		Name: &newName,
	}

	_, err := service.UpdateRole(ctx, systemRole.ID, updateReq)

	if err == nil {
		t.Fatal("expected error when updating system role")
	}

	if err.Code != errors.ErrCodeForbidden {
		t.Errorf("expected forbidden error, got %s", err.Code)
	}
}

func TestUpdateRole_Error_CircularHierarchy(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	req := &models.CreateRoleRequest{
		Name:        "role1",
		Description: "Role 1",
	}
	role, _ := service.CreateRole(ctx, req, "admin_123")

	// Try to set role as its own parent
	updateReq := &models.UpdateRoleRequest{
		ParentRoleID: &role.ID,
	}

	_, err := service.UpdateRole(ctx, role.ID, updateReq)

	if err == nil {
		t.Fatal("expected error for circular hierarchy")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestDeleteRole_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	req := &models.CreateRoleRequest{
		Name:        "temp_role",
		Description: "Temporary role",
	}
	role, _ := service.CreateRole(ctx, req, "admin_123")

	// Delete role
	err := service.DeleteRole(ctx, role.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify role is deleted
	_, getErr := repo.GetRoleByID(ctx, role.ID)
	if getErr == nil {
		t.Error("expected role to be deleted")
	}
}

func TestGetRoleWithHierarchy_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role hierarchy: grandparent -> parent -> child
	grandparentReq := &models.CreateRoleRequest{
		Name:        "user",
		Description: "Basic user",
	}
	grandparent, _ := service.CreateRole(ctx, grandparentReq, "system")

	parentReq := &models.CreateRoleRequest{
		Name:         "editor",
		Description:  "Editor",
		ParentRoleID: &grandparent.ID,
	}
	parent, _ := service.CreateRole(ctx, parentReq, "admin")

	childReq := &models.CreateRoleRequest{
		Name:         "senior_editor",
		Description:  "Senior Editor",
		ParentRoleID: &parent.ID,
	}
	child, _ := service.CreateRole(ctx, childReq, "admin")

	// Get hierarchy
	hierarchy, err := service.GetRoleWithHierarchy(ctx, child.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should return: child, parent, grandparent
	if len(hierarchy) != 3 {
		t.Fatalf("expected 3 roles in hierarchy, got %d", len(hierarchy))
	}

	if hierarchy[0].Name != "senior_editor" {
		t.Errorf("expected first role 'senior_editor', got %s", hierarchy[0].Name)
	}

	if hierarchy[1].Name != "editor" {
		t.Errorf("expected second role 'editor', got %s", hierarchy[1].Name)
	}

	if hierarchy[2].Name != "user" {
		t.Errorf("expected third role 'user', got %s", hierarchy[2].Name)
	}
}

// ============================================================================
// Tests: Permission Operations
// ============================================================================

func TestCreatePermission_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	req := &models.CreatePermissionRequest{
		Name:        "wallet:balance:read",
		Service:     "wallet",
		Resource:    "balance",
		Action:      "read",
		Description: "Read wallet balance",
	}

	perm, err := service.CreatePermission(ctx, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if perm.Name != "wallet:balance:read" {
		t.Errorf("expected permission name 'wallet:balance:read', got %s", perm.Name)
	}

	if perm.Service != "wallet" {
		t.Errorf("expected service 'wallet', got %s", perm.Service)
	}
}

func TestCreatePermission_Error_InvalidNameFormat(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	req := &models.CreatePermissionRequest{
		Name:        "invalid_name",
		Service:     "wallet",
		Resource:    "balance",
		Action:      "read",
		Description: "Read wallet balance",
	}

	_, err := service.CreatePermission(ctx, req)

	if err == nil {
		t.Fatal("expected error for invalid permission name format")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Role-Permission Assignment
// ============================================================================

func TestAssignPermissionToRole_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	roleReq := &models.CreateRoleRequest{
		Name:        "viewer",
		Description: "Viewer role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Create permission
	permReq := &models.CreatePermissionRequest{
		Name:        "wallet:balance:read",
		Service:     "wallet",
		Resource:    "balance",
		Action:      "read",
		Description: "Read wallet balance",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	// Assign permission to role
	assignedBy := "admin_123"
	err := service.AssignPermissionToRole(ctx, role.ID, perm.ID, &assignedBy)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify assignment
	perms, _ := service.GetRolePermissions(ctx, role.ID)
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(perms))
	}

	if perms[0].Name != "wallet:balance:read" {
		t.Errorf("expected permission 'wallet:balance:read', got %s", perms[0].Name)
	}
}

func TestAssignPermissionToRole_Error_SystemRole(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create system role
	systemRole := &models.Role{
		Name:     "system_role",
		IsSystem: true,
		IsActive: true,
	}
	repo.CreateRole(ctx, systemRole)

	// Create permission
	permReq := &models.CreatePermissionRequest{
		Name:        "test:resource:action",
		Service:     "test",
		Resource:    "resource",
		Action:      "action",
		Description: "Test permission",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	// Try to assign permission to system role
	assignedBy := "admin_123"
	err := service.AssignPermissionToRole(ctx, systemRole.ID, perm.ID, &assignedBy)

	if err == nil {
		t.Fatal("expected error when assigning permission to system role")
	}

	if err.Code != errors.ErrCodeForbidden {
		t.Errorf("expected forbidden error, got %s", err.Code)
	}
}

func TestRemovePermissionFromRole_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role and permission
	roleReq := &models.CreateRoleRequest{
		Name:        "editor",
		Description: "Editor role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	permReq := &models.CreatePermissionRequest{
		Name:        "article:content:edit",
		Service:     "article",
		Resource:    "content",
		Action:      "edit",
		Description: "Edit article content",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	// Assign and then remove
	assignedBy := "admin_123"
	service.AssignPermissionToRole(ctx, role.ID, perm.ID, &assignedBy)

	err := service.RemovePermissionFromRole(ctx, role.ID, perm.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify removal
	perms, _ := service.GetRolePermissions(ctx, role.ID)
	if len(perms) != 0 {
		t.Errorf("expected 0 permissions after removal, got %d", len(perms))
	}
}

func TestGetRolePermissionsWithHierarchy_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role hierarchy
	parentReq := &models.CreateRoleRequest{
		Name:        "user",
		Description: "User role",
	}
	parent, _ := service.CreateRole(ctx, parentReq, "system")

	childReq := &models.CreateRoleRequest{
		Name:         "admin",
		Description:  "Admin role",
		ParentRoleID: &parent.ID,
	}
	child, _ := service.CreateRole(ctx, childReq, "system")

	// Create permissions
	readPerm := &models.CreatePermissionRequest{
		Name:        "data:resource:read",
		Service:     "data",
		Resource:    "resource",
		Action:      "read",
		Description: "Read data",
	}
	perm1, _ := service.CreatePermission(ctx, readPerm)

	writePerm := &models.CreatePermissionRequest{
		Name:        "data:resource:write",
		Service:     "data",
		Resource:    "resource",
		Action:      "write",
		Description: "Write data",
	}
	perm2, _ := service.CreatePermission(ctx, writePerm)

	// Assign read to parent, write to child
	assignedBy := "system"
	service.AssignPermissionToRole(ctx, parent.ID, perm1.ID, &assignedBy)
	service.AssignPermissionToRole(ctx, child.ID, perm2.ID, &assignedBy)

	// Get child permissions with hierarchy
	perms, err := service.GetRolePermissionsWithHierarchy(ctx, child.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have both permissions (inherited + direct)
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions (inherited + direct), got %d", len(perms))
	}

	permNames := make(map[string]bool)
	for _, p := range perms {
		permNames[p.Name] = true
	}

	if !permNames["data:resource:read"] {
		t.Error("expected inherited 'read' permission")
	}

	if !permNames["data:resource:write"] {
		t.Error("expected direct 'write' permission")
	}
}

// ============================================================================
// Tests: User-Role Assignment
// ============================================================================

func TestAssignRoleToUser_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	roleReq := &models.CreateRoleRequest{
		Name:        "member",
		Description: "Member role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Assign role to user
	assignReq := &models.AssignRoleToUserRequest{
		UserID: "user_123",
		RoleID: role.ID,
	}
	assignedBy := "admin_456"

	userRole, err := service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userRole.UserID != "user_123" {
		t.Errorf("expected user ID 'user_123', got %s", userRole.UserID)
	}

	if userRole.RoleID != role.ID {
		t.Errorf("expected role ID %s, got %s", role.ID, userRole.RoleID)
	}

	if !userRole.IsActive {
		t.Error("assigned role should be active")
	}
}

func TestAssignRoleToUser_WithExpiry_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	roleReq := &models.CreateRoleRequest{
		Name:        "trial_member",
		Description: "Trial member role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Assign with expiry
	futureTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	assignReq := &models.AssignRoleToUserRequest{
		UserID:    "user_456",
		RoleID:    role.ID,
		ExpiresAt: &futureTime,
	}
	assignedBy := "admin_789"

	userRole, err := service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if userRole.ExpiresAt == nil {
		t.Error("expected expiry time to be set")
	}
}

func TestAssignRoleToUser_Error_InactiveRole(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create and deactivate role
	roleReq := &models.CreateRoleRequest{
		Name:        "inactive_role",
		Description: "Inactive role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	isActive := false
	service.UpdateRole(ctx, role.ID, &models.UpdateRoleRequest{
		IsActive: &isActive,
	})

	// Try to assign inactive role
	assignReq := &models.AssignRoleToUserRequest{
		UserID: "user_789",
		RoleID: role.ID,
	}
	assignedBy := "admin_123"

	_, err := service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	if err == nil {
		t.Fatal("expected error when assigning inactive role")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestAssignRoleToUser_Error_PastExpiry(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	roleReq := &models.CreateRoleRequest{
		Name:        "expired_role",
		Description: "Expired role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Try to assign with past expiry
	pastTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	assignReq := &models.AssignRoleToUserRequest{
		UserID:    "user_999",
		RoleID:    role.ID,
		ExpiresAt: &pastTime,
	}
	assignedBy := "admin_123"

	_, err := service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	if err == nil {
		t.Fatal("expected error for past expiry time")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestRemoveRoleFromUser_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create and assign role
	roleReq := &models.CreateRoleRequest{
		Name:        "temp_member",
		Description: "Temporary member",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	assignReq := &models.AssignRoleToUserRequest{
		UserID: "user_temp",
		RoleID: role.ID,
	}
	assignedBy := "admin_123"
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Remove role
	err := service.RemoveRoleFromUser(ctx, "user_temp", role.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify removal
	userRoles, _ := service.GetUserRoles(ctx, "user_temp")
	if len(userRoles) != 0 {
		t.Errorf("expected 0 roles after removal, got %d", len(userRoles))
	}
}

func TestGetUserPermissions_Success(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role
	roleReq := &models.CreateRoleRequest{
		Name:        "analyst",
		Description: "Analyst role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Create and assign permission
	permReq := &models.CreatePermissionRequest{
		Name:        "analytics:reports:read",
		Service:     "analytics",
		Resource:    "reports",
		Action:      "read",
		Description: "Read analytics reports",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	assignedBy := "admin_123"
	service.AssignPermissionToRole(ctx, role.ID, perm.ID, &assignedBy)

	// Assign role to user
	assignReq := &models.AssignRoleToUserRequest{
		UserID: "analyst_001",
		RoleID: role.ID,
	}
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Get user permissions
	response, err := service.GetUserPermissions(ctx, "analyst_001")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.UserID != "analyst_001" {
		t.Errorf("expected user ID 'analyst_001', got %s", response.UserID)
	}

	if len(response.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(response.Roles))
	}

	if len(response.Permissions) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(response.Permissions))
	}

	if response.Permissions[0].Name != "analytics:reports:read" {
		t.Errorf("expected permission 'analytics:reports:read', got %s", response.Permissions[0].Name)
	}
}

// ============================================================================
// Tests: Permission Checking (Core RBAC Logic)
// ============================================================================

func TestCheckPermission_Allowed(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Setup: Create role, permission, assign to user
	roleReq := &models.CreateRoleRequest{
		Name:        "operator",
		Description: "Operator role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	permReq := &models.CreatePermissionRequest{
		Name:        "system:config:read",
		Service:     "system",
		Resource:    "config",
		Action:      "read",
		Description: "Read system config",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	assignedBy := "admin_123"
	service.AssignPermissionToRole(ctx, role.ID, perm.ID, &assignedBy)

	assignReq := &models.AssignRoleToUserRequest{
		UserID: "operator_001",
		RoleID: role.ID,
	}
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Check permission
	checkReq := &models.CheckPermissionRequest{
		UserID:     "operator_001",
		Permission: "system:config:read",
	}

	response, err := service.CheckPermission(ctx, checkReq)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !response.Allowed {
		t.Error("expected permission to be allowed")
	}

	if len(response.Roles) == 0 {
		t.Error("expected roles to be populated")
	}
}

func TestCheckPermission_Denied(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create user with no permissions
	roleReq := &models.CreateRoleRequest{
		Name:        "guest",
		Description: "Guest role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	assignReq := &models.AssignRoleToUserRequest{
		UserID: "guest_001",
		RoleID: role.ID,
	}
	assignedBy := "admin_123"
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Check permission user doesn't have
	checkReq := &models.CheckPermissionRequest{
		UserID:     "guest_001",
		Permission: "admin:users:delete",
	}

	response, err := service.CheckPermission(ctx, checkReq)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Allowed {
		t.Error("expected permission to be denied")
	}

	if response.Reason == "" {
		t.Error("expected reason to be populated")
	}
}

func TestCheckPermissions_Batch(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Setup role with multiple permissions
	roleReq := &models.CreateRoleRequest{
		Name:        "developer",
		Description: "Developer role",
	}
	role, _ := service.CreateRole(ctx, roleReq, "admin")

	// Create permissions
	readPerm := &models.CreatePermissionRequest{
		Name:        "code:repo:read",
		Service:     "code",
		Resource:    "repo",
		Action:      "read",
		Description: "Read code repository",
	}
	perm1, _ := service.CreatePermission(ctx, readPerm)

	writePerm := &models.CreatePermissionRequest{
		Name:        "code:repo:write",
		Service:     "code",
		Resource:    "repo",
		Action:      "write",
		Description: "Write to code repository",
	}
	service.CreatePermission(ctx, writePerm)

	// Assign only read permission
	assignedBy := "admin_123"
	service.AssignPermissionToRole(ctx, role.ID, perm1.ID, &assignedBy)

	// Assign role to user
	assignReq := &models.AssignRoleToUserRequest{
		UserID: "dev_001",
		RoleID: role.ID,
	}
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Batch check
	checkReq := &models.CheckPermissionsRequest{
		UserID: "dev_001",
		Permissions: []string{
			"code:repo:read",
			"code:repo:write",
			"code:repo:delete",
		},
	}

	response, err := service.CheckPermissions(ctx, checkReq)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(response.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(response.Results))
	}

	if !response.Results["code:repo:read"] {
		t.Error("expected 'read' permission to be allowed")
	}

	if response.Results["code:repo:write"] {
		t.Error("expected 'write' permission to be denied")
	}

	if response.Results["code:repo:delete"] {
		t.Error("expected 'delete' permission to be denied")
	}
}

func TestHasPermission_WithHierarchy(t *testing.T) {
	repo := newMockRBACRepository()
	service := NewRBACService(repo)
	ctx := context.Background()

	// Create role hierarchy
	parentReq := &models.CreateRoleRequest{
		Name:        "viewer",
		Description: "Viewer role",
	}
	parent, _ := service.CreateRole(ctx, parentReq, "system")

	childReq := &models.CreateRoleRequest{
		Name:         "editor",
		Description:  "Editor role",
		ParentRoleID: &parent.ID,
	}
	child, _ := service.CreateRole(ctx, childReq, "system")

	// Create permission and assign to parent
	permReq := &models.CreatePermissionRequest{
		Name:        "content:article:read",
		Service:     "content",
		Resource:    "article",
		Action:      "read",
		Description: "Read articles",
	}
	perm, _ := service.CreatePermission(ctx, permReq)

	assignedBy := "system"
	service.AssignPermissionToRole(ctx, parent.ID, perm.ID, &assignedBy)

	// Assign child role to user (should inherit parent permissions)
	assignReq := &models.AssignRoleToUserRequest{
		UserID: "editor_001",
		RoleID: child.ID,
	}
	service.AssignRoleToUser(ctx, assignReq, &assignedBy)

	// Check inherited permission
	hasPermission, err := service.HasPermission(ctx, "editor_001", "content:article:read")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !hasPermission {
		t.Error("expected user to have inherited permission from parent role")
	}
}
