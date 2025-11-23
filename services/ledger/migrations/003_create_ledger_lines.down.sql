-- Drop views
DROP VIEW IF EXISTS general_ledger;
DROP VIEW IF EXISTS account_balances;

-- Drop triggers
DROP TRIGGER IF EXISTS validate_entry_balanced_on_post ON journal_entries;
DROP TRIGGER IF EXISTS update_account_balances_on_post ON journal_entries;

-- Drop functions
DROP FUNCTION IF EXISTS validate_journal_entry_balanced();
DROP FUNCTION IF EXISTS update_account_balances();

-- Drop indexes
DROP INDEX IF EXISTS idx_ledger_lines_created_at;
DROP INDEX IF EXISTS idx_ledger_lines_account_id;
DROP INDEX IF EXISTS idx_ledger_lines_entry_id;

-- Drop ledger_lines table
DROP TABLE IF EXISTS ledger_lines;
