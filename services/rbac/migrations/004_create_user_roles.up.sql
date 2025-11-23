-- User-Role assignment table
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,

    PRIMARY KEY (user_id, role_id),

    -- Constraints
    CONSTRAINT check_expiry_in_future CHECK (expires_at IS NULL OR expires_at > assigned_at)
);

-- Indexes for efficient lookups
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_active ON user_roles(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_user_roles_expires_at ON user_roles(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_user_roles_assigned_at ON user_roles(assigned_at DESC);

-- Function to check if a role assignment is expired
CREATE OR REPLACE FUNCTION is_user_role_expired(expires_at_value TIMESTAMP WITH TIME ZONE)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN expires_at_value IS NOT NULL AND expires_at_value < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Comments
COMMENT ON TABLE user_roles IS 'User-role assignments with optional expiry (user_id references identity service)';
COMMENT ON COLUMN user_roles.user_id IS 'User ID from Identity service (no FK constraint for decoupling)';
COMMENT ON COLUMN user_roles.assigned_by IS 'Admin user ID who assigned this role (for audit trail)';
COMMENT ON COLUMN user_roles.expires_at IS 'Optional expiry timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.is_active IS 'Active flag for soft-deletion without losing audit trail';
