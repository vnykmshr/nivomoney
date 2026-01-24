-- Ledger Service Schema Rollback

DROP VIEW IF EXISTS general_ledger CASCADE;
DROP VIEW IF EXISTS account_balances CASCADE;

DROP TABLE IF EXISTS ledger_lines CASCADE;
DROP TABLE IF EXISTS journal_entries CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS generate_entry_number() CASCADE;
DROP FUNCTION IF EXISTS update_account_balances() CASCADE;
DROP FUNCTION IF EXISTS validate_journal_entry_balanced() CASCADE;
