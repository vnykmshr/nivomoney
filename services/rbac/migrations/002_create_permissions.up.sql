-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    service VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT permission_name_not_empty CHECK (name <> ''),
    CONSTRAINT permission_service_not_empty CHECK (service <> ''),
    CONSTRAINT permission_resource_not_empty CHECK (resource <> ''),
    CONSTRAINT permission_action_not_empty CHECK (action <> ''),
    CONSTRAINT permission_name_format CHECK (name ~ '^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$')
);

-- Indexes for permission lookups
CREATE INDEX idx_permissions_service ON permissions(service);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_system ON permissions(is_system) WHERE is_system = true;
CREATE INDEX idx_permissions_service_resource ON permissions(service, resource);

-- Comments
COMMENT ON TABLE permissions IS 'Granular permissions in format service:resource:action';
COMMENT ON COLUMN permissions.name IS 'Permission identifier in format service:resource:action (e.g., identity:kyc:verify)';
COMMENT ON COLUMN permissions.service IS 'Service name (identity, wallet, ledger, transaction, rbac)';
COMMENT ON COLUMN permissions.resource IS 'Resource type (kyc, wallet, account, transaction, role)';
COMMENT ON COLUMN permissions.action IS 'Action type (create, read, update, delete, verify, approve, etc.)';
COMMENT ON COLUMN permissions.is_system IS 'System-defined permissions cannot be deleted';
