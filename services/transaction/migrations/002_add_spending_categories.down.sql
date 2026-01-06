-- Drop category index
DROP INDEX IF EXISTS idx_transactions_category;

-- Drop category_patterns table
DROP TABLE IF EXISTS category_patterns;

-- Remove category column from transactions
ALTER TABLE transactions DROP COLUMN IF EXISTS category;

-- Drop spending category type
DROP TYPE IF EXISTS spending_category;
