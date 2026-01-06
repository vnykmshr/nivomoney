-- Migration: 005_create_user_admin_pairs
-- Description: Creates the user_admin_pairs table for User-Admin paired account system
-- This enables self-service verification flows without external SMS/email dependencies

-- Create user_admin_pairs table
CREATE TABLE IF NOT EXISTS user_admin_pairs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    admin_user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Foreign keys to users table
    CONSTRAINT fk_user_admin_pairs_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_admin_pairs_admin
        FOREIGN KEY (admin_user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Each user can only have one User-Admin, and each User-Admin is for one user
    CONSTRAINT uq_user_admin_pairs_user UNIQUE (user_id),
    CONSTRAINT uq_user_admin_pairs_admin UNIQUE (admin_user_id)
);

-- Indexes for fast lookups
CREATE INDEX IF NOT EXISTS idx_user_admin_pairs_user_id
    ON user_admin_pairs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_admin_pairs_admin_user_id
    ON user_admin_pairs(admin_user_id);

-- Add account_type column to users table to distinguish account types
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_type VARCHAR(20) DEFAULT 'user';

-- Add index for account_type queries
CREATE INDEX IF NOT EXISTS idx_users_account_type ON users(account_type);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_user_admin_pairs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_user_admin_pairs_updated_at
    BEFORE UPDATE ON user_admin_pairs
    FOR EACH ROW
    EXECUTE FUNCTION update_user_admin_pairs_updated_at();

-- Comments for documentation
COMMENT ON TABLE user_admin_pairs IS 'Maps regular users to their paired User-Admin accounts for self-service verification';
COMMENT ON COLUMN user_admin_pairs.user_id IS 'The regular user account ID';
COMMENT ON COLUMN user_admin_pairs.admin_user_id IS 'The paired User-Admin account ID';
COMMENT ON COLUMN users.account_type IS 'Account type: user, user_admin, admin, super_admin';
