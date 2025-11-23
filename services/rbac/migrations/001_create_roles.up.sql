-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Roles table with hierarchical support
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT role_name_not_empty CHECK (name <> ''),
    CONSTRAINT no_self_reference CHECK (id != parent_role_id)
);

-- Index for role hierarchy lookups
CREATE INDEX idx_roles_parent_role_id ON roles(parent_role_id) WHERE parent_role_id IS NOT NULL;
CREATE INDEX idx_roles_active ON roles(is_active) WHERE is_active = true;
CREATE INDEX idx_roles_system ON roles(is_system) WHERE is_system = true;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE roles IS 'System roles with hierarchical support (parent_role_id)';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for inheritance - child inherits parent permissions';
COMMENT ON COLUMN roles.is_system IS 'System-defined roles cannot be deleted';
COMMENT ON COLUMN roles.is_active IS 'Inactive roles cannot be assigned to users';
