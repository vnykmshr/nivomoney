-- Revert to original wallet type constraint
ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_type_check;
ALTER TABLE wallets ADD CONSTRAINT wallets_type_check CHECK (type IN ('savings', 'current', 'fixed'));
