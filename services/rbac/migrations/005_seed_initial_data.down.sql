-- Remove all seeded data (in reverse order of dependencies)

-- Remove role-permission assignments
DELETE FROM role_permissions WHERE role_id IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000005',
    '00000000-0000-0000-0000-000000000006'
);

-- Remove permissions
DELETE FROM permissions WHERE is_system = true;

-- Remove roles
DELETE FROM roles WHERE is_system = true;
