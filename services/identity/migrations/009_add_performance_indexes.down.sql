-- Rollback performance indexes

DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_user_expires;
DROP INDEX CONCURRENTLY IF EXISTS idx_sessions_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_status_created;
