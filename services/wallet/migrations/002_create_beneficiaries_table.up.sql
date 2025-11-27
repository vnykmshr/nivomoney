-- Create beneficiaries table
CREATE TABLE IF NOT EXISTS beneficiaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL,  -- User who saved this beneficiary
    beneficiary_user_id UUID NOT NULL,  -- User being saved as beneficiary
    beneficiary_wallet_id UUID NOT NULL,  -- Default wallet for transfers
    nickname VARCHAR(100) NOT NULL,  -- Friendly name (e.g., "Mom", "John - Rent")
    beneficiary_phone VARCHAR(20) NOT NULL,  -- Phone for display
    metadata JSONB,  -- Additional metadata (last used, frequency, etc.)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT beneficiaries_owner_check CHECK (owner_user_id != beneficiary_user_id),
    CONSTRAINT beneficiaries_nickname_length CHECK (LENGTH(nickname) >= 1 AND LENGTH(nickname) <= 100)
);

-- Create indexes
CREATE INDEX idx_beneficiaries_owner ON beneficiaries(owner_user_id, created_at DESC);
CREATE INDEX idx_beneficiaries_beneficiary_user ON beneficiaries(beneficiary_user_id);

-- Create unique constraint: one beneficiary per user (prevent duplicates)
CREATE UNIQUE INDEX idx_beneficiaries_unique_user
    ON beneficiaries(owner_user_id, beneficiary_user_id);

-- Create unique constraint: nicknames must be unique per owner
CREATE UNIQUE INDEX idx_beneficiaries_unique_nickname
    ON beneficiaries(owner_user_id, LOWER(nickname));

-- Create trigger to update updated_at
CREATE TRIGGER update_beneficiaries_updated_at
    BEFORE UPDATE ON beneficiaries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
