-- Fix: Replace trigger that resets limits to only fire on SELECT FOR UPDATE (read operations)
-- The previous trigger had a critical bug: it fired on every UPDATE, causing spent amounts
-- to reset immediately after being incremented.

-- Drop the broken trigger
DROP TRIGGER IF EXISTS reset_wallet_limit_trigger ON wallet_limits;
DROP FUNCTION IF EXISTS reset_wallet_limit_if_expired();

-- New approach: Reset limits on read (SELECT FOR UPDATE) via a check function
-- This function checks if limits need reset and returns the corrected values
CREATE OR REPLACE FUNCTION check_and_reset_wallet_limits(
    p_wallet_id UUID,
    p_daily_limit BIGINT,
    p_daily_spent BIGINT,
    p_daily_reset_at TIMESTAMP WITH TIME ZONE,
    p_monthly_limit BIGINT,
    p_monthly_spent BIGINT,
    p_monthly_reset_at TIMESTAMP WITH TIME ZONE,
    OUT new_daily_spent BIGINT,
    OUT new_daily_reset_at TIMESTAMP WITH TIME ZONE,
    OUT new_monthly_spent BIGINT,
    OUT new_monthly_reset_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    -- Check if daily limit needs reset
    IF NOW() >= p_daily_reset_at THEN
        new_daily_spent := 0;
        new_daily_reset_at := DATE_TRUNC('day', NOW() + INTERVAL '1 day');
    ELSE
        new_daily_spent := p_daily_spent;
        new_daily_reset_at := p_daily_reset_at;
    END IF;

    -- Check if monthly limit needs reset
    IF NOW() >= p_monthly_reset_at THEN
        new_monthly_spent := 0;
        new_monthly_reset_at := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    ELSE
        new_monthly_spent := p_monthly_spent;
        new_monthly_reset_at := p_monthly_reset_at;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Add comment explaining the design
COMMENT ON FUNCTION check_and_reset_wallet_limits IS
'Checks if wallet limits have expired and returns reset values. Call this function when reading limits before checking/reserving transfer amounts. This ensures limits are reset on read operations only, not on every write.';
