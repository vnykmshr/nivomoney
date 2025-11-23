-- Drop entry number generation function
DROP FUNCTION IF EXISTS generate_entry_number();

-- Drop updated_at trigger
DROP TRIGGER IF EXISTS update_journal_entries_updated_at ON journal_entries;

-- Drop indexes
DROP INDEX IF EXISTS idx_journal_entries_created_at;
DROP INDEX IF EXISTS idx_journal_entries_reference;
DROP INDEX IF EXISTS idx_journal_entries_posted_at;
DROP INDEX IF EXISTS idx_journal_entries_type;
DROP INDEX IF EXISTS idx_journal_entries_status;
DROP INDEX IF EXISTS idx_journal_entries_entry_number;

-- Drop journal_entries table
DROP TABLE IF EXISTS journal_entries CASCADE;
