-- Down Migration: 005_create_user_admin_pairs
-- Reverts the user_admin_pairs table and related changes

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_user_admin_pairs_updated_at ON user_admin_pairs;
DROP FUNCTION IF EXISTS update_user_admin_pairs_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_user_admin_pairs_user_id;
DROP INDEX IF EXISTS idx_user_admin_pairs_admin_user_id;
DROP INDEX IF EXISTS idx_users_account_type;

-- Drop the user_admin_pairs table
DROP TABLE IF EXISTS user_admin_pairs;

-- Remove account_type column from users
ALTER TABLE users DROP COLUMN IF EXISTS account_type;
