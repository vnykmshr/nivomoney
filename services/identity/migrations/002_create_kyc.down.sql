-- Drop address validation trigger
DROP TRIGGER IF EXISTS validate_kyc_address ON user_kyc;

-- Drop address validation function
DROP FUNCTION IF EXISTS validate_address_jsonb();

-- Drop updated_at trigger
DROP TRIGGER IF EXISTS update_user_kyc_updated_at ON user_kyc;

-- Drop indexes
DROP INDEX IF EXISTS idx_kyc_status;
DROP INDEX IF EXISTS idx_kyc_pan;

-- Drop user_kyc table
DROP TABLE IF EXISTS user_kyc;
