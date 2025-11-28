-- Remove suspension tracking fields from users table
DROP INDEX IF EXISTS idx_users_suspended_at;

ALTER TABLE users
DROP COLUMN IF EXISTS suspended_by,
DROP COLUMN IF EXISTS suspension_reason,
DROP COLUMN IF EXISTS suspended_at;
