-- Drop trigger
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;

-- Drop indexes
DROP INDEX IF EXISTS idx_roles_parent_role_id;
DROP INDEX IF EXISTS idx_roles_active;
DROP INDEX IF EXISTS idx_roles_system;

-- Drop table
DROP TABLE IF EXISTS roles CASCADE;
