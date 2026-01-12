-- Notification Service Schema Rollback

DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS notification_templates CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
