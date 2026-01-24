-- RBAC Service Initial Schema
-- Consolidated migration for pre-release

-- ============================================================================
-- Extensions
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
-- Roles Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    parent_role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    is_system BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT role_name_not_empty CHECK (name <> ''),
    CONSTRAINT no_self_reference CHECK (id != parent_role_id)
);

CREATE INDEX idx_roles_parent_role_id ON roles(parent_role_id) WHERE parent_role_id IS NOT NULL;
CREATE INDEX idx_roles_active ON roles(is_active) WHERE is_active = true;
CREATE INDEX idx_roles_system ON roles(is_system) WHERE is_system = true;

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE roles IS 'System roles with hierarchical support (parent_role_id)';
COMMENT ON COLUMN roles.parent_role_id IS 'Parent role for inheritance - child inherits parent permissions';
COMMENT ON COLUMN roles.is_system IS 'System-defined roles cannot be deleted';
COMMENT ON COLUMN roles.is_active IS 'Inactive roles cannot be assigned to users';

-- ============================================================================
-- Permissions Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    service VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT permission_name_not_empty CHECK (name <> ''),
    CONSTRAINT permission_service_not_empty CHECK (service <> ''),
    CONSTRAINT permission_resource_not_empty CHECK (resource <> ''),
    CONSTRAINT permission_action_not_empty CHECK (action <> ''),
    CONSTRAINT permission_name_format CHECK (name ~ '^[a-z0-9_]+:[a-z0-9_]+:[a-z0-9_]+$')
);

CREATE INDEX idx_permissions_service ON permissions(service);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_action ON permissions(action);
CREATE INDEX idx_permissions_system ON permissions(is_system) WHERE is_system = true;
CREATE INDEX idx_permissions_service_resource ON permissions(service, resource);

COMMENT ON TABLE permissions IS 'Granular permissions in format service:resource:action';
COMMENT ON COLUMN permissions.name IS 'Permission identifier in format service:resource:action (e.g., identity:kyc:verify)';
COMMENT ON COLUMN permissions.service IS 'Service name (identity, wallet, ledger, transaction, rbac)';
COMMENT ON COLUMN permissions.resource IS 'Resource type (kyc, wallet, account, transaction, role)';
COMMENT ON COLUMN permissions.action IS 'Action type (create, read, update, delete, verify, approve, etc.)';
COMMENT ON COLUMN permissions.is_system IS 'System-defined permissions cannot be deleted';

-- ============================================================================
-- Role-Permission Mapping Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted_by UUID,
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);
CREATE INDEX idx_role_permissions_granted_at ON role_permissions(granted_at DESC);

COMMENT ON TABLE role_permissions IS 'Many-to-many mapping between roles and permissions';
COMMENT ON COLUMN role_permissions.granted_by IS 'User ID of admin who granted this permission (for audit trail)';

-- ============================================================================
-- User-Role Assignment Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,

    PRIMARY KEY (user_id, role_id),

    CONSTRAINT check_expiry_in_future CHECK (expires_at IS NULL OR expires_at > assigned_at)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_active ON user_roles(user_id, is_active) WHERE is_active = true;
CREATE INDEX idx_user_roles_expires_at ON user_roles(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_user_roles_assigned_at ON user_roles(assigned_at DESC);

CREATE OR REPLACE FUNCTION is_user_role_expired(expires_at_value TIMESTAMP WITH TIME ZONE)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN expires_at_value IS NOT NULL AND expires_at_value < CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMENT ON TABLE user_roles IS 'User-role assignments with optional expiry (user_id references identity service)';
COMMENT ON COLUMN user_roles.user_id IS 'User ID from Identity service (no FK constraint for decoupling)';
COMMENT ON COLUMN user_roles.assigned_by IS 'Admin user ID who assigned this role (for audit trail)';
COMMENT ON COLUMN user_roles.expires_at IS 'Optional expiry timestamp for temporary role assignments';
COMMENT ON COLUMN user_roles.is_active IS 'Active flag for soft-deletion without losing audit trail';

-- ============================================================================
-- Seed Data: Roles
-- ============================================================================

INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000001', 'user', 'Regular user with basic permissions', NULL, true, true),
('00000000-0000-0000-0000-000000000002', 'support', 'Customer support with read-only access', '00000000-0000-0000-0000-000000000001', true, true),
('00000000-0000-0000-0000-000000000003', 'accountant', 'Financial operations and reporting', '00000000-0000-0000-0000-000000000002', true, true),
('00000000-0000-0000-0000-000000000004', 'compliance_officer', 'KYC/AML verification and compliance', '00000000-0000-0000-0000-000000000002', true, true),
('00000000-0000-0000-0000-000000000005', 'admin', 'System administrator with elevated permissions', '00000000-0000-0000-0000-000000000004', true, true),
('00000000-0000-0000-0000-000000000006', 'super_admin', 'Super administrator with all permissions', '00000000-0000-0000-0000-000000000005', true, true),
('00000000-0000-0000-0000-000000000007', 'user_admin', 'Regulated admin for paired user verification operations', '00000000-0000-0000-0000-000000000001', true, true);

-- ============================================================================
-- Seed Data: Permissions
-- ============================================================================

-- Identity Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('10000000-0000-0000-0000-000000000001', 'identity:auth:login', 'identity', 'auth', 'login', 'Login to the system', true),
('10000000-0000-0000-0000-000000000002', 'identity:auth:logout', 'identity', 'auth', 'logout', 'Logout from the system', true),
('10000000-0000-0000-0000-000000000003', 'identity:auth:refresh', 'identity', 'auth', 'refresh', 'Refresh authentication token', true),
('10000000-0000-0000-0000-000000000010', 'identity:profile:read', 'identity', 'profile', 'read', 'Read own profile', true),
('10000000-0000-0000-0000-000000000011', 'identity:profile:update', 'identity', 'profile', 'update', 'Update own profile', true),
('10000000-0000-0000-0000-000000000012', 'identity:profile:delete', 'identity', 'profile', 'delete', 'Delete own profile', true),
('10000000-0000-0000-0000-000000000020', 'identity:users:read', 'identity', 'users', 'read', 'Read all users', true),
('10000000-0000-0000-0000-000000000021', 'identity:users:create', 'identity', 'users', 'create', 'Create new users', true),
('10000000-0000-0000-0000-000000000022', 'identity:users:update', 'identity', 'users', 'update', 'Update any user', true),
('10000000-0000-0000-0000-000000000023', 'identity:users:delete', 'identity', 'users', 'delete', 'Delete any user', true),
('10000000-0000-0000-0000-000000000030', 'identity:kyc:submit', 'identity', 'kyc', 'submit', 'Submit KYC documents', true),
('10000000-0000-0000-0000-000000000031', 'identity:kyc:read', 'identity', 'kyc', 'read', 'Read KYC documents', true),
('10000000-0000-0000-0000-000000000032', 'identity:kyc:verify', 'identity', 'kyc', 'verify', 'Verify KYC documents', true),
('10000000-0000-0000-0000-000000000033', 'identity:kyc:reject', 'identity', 'kyc', 'reject', 'Reject KYC documents', true),
('10000000-0000-0000-0000-000000000040', 'identity:verification:create', 'identity', 'verification', 'create', 'Create verification request', true),
('10000000-0000-0000-0000-000000000041', 'identity:verification:read', 'identity', 'verification', 'read', 'Read own verification requests', true),
('10000000-0000-0000-0000-000000000042', 'identity:verification:verify', 'identity', 'verification', 'verify', 'Verify OTP code', true),
('10000000-0000-0000-0000-000000000043', 'identity:verification:cancel', 'identity', 'verification', 'cancel', 'Cancel verification request', true),
('10000000-0000-0000-0000-000000000044', 'identity:verification:view_pending', 'identity', 'verification', 'view_pending', 'View pending verifications with OTP (User-Admin)', true),
('10000000-0000-0000-0000-000000000045', 'identity:paired_user:view', 'identity', 'paired_user', 'view', 'View paired user profile (User-Admin)', true),
('10000000-0000-0000-0000-000000000046', 'identity:admin_portal:access', 'identity', 'admin_portal', 'access', 'Access admin portal info', true);

-- Wallet Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('20000000-0000-0000-0000-000000000001', 'wallet:wallet:create', 'wallet', 'wallet', 'create', 'Create a new wallet', true),
('20000000-0000-0000-0000-000000000002', 'wallet:wallet:read', 'wallet', 'wallet', 'read', 'Read wallet details', true),
('20000000-0000-0000-0000-000000000003', 'wallet:wallet:update', 'wallet', 'wallet', 'update', 'Update wallet details', true),
('20000000-0000-0000-0000-000000000004', 'wallet:wallet:delete', 'wallet', 'wallet', 'delete', 'Delete a wallet', true),
('20000000-0000-0000-0000-000000000005', 'wallet:wallet:freeze', 'wallet', 'wallet', 'freeze', 'Freeze a wallet', true),
('20000000-0000-0000-0000-000000000006', 'wallet:wallet:unfreeze', 'wallet', 'wallet', 'unfreeze', 'Unfreeze a wallet', true),
('20000000-0000-0000-0000-000000000007', 'wallet:wallet:list', 'wallet', 'wallet', 'list', 'List all wallets', true),
('20000000-0000-0000-0000-000000000008', 'wallet:beneficiary:manage', 'wallet', 'beneficiary', 'manage', 'Manage saved beneficiaries', true);

-- Ledger Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('30000000-0000-0000-0000-000000000001', 'ledger:account:create', 'ledger', 'account', 'create', 'Create ledger account', true),
('30000000-0000-0000-0000-000000000002', 'ledger:account:read', 'ledger', 'account', 'read', 'Read ledger account', true),
('30000000-0000-0000-0000-000000000003', 'ledger:account:update', 'ledger', 'account', 'update', 'Update ledger account', true),
('30000000-0000-0000-0000-000000000004', 'ledger:account:list', 'ledger', 'account', 'list', 'List all accounts', true),
('30000000-0000-0000-0000-000000000010', 'ledger:journal:create', 'ledger', 'journal', 'create', 'Create journal entry', true),
('30000000-0000-0000-0000-000000000011', 'ledger:journal:read', 'ledger', 'journal', 'read', 'Read journal entry', true),
('30000000-0000-0000-0000-000000000012', 'ledger:journal:list', 'ledger', 'journal', 'list', 'List journal entries', true),
('30000000-0000-0000-0000-000000000013', 'ledger:journal:reverse', 'ledger', 'journal', 'reverse', 'Reverse journal entry', true),
('30000000-0000-0000-0000-000000000020', 'ledger:balance:read', 'ledger', 'balance', 'read', 'Read account balance', true);

-- Transaction Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('40000000-0000-0000-0000-000000000001', 'transaction:transaction:create', 'transaction', 'transaction', 'create', 'Create transaction', true),
('40000000-0000-0000-0000-000000000002', 'transaction:transaction:read', 'transaction', 'transaction', 'read', 'Read transaction', true),
('40000000-0000-0000-0000-000000000003', 'transaction:transaction:list', 'transaction', 'transaction', 'list', 'List transactions', true),
('40000000-0000-0000-0000-000000000004', 'transaction:transaction:reverse', 'transaction', 'transaction', 'reverse', 'Reverse transaction', true),
('40000000-0000-0000-0000-000000000005', 'transaction:transfer:create', 'transaction', 'transfer', 'create', 'Create transfer', true),
('40000000-0000-0000-0000-000000000006', 'transaction:deposit:create', 'transaction', 'deposit', 'create', 'Create deposit', true),
('40000000-0000-0000-0000-000000000007', 'transaction:withdrawal:create', 'transaction', 'withdrawal', 'create', 'Create withdrawal', true);

-- RBAC Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('50000000-0000-0000-0000-000000000001', 'rbac:role:create', 'rbac', 'role', 'create', 'Create new role', true),
('50000000-0000-0000-0000-000000000002', 'rbac:role:read', 'rbac', 'role', 'read', 'Read role details', true),
('50000000-0000-0000-0000-000000000003', 'rbac:role:update', 'rbac', 'role', 'update', 'Update role', true),
('50000000-0000-0000-0000-000000000004', 'rbac:role:delete', 'rbac', 'role', 'delete', 'Delete role', true),
('50000000-0000-0000-0000-000000000005', 'rbac:role:list', 'rbac', 'role', 'list', 'List all roles', true),
('50000000-0000-0000-0000-000000000010', 'rbac:permission:create', 'rbac', 'permission', 'create', 'Create new permission', true),
('50000000-0000-0000-0000-000000000011', 'rbac:permission:read', 'rbac', 'permission', 'read', 'Read permission details', true),
('50000000-0000-0000-0000-000000000012', 'rbac:permission:list', 'rbac', 'permission', 'list', 'List all permissions', true),
('50000000-0000-0000-0000-000000000020', 'rbac:assignment:create', 'rbac', 'assignment', 'create', 'Assign role to user', true),
('50000000-0000-0000-0000-000000000021', 'rbac:assignment:read', 'rbac', 'assignment', 'read', 'Read user role assignments', true),
('50000000-0000-0000-0000-000000000022', 'rbac:assignment:delete', 'rbac', 'assignment', 'delete', 'Remove role from user', true),
('50000000-0000-0000-0000-000000000030', 'rbac:check:permission', 'rbac', 'check', 'permission', 'Check user permissions', true);

-- Notification Service Permissions
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('60000000-0000-0000-0000-000000000010', 'notification:verification:view', 'notification', 'verification', 'view', 'View verification notifications', true);

-- ============================================================================
-- Seed Data: Role-Permission Assignments
-- ============================================================================

-- USER Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000010'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000011'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000030'),
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001'),
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008'),
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005'),
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000006'),
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000007'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000040'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000041'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000042'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000043'),
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000046');

-- ACCOUNTANT Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004'),
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000011'),
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000012'),
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000020'),
('00000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000007');

-- COMPLIANCE_OFFICER Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000031'),
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000032'),
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000033'),
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000020');

-- ADMIN Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021'),
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000022'),
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000023'),
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000004'),
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000005'),
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000006'),
('00000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000004'),
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000001'),
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000010'),
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000013'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000005'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000011'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000012'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000020'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000021'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000022'),
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000030');

-- SUPER_ADMIN Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000001'),
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000003'),
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000004'),
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000010'),
('00000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000012');

-- USER_ADMIN Role Permissions
INSERT INTO role_permissions (role_id, permission_id) VALUES
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000044'),
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000045'),
('00000000-0000-0000-0000-000000000007', '60000000-0000-0000-0000-000000000010'),
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000041'),
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000010'),
('00000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000002'),
('00000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000003');
