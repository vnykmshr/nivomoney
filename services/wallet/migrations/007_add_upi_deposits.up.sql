-- Add UPI VPA to wallets
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS upi_vpa VARCHAR(50);

-- Create unique index for UPI VPA
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_upi_vpa ON wallets(upi_vpa) WHERE upi_vpa IS NOT NULL;

-- Create UPI deposits table
CREATE TABLE IF NOT EXISTS upi_deposits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    amount BIGINT NOT NULL, -- Amount in paise
    upi_reference VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, completed, failed, expired
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() + INTERVAL '5 minutes',
    completed_at TIMESTAMP WITH TIME ZONE,
    failed_reason VARCHAR(255)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_upi_deposits_wallet ON upi_deposits(wallet_id);
CREATE INDEX IF NOT EXISTS idx_upi_deposits_user ON upi_deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_upi_deposits_status ON upi_deposits(status);
CREATE INDEX IF NOT EXISTS idx_upi_deposits_expires ON upi_deposits(expires_at) WHERE status = 'pending';

-- Function to generate UPI VPA for wallet
CREATE OR REPLACE FUNCTION generate_upi_vpa(wallet_id UUID)
RETURNS VARCHAR(50) AS $$
DECLARE
    short_id VARCHAR(8);
BEGIN
    -- Generate a short ID from the wallet UUID
    short_id := LOWER(SUBSTRING(REPLACE(wallet_id::TEXT, '-', '') FROM 1 FOR 8));
    RETURN short_id || '@nivo';
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-generate UPI VPA on wallet activation
CREATE OR REPLACE FUNCTION auto_generate_upi_vpa()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'active' AND NEW.upi_vpa IS NULL THEN
        NEW.upi_vpa := generate_upi_vpa(NEW.id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_auto_upi_vpa ON wallets;
CREATE TRIGGER trigger_auto_upi_vpa
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    WHEN (NEW.status = 'active' AND OLD.status != 'active')
    EXECUTE FUNCTION auto_generate_upi_vpa();

-- Update existing active wallets with UPI VPA
UPDATE wallets
SET upi_vpa = generate_upi_vpa(id)
WHERE status = 'active' AND upi_vpa IS NULL;
