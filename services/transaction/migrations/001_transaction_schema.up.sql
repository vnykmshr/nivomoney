-- Transaction Service Initial Schema
-- Consolidated migration for pre-release

-- ============================================================================
-- Helper Functions
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Spending Category Type
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'spending_category') THEN
        CREATE TYPE spending_category AS ENUM (
            'food', 'transport', 'utilities', 'entertainment',
            'shopping', 'health', 'education', 'transfer', 'other'
        );
    END IF;
END$$;

-- ============================================================================
-- Transactions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    source_wallet_id UUID,
    destination_wallet_id UUID,
    amount BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    description TEXT NOT NULL,
    reference VARCHAR(100),
    ledger_entry_id UUID,
    parent_transaction_id UUID,
    category spending_category DEFAULT 'other',
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

CREATE INDEX idx_transactions_source_wallet ON transactions(source_wallet_id) WHERE source_wallet_id IS NOT NULL;
CREATE INDEX idx_transactions_destination_wallet ON transactions(destination_wallet_id) WHERE destination_wallet_id IS NOT NULL;
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_reference ON transactions(reference) WHERE reference IS NOT NULL;
CREATE INDEX idx_transactions_ledger_entry ON transactions(ledger_entry_id) WHERE ledger_entry_id IS NOT NULL;
CREATE INDEX idx_transactions_parent ON transactions(parent_transaction_id) WHERE parent_transaction_id IS NOT NULL;
CREATE INDEX idx_transactions_wallet_history ON transactions(source_wallet_id, destination_wallet_id, created_at DESC);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_status_type ON transactions(status, type);
CREATE INDEX idx_transactions_source_wallet_created ON transactions(source_wallet_id, created_at DESC) WHERE source_wallet_id IS NOT NULL;
CREATE INDEX idx_transactions_dest_wallet_created ON transactions(destination_wallet_id, created_at DESC) WHERE destination_wallet_id IS NOT NULL;

CREATE TRIGGER update_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Category Patterns Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS category_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern VARCHAR(100) NOT NULL,
    category spending_category NOT NULL,
    priority INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_category_patterns_priority ON category_patterns(priority DESC);

-- ============================================================================
-- Seed Data: Category Patterns
-- ============================================================================

INSERT INTO category_patterns (pattern, category, priority) VALUES
    ('swiggy', 'food', 10),
    ('zomato', 'food', 10),
    ('domino', 'food', 8),
    ('mcdonald', 'food', 8),
    ('pizza', 'food', 6),
    ('restaurant', 'food', 5),
    ('cafe', 'food', 5),
    ('uber', 'transport', 10),
    ('ola', 'transport', 10),
    ('rapido', 'transport', 8),
    ('petrol', 'transport', 7),
    ('fuel', 'transport', 7),
    ('metro', 'transport', 6),
    ('bus', 'transport', 5),
    ('flipkart', 'shopping', 10),
    ('amazon', 'shopping', 10),
    ('myntra', 'shopping', 8),
    ('ajio', 'shopping', 8),
    ('mall', 'shopping', 5),
    ('store', 'shopping', 4),
    ('electricity', 'utilities', 10),
    ('water bill', 'utilities', 10),
    ('gas bill', 'utilities', 10),
    ('internet', 'utilities', 8),
    ('broadband', 'utilities', 8),
    ('mobile recharge', 'utilities', 7),
    ('netflix', 'entertainment', 10),
    ('spotify', 'entertainment', 10),
    ('hotstar', 'entertainment', 8),
    ('prime video', 'entertainment', 8),
    ('movie', 'entertainment', 6),
    ('cinema', 'entertainment', 6),
    ('hospital', 'health', 10),
    ('pharmacy', 'health', 10),
    ('medical', 'health', 8),
    ('doctor', 'health', 8),
    ('clinic', 'health', 7),
    ('apollo', 'health', 7),
    ('school', 'education', 10),
    ('college', 'education', 10),
    ('university', 'education', 8),
    ('course', 'education', 6),
    ('tuition', 'education', 7),
    ('coaching', 'education', 6)
ON CONFLICT DO NOTHING;
