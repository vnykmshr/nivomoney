-- ============================================================================
-- Add User-Admin Role and Permissions
-- Description: Creates the user_admin role for self-service verification flows
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. Create User-Admin Role (inherits from user)
-- ----------------------------------------------------------------------------

INSERT INTO roles (id, name, description, parent_role_id, is_system, is_active) VALUES
('00000000-0000-0000-0000-000000000007', 'user_admin', 'Regulated admin for paired user verification operations', '00000000-0000-0000-0000-000000000001', true, true)
ON CONFLICT (name) DO NOTHING;

-- ----------------------------------------------------------------------------
-- 2. Create Verification Service Permissions
-- ----------------------------------------------------------------------------

INSERT INTO permissions (id, name, service, resource, action, description, is_system) VALUES
-- Verification request permissions
('10000000-0000-0000-0000-000000000040', 'identity:verification:create', 'identity', 'verification', 'create', 'Create verification request', true),
('10000000-0000-0000-0000-000000000041', 'identity:verification:read', 'identity', 'verification', 'read', 'Read own verification requests', true),
('10000000-0000-0000-0000-000000000042', 'identity:verification:verify', 'identity', 'verification', 'verify', 'Verify OTP code', true),
('10000000-0000-0000-0000-000000000043', 'identity:verification:cancel', 'identity', 'verification', 'cancel', 'Cancel verification request', true),

-- User-Admin specific permissions (scoped to paired user)
('10000000-0000-0000-0000-000000000044', 'identity:verification:view_pending', 'identity', 'verification', 'view_pending', 'View pending verifications with OTP (User-Admin)', true),
('10000000-0000-0000-0000-000000000045', 'identity:paired_user:view', 'identity', 'paired_user', 'view', 'View paired user profile (User-Admin)', true),
('10000000-0000-0000-0000-000000000046', 'identity:admin_portal:access', 'identity', 'admin_portal', 'access', 'Access admin portal info', true),

-- Notification permissions for verification
('60000000-0000-0000-0000-000000000010', 'notification:verification:view', 'notification', 'verification', 'view', 'View verification notifications', true)
ON CONFLICT (name) DO NOTHING;

-- ----------------------------------------------------------------------------
-- 3. Assign Permissions to User Role (regular users)
-- ----------------------------------------------------------------------------

-- Regular users can create and manage their own verification requests
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- Verification permissions for regular users
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000040'), -- create verification
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000041'), -- read own verifications
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000042'), -- verify OTP
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000043'), -- cancel verification
('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000046')  -- access admin portal info
ON CONFLICT DO NOTHING;

-- ----------------------------------------------------------------------------
-- 4. Assign Permissions to User-Admin Role
-- ----------------------------------------------------------------------------

-- User-Admin has special permissions scoped to their paired user
INSERT INTO role_permissions (role_id, permission_id) VALUES
-- User-Admin specific permissions
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000044'), -- view pending with OTP
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000045'), -- view paired user
('00000000-0000-0000-0000-000000000007', '60000000-0000-0000-0000-000000000010'), -- view verification notifications
-- Inherits user permissions through role hierarchy:
-- - profile:read, profile:update
-- - wallet:read (limited to paired user in middleware)
-- - transaction:read, transaction:list (limited to paired user in middleware)
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000041'), -- read verifications
('00000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000010'), -- profile:read
('00000000-0000-0000-0000-000000000007', '20000000-0000-0000-0000-000000000002'), -- wallet:read
('00000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000002'), -- transaction:read
('00000000-0000-0000-0000-000000000007', '40000000-0000-0000-0000-000000000003')  -- transaction:list
ON CONFLICT DO NOTHING;

-- ----------------------------------------------------------------------------
-- 5. Documentation
-- ----------------------------------------------------------------------------

COMMENT ON TABLE roles IS 'Roles including user_admin for self-service verification';
