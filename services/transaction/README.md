# Transaction Service

The Transaction Service handles all money movement operations for Nivo, including transfers, deposits, withdrawals, and reversals. It integrates with Wallet, Ledger, and Risk services to ensure secure and compliant transactions.

## Features

- **Transfers**: Wallet-to-wallet transfers with limit checking
- **Deposits**: Direct deposits and UPI deposit simulation
- **Withdrawals**: Withdrawal requests with balance verification
- **Reversals**: Transaction reversal for refunds and corrections
- **Risk Integration**: All transactions evaluated by Risk Service
- **Rate Limiting**: Strict rate limits on money movement operations
- **Transaction History**: Full audit trail with filtering and search

## API Endpoints

### Protected Endpoints (Requires Authentication)

All protected endpoints require an `Authorization` header with a Bearer token:
```http
Authorization: Bearer <jwt_token>
```

### Transfer Operations

#### Create Transfer
```http
POST /api/v1/transactions/transfer
Content-Type: application/json

{
  "source_wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "destination_wallet_id": "660e8400-e29b-41d4-a716-446655440000",
  "amount": 100000,
  "currency": "INR",
  "description": "Payment for services",
  "reference": "INV-2024-001"
}
```

**Response:**
```json
{
  "success": true,
  "message": "transfer completed successfully",
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "type": "transfer",
    "status": "completed",
    "source_wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "destination_wallet_id": "660e8400-e29b-41d4-a716-446655440000",
    "amount": 100000,
    "currency": "INR",
    "description": "Payment for services",
    "reference": "INV-2024-001",
    "created_at": "2024-01-15T10:30:00Z",
    "completed_at": "2024-01-15T10:30:01Z"
  }
}
```

### Deposit Operations

#### Create Direct Deposit
```http
POST /api/v1/transactions/deposit
Content-Type: application/json

{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 500000,
  "currency": "INR",
  "description": "Bank transfer deposit",
  "reference": "NEFT-123456"
}
```

#### Initiate UPI Deposit
```http
POST /api/v1/transactions/deposit/upi
Content-Type: application/json

{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 100000,
  "currency": "INR",
  "description": "UPI deposit"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": "880e8400-e29b-41d4-a716-446655440000",
      "type": "deposit",
      "status": "pending",
      "amount": 100000
    },
    "virtual_upi_id": "nivo.550e8400@ybl",
    "qr_code": "data:image/png;base64,...",
    "expires_at": "2024-01-15T10:45:00Z",
    "instructions": [
      "Open your UPI app",
      "Scan the QR code or use the VPA",
      "Enter amount: ₹1,000.00",
      "Complete the payment"
    ]
  }
}
```

#### Complete UPI Deposit (Webhook Simulation)
```http
POST /api/v1/transactions/deposit/upi/complete
Content-Type: application/json

{
  "transaction_id": "880e8400-e29b-41d4-a716-446655440000",
  "upi_transaction_id": "UPI-REF-123456789",
  "status": "success"
}
```

### Withdrawal Operations

#### Create Withdrawal
```http
POST /api/v1/transactions/withdrawal
Content-Type: application/json

{
  "wallet_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 200000,
  "currency": "INR",
  "description": "Withdrawal to bank account",
  "reference": "BANK-ACC-XXXX1234"
}
```

### Transaction Retrieval

#### Get Transaction
```http
GET /api/v1/transactions/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "type": "transfer",
    "status": "completed",
    "source_wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "destination_wallet_id": "660e8400-e29b-41d4-a716-446655440000",
    "amount": 100000,
    "currency": "INR",
    "description": "Payment for services",
    "ledger_entry_id": "990e8400-e29b-41d4-a716-446655440000",
    "created_at": "2024-01-15T10:30:00Z",
    "processed_at": "2024-01-15T10:30:00Z",
    "completed_at": "2024-01-15T10:30:01Z"
  }
}
```

#### List Wallet Transactions
```http
GET /api/v1/wallets/{walletId}/transactions?limit=20&offset=0&status=completed
```

Query Parameters:
- `limit`: Number of results (default: 20, max: 100)
- `offset`: Pagination offset
- `status`: Filter by status (pending, completed, failed, reversed)
- `type`: Filter by type (transfer, deposit, withdrawal)
- `start_date`: Filter from date (ISO 8601)
- `end_date`: Filter to date (ISO 8601)

### Admin Operations

#### Search All Transactions
```http
GET /api/v1/admin/transactions/search?user_id={userId}&min_amount=100000
```

Query Parameters:
- `transaction_id`: Exact transaction ID match
- `user_id`: Filter by user
- `status`: Transaction status
- `type`: Transaction type
- `min_amount`: Minimum amount (paise)
- `max_amount`: Maximum amount (paise)
- `search`: Search in description/reference

#### Reverse Transaction
```http
POST /api/v1/transactions/{id}/reverse
Content-Type: application/json

{
  "reason": "Customer dispute - duplicate payment identified by support team"
}
```

### Health Check
```http
GET /health
```

## Transaction Types

| Type | Description |
|------|-------------|
| `transfer` | Wallet-to-wallet transfer |
| `deposit` | Money added to wallet |
| `withdrawal` | Money removed from wallet |
| `reversal` | Reversal of a previous transaction |
| `fee` | Fee charge |
| `refund` | Refund to customer |

## Transaction Status Workflow

```
pending → processing → completed
              ↓
            failed

completed → reversed (via reversal)
```

| Status | Description |
|--------|-------------|
| `pending` | Transaction initiated, awaiting processing |
| `processing` | Transaction being processed |
| `completed` | Transaction successful |
| `failed` | Transaction failed (with failure_reason) |
| `reversed` | Transaction reversed |
| `cancelled` | Transaction cancelled before processing |

## Rate Limiting

Money movement endpoints have strict rate limiting to prevent abuse:

| Endpoint | Limit |
|----------|-------|
| Transfer | 10 requests/minute per user |
| Deposit | 10 requests/minute per user |
| Withdrawal | 5 requests/minute per user |
| Reverse | 3 requests/minute per admin |

## Integration Points

### Wallet Service
- Verifies wallet ownership
- Checks available balance
- Processes balance updates

### Ledger Service
- Creates double-entry journal entries
- Maintains audit trail

### Risk Service
- Evaluates transaction risk before processing
- May block or flag suspicious transactions

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Running Wallet, Ledger, and Risk services

### Environment Variables

Required:
- `SERVICE_PORT`: Server port (default: 8084)
- `DATABASE_PASSWORD`: PostgreSQL password
- `JWT_SECRET`: Secret for JWT validation

Optional:
- `DATABASE_HOST`: Database host (default: localhost)
- `WALLET_SERVICE_URL`: Wallet service URL (default: http://localhost:8083)
- `LEDGER_SERVICE_URL`: Ledger service URL (default: http://localhost:8081)
- `RISK_SERVICE_URL`: Risk service URL (default: http://localhost:8085)

### Running the Service

```bash
cd services/transaction
go run cmd/server/main.go
```

## Architecture

```
services/transaction/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   │   └── transaction_handler.go
│   ├── service/         # Business logic
│   │   ├── transaction_service.go
│   │   ├── wallet_client.go
│   │   ├── ledger_client.go
│   │   └── risk_client.go
│   ├── repository/      # Database operations
│   │   └── transaction_repository.go
│   ├── models/          # Domain models
│   │   └── transaction.go
│   └── router/          # Route configuration
├── Makefile
└── README.md
```

## Security Features

- **JWT Authentication**: All endpoints require valid JWT
- **RBAC Permissions**: Granular permission checks per operation
- **Rate Limiting**: Strict limits on money movement
- **Idempotency**: Reference field prevents duplicate transactions
- **Risk Evaluation**: All transactions checked before processing

## Future Enhancements

- [ ] Scheduled/recurring transfers
- [ ] Batch transfers for payroll
- [ ] Real UPI integration
- [ ] IMPS/NEFT/RTGS support
- [ ] International transfers
- [ ] Transaction dispute management
