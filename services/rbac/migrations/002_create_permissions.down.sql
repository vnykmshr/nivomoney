-- Drop indexes
DROP INDEX IF EXISTS idx_permissions_service;
DROP INDEX IF EXISTS idx_permissions_resource;
DROP INDEX IF EXISTS idx_permissions_action;
DROP INDEX IF EXISTS idx_permissions_system;
DROP INDEX IF EXISTS idx_permissions_service_resource;

-- Drop table
DROP TABLE IF EXISTS permissions CASCADE;
