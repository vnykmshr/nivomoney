-- Add table to track processed transfers for idempotency
-- This prevents double-processing if the transaction service retries

CREATE TABLE IF NOT EXISTS processed_transfers (
    transaction_id UUID PRIMARY KEY,
    source_wallet_id UUID NOT NULL,
    destination_wallet_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Index for cleanup queries
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on processed_at for efficient cleanup of old records
CREATE INDEX idx_processed_transfers_processed_at ON processed_transfers(processed_at);

-- Add comment
COMMENT ON TABLE processed_transfers IS
'Tracks processed transfers to ensure idempotency. Prevents duplicate execution if transaction service retries.';

COMMENT ON COLUMN processed_transfers.transaction_id IS
'The transaction ID from the transaction service. Used as idempotency key.';
