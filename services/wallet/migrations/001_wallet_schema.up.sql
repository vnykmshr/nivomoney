-- Wallet Service Initial Schema
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
-- Wallets Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'savings',
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    balance BIGINT NOT NULL DEFAULT 0,
    available_balance BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    ledger_account_id UUID NOT NULL,
    upi_vpa VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    closed_reason TEXT,

    CONSTRAINT wallets_type_check CHECK (type IN ('default', 'savings', 'current', 'fixed')),
    CONSTRAINT wallets_status_check CHECK (status IN ('active', 'frozen', 'closed', 'inactive')),
    CONSTRAINT wallets_balance_check CHECK (balance >= 0),
    CONSTRAINT wallets_available_balance_check CHECK (available_balance >= 0 AND available_balance <= balance),
    CONSTRAINT wallets_closed_check CHECK (
        (status = 'closed' AND closed_at IS NOT NULL AND closed_reason IS NOT NULL) OR
        (status != 'closed' AND closed_at IS NULL AND closed_reason IS NULL)
    )
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_status ON wallets(status);
CREATE INDEX idx_wallets_ledger_account ON wallets(ledger_account_id);
CREATE INDEX idx_wallets_created_at ON wallets(created_at DESC);
CREATE INDEX idx_wallets_user_created ON wallets(user_id, created_at DESC);
CREATE INDEX idx_wallets_user_type_status ON wallets(user_id, type, status);

CREATE UNIQUE INDEX idx_wallets_unique_active
    ON wallets(user_id, type, currency)
    WHERE status IN ('active', 'frozen', 'inactive');

CREATE UNIQUE INDEX idx_wallets_upi_vpa ON wallets(upi_vpa) WHERE upi_vpa IS NOT NULL;

CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION sync_wallet_available_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        NEW.available_balance := NEW.balance;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_wallet_available_balance_trigger
    BEFORE INSERT ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION sync_wallet_available_balance();

-- UPI VPA Functions
CREATE OR REPLACE FUNCTION generate_upi_vpa(wallet_id UUID)
RETURNS VARCHAR(50) AS $$
DECLARE
    short_id VARCHAR(8);
BEGIN
    short_id := LOWER(SUBSTRING(REPLACE(wallet_id::TEXT, '-', '') FROM 1 FOR 8));
    RETURN short_id || '@nivo';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION auto_generate_upi_vpa()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'active' AND NEW.upi_vpa IS NULL THEN
        NEW.upi_vpa := generate_upi_vpa(NEW.id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_auto_upi_vpa
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    WHEN (NEW.status = 'active' AND OLD.status != 'active')
    EXECUTE FUNCTION auto_generate_upi_vpa();

-- ============================================================================
-- Beneficiaries Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL,
    beneficiary_user_id UUID NOT NULL,
    beneficiary_wallet_id UUID NOT NULL,
    nickname VARCHAR(100) NOT NULL,
    beneficiary_phone VARCHAR(20) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT beneficiaries_owner_check CHECK (owner_user_id != beneficiary_user_id),
    CONSTRAINT beneficiaries_nickname_length CHECK (LENGTH(nickname) >= 1 AND LENGTH(nickname) <= 100)
);

CREATE INDEX idx_beneficiaries_owner ON beneficiaries(owner_user_id, created_at DESC);
CREATE INDEX idx_beneficiaries_beneficiary_user ON beneficiaries(beneficiary_user_id);

CREATE UNIQUE INDEX idx_beneficiaries_unique_user
    ON beneficiaries(owner_user_id, beneficiary_user_id);

CREATE UNIQUE INDEX idx_beneficiaries_unique_nickname
    ON beneficiaries(owner_user_id, LOWER(nickname));

CREATE TRIGGER update_beneficiaries_updated_at
    BEFORE UPDATE ON beneficiaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Wallet Limits Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS wallet_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,

    daily_limit BIGINT NOT NULL DEFAULT 1000000,
    daily_spent BIGINT NOT NULL DEFAULT 0,
    daily_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT DATE_TRUNC('day', NOW() + INTERVAL '1 day'),

    monthly_limit BIGINT NOT NULL DEFAULT 10000000,
    monthly_spent BIGINT NOT NULL DEFAULT 0,
    monthly_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT DATE_TRUNC('month', NOW() + INTERVAL '1 month'),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT wallet_limits_daily_limit_check CHECK (daily_limit > 0),
    CONSTRAINT wallet_limits_monthly_limit_check CHECK (monthly_limit > 0),
    CONSTRAINT wallet_limits_daily_spent_check CHECK (daily_spent >= 0),
    CONSTRAINT wallet_limits_monthly_spent_check CHECK (monthly_spent >= 0),
    CONSTRAINT wallet_limits_daily_lte_monthly_check CHECK (daily_limit <= monthly_limit)
);

CREATE UNIQUE INDEX idx_wallet_limits_wallet_id ON wallet_limits(wallet_id);
CREATE INDEX idx_wallet_limits_daily_reset ON wallet_limits(daily_reset_at);
CREATE INDEX idx_wallet_limits_monthly_reset ON wallet_limits(monthly_reset_at);

CREATE TRIGGER update_wallet_limits_updated_at
    BEFORE UPDATE ON wallet_limits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to check and reset wallet limits on read
CREATE OR REPLACE FUNCTION check_and_reset_wallet_limits(
    p_wallet_id UUID,
    p_daily_limit BIGINT,
    p_daily_spent BIGINT,
    p_daily_reset_at TIMESTAMP WITH TIME ZONE,
    p_monthly_limit BIGINT,
    p_monthly_spent BIGINT,
    p_monthly_reset_at TIMESTAMP WITH TIME ZONE,
    OUT new_daily_spent BIGINT,
    OUT new_daily_reset_at TIMESTAMP WITH TIME ZONE,
    OUT new_monthly_spent BIGINT,
    OUT new_monthly_reset_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    IF NOW() >= p_daily_reset_at THEN
        new_daily_spent := 0;
        new_daily_reset_at := DATE_TRUNC('day', NOW() + INTERVAL '1 day');
    ELSE
        new_daily_spent := p_daily_spent;
        new_daily_reset_at := p_daily_reset_at;
    END IF;

    IF NOW() >= p_monthly_reset_at THEN
        new_monthly_spent := 0;
        new_monthly_reset_at := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    ELSE
        new_monthly_spent := p_monthly_spent;
        new_monthly_reset_at := p_monthly_reset_at;
    END IF;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_and_reset_wallet_limits IS
'Checks if wallet limits have expired and returns reset values. Call this function when reading limits before checking/reserving transfer amounts.';

COMMENT ON CONSTRAINT wallet_limits_daily_lte_monthly_check ON wallet_limits IS
'Ensures daily transfer limit cannot exceed monthly limit. Enforced at database level for defense in depth.';

-- ============================================================================
-- Processed Transfers Table (Idempotency)
-- ============================================================================

CREATE TABLE IF NOT EXISTS processed_transfers (
    transaction_id UUID PRIMARY KEY,
    source_wallet_id UUID NOT NULL,
    destination_wallet_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processed_transfers_processed_at ON processed_transfers(processed_at);

COMMENT ON TABLE processed_transfers IS
'Tracks processed transfers to ensure idempotency. Prevents duplicate execution if transaction service retries.';

COMMENT ON COLUMN processed_transfers.transaction_id IS
'The transaction ID from the transaction service. Used as idempotency key.';

-- ============================================================================
-- UPI Deposits Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS upi_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    upi_reference VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '5 minutes',
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_reason VARCHAR(255)
);

CREATE INDEX idx_upi_deposits_wallet ON upi_deposits(wallet_id);
CREATE INDEX idx_upi_deposits_user ON upi_deposits(user_id);
CREATE INDEX idx_upi_deposits_status ON upi_deposits(status);
CREATE INDEX idx_upi_deposits_expires ON upi_deposits(expires_at) WHERE status = 'pending';

-- ============================================================================
-- Virtual Cards
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'card_status') THEN
        CREATE TYPE card_status AS ENUM ('active', 'frozen', 'expired', 'cancelled');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'card_type') THEN
        CREATE TYPE card_type AS ENUM ('virtual', 'physical');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS virtual_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    card_number VARCHAR(16) NOT NULL,
    card_holder_name VARCHAR(100) NOT NULL,
    expiry_month INT NOT NULL CHECK (expiry_month >= 1 AND expiry_month <= 12),
    expiry_year INT NOT NULL CHECK (expiry_year >= 2024 AND expiry_year <= 2040),
    cvv VARCHAR(64) NOT NULL,
    card_type card_type NOT NULL DEFAULT 'virtual',
    status card_status NOT NULL DEFAULT 'active',
    daily_limit BIGINT DEFAULT 5000000,
    monthly_limit BIGINT DEFAULT 20000000,
    per_transaction_limit BIGINT DEFAULT 1000000,
    daily_spent BIGINT DEFAULT 0,
    monthly_spent BIGINT DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    frozen_at TIMESTAMP WITH TIME ZONE,
    frozen_reason VARCHAR(500),
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancelled_reason VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_virtual_cards_wallet_id ON virtual_cards(wallet_id);
CREATE INDEX idx_virtual_cards_user_id ON virtual_cards(user_id);
CREATE INDEX idx_virtual_cards_status ON virtual_cards(status);
CREATE UNIQUE INDEX idx_virtual_cards_number ON virtual_cards(card_number);

CREATE OR REPLACE FUNCTION update_virtual_cards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER virtual_cards_updated_at
    BEFORE UPDATE ON virtual_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_virtual_cards_updated_at();
