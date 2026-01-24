-- Ledger Service Initial Schema
-- Consolidated migration for pre-release

-- ============================================================================
-- Helper Functions
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Accounts Table (Chart of Accounts)
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    parent_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    balance BIGINT NOT NULL DEFAULT 0,
    debit_total BIGINT NOT NULL DEFAULT 0,
    credit_total BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT accounts_type_check CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    CONSTRAINT accounts_status_check CHECK (status IN ('active', 'inactive', 'closed')),
    CONSTRAINT accounts_currency_check CHECK (currency ~* '^[A-Z]{3}$')
);

CREATE INDEX idx_accounts_code ON accounts(code);
CREATE INDEX idx_accounts_type ON accounts(type);
CREATE INDEX idx_accounts_parent_id ON accounts(parent_id);
CREATE INDEX idx_accounts_status ON accounts(status);

CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Journal Entries Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_number VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL DEFAULT 'standard',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    description TEXT NOT NULL,
    reference_type VARCHAR(50),
    reference_id VARCHAR(100),
    posted_at TIMESTAMP WITH TIME ZONE,
    posted_by UUID,
    voided_at TIMESTAMP WITH TIME ZONE,
    voided_by UUID,
    void_reason TEXT,
    reversal_entry_id UUID REFERENCES journal_entries(id),
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT journal_entries_type_check CHECK (type IN ('standard', 'opening', 'closing', 'adjusting', 'reversing')),
    CONSTRAINT journal_entries_status_check CHECK (status IN ('draft', 'posted', 'voided', 'reversed')),
    CONSTRAINT journal_entries_posted_check CHECK (
        (status = 'posted' AND posted_at IS NOT NULL AND posted_by IS NOT NULL) OR
        (status != 'posted' AND posted_at IS NULL AND posted_by IS NULL)
    ),
    CONSTRAINT journal_entries_voided_check CHECK (
        (status = 'voided' AND voided_at IS NOT NULL AND voided_by IS NOT NULL AND void_reason IS NOT NULL) OR
        (status != 'voided')
    )
);

CREATE INDEX idx_journal_entries_entry_number ON journal_entries(entry_number);
CREATE INDEX idx_journal_entries_status ON journal_entries(status);
CREATE INDEX idx_journal_entries_type ON journal_entries(type);
CREATE INDEX idx_journal_entries_posted_at ON journal_entries(posted_at);
CREATE INDEX idx_journal_entries_reference ON journal_entries(reference_type, reference_id);
CREATE INDEX idx_journal_entries_created_at ON journal_entries(created_at DESC);

CREATE TRIGGER update_journal_entries_updated_at
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE FUNCTION generate_entry_number()
RETURNS TEXT AS $$
DECLARE
    current_year TEXT;
    next_number INTEGER;
    new_entry_number TEXT;
BEGIN
    current_year := TO_CHAR(NOW(), 'YYYY');
    SELECT COALESCE(MAX(
        CAST(
            SUBSTRING(entry_number FROM 'JE-' || current_year || '-(\d+)') AS INTEGER
        )
    ), 0) + 1
    INTO next_number
    FROM journal_entries
    WHERE entry_number LIKE 'JE-' || current_year || '-%';
    new_entry_number := 'JE-' || current_year || '-' || LPAD(next_number::TEXT, 5, '0');
    RETURN new_entry_number;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Ledger Lines Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS ledger_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    debit_amount BIGINT NOT NULL DEFAULT 0,
    credit_amount BIGINT NOT NULL DEFAULT 0,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT ledger_lines_amount_check CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

CREATE INDEX idx_ledger_lines_entry_id ON ledger_lines(entry_id);
CREATE INDEX idx_ledger_lines_account_id ON ledger_lines(account_id);
CREATE INDEX idx_ledger_lines_created_at ON ledger_lines(created_at DESC);

-- ============================================================================
-- Ledger Functions and Triggers
-- ============================================================================

CREATE OR REPLACE FUNCTION update_account_balances()
RETURNS TRIGGER AS $$
DECLARE
    line RECORD;
    account RECORD;
    new_balance BIGINT;
BEGIN
    IF NEW.status = 'posted' AND (OLD.status IS NULL OR OLD.status != 'posted') THEN
        FOR line IN
            SELECT account_id, debit_amount, credit_amount
            FROM ledger_lines
            WHERE entry_id = NEW.id
        LOOP
            SELECT id, type, balance, debit_total, credit_total
            INTO account
            FROM accounts
            WHERE id = line.account_id
            FOR UPDATE;

            UPDATE accounts
            SET
                debit_total = debit_total + line.debit_amount,
                credit_total = credit_total + line.credit_amount
            WHERE id = account.id;

            IF account.type IN ('asset', 'expense') THEN
                new_balance := account.balance + line.debit_amount - line.credit_amount;
            ELSE
                new_balance := account.balance + line.credit_amount - line.debit_amount;
            END IF;

            UPDATE accounts
            SET balance = new_balance
            WHERE id = account.id;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_account_balances_on_post
    AFTER UPDATE ON journal_entries
    FOR EACH ROW
    WHEN (NEW.status = 'posted')
    EXECUTE FUNCTION update_account_balances();

CREATE OR REPLACE FUNCTION validate_journal_entry_balanced()
RETURNS TRIGGER AS $$
DECLARE
    total_debits BIGINT;
    total_credits BIGINT;
BEGIN
    IF NEW.status = 'posted' AND (OLD.status IS NULL OR OLD.status != 'posted') THEN
        SELECT
            COALESCE(SUM(debit_amount), 0),
            COALESCE(SUM(credit_amount), 0)
        INTO total_debits, total_credits
        FROM ledger_lines
        WHERE entry_id = NEW.id;

        IF total_debits != total_credits THEN
            RAISE EXCEPTION 'Journal entry % is not balanced: debits=%, credits=%',
                NEW.entry_number, total_debits, total_credits;
        END IF;

        IF (SELECT COUNT(*) FROM ledger_lines WHERE entry_id = NEW.id) < 2 THEN
            RAISE EXCEPTION 'Journal entry % must have at least 2 lines', NEW.entry_number;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_entry_balanced_on_post
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    WHEN (NEW.status = 'posted')
    EXECUTE FUNCTION validate_journal_entry_balanced();

-- ============================================================================
-- Views
-- ============================================================================

CREATE OR REPLACE VIEW account_balances AS
SELECT
    a.id,
    a.code,
    a.name,
    a.type,
    a.currency,
    a.balance,
    a.debit_total,
    a.credit_total,
    a.status,
    a.created_at,
    a.updated_at,
    CASE
        WHEN a.type IN ('asset', 'expense') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('asset', 'expense') AND a.balance < 0 THEN 'abnormal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance < 0 THEN 'abnormal'
        ELSE NULL
    END AS balance_status
FROM accounts a;

CREATE OR REPLACE VIEW general_ledger AS
SELECT
    ll.id AS line_id,
    ll.entry_id,
    je.entry_number,
    je.type AS entry_type,
    je.description AS entry_description,
    je.posted_at,
    ll.account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    ll.debit_amount,
    ll.credit_amount,
    ll.description AS line_description,
    ll.created_at
FROM ledger_lines ll
JOIN journal_entries je ON ll.entry_id = je.id
JOIN accounts a ON ll.account_id = a.id
WHERE je.status = 'posted'
ORDER BY je.posted_at DESC, je.entry_number DESC, ll.created_at;

-- ============================================================================
-- Seed Data: Standard Chart of Accounts for Indian Neobank
-- ============================================================================

-- Assets (1000-1999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('1000', 'Cash and Bank Accounts', 'asset', 'INR', 'active'),
('1100', 'Accounts Receivable', 'asset', 'INR', 'active'),
('1200', 'Loans Receivable', 'asset', 'INR', 'active'),
('1300', 'Investments', 'asset', 'INR', 'active'),
('1400', 'Fixed Assets', 'asset', 'INR', 'active'),
('1500', 'Other Assets', 'asset', 'INR', 'active');

-- Liabilities (2000-2999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('2000', 'Accounts Payable', 'liability', 'INR', 'active'),
('2100', 'Customer Deposits', 'liability', 'INR', 'active'),
('2200', 'Borrowings', 'liability', 'INR', 'active'),
('2300', 'Taxes Payable', 'liability', 'INR', 'active'),
('2400', 'Other Liabilities', 'liability', 'INR', 'active');

-- Equity (3000-3999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('3000', 'Share Capital', 'equity', 'INR', 'active'),
('3100', 'Retained Earnings', 'equity', 'INR', 'active'),
('3200', 'Reserves', 'equity', 'INR', 'active');

-- Revenue (4000-4999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('4000', 'Interest Income', 'revenue', 'INR', 'active'),
('4100', 'Fee Income', 'revenue', 'INR', 'active'),
('4200', 'Transaction Fees', 'revenue', 'INR', 'active'),
('4300', 'Other Income', 'revenue', 'INR', 'active');

-- Expenses (5000-5999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('5000', 'Interest Expense', 'expense', 'INR', 'active'),
('5100', 'Operating Expenses', 'expense', 'INR', 'active'),
('5200', 'Salary and Wages', 'expense', 'INR', 'active'),
('5300', 'Technology Expenses', 'expense', 'INR', 'active'),
('5400', 'Marketing Expenses', 'expense', 'INR', 'active'),
('5500', 'Compliance and Legal', 'expense', 'INR', 'active'),
('5600', 'Other Expenses', 'expense', 'INR', 'active');
