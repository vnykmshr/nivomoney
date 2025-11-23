-- Drop function
DROP FUNCTION IF EXISTS is_user_role_expired(TIMESTAMP WITH TIME ZONE);

-- Drop indexes
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_active;
DROP INDEX IF EXISTS idx_user_roles_expires_at;
DROP INDEX IF EXISTS idx_user_roles_assigned_at;

-- Drop table
DROP TABLE IF EXISTS user_roles CASCADE;
