-- ============================================================================
-- Seed Initial Roles and Permissions
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. Create Roles with Hierarchy
-- ----------------------------------------------------------------------------

-- Base user role (no parent)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000001', 'user', 'Regular user with basic permissions', NULL, true, true);

-- Support role (inherits from user)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000002', 'support', 'Customer support with read-only access', '00000000-0000-0000-0000-000000000001', true, true);

-- Accountant role (inherits from support)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000003', 'accountant', 'Financial operations and reporting', '00000000-0000-0000-0000-000000000002', true, true);

-- Compliance officer role (inherits from support)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000004', 'compliance_officer', 'KYC/AML verification and compliance', '00000000-0000-0000-0000-000000000002', true, true);

-- Admin role (inherits from compliance_officer)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000005', 'admin', 'System administrator with elevated permissions', '00000000-0000-0000-0000-000000000004', true, true);

-- Super admin role (inherits from admin)
INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000006', 'super_admin', 'Super administrator with all permissions', '00000000-0000-0000-0000-000000000005', true, true);

-- ----------------------------------------------------------------------------
-- 2. Create Permissions
-- ----------------------------------------------------------------------------

-- === Identity Service Permissions ===
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
-- Auth
('10000000-0000-0000-0000-000000000001', 'identity:auth:login', 'identity', 'auth', 'login', 'Login to the system', true),
('10000000-0000-0000-0000-000000000002', 'identity:auth:logout', 'identity', 'auth', 'logout', 'Logout from the system', true),
('10000000-0000-0000-0000-000000000003', 'identity:auth:refresh', 'identity', 'auth', 'refresh', 'Refresh authentication token', true),

-- Profile
('10000000-0000-0000-0000-000000000010', 'identity:profile:read', 'identity', 'profile', 'read', 'Read own profile', true),
('10000000-0000-0000-0000-000000000011', 'identity:profile:update', 'identity', 'profile', 'update', 'Update own profile', true),
('10000000-0000-0000-0000-000000000012', 'identity:profile:delete', 'identity', 'profile', 'delete', 'Delete own profile', true),

-- Users (Admin operations)
('10000000-0000-0000-0000-000000000020', 'identity:users:read', 'identity', 'users', 'read', 'Read all users', true),
('10000000-0000-0000-0000-000000000021', 'identity:users:create', 'identity', 'users', 'create', 'Create new users', true),
('10000000-0000-0000-0000-000000000022', 'identity:users:update', 'identity', 'users', 'update', 'Update any user', true),
('10000000-0000-0000-0000-000000000023', 'identity:users:delete', 'identity', 'users', 'delete', 'Delete any user', true),

-- KYC
('10000000-0000-0000-0000-000000000030', 'identity:kyc:submit', 'identity', 'kyc', 'submit', 'Submit KYC documents', true),
('10000000-0000-0000-0000-000000000031', 'identity:kyc:read', 'identity', 'kyc', 'read', 'Read KYC documents', true),
('10000000-0000-0000-0000-000000000032', 'identity:kyc:verify', 'identity', 'kyc', 'verify', 'Verify KYC documents', true),
('10000000-0000-0000-0000-000000000033', 'identity:kyc:reject', 'identity', 'kyc', 'reject', 'Reject KYC documents', true);

-- === Wallet Service Permissions ===
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('20000000-0000-0000-0000-000000000001', 'wallet:wallet:create', 'wallet', 'wallet', 'create', 'Create a new wallet', true),
('20000000-0000-0000-0000-000000000002', 'wallet:wallet:read', 'wallet', 'wallet', 'read', 'Read wallet details', true),
('20000000-0000-0000-0000-000000000003', 'wallet:wallet:update', 'wallet', 'wallet', 'update', 'Update wallet details', true),
('20000000-0000-0000-0000-000000000004', 'wallet:wallet:delete', 'wallet', 'wallet', 'delete', 'Delete a wallet', true),
('20000000-0000-0000-0000-000000000005', 'wallet:wallet:freeze', 'wallet', 'wallet', 'freeze', 'Freeze a wallet', true),
('20000000-0000-0000-0000-000000000006', 'wallet:wallet:unfreeze', 'wallet', 'wallet', 'unfreeze', 'Unfreeze a wallet', true),
('20000000-0000-0000-0000-000000000007', 'wallet:wallet:list', 'wallet', 'wallet', 'list', 'List all wallets', true),
('20000000-0000-0000-0000-000000000008', 'wallet:beneficiary:manage', 'wallet', 'beneficiary', 'manage', 'Manage saved beneficiaries', true);

-- === Ledger Service Permissions ===
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
-- Accounts
('30000000-0000-0000-0000-000000000001', 'ledger:account:create', 'ledger', 'account', 'create', 'Create ledger account', true),
('30000000-0000-0000-0000-000000000002', 'ledger:account:read', 'ledger', 'account', 'read', 'Read ledger account', true),
('30000000-0000-0000-0000-000000000003', 'ledger:account:update', 'ledger', 'account', 'update', 'Update ledger account', true),
('30000000-0000-0000-0000-000000000004', 'ledger:account:list', 'ledger', 'account', 'list', 'List all accounts', true),

-- Journal Entries
('30000000-0000-0000-0000-000000000010', 'ledger:journal:create', 'ledger', 'journal', 'create', 'Create journal entry', true),
('30000000-0000-0000-0000-000000000011', 'ledger:journal:read', 'ledger', 'journal', 'read', 'Read journal entry', true),
('30000000-0000-0000-0000-000000000012', 'ledger:journal:list', 'ledger', 'journal', 'list', 'List journal entries', true),
('30000000-0000-0000-0000-000000000013', 'ledger:journal:reverse', 'ledger', 'journal', 'reverse', 'Reverse journal entry', true),

-- Balance Operations
('30000000-0000-0000-0000-000000000020', 'ledger:balance:read', 'ledger', 'balance', 'read', 'Read account balance', true);

-- === Transaction Service Permissions ===
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
('40000000-0000-0000-0000-000000000001', 'transaction:transaction:create', 'transaction', 'transaction', 'create', 'Create transaction', true),
('40000000-0000-0000-0000-000000000002', 'transaction:transaction:read', 'transaction', 'transaction', 'read', 'Read transaction', true),
('40000000-0000-0000-0000-000000000003', 'transaction:transaction:list', 'transaction', 'transaction', 'list', 'List transactions', true),
('40000000-0000-0000-0000-000000000004', 'transaction:transaction:reverse', 'transaction', 'transaction', 'reverse', 'Reverse transaction', true),
('40000000-0000-0000-0000-000000000005', 'transaction:transfer:create', 'transaction', 'transfer', 'create', 'Create transfer', true),
('40000000-0000-0000-0000-000000000006', 'transaction:deposit:create', 'transaction', 'deposit', 'create', 'Create deposit', true),
('40000000-0000-0000-0000-000000000007', 'transaction:withdrawal:create', 'transaction', 'withdrawal', 'create', 'Create withdrawal', true);

-- === RBAC Service Permissions ===
INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
-- Roles
('50000000-0000-0000-0000-000000000001', 'rbac:role:create', 'rbac', 'role', 'create', 'Create new role', true),
('50000000-0000-0000-0000-000000000002', 'rbac:role:read', 'rbac', 'role', 'read', 'Read role details', true),
('50000000-0000-0000-0000-000000000003', 'rbac:role:update', 'rbac', 'role', 'update', 'Update role', true),
('50000000-0000-0000-0000-000000000004', 'rbac:role:delete', 'rbac', 'role', 'delete', 'Delete role', true),
('50000000-0000-0000-0000-000000000005', 'rbac:role:list', 'rbac', 'role', 'list', 'List all roles', true),

-- Permissions
('50000000-0000-0000-0000-000000000010', 'rbac:permission:create', 'rbac', 'permission', 'create', 'Create new permission', true),
('50000000-0000-0000-0000-000000000011', 'rbac:permission:read', 'rbac', 'permission', 'read', 'Read permission details', true),
('50000000-0000-0000-0000-000000000012', 'rbac:permission:list', 'rbac', 'permission', 'list', 'List all permissions', true),

-- Role Assignments
('50000000-0000-0000-0000-000000000020', 'rbac:assignment:create', 'rbac', 'assignment', 'create', 'Assign role to user', true),
('50000000-0000-0000-0000-000000000021', 'rbac:assignment:read', 'rbac', 'assignment', 'read', 'Read user role assignments', true),
('50000000-0000-0000-0000-000000000022', 'rbac:assignment:delete', 'rbac', 'assignment', 'delete', 'Remove role from user', true),

-- Permission Checks
('50000000-0000-0000-0000-000000000030', 'rbac:check:permission', 'rbac', 'check', 'permission', 'Check user permissions', true);

-- ----------------------------------------------------------------------------
-- 3. Assign Permissions to Roles
-- ----------------------------------------------------------------------------

-- === USER Role Permissions (Base Level) ===
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Identity: Basic auth and profile
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001'), -- login
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000002'), -- logout
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000003'), -- refresh
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000010'), -- read profile
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000011'), -- update profile
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000030'), -- submit KYC

-- Wallet: Own wallet operations
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001'), -- create wallet
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000002'), -- read wallet
('00000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000008'), -- manage beneficiaries

-- Transaction: Own transactions
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000002'), -- read transaction
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000003'), -- list transactions
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000005'), -- create transfer
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000006'), -- create deposit
('00000000-0000-0000-0000-000000000001', '40000000-0000-0000-0000-000000000007'); -- create withdrawal

-- === SUPPORT Role Permissions (Inherits USER + Read-Only) ===
-- Note: Support inherits all user permissions via hierarchy

-- === ACCOUNTANT Role Permissions (Inherits SUPPORT + Financial) ===
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Ledger: Full read access
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002'), -- read account
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004'), -- list accounts
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000011'), -- read journal
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000012'), -- list journal
('00000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000020'), -- read balance

-- Wallet: View all wallets
('00000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000007'); -- list wallets

-- === COMPLIANCE_OFFICER Role Permissions (Inherits SUPPORT + KYC) ===
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Identity: KYC operations
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000031'), -- read KYC
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000032'), -- verify KYC
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000033'), -- reject KYC
('00000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000020'); -- read users

-- === ADMIN Role Permissions (Inherits COMPLIANCE_OFFICER + Admin Ops) ===
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Identity: User management
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000021'), -- create users
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000022'), -- update users
('00000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000023'), -- delete users

-- Wallet: Admin operations
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000003'), -- update wallet
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000004'), -- delete wallet
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000005'), -- freeze wallet
('00000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000006'), -- unfreeze wallet

-- Transaction: Admin operations
('00000000-0000-0000-0000-000000000005', '40000000-0000-0000-0000-000000000004'), -- reverse transaction

-- Ledger: Write operations
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000001'), -- create account
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000003'), -- update account
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000010'), -- create journal
('00000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000013'), -- reverse journal

-- RBAC: Role and assignment management
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000002'), -- read role
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000005'), -- list roles
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000011'), -- read permission
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000012'), -- list permissions
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000020'), -- create assignment
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000021'), -- read assignment
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000022'), -- delete assignment
('00000000-0000-0000-0000-000000000005', '50000000-0000-0000-0000-000000000030'); -- check permission

-- === SUPER_ADMIN Role Permissions (Inherits ADMIN + Full RBAC) ===
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- RBAC: Full control
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000001'), -- create role
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000003'), -- update role
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000004'), -- delete role
('00000000-0000-0000-0000-000000000006', '50000000-0000-0000-0000-000000000010'), -- create permission

-- Identity: Profile deletion
('00000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000012'); -- delete profile
