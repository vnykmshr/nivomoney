-- Identity Service Initial Schema
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
-- Users Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    full_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    account_type VARCHAR(20) NOT NULL DEFAULT 'user',

    -- Suspension tracking
    suspended_at TIMESTAMP WITH TIME ZONE,
    suspension_reason TEXT,
    suspended_by UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_check CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT users_phone_check CHECK (phone IS NULL OR phone ~* '^\+91[6-9][0-9]{9}$'),
    CONSTRAINT users_status_check CHECK (status IN ('pending', 'active', 'suspended', 'closed')),
    CONSTRAINT chk_users_phone_required CHECK (
        account_type IN ('user_admin', 'admin', 'super_admin') OR phone IS NOT NULL
    )
);

-- Self-referencing FK for suspended_by
ALTER TABLE users ADD CONSTRAINT fk_users_suspended_by
    FOREIGN KEY (suspended_by) REFERENCES users(id);

-- Composite unique index on (email, account_type)
CREATE UNIQUE INDEX idx_users_email_account_type_unique ON users(email, account_type);

-- Partial unique index for phone (only non-null)
CREATE UNIQUE INDEX idx_users_phone_unique ON users(phone) WHERE phone IS NOT NULL;

-- Query indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_account_type ON users(account_type);
CREATE INDEX idx_users_suspended_at ON users(suspended_at) WHERE suspended_at IS NOT NULL;
CREATE INDEX idx_users_status_created ON users(status, created_at DESC);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- KYC Table (India-specific)
-- ============================================================================

CREATE OR REPLACE FUNCTION validate_address_jsonb()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT (
        NEW.address ? 'street' AND
        NEW.address ? 'city' AND
        NEW.address ? 'state' AND
        NEW.address ? 'pin' AND
        NEW.address ? 'country'
    ) THEN
        RAISE EXCEPTION 'Address must contain street, city, state, pin, and country';
    END IF;
    IF NOT (NEW.address->>'pin' ~* '^[1-9][0-9]{5}$') THEN
        RAISE EXCEPTION 'Invalid PIN code format';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS user_kyc (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pan VARCHAR(10) NOT NULL,
    aadhaar VARCHAR(12) NOT NULL,
    date_of_birth DATE NOT NULL,
    address JSONB NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT kyc_status_check CHECK (status IN ('pending', 'verified', 'rejected', 'expired')),
    CONSTRAINT kyc_pan_check CHECK (pan ~* '^[A-Z]{5}[0-9]{4}[A-Z]$'),
    CONSTRAINT kyc_aadhaar_check CHECK (aadhaar ~* '^[2-9][0-9]{11}$'),
    CONSTRAINT kyc_pan_unique UNIQUE (pan),
    CONSTRAINT kyc_aadhaar_unique UNIQUE (aadhaar)
);

CREATE INDEX idx_kyc_pan ON user_kyc(pan);
CREATE INDEX idx_kyc_status ON user_kyc(status);

CREATE TRIGGER update_user_kyc_updated_at
    BEFORE UPDATE ON user_kyc
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER validate_kyc_address
    BEFORE INSERT OR UPDATE ON user_kyc
    FOR EACH ROW
    EXECUTE FUNCTION validate_address_jsonb();

-- ============================================================================
-- Sessions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT sessions_expires_future CHECK (expires_at > created_at)
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_user_expires ON sessions(user_id, expires_at DESC);
-- Note: Partial index with NOW() not supported - use idx_sessions_user_expires for active session queries

CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- User-Admin Pairs Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_admin_pairs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admin_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_user_admin_pairs_user UNIQUE (user_id),
    CONSTRAINT uq_user_admin_pairs_admin UNIQUE (admin_user_id)
);

CREATE INDEX idx_user_admin_pairs_user_id ON user_admin_pairs(user_id);
CREATE INDEX idx_user_admin_pairs_admin_user_id ON user_admin_pairs(admin_user_id);

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

-- ============================================================================
-- Verification Requests Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS verification_requests (
    id VARCHAR(50) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,
    otp_code VARCHAR(10) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,
    attempt_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT chk_verification_status CHECK (status IN ('pending', 'verified', 'expired', 'cancelled'))
);

CREATE INDEX idx_verification_user_status ON verification_requests(user_id, status);
CREATE INDEX idx_verification_expires ON verification_requests(expires_at) WHERE status = 'pending';
CREATE INDEX idx_verification_pending ON verification_requests(user_id, created_at DESC) WHERE status = 'pending';
CREATE INDEX idx_verification_operation ON verification_requests(user_id, operation_type, created_at DESC);
