-- Remove trigger
DROP TRIGGER IF EXISTS trigger_auto_upi_vpa ON wallets;

-- Remove functions
DROP FUNCTION IF EXISTS auto_generate_upi_vpa();
DROP FUNCTION IF EXISTS generate_upi_vpa(UUID);

-- Drop UPI deposits table
DROP TABLE IF EXISTS upi_deposits;

-- Remove UPI VPA from wallets
DROP INDEX IF EXISTS idx_wallets_upi_vpa;
ALTER TABLE wallets DROP COLUMN IF EXISTS upi_vpa;
