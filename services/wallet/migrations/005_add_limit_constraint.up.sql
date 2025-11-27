-- Add constraint to enforce daily_limit <= monthly_limit at database level
-- This provides defense in depth beyond service layer validation

ALTER TABLE wallet_limits
ADD CONSTRAINT wallet_limits_daily_lte_monthly_check
CHECK (daily_limit <= monthly_limit);

-- Add comment
COMMENT ON CONSTRAINT wallet_limits_daily_lte_monthly_check ON wallet_limits IS
'Ensures daily transfer limit cannot exceed monthly limit. Enforced at database level for defense in depth.';
