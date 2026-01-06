-- Drop trigger
DROP TRIGGER IF EXISTS virtual_cards_updated_at ON virtual_cards;
DROP FUNCTION IF EXISTS update_virtual_cards_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_virtual_cards_number;
DROP INDEX IF EXISTS idx_virtual_cards_status;
DROP INDEX IF EXISTS idx_virtual_cards_user_id;
DROP INDEX IF EXISTS idx_virtual_cards_wallet_id;

-- Drop virtual cards table
DROP TABLE IF EXISTS virtual_cards;

-- Drop enums
DROP TYPE IF EXISTS card_type;
DROP TYPE IF EXISTS card_status;
