-- ============================================================================
-- Allow NULL phone for User-Admin accounts
-- Description: User-Admin accounts login via email only, not phone
-- ============================================================================

-- Step 1: Drop the existing unique constraint on phone
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_key;

-- Step 2: Make phone column nullable
ALTER TABLE users ALTER COLUMN phone DROP NOT NULL;

-- Step 3: Create partial unique index (only enforce uniqueness for non-null phones)
-- This allows NULL phones for User-Admin accounts while maintaining uniqueness for regular users
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique
ON users(phone)
WHERE phone IS NOT NULL;

-- Step 4: Add a check constraint to ensure regular users have a phone
-- User-Admin and other admin types can have NULL phone
ALTER TABLE users ADD CONSTRAINT chk_users_phone_required
CHECK (
    account_type IN ('user_admin', 'admin', 'super_admin')
    OR phone IS NOT NULL
);

COMMENT ON COLUMN users.phone IS 'Phone number (required for regular users, NULL for admin accounts)';
