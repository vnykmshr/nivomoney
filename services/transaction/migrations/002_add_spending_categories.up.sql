-- Create spending category enum type
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'spending_category') THEN
        CREATE TYPE spending_category AS ENUM (
            'food', 'transport', 'utilities', 'entertainment',
            'shopping', 'health', 'education', 'transfer', 'other'
        );
    END IF;
END$$;

-- Add category column to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS category spending_category DEFAULT 'other';

-- Create category detection patterns table
CREATE TABLE IF NOT EXISTS category_patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pattern VARCHAR(100) NOT NULL,
    category spending_category NOT NULL,
    priority INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create index for pattern matching
CREATE INDEX IF NOT EXISTS idx_category_patterns_priority ON category_patterns(priority DESC);

-- Seed common patterns
INSERT INTO category_patterns (pattern, category, priority) VALUES
    ('swiggy', 'food', 10),
    ('zomato', 'food', 10),
    ('domino', 'food', 8),
    ('mcdonald', 'food', 8),
    ('pizza', 'food', 6),
    ('restaurant', 'food', 5),
    ('cafe', 'food', 5),
    ('uber', 'transport', 10),
    ('ola', 'transport', 10),
    ('rapido', 'transport', 8),
    ('petrol', 'transport', 7),
    ('fuel', 'transport', 7),
    ('metro', 'transport', 6),
    ('bus', 'transport', 5),
    ('flipkart', 'shopping', 10),
    ('amazon', 'shopping', 10),
    ('myntra', 'shopping', 8),
    ('ajio', 'shopping', 8),
    ('mall', 'shopping', 5),
    ('store', 'shopping', 4),
    ('electricity', 'utilities', 10),
    ('water bill', 'utilities', 10),
    ('gas bill', 'utilities', 10),
    ('internet', 'utilities', 8),
    ('broadband', 'utilities', 8),
    ('mobile recharge', 'utilities', 7),
    ('netflix', 'entertainment', 10),
    ('spotify', 'entertainment', 10),
    ('hotstar', 'entertainment', 8),
    ('prime video', 'entertainment', 8),
    ('movie', 'entertainment', 6),
    ('cinema', 'entertainment', 6),
    ('hospital', 'health', 10),
    ('pharmacy', 'health', 10),
    ('medical', 'health', 8),
    ('doctor', 'health', 8),
    ('clinic', 'health', 7),
    ('apollo', 'health', 7),
    ('school', 'education', 10),
    ('college', 'education', 10),
    ('university', 'education', 8),
    ('course', 'education', 6),
    ('tuition', 'education', 7),
    ('coaching', 'education', 6)
ON CONFLICT DO NOTHING;

-- Create index on category for faster aggregations
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
