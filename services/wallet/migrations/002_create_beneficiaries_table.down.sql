-- Drop beneficiaries table
DROP TRIGGER IF EXISTS update_beneficiaries_updated_at ON beneficiaries;
DROP INDEX IF EXISTS idx_beneficiaries_unique_nickname;
DROP INDEX IF EXISTS idx_beneficiaries_unique_user;
DROP INDEX IF EXISTS idx_beneficiaries_beneficiary_user;
DROP INDEX IF EXISTS idx_beneficiaries_owner;
DROP TABLE IF EXISTS beneficiaries;
