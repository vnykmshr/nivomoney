-- Drop updated_at trigger
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;

-- Drop indexes
DROP INDEX IF EXISTS idx_accounts_status;
DROP INDEX IF EXISTS idx_accounts_parent_id;
DROP INDEX IF EXISTS idx_accounts_type;
DROP INDEX IF EXISTS idx_accounts_code;

-- Drop accounts table (including all seed data)
DROP TABLE IF EXISTS accounts CASCADE;

-- Drop shared trigger function (only if no other tables use it)
-- Note: This function is shared across multiple tables, so only drop after all tables are dropped
DROP FUNCTION IF EXISTS update_updated_at_column();
