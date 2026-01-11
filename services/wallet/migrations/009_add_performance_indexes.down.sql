-- Rollback performance indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_wallets_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_wallets_user_type_status;
