-- Drop wallet limits table
DROP TRIGGER IF EXISTS reset_wallet_limit_trigger ON wallet_limits;
DROP FUNCTION IF EXISTS reset_wallet_limit_if_expired();
DROP TRIGGER IF EXISTS update_wallet_limits_updated_at ON wallet_limits;
DROP TABLE IF EXISTS wallet_limits;
