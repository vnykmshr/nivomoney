-- ============================================================================
-- Change email uniqueness to composite (email, account_type)
-- Description: Allows same email for user and user_admin accounts (same person, different portals)
-- ============================================================================

-- Step 1: Drop existing unique constraint on email
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

-- Step 2: Create composite unique index on (email, account_type)
-- This allows: user@example.com with account_type='user' AND user@example.com with account_type='user_admin'
CREATE UNIQUE INDEX idx_users_email_account_type_unique
ON users(email, account_type);

-- Step 3: Add index for efficient lookups by email + account_type
CREATE INDEX IF NOT EXISTS idx_users_email_lookup
ON users(email);

COMMENT ON INDEX idx_users_email_account_type_unique IS 'Allows same email for different account types (user vs user_admin)';
