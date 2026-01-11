-- Performance indexes for common query patterns
-- Phase 4: v1.4.0 Production Hardening

-- Composite index for user session validation
-- Supports: Finding valid sessions for a user (expires_at > NOW())
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_expires
    ON sessions(user_id, expires_at DESC);

-- Partial index for active sessions only
-- Optimizes: Session validation queries that check expiry
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_active
    ON sessions(user_id, token_hash)
    WHERE expires_at > NOW();

-- Composite index for user lookup by status
-- Supports: Admin queries filtering users by status
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_status_created
    ON users(status, created_at DESC);
