-- Identity Service Schema Rollback

DROP TABLE IF EXISTS verification_requests CASCADE;
DROP TABLE IF EXISTS user_admin_pairs CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS user_kyc CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS validate_address_jsonb() CASCADE;
DROP FUNCTION IF EXISTS cleanup_expired_sessions() CASCADE;
DROP FUNCTION IF EXISTS update_user_admin_pairs_updated_at() CASCADE;
