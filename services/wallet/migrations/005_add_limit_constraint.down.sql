-- Remove constraint
ALTER TABLE wallet_limits
DROP CONSTRAINT IF EXISTS wallet_limits_daily_lte_monthly_check;
