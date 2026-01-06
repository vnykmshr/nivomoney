

# Ledger Service

The Ledger Service implements double-entry bookkeeping for Nivo's financial operations.

## Features

✅ **Double-Entry Bookkeeping**: Enforces debits = credits for every transaction
✅ **Chart of Accounts**: Hierarchical account structure with 5 account types
✅ **Journal Entries**: Draft → Posted → Voided/Reversed workflow
✅ **Automatic Balance Updates**: Database triggers update balances on posting
✅ **Audit Trail**: Complete history of all financial transactions
✅ **India-Centric**: Default INR currency, standard Indian chart of accounts

## Architecture

### Domain Models

**Account** - Ledger account in the chart of accounts
- Types: Asset, Liability, Equity, Revenue, Expense
- Tracks: Balance, Debit Total, Credit Total
- Hierarchical structure with parent accounts

**JournalEntry** - Complete transaction with multiple lines
- Statuses: Draft, Posted, Voided, Reversed
- Types: Standard, Opening, Closing, Adjusting, Reversing
- Metadata and reference tracking

**LedgerLine** - Individual debit/credit in a journal entry
- Each line affects one account
- Either debit OR credit (never both)

### Database Schema

```sql
accounts
├── id, code, name, type, currency
├── parent_id (hierarchical)
├── balance, debit_total, credit_total
└── status, metadata

journal_entries
├── id, entry_number, type, status
├── description, reference_type, reference_id
├── posted_at, posted_by
├── voided_at, voided_by, void_reason
└── metadata

ledger_lines
├── id, entry_id, account_id
├── debit_amount, credit_amount
└── description, metadata
```

### Business Rules

1. **Balance Equation**: Debits must equal Credits in every journal entry
2. **Account Types**:
   - Assets & Expenses: Debit normal (increase with debits)
   - Liabilities, Equity & Revenue: Credit normal (increase with credits)
3. **Posting Workflow**: Draft → Validated → Posted (immutable)
4. **Reversals**: Create opposite entry to undo posted transactions

### Standard Chart of Accounts (India)

```
1000-1999: Assets
  1000: Cash and Bank Accounts
  1100: Accounts Receivable
  1200: Loans Receivable
  1300: Investments
  1400: Fixed Assets

2000-2999: Liabilities
  2000: Accounts Payable
  2100: Customer Deposits
  2200: Borrowings
  2300: Taxes Payable

3000-3999: Equity
  3000: Share Capital
  3100: Retained Earnings
  3200: Reserves

4000-4999: Revenue
  4000: Interest Income
  4100: Fee Income
  4200: Transaction Fees

5000-5999: Expenses
  5000: Interest Expense
  5100: Operating Expenses
  5200: Salary and Wages
  5300: Technology Expenses
```

## API Endpoints

### Accounts

- `POST /api/v1/accounts` - Create account
- `GET /api/v1/accounts/:id` - Get account
- `GET /api/v1/accounts` - List accounts
- `PUT /api/v1/accounts/:id` - Update account
- `GET /api/v1/accounts/:id/balance` - Get balance

### Journal Entries

- `POST /api/v1/journal-entries` - Create entry (draft)
- `GET /api/v1/journal-entries/:id` - Get entry with lines
- `GET /api/v1/journal-entries` - List entries
- `POST /api/v1/journal-entries/:id/post` - Post entry
- `POST /api/v1/journal-entries/:id/void` - Void entry
- `POST /api/v1/journal-entries/:id/reverse` - Reverse entry

## Example: Recording a Transaction

```json
POST /api/v1/journal-entries
{
  "type": "standard",
  "description": "Customer payment received",
  "reference_type": "transaction",
  "reference_id": "txn_123",
  "lines": [
    {
      "account_id": "acc_cash",
      "debit_amount": 100000,
      "credit_amount": 0,
      "description": "Cash received"
    },
    {
      "account_id": "acc_revenue",
      "debit_amount": 0,
      "credit_amount": 100000,
      "description": "Service revenue"
    }
  ]
}
```

## Database Triggers

**Validation Trigger**: Ensures entry is balanced before posting
**Balance Update Trigger**: Updates account balances when entry is posted
**Entry Number Generation**: Auto-generates sequential numbers (JE-2024-00001)

## Views

**account_balances**: Real-time account balances with normal/abnormal status
**general_ledger**: All posted transactions for reporting

### Internal Endpoints (Service-to-Service)

No authentication required. Used by Wallet Service.

- `POST /internal/v1/accounts` - Create ledger account (for wallet creation)
- `GET /internal/v1/accounts/by-code/{code}` - Get account by code

### Health Check

```http
GET /health
```

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | Server port | 8081 |
| `DATABASE_PASSWORD` | PostgreSQL password | (required) |
| `JWT_SECRET` | JWT validation secret | (required) |

### Running the Service

```bash
cd services/ledger
go run cmd/server/main.go
```

## Architecture

```
services/ledger/
├── cmd/
│   └── server/           # Server entry point
├── internal/
│   ├── handler/          # HTTP handlers
│   │   ├── ledger_handler.go
│   │   └── routes.go
│   ├── service/          # Business logic
│   │   └── ledger_service.go
│   ├── repository/       # Database operations
│   │   ├── account_repository.go
│   │   └── journal_repository.go
│   └── models/           # Domain models
│       ├── account.go
│       └── journal_entry.go
├── migrations/           # SQL migrations
└── README.md
```

## Future Enhancements

- [ ] Trial balance report endpoint
- [ ] Balance sheet generation
- [ ] Profit & Loss statement
- [ ] Multi-currency support
- [ ] Fiscal year closing automation
