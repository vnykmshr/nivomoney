-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    source_wallet_id UUID,  -- NULL for deposits
    destination_wallet_id UUID,  -- NULL for withdrawals
    amount BIGINT NOT NULL,  -- In smallest unit (paise for INR)
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    description TEXT NOT NULL,
    reference VARCHAR(100),  -- External reference ID
    ledger_entry_id UUID,  -- Reference to ledger service entry
    parent_transaction_id UUID,  -- For reversals and refunds
    metadata JSONB,
    failure_reason TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT transactions_type_check CHECK (type IN ('transfer', 'deposit', 'withdrawal', 'reversal', 'fee', 'refund')),
    CONSTRAINT transactions_status_check CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'reversed', 'cancelled')),
    CONSTRAINT transactions_amount_check CHECK (amount > 0),
    CONSTRAINT transactions_transfer_check CHECK (
        (type = 'transfer' AND source_wallet_id IS NOT NULL AND destination_wallet_id IS NOT NULL) OR
        (type = 'deposit' AND destination_wallet_id IS NOT NULL) OR
        (type = 'withdrawal' AND source_wallet_id IS NOT NULL) OR
        (type IN ('reversal', 'fee', 'refund'))
    ),
    CONSTRAINT transactions_status_reason_check CHECK (
        (status = 'failed' AND failure_reason IS NOT NULL) OR
        (status != 'failed')
    ),
    CONSTRAINT transactions_parent_check CHECK (
        (type IN ('reversal', 'refund') AND parent_transaction_id IS NOT NULL) OR
        (type NOT IN ('reversal', 'refund'))
    ),
    CONSTRAINT transactions_self_transfer_check CHECK (
        source_wallet_id IS NULL OR destination_wallet_id IS NULL OR
        source_wallet_id != destination_wallet_id
    )
);

-- Create indexes
CREATE INDEX idx_transactions_source_wallet ON transactions(source_wallet_id) WHERE source_wallet_id IS NOT NULL;
CREATE INDEX idx_transactions_destination_wallet ON transactions(destination_wallet_id) WHERE destination_wallet_id IS NOT NULL;
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_reference ON transactions(reference) WHERE reference IS NOT NULL;
CREATE INDEX idx_transactions_ledger_entry ON transactions(ledger_entry_id) WHERE ledger_entry_id IS NOT NULL;
CREATE INDEX idx_transactions_parent ON transactions(parent_transaction_id) WHERE parent_transaction_id IS NOT NULL;

-- Create composite index for wallet transaction history
CREATE INDEX idx_transactions_wallet_history ON transactions(source_wallet_id, destination_wallet_id, created_at DESC);

-- Create trigger to update updated_at
CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
