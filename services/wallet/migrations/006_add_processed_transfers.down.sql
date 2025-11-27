-- Remove processed transfers tracking table
DROP INDEX IF EXISTS idx_processed_transfers_processed_at;
DROP TABLE IF EXISTS processed_transfers;
