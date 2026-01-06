-- Down Migration: 007_email_account_type_unique
-- Reverts to single email uniqueness (will fail if duplicates exist)

-- Drop composite unique index
DROP INDEX IF EXISTS idx_users_email_account_type_unique;

-- Drop email lookup index
DROP INDEX IF EXISTS idx_users_email_lookup;

-- Restore original unique constraint on email
-- NOTE: This will fail if there are duplicate emails with different account_types
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
