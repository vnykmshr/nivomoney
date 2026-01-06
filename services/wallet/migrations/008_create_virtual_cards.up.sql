-- Create virtual card status enum
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'card_status') THEN
        CREATE TYPE card_status AS ENUM ('active', 'frozen', 'expired', 'cancelled');
    END IF;
END$$;

-- Create virtual card type enum
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'card_type') THEN
        CREATE TYPE card_type AS ENUM ('virtual', 'physical');
    END IF;
END$$;

-- Create virtual cards table
CREATE TABLE IF NOT EXISTS virtual_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    card_number VARCHAR(16) NOT NULL,
    card_holder_name VARCHAR(100) NOT NULL,
    expiry_month INT NOT NULL CHECK (expiry_month >= 1 AND expiry_month <= 12),
    expiry_year INT NOT NULL CHECK (expiry_year >= 2024 AND expiry_year <= 2040),
    cvv VARCHAR(64) NOT NULL, -- Stored encrypted/hashed
    card_type card_type NOT NULL DEFAULT 'virtual',
    status card_status NOT NULL DEFAULT 'active',
    daily_limit BIGINT DEFAULT 5000000, -- Default ₹50,000 daily limit (in paise)
    monthly_limit BIGINT DEFAULT 20000000, -- Default ₹2,00,000 monthly limit (in paise)
    per_transaction_limit BIGINT DEFAULT 1000000, -- Default ₹10,000 per transaction (in paise)
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

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_virtual_cards_wallet_id ON virtual_cards(wallet_id);
CREATE INDEX IF NOT EXISTS idx_virtual_cards_user_id ON virtual_cards(user_id);
CREATE INDEX IF NOT EXISTS idx_virtual_cards_status ON virtual_cards(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_virtual_cards_number ON virtual_cards(card_number);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_virtual_cards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS virtual_cards_updated_at ON virtual_cards;
CREATE TRIGGER virtual_cards_updated_at
    BEFORE UPDATE ON virtual_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_virtual_cards_updated_at();
