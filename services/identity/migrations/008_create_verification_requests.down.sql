-- Down Migration: 008_create_verification_requests
-- Removes the verification_requests table

DROP INDEX IF EXISTS idx_verification_operation;
DROP INDEX IF EXISTS idx_verification_pending;
DROP INDEX IF EXISTS idx_verification_expires;
DROP INDEX IF EXISTS idx_verification_user_status;
DROP TABLE IF EXISTS verification_requests;
