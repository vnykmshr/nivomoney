# Seed Service

The Seed Service populates the database with demo accounts for development and demonstration. It creates complete user accounts with verified KYC, active wallets, and initial balances using proper double-entry bookkeeping.

## Features

- **Complete Account Setup**: Creates users with KYC, wallets, and balances
- **Idempotent**: Safe to run multiple times without duplicating data
- **Clean Mode**: Option to reset database before seeding
- **Double-Entry Ledger**: Initial balances recorded with proper journal entries
- **India-Specific Data**: Demo data with Indian addresses, PAN, Aadhaar

## Usage

### Basic Seed
```bash
# Run seed service
go run services/seed/cmd/server/main.go
```

### Clean and Seed
Resets the database before seeding (useful for fresh start):
```bash
go run services/seed/cmd/server/main.go --clean
```

## Demo Accounts

The seed creates 6 demo accounts ready for testing:

| Name | Email | Password | Initial Balance |
|------|-------|----------|-----------------|
| Admin User | admin@nivo.local | *(generated)* | ₹1,00,000 |
| Raj Kumar | raj.kumar@gmail.com | raj123 | ₹50,000 |
| Priya Sharma | priya.electronics@business.com | priya123 | ₹1,50,000 |
| Arjun Patel | arjun.design@freelance.com | arjun123 | ₹75,000 |
| Neha Singh | neha.singh@student.com | neha123 | ₹25,000 |
| Vikram Malhotra | vikram.m@corporate.com | vikram123 | ₹2,00,000 |

**Note**: Demo user credentials use fixed passwords for convenience. The admin password is **generated at runtime** for security and saved to `.secrets/credentials.txt`.

**Admin user roles**: The admin user (`admin@nivo.local`) is created as a regular `user` account type with `user`, `admin`, and `super_admin` roles. This allows login via admin.nivomoney.com with full platform administration permissions.

## Generated Credentials

After running the seed, admin credentials are written to `.secrets/credentials.txt`:

```
.secrets/
└── credentials.txt    # Admin credentials (git-ignored)
```

The file contains all seeded user credentials including the generated admin password. This file is excluded from git via `.gitignore`.

**Why generated admin password?**
- Prevents accidental commit of admin credentials to public repositories
- Each environment gets a unique admin password
- Credentials displayed in seed output and saved to local file

## What Gets Created

For each user, the seed service creates:

1. **User Record**
   - Email, phone, name
   - Bcrypt-hashed password
   - Status: `pending` → `active`

2. **KYC Record**
   - PAN card number
   - Aadhaar number (masked in responses)
   - Date of birth
   - Address (Indian format)
   - Status: `verified`

3. **Ledger Account**
   - Code: `WALLET-{user_id}`
   - Type: `liability` (customer deposit)
   - Currency: `INR`

4. **Wallet**
   - Type: `default`
   - Currency: `INR`
   - Status: `active`
   - Linked to ledger account

5. **Initial Balance** (if specified)
   - Journal entry with reference `SEED-{date}-{timestamp}`
   - Debit: Cash account (1000)
   - Credit: Customer wallet account

## Seed Data Format

Located at `cmd/server/data/users.yaml`:

```yaml
users:
  - id: john
    full_name: "John Doe"
    email: "john@example.com"
    password: "password123"
    phone: "+919876543210"
    pan: "ABCDE1234F"
    aadhaar: "234567890123"
    date_of_birth: "1990-01-15"
    address:
      street: "123 Main Street"
      city: "Mumbai"
      state: "Maharashtra"
      pin: "400001"
      country: "India"
    initial_balance: 10000000  # ₹100,000.00 in paise
```

## Clean Mode Behavior

When `--clean` flag is used:

**Truncated Tables:**
- beneficiaries
- processed_transfers
- transactions
- wallet_limits
- wallets
- user_kyc
- user_roles
- role_permissions
- sessions
- notifications
- risk_events
- ledger_lines
- journal_entries

**Preserved:**
- Chart of accounts (account codes 1000-5999)
- System users (@vnykmshr.com domain)

**Reset:**
- Account balances reset to zero
- Wallet-specific ledger accounts deleted

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_PASSWORD` | PostgreSQL password | (required) |
| `DATABASE_HOST` | Database host | localhost |
| `DATABASE_PORT` | Database port | 5432 |
| `DATABASE_USER` | Database user | nivo |
| `DATABASE_NAME` | Database name | nivo |
| `JWT_SECRET` | JWT secret (required by config) | (required) |

## Output Example

```
[seed] ========================================
[seed]   Nivo Money - Database Seed Script
[seed] ========================================
[seed] Clean mode: false
[seed] Connected to database successfully
[seed] Loaded 6 users from seed data

[seed] ========== Seeding Complete User Accounts ==========

[seed] [1/6] Processing: Admin User (admin@nivo.local)
[seed]   → User created: 550e8400-e29b-41d4-a716-446655440000
[seed]   → KYC created and verified
[seed]   → User activated
[seed]   → Ledger account created: 660e8400-e29b-41d4-a716-446655440000
[seed]   → Wallet created: 770e8400-e29b-41d4-a716-446655440000
[seed]   → Initial balance added: ₹1,00,000.00
[seed]   ✓ Complete account ready

[seed] [2/6] Processing: Raj Kumar (raj.kumar@gmail.com)
...

[seed] ========================================
[seed] Seed completed successfully!
[seed] Created/verified 6 ready-to-use accounts
[seed] ========================================
```

## Architecture

```
services/seed/
├── cmd/
│   └── server/
│       ├── main.go       # Entry point with seeding logic
│       └── data/
│           └── users.yaml # Demo user data (embedded)
└── README.md
```

## Double-Entry Bookkeeping

Initial balances are recorded with proper accounting:

```
Journal Entry: SEED-20240115-1705312200
Type: opening
Status: posted

Ledger Lines:
  1. Debit  | Account 1000 (Cash)           | ₹1,00,000
  2. Credit | Account WALLET-{user_id}      | ₹1,00,000
```

This ensures:
- Complete audit trail
- Balanced books from day one
- Proper accounting for demo funds

## Troubleshooting

### "User already exists"
This is normal for idempotent runs. The seed skips existing users.

### "Failed to create KYC"
Check that the user wasn't partially created. Run with `--clean` to reset.

### "Journal entry creation failed"
Ensure the chart of accounts is set up (account 1000 must exist).

### Missing initial balance
Check that `initial_balance` is set in users.yaml (in paise).

## Future Enhancements

- [ ] Configurable number of demo users
- [ ] Random transaction history generation
- [ ] Multiple currency support
- [ ] API endpoint for on-demand seeding
- [ ] Seed data validation
