-- Drop updated_at trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;

-- Drop users table
DROP TABLE IF EXISTS users CASCADE;

-- Drop shared trigger function (only if no other tables use it)
-- Note: This function is shared across multiple tables, so only drop after all tables are dropped
DROP FUNCTION IF EXISTS update_updated_at_column();
