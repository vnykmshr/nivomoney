-- Drop cleanup function
DROP FUNCTION IF EXISTS cleanup_expired_sessions();

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_sessions_user_id;

-- Drop sessions table
DROP TABLE IF EXISTS sessions;
