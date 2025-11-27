-- Create wallet_limits table to track transfer limits
CREATE TABLE IF NOT EXISTS wallet_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,

    -- Daily limits
    daily_limit BIGINT NOT NULL DEFAULT 1000000,  -- Default ₹10,000 in paise
    daily_spent BIGINT NOT NULL DEFAULT 0,
    daily_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT DATE_TRUNC('day', NOW() + INTERVAL '1 day'),

    -- Monthly limits
    monthly_limit BIGINT NOT NULL DEFAULT 10000000,  -- Default ₹100,000 in paise
    monthly_spent BIGINT NOT NULL DEFAULT 0,
    monthly_reset_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT DATE_TRUNC('month', NOW() + INTERVAL '1 month'),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT wallet_limits_daily_limit_check CHECK (daily_limit > 0),
    CONSTRAINT wallet_limits_monthly_limit_check CHECK (monthly_limit > 0),
    CONSTRAINT wallet_limits_daily_spent_check CHECK (daily_spent >= 0),
    CONSTRAINT wallet_limits_monthly_spent_check CHECK (monthly_spent >= 0)
);

-- One set of limits per wallet
CREATE UNIQUE INDEX idx_wallet_limits_wallet_id ON wallet_limits(wallet_id);

-- Index for checking expired limits
CREATE INDEX idx_wallet_limits_daily_reset ON wallet_limits(daily_reset_at);
CREATE INDEX idx_wallet_limits_monthly_reset ON wallet_limits(monthly_reset_at);

-- Create trigger to update updated_at
CREATE TRIGGER update_wallet_limits_updated_at
    BEFORE UPDATE ON wallet_limits
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically reset daily/monthly spent amounts
CREATE OR REPLACE FUNCTION reset_wallet_limit_if_expired()
RETURNS TRIGGER AS $$
BEGIN
    -- Reset daily spent if past reset time
    IF NOW() >= OLD.daily_reset_at THEN
        NEW.daily_spent := 0;
        NEW.daily_reset_at := DATE_TRUNC('day', NOW() + INTERVAL '1 day');
    END IF;

    -- Reset monthly spent if past reset time
    IF NOW() >= OLD.monthly_reset_at THEN
        NEW.monthly_spent := 0;
        NEW.monthly_reset_at := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reset_wallet_limit_trigger
    BEFORE UPDATE ON wallet_limits
    FOR EACH ROW
    EXECUTE FUNCTION reset_wallet_limit_if_expired();

-- Create default limits for existing wallets
INSERT INTO wallet_limits (wallet_id)
SELECT id FROM wallets
WHERE NOT EXISTS (
    SELECT 1 FROM wallet_limits WHERE wallet_limits.wallet_id = wallets.id
);
