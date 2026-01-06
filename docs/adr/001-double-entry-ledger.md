---
layout: default
title: "ADR-001: Double-Entry Ledger"
parent: Architecture Decision Records
nav_order: 1
---

# ADR-001: Double-Entry Ledger for Financial Transactions

**Status**: Accepted
**Date**: 2024-01-15
**Decision Makers**: Engineering Team

## Context

Nivo is a neobank platform that handles real money movements: deposits, withdrawals, and peer-to-peer transfers. We need a system to track all financial transactions that:

1. Maintains accurate balances at all times
2. Provides a complete audit trail for compliance
3. Prevents money from being "created" or "lost" due to bugs
4. Supports financial reporting and reconciliation

## Decision

Implement a **double-entry bookkeeping system** where every transaction creates balanced journal entries with debits equaling credits.

### Core Concepts

**Chart of Accounts**: Hierarchical account structure following Indian accounting standards
```
1000-1999: Assets (Cash, Receivables)
2000-2999: Liabilities (Customer Deposits, Payables)
3000-3999: Equity (Capital, Retained Earnings)
4000-4999: Revenue (Fees, Interest)
5000-5999: Expenses (Operations, Technology)
```

**Journal Entries**: Every financial event creates a balanced entry
```
Transfer ₹1,000 from User A to User B:

  Debit:  User A Wallet (Liability)  ₹1,000
  Credit: User B Wallet (Liability)  ₹1,000
  ─────────────────────────────────────────
  Total Debits = Total Credits ✓
```

**Wallet as Liability**: Customer wallet balances are liabilities to the company (we owe them that money).

### Implementation

```go
// JournalEntry must always be balanced
type JournalEntry struct {
    ID          string
    EntryNumber string        // Sequential: JE-2024-00001
    Type        EntryType     // standard, opening, reversing
    Status      EntryStatus   // draft → posted → voided
    Lines       []LedgerLine  // Debit and credit lines
}

// IsBalanced enforces the fundamental equation
func (j *JournalEntry) IsBalanced() bool {
    var totalDebits, totalCredits int64
    for _, line := range j.Lines {
        totalDebits += line.DebitAmount
        totalCredits += line.CreditAmount
    }
    return totalDebits == totalCredits
}
```

### Database Constraints

```sql
-- Trigger ensures entries are balanced before posting
CREATE FUNCTION validate_journal_entry_balance()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT SUM(debit_amount) FROM ledger_lines WHERE entry_id = NEW.id)
       != (SELECT SUM(credit_amount) FROM ledger_lines WHERE entry_id = NEW.id)
    THEN
        RAISE EXCEPTION 'Journal entry is not balanced';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Alternatives Considered

### 1. Simple Balance Tracking
Store only wallet balances, increment/decrement on transactions.

**Rejected because:**
- No audit trail for compliance
- Easy to create/lose money through bugs
- Can't reconcile or investigate discrepancies
- Doesn't meet financial services regulatory requirements

### 2. Event Sourcing
Store all events and compute balances from event stream.

**Rejected because:**
- More complex infrastructure (event store, projections)
- Overkill for MVP scope
- Double-entry provides same auditability with simpler model
- Can add event sourcing later if needed

### 3. Third-Party Ledger Service
Use a service like Modern Treasury or Moov.

**Rejected because:**
- Adds external dependency and cost
- Less control over implementation details
- Portfolio project should demonstrate our own implementation
- Can integrate later for production if needed

## Consequences

### Positive

- **Data Integrity**: Mathematically impossible to have unbalanced transactions
- **Audit Trail**: Complete history of every financial movement
- **Regulatory Compliance**: Standard approach accepted by auditors
- **Debugging**: Easy to trace any balance to its source transactions
- **Reporting**: Can generate standard financial reports (trial balance, P&L)

### Negative

- **Complexity**: More complex than simple balance tracking
- **Learning Curve**: Team needs to understand accounting basics
- **Performance**: Slightly more writes per transaction (multiple lines)

### Mitigations

- **Complexity**: Clear abstractions hide accounting details from other services
- **Learning Curve**: Documentation and code comments explain concepts
- **Performance**: Acceptable for demo scale; indexing handles query performance

## Related Decisions

- Wallet Service delegates balance changes to Ledger Service
- Transaction Service orchestrates multi-step flows with ledger entries
- All amounts stored in paise (smallest currency unit) to avoid floating-point issues

## References

- [Double-Entry Bookkeeping (Wikipedia)](https://en.wikipedia.org/wiki/Double-entry_bookkeeping)
- [Martin Fowler: Accounting Patterns](https://martinfowler.com/eaaDev/AccountingNarrative.html)
- [Stripe: Money Movement with Double-Entry](https://stripe.com/blog/engineering-journal)
