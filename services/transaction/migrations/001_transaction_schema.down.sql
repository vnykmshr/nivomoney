-- Transaction Service Schema Rollback

DROP TABLE IF EXISTS category_patterns CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;

DROP TYPE IF EXISTS spending_category CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
