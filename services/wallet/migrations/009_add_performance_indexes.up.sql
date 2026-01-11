-- Performance indexes for common query patterns
-- Phase 4: v1.4.0 Production Hardening

-- Composite index for user wallet listing sorted by creation date
-- Supports: GET /wallets?user_id=X (with default order by created_at)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wallets_user_created
    ON wallets(user_id, created_at DESC);

-- Composite index for user wallet lookup by type
-- Supports: Finding user's savings/current wallet
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wallets_user_type_status
    ON wallets(user_id, type, status);
