-- Rollback: Remove the check function and restore the original (buggy) trigger

DROP FUNCTION IF EXISTS check_and_reset_wallet_limits;

-- Restore original trigger logic
CREATE OR REPLACE FUNCTION reset_wallet_limit_if_expired()
RETURNS TRIGGER AS $$
BEGIN
    -- Reset daily spent if past reset time
    IF NOW() >= OLD.daily_reset_at THEN
        NEW.daily_spent := 0;
        NEW.daily_reset_at := DATE_TRUNC('day', NOW() + INTERVAL '1 day');
    END IF;

    -- Reset monthly spent if past reset time
    IF NOW() >= OLD.monthly_reset_at THEN
        NEW.monthly_spent := 0;
        NEW.monthly_reset_at := DATE_TRUNC('month', NOW() + INTERVAL '1 month');
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reset_wallet_limit_trigger
    BEFORE UPDATE ON wallet_limits
    FOR EACH ROW
    EXECUTE FUNCTION reset_wallet_limit_if_expired();
