-- Performance indexes for common query patterns
-- Phase 4: v1.4.0 Production Hardening

-- Composite index for filtered transaction queries (status + type)
-- Supports: GET /transactions?status=completed&type=transfer
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_status_type
    ON transactions(status, type);

-- Composite index for wallet transaction history sorted by date
-- Supports: User transaction history with source wallet filter
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_source_wallet_created
    ON transactions(source_wallet_id, created_at DESC)
    WHERE source_wallet_id IS NOT NULL;

-- Composite index for destination wallet lookups with date
-- Supports: Incoming transaction history
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_transactions_dest_wallet_created
    ON transactions(destination_wallet_id, created_at DESC)
    WHERE destination_wallet_id IS NOT NULL;
