-- Down Migration: 006_allow_null_phone_for_user_admin
-- Reverts phone column changes

-- Remove check constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_phone_required;

-- Drop partial unique index
DROP INDEX IF EXISTS idx_users_phone_unique;

-- Make phone NOT NULL again (will fail if there are NULL phones)
-- In production, you would need to handle NULL values before this
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;

-- Recreate the unique constraint
ALTER TABLE users ADD CONSTRAINT users_phone_key UNIQUE (phone);
