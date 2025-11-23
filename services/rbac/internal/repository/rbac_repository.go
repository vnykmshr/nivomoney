package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vnykmshr/nivo/services/rbac/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// RBACRepository handles all RBAC database operations.
type RBACRepository struct {
	db *sql.DB
}

// NewRBACRepository creates a new RBAC repository.
func NewRBACRepository(db *sql.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

// ============================================================================
// Role Operations
// ============================================================================

// CreateRole creates a new role.
func (r *RBACRepository) CreateRole(ctx context.Context, role *models.Role) *errors.Error {
	query := `
		INSERT INTO roles (name, description, parent_role_id, is_system, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		role.Name,
		role.Description,
		role.ParentRoleID,
		role.IsSystem,
		role.IsActive,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("role with this name already exists")
		}
		if isForeignKeyViolation(err) {
			return errors.BadRequest("parent role does not exist")
		}
		return errors.DatabaseWrap(err, "failed to create role")
	}

	return nil
}

// GetRoleByID retrieves a role by ID.
func (r *RBACRepository) GetRoleByID(ctx context.Context, id string) (*models.Role, *errors.Error) {
	role := &models.Role{}

	query := `
		SELECT id, name, description, parent_role_id, is_system, is_active, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.ParentRoleID,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("role", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get role")
	}

	return role, nil
}

// GetRoleByName retrieves a role by name.
func (r *RBACRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, *errors.Error) {
	role := &models.Role{}

	query := `
		SELECT id, name, description, parent_role_id, is_system, is_active, created_at, updated_at
		FROM roles
		WHERE name = $1
	`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.ParentRoleID,
		&role.IsSystem,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound(fmt.Sprintf("role '%s' not found", name))
		}
		return nil, errors.DatabaseWrap(err, "failed to get role")
	}

	return role, nil
}

// ListRoles retrieves all roles.
func (r *RBACRepository) ListRoles(ctx context.Context, activeOnly bool) ([]models.Role, *errors.Error) {
	query := `
		SELECT id, name, description, parent_role_id, is_system, is_active, created_at, updated_at
		FROM roles
	`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list roles")
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.ParentRoleID,
			&role.IsSystem,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan role")
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// UpdateRole updates a role.
func (r *RBACRepository) UpdateRole(ctx context.Context, id string, updates map[string]interface{}) *errors.Error {
	if len(updates) == 0 {
		return errors.BadRequest("no fields to update")
	}

	// Build dynamic UPDATE query
	query := "UPDATE roles SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argPos := 1

	for field, value := range updates {
		query += fmt.Sprintf(", %s = $%d", field, argPos)
		args = append(args, value)
		argPos++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("role with this name already exists")
		}
		if isForeignKeyViolation(err) {
			return errors.BadRequest("parent role does not exist")
		}
		return errors.DatabaseWrap(err, "failed to update role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFoundWithID("role", id)
	}

	return nil
}

// DeleteRole deletes a role (only if not system role).
func (r *RBACRepository) DeleteRole(ctx context.Context, id string) *errors.Error {
	query := `DELETE FROM roles WHERE id = $1 AND is_system = false`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to delete role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.BadRequest("cannot delete system role or role not found")
	}

	return nil
}

// GetRoleHierarchy retrieves a role with all its ancestors (parent chain).
func (r *RBACRepository) GetRoleHierarchy(ctx context.Context, roleID string) ([]models.Role, *errors.Error) {
	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Base case: start with the given role
			SELECT id, name, description, parent_role_id, is_system, is_active, created_at, updated_at, 0 as level
			FROM roles
			WHERE id = $1

			UNION ALL

			-- Recursive case: get parent roles
			SELECT r.id, r.name, r.description, r.parent_role_id, r.is_system, r.is_active, r.created_at, r.updated_at, rh.level + 1
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
		)
		SELECT id, name, description, parent_role_id, is_system, is_active, created_at, updated_at
		FROM role_hierarchy
		ORDER BY level ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get role hierarchy")
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.ParentRoleID,
			&role.IsSystem,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan role hierarchy")
		}
		roles = append(roles, role)
	}

	if len(roles) == 0 {
		return nil, errors.NotFoundWithID("role", roleID)
	}

	return roles, nil
}

// ============================================================================
// Permission Operations
// ============================================================================

// CreatePermission creates a new permission.
func (r *RBACRepository) CreatePermission(ctx context.Context, perm *models.Permission) *errors.Error {
	query := `
		INSERT INTO permissions (name, service, resource, action, description, is_system)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		perm.Name,
		perm.Service,
		perm.Resource,
		perm.Action,
		perm.Description,
		perm.IsSystem,
	).Scan(&perm.ID, &perm.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return errors.Conflict("permission with this name already exists")
		}
		return errors.DatabaseWrap(err, "failed to create permission")
	}

	return nil
}

// GetPermissionByID retrieves a permission by ID.
func (r *RBACRepository) GetPermissionByID(ctx context.Context, id string) (*models.Permission, *errors.Error) {
	perm := &models.Permission{}

	query := `
		SELECT id, name, service, resource, action, description, is_system, created_at
		FROM permissions
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Service,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.IsSystem,
		&perm.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundWithID("permission", id)
		}
		return nil, errors.DatabaseWrap(err, "failed to get permission")
	}

	return perm, nil
}

// GetPermissionByName retrieves a permission by name.
func (r *RBACRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, *errors.Error) {
	perm := &models.Permission{}

	query := `
		SELECT id, name, service, resource, action, description, is_system, created_at
		FROM permissions
		WHERE name = $1
	`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Service,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.IsSystem,
		&perm.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFound(fmt.Sprintf("permission '%s' not found", name))
		}
		return nil, errors.DatabaseWrap(err, "failed to get permission")
	}

	return perm, nil
}

// ListPermissions retrieves all permissions.
func (r *RBACRepository) ListPermissions(ctx context.Context, service string) ([]models.Permission, *errors.Error) {
	query := `
		SELECT id, name, service, resource, action, description, is_system, created_at
		FROM permissions
	`

	args := []interface{}{}
	if service != "" {
		query += " WHERE service = $1"
		args = append(args, service)
	}

	query += " ORDER BY service, resource, action"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to list permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Service,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.IsSystem,
			&perm.CreatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan permission")
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// ============================================================================
// Role-Permission Assignment Operations
// ============================================================================

// AssignPermissionToRole assigns a permission to a role.
func (r *RBACRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string, grantedBy *string) *errors.Error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, grantedBy)
	if err != nil {
		if isForeignKeyViolation(err) {
			return errors.BadRequest("role or permission does not exist")
		}
		return errors.DatabaseWrap(err, "failed to assign permission to role")
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role.
func (r *RBACRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) *errors.Error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to remove permission from role")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("permission assignment not found")
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role (direct only, no hierarchy).
func (r *RBACRepository) GetRolePermissions(ctx context.Context, roleID string) ([]models.Permission, *errors.Error) {
	query := `
		SELECT p.id, p.name, p.service, p.resource, p.action, p.description, p.is_system, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.service, p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get role permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Service,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.IsSystem,
			&perm.CreatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan permission")
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// ============================================================================
// User-Role Assignment Operations
// ============================================================================

// AssignRoleToUser assigns a role to a user.
func (r *RBACRepository) AssignRoleToUser(ctx context.Context, userRole *models.UserRole) *errors.Error {
	query := `
		INSERT INTO user_roles (user_id, role_id, assigned_by, expires_at, is_active)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, role_id) DO UPDATE
		SET is_active = true, expires_at = EXCLUDED.expires_at, assigned_at = CURRENT_TIMESTAMP
		RETURNING assigned_at
	`

	err := r.db.QueryRowContext(ctx, query,
		userRole.UserID,
		userRole.RoleID,
		userRole.AssignedBy,
		userRole.ExpiresAt,
		userRole.IsActive,
	).Scan(&userRole.AssignedAt)

	if err != nil {
		if isForeignKeyViolation(err) {
			return errors.BadRequest("role does not exist")
		}
		return errors.DatabaseWrap(err, "failed to assign role to user")
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user (soft delete).
func (r *RBACRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) *errors.Error {
	query := `UPDATE user_roles SET is_active = false WHERE user_id = $1 AND role_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		return errors.DatabaseWrap(err, "failed to remove role from user")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NotFound("user role assignment not found")
	}

	return nil
}

// GetUserRoles retrieves all active roles for a user.
func (r *RBACRepository) GetUserRoles(ctx context.Context, userID string) ([]models.UserRole, *errors.Error) {
	query := `
		SELECT ur.user_id, ur.role_id, ur.assigned_by, ur.assigned_at, ur.expires_at, ur.is_active
		FROM user_roles ur
		WHERE ur.user_id = $1
		  AND ur.is_active = true
		  AND (ur.expires_at IS NULL OR ur.expires_at > CURRENT_TIMESTAMP)
		ORDER BY ur.assigned_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get user roles")
	}
	defer rows.Close()

	var userRoles []models.UserRole
	for rows.Next() {
		var ur models.UserRole
		if err := rows.Scan(
			&ur.UserID,
			&ur.RoleID,
			&ur.AssignedBy,
			&ur.AssignedAt,
			&ur.ExpiresAt,
			&ur.IsActive,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan user role")
		}
		userRoles = append(userRoles, ur)
	}

	return userRoles, nil
}

// GetUserPermissions retrieves all permissions for a user (includes hierarchy).
func (r *RBACRepository) GetUserPermissions(ctx context.Context, userID string) ([]models.Permission, *errors.Error) {
	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Get user's direct roles
			SELECT r.id, r.name, r.parent_role_id
			FROM roles r
			INNER JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1
			  AND ur.is_active = true
			  AND (ur.expires_at IS NULL OR ur.expires_at > CURRENT_TIMESTAMP)

			UNION ALL

			-- Get parent roles recursively
			SELECT r.id, r.name, r.parent_role_id
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
		)
		SELECT DISTINCT p.id, p.name, p.service, p.resource, p.action, p.description, p.is_system, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN role_hierarchy rh ON rp.role_id = rh.id
		ORDER BY p.service, p.resource, p.action
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.DatabaseWrap(err, "failed to get user permissions")
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Service,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.IsSystem,
			&perm.CreatedAt,
		); err != nil {
			return nil, errors.DatabaseWrap(err, "failed to scan permission")
		}
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission (includes hierarchy).
func (r *RBACRepository) HasPermission(ctx context.Context, userID, permissionName string) (bool, *errors.Error) {
	query := `
		WITH RECURSIVE role_hierarchy AS (
			-- Get user's direct roles
			SELECT r.id, r.parent_role_id
			FROM roles r
			INNER JOIN user_roles ur ON r.id = ur.role_id
			WHERE ur.user_id = $1
			  AND ur.is_active = true
			  AND (ur.expires_at IS NULL OR ur.expires_at > CURRENT_TIMESTAMP)

			UNION ALL

			-- Get parent roles recursively
			SELECT r.id, r.parent_role_id
			FROM roles r
			INNER JOIN role_hierarchy rh ON r.id = rh.parent_role_id
		)
		SELECT EXISTS (
			SELECT 1
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN role_hierarchy rh ON rp.role_id = rh.id
			WHERE p.name = $2
		)
	`

	var hasPermission bool
	err := r.db.QueryRowContext(ctx, query, userID, permissionName).Scan(&hasPermission)
	if err != nil {
		return false, errors.DatabaseWrap(err, "failed to check permission")
	}

	return hasPermission, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

func isUniqueViolation(err error) bool {
	return err != nil && (err.Error() == "pq: duplicate key value violates unique constraint" ||
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "UNIQUE constraint failed"))
}

func isForeignKeyViolation(err error) bool {
	return err != nil && (contains(err.Error(), "foreign key constraint") ||
		contains(err.Error(), "FOREIGN KEY constraint failed"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
