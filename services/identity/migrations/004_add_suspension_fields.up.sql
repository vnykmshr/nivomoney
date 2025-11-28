-- Add suspension tracking fields to users table
ALTER TABLE users
ADD COLUMN suspended_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN suspension_reason TEXT,
ADD COLUMN suspended_by UUID REFERENCES users(id);

-- Add index for suspended users lookup
CREATE INDEX idx_users_suspended_at ON users(suspended_at) WHERE suspended_at IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.suspended_at IS 'Timestamp when user was suspended (NULL if never suspended or currently active)';
COMMENT ON COLUMN users.suspension_reason IS 'Admin-provided reason for suspension';
COMMENT ON COLUMN users.suspended_by IS 'Admin user ID who performed the suspension';
