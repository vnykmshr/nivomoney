-- Drop indexes
DROP INDEX IF EXISTS idx_role_permissions_role_id;
DROP INDEX IF EXISTS idx_role_permissions_permission_id;
DROP INDEX IF EXISTS idx_role_permissions_granted_at;

-- Drop table
DROP TABLE IF EXISTS role_permissions CASCADE;
