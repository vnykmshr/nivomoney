-- ============================================================================
-- Create verification_requests table for OTP-based verification system
-- Description: Stores verification requests for sensitive operations with OTP codes
-- ============================================================================

CREATE TABLE IF NOT EXISTS verification_requests (
    id VARCHAR(50) PRIMARY KEY,
    user_id UUID NOT NULL,

    -- Request details
    operation_type VARCHAR(50) NOT NULL,  -- 'password_change', 'email_change', 'high_value_transfer', etc.
    otp_code VARCHAR(10) NOT NULL,

    -- Status tracking
    status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'verified', 'expired', 'cancelled'

    -- Operation context (what to do after verification)
    metadata JSONB DEFAULT '{}',

    -- Timing
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE,

    -- Attempt tracking (for rate limiting)
    attempt_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,

    -- Foreign keys
    CONSTRAINT fk_verification_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Status validation
    CONSTRAINT chk_verification_status
        CHECK (status IN ('pending', 'verified', 'expired', 'cancelled'))
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_verification_user_status
    ON verification_requests(user_id, status);

CREATE INDEX IF NOT EXISTS idx_verification_expires
    ON verification_requests(expires_at)
    WHERE status = 'pending';

-- Partial index for pending verifications (most common query)
CREATE INDEX IF NOT EXISTS idx_verification_pending
    ON verification_requests(user_id, created_at DESC)
    WHERE status = 'pending';

-- Index for operation type lookups
CREATE INDEX IF NOT EXISTS idx_verification_operation
    ON verification_requests(user_id, operation_type, created_at DESC);

COMMENT ON TABLE verification_requests IS 'OTP-based verification requests for sensitive operations';
COMMENT ON COLUMN verification_requests.otp_code IS 'Cryptographically generated 6-digit OTP code';
COMMENT ON COLUMN verification_requests.metadata IS 'Operation-specific context data (e.g., new email for email_change)';
