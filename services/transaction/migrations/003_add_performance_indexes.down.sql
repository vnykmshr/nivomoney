-- Rollback performance indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_status_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_source_wallet_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_transactions_dest_wallet_created;
