# Wallet Service

The Wallet Service manages user wallets, balances, transfer limits, and beneficiaries for Nivo, an India-centric neobank platform.

## Features

- **Wallet Management**: Create, activate, freeze, and close user wallets
- **Balance Tracking**: Real-time balance and available balance management
- **Transfer Limits**: Configurable daily and monthly transfer limits
- **Beneficiary Management**: Save and manage frequent transfer recipients
- **Ledger Integration**: Links to double-entry ledger accounts for audit trails
- **Status Workflow**: Full lifecycle management (inactive → active → frozen → closed)

## API Endpoints

### Protected Endpoints (Requires Authentication)

All protected endpoints require an `Authorization` header with a Bearer token:
```http
Authorization: Bearer <jwt_token>
```

#### Create Wallet
```http
POST /api/v1/wallets
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "default",
  "currency": "INR"
}
```

**Response:**
```json
{
  "success": true,
  "message": "wallet created successfully",
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "default",
    "currency": "INR",
    "balance": 0,
    "available_balance": 0,
    "status": "inactive",
    "ledger_account_id": "770e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Get Wallet
```http
GET /api/v1/wallets/{id}
```

#### Get Wallet Balance
```http
GET /api/v1/wallets/{id}/balance
```

**Response:**
```json
{
  "success": true,
  "data": {
    "wallet_id": "660e8400-e29b-41d4-a716-446655440000",
    "balance": 100000,
    "available_balance": 95000,
    "held_amount": 5000
  }
}
```

#### Get Wallet Limits
```http
GET /api/v1/wallets/{id}/limits
```

**Response:**
```json
{
  "success": true,
  "data": {
    "wallet_id": "660e8400-e29b-41d4-a716-446655440000",
    "daily_limit": 10000000,
    "daily_spent": 500000,
    "daily_remaining": 9500000,
    "monthly_limit": 100000000,
    "monthly_spent": 2500000,
    "monthly_remaining": 97500000
  }
}
```

#### Update Wallet Limits
```http
PUT /api/v1/wallets/{id}/limits
Content-Type: application/json

{
  "daily_limit": 20000000,
  "monthly_limit": 200000000
}
```

#### List My Wallets
```http
GET /api/v1/wallets
```

#### List User Wallets (Admin)
```http
GET /api/v1/users/{userId}/wallets
```

### Wallet Status Management

#### Activate Wallet
```http
POST /api/v1/wallets/{id}/activate
```

#### Freeze Wallet
```http
POST /api/v1/wallets/{id}/freeze
Content-Type: application/json

{
  "reason": "Suspicious activity detected - manual review required"
}
```

#### Unfreeze Wallet
```http
POST /api/v1/wallets/{id}/unfreeze
```

#### Close Wallet
```http
POST /api/v1/wallets/{id}/close
Content-Type: application/json

{
  "reason": "Account closure requested by user via customer support"
}
```

### Beneficiary Endpoints

#### Add Beneficiary
```http
POST /api/v1/beneficiaries
Content-Type: application/json

{
  "phone": "+919876543210",
  "nickname": "Mom"
}
```

**Response:**
```json
{
  "success": true,
  "message": "beneficiary added successfully",
  "data": {
    "id": "880e8400-e29b-41d4-a716-446655440000",
    "nickname": "Mom",
    "phone": "+919876543210",
    "wallet_id": "990e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### List Beneficiaries
```http
GET /api/v1/beneficiaries
```

#### Get Beneficiary
```http
GET /api/v1/beneficiaries/{id}
```

#### Update Beneficiary
```http
PUT /api/v1/beneficiaries/{id}
Content-Type: application/json

{
  "nickname": "Mother"
}
```

#### Delete Beneficiary
```http
DELETE /api/v1/beneficiaries/{id}
```

### Internal Endpoints (Service-to-Service)

These endpoints are called by the Transaction Service to execute transfers:

#### Process Transfer
```http
POST /internal/v1/wallets/transfer
Content-Type: application/json

{
  "source_wallet_id": "660e8400-e29b-41d4-a716-446655440000",
  "destination_wallet_id": "770e8400-e29b-41d4-a716-446655440000",
  "amount": 100000,
  "transaction_id": "880e8400-e29b-41d4-a716-446655440000"
}
```

#### Process Deposit
```http
POST /internal/v1/wallets/deposit
Content-Type: application/json

{
  "wallet_id": "660e8400-e29b-41d4-a716-446655440000",
  "amount": 100000,
  "transaction_id": "880e8400-e29b-41d4-a716-446655440000",
  "description": "UPI deposit"
}
```

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "wallet"
}
```

## Wallet Status Workflow

```
[Created] → inactive → active ↔ frozen → closed
                         ↓
                       closed
```

1. **Inactive**: New wallet, KYC pending or not activated
2. **Active**: Fully operational, can perform transactions
3. **Frozen**: Temporarily suspended (compliance, fraud investigation)
4. **Closed**: Permanently closed, no operations allowed

## Balance Model

The wallet tracks two balance types:

| Field | Description |
|-------|-------------|
| `balance` | Total balance in the wallet |
| `available_balance` | Balance available for transactions (balance minus holds) |
| `held_amount` | Difference between balance and available_balance |

All amounts are stored in **paise** (smallest currency unit for INR).

Example: ₹1,000.00 = 100000 paise

## Transfer Limits

| Limit Type | Default | Description |
|------------|---------|-------------|
| Daily Limit | ₹1,00,000 | Maximum transfer amount per day |
| Monthly Limit | ₹10,00,000 | Maximum transfer amount per month |

Limits reset at midnight IST (daily) and first of month (monthly).

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Running Ledger Service (for account creation)

### Environment Variables

Required:
- `SERVICE_PORT`: Server port (default: 8083)
- `DATABASE_PASSWORD`: PostgreSQL password
- `JWT_SECRET`: Secret for JWT validation

Optional:
- `DATABASE_HOST`: Database host (default: localhost)
- `DATABASE_PORT`: Database port (default: 5432)
- `DATABASE_USER`: Database user (default: nivo)
- `DATABASE_NAME`: Database name (default: nivo)
- `LEDGER_SERVICE_URL`: Ledger service URL (default: http://localhost:8081)
- `IDENTITY_SERVICE_URL`: Identity service URL (default: http://localhost:8080)

### Running the Service

```bash
# From repository root
cd services/wallet

# Run directly
go run cmd/server/main.go

# Or use make
make run
```

## Architecture

```
services/wallet/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   │   ├── wallet_handler.go
│   │   └── beneficiary_handler.go
│   ├── service/         # Business logic
│   │   ├── wallet_service.go
│   │   ├── beneficiary_service.go
│   │   ├── ledger_client.go
│   │   └── identity_client.go
│   ├── repository/      # Database operations
│   │   ├── wallet_repository.go
│   │   └── beneficiary_repository.go
│   ├── models/          # Domain models
│   │   ├── wallet.go
│   │   └── beneficiary.go
│   └── router/          # Route configuration
├── Makefile
└── README.md
```

## Dependencies

- `github.com/vnykmshr/nivo/shared`: Shared utilities and middleware
- `github.com/vnykmshr/gopantic`: Request validation
- `github.com/lib/pq`: PostgreSQL driver

## Security Features

- **JWT Authentication**: All endpoints require valid JWT
- **RBAC Permissions**: Fine-grained permission checks
- **Rate Limiting**: Beneficiary operations rate-limited to prevent abuse
- **Ownership Verification**: Users can only access their own wallets

## Future Enhancements

- [ ] Multi-currency wallet support
- [ ] Wallet-to-wallet instant transfer optimization
- [ ] Scheduled transfers
- [ ] Wallet statements and export
- [ ] Sub-wallets for budgeting
