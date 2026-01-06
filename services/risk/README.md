# Risk Service

The Risk Service evaluates transactions for fraud and compliance risk. It implements configurable risk rules and maintains an audit trail of all risk evaluations.

## Features

- **Transaction Evaluation**: Real-time risk scoring for all transactions
- **Configurable Rules**: Create and manage risk rules with different thresholds
- **Rule Types**: Velocity checks, daily limits, amount thresholds
- **Risk Actions**: Allow, block, or flag transactions for review
- **Audit Trail**: Complete history of all risk evaluations
- **Risk Events**: Detailed logging for compliance and investigation

## API Endpoints

### Transaction Evaluation

#### Evaluate Transaction
Called by Transaction Service before processing any money movement.

```http
POST /api/v1/risk/evaluate
Content-Type: application/json

{
  "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "660e8400-e29b-41d4-a716-446655440000",
  "amount": 1000000,
  "currency": "INR",
  "transaction_type": "transfer",
  "from_wallet_id": "770e8400-e29b-41d4-a716-446655440000",
  "to_wallet_id": "880e8400-e29b-41d4-a716-446655440000"
}
```

**Response (Allowed):**
```json
{
  "success": true,
  "data": {
    "allowed": true,
    "action": "allow",
    "risk_score": 15,
    "reason": "Transaction within normal parameters",
    "triggered_rules": [],
    "event_id": "990e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response (Blocked):**
```json
{
  "success": true,
  "data": {
    "allowed": false,
    "action": "block",
    "risk_score": 85,
    "reason": "Daily transfer limit exceeded",
    "triggered_rules": ["rule-daily-limit-001"],
    "event_id": "990e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Risk Rules Management

#### List All Rules
```http
GET /api/v1/risk/rules
GET /api/v1/risk/rules?enabled=true
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "rule-001",
      "rule_type": "velocity",
      "name": "High Frequency Check",
      "parameters": {
        "max_transactions": 10,
        "time_window_mins": 5,
        "per_user": true
      },
      "action": "block",
      "enabled": true,
      "created_at": "2024-01-15T10:00:00Z"
    },
    {
      "id": "rule-002",
      "rule_type": "daily_limit",
      "name": "Daily Transfer Limit",
      "parameters": {
        "max_amount": 10000000,
        "currency": "INR",
        "per_user": true
      },
      "action": "block",
      "enabled": true,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### Get Rule by ID
```http
GET /api/v1/risk/rules/{id}
```

#### Create Rule
```http
POST /api/v1/risk/rules
Content-Type: application/json

{
  "rule_type": "threshold",
  "name": "Large Transaction Alert",
  "parameters": {
    "min_amount": 5000000,
    "max_amount": 0,
    "currency": "INR"
  },
  "action": "flag",
  "enabled": true
}
```

#### Update Rule
```http
PUT /api/v1/risk/rules/{id}
Content-Type: application/json

{
  "name": "Updated Rule Name",
  "parameters": {
    "max_transactions": 15
  },
  "enabled": false
}
```

#### Delete Rule
```http
DELETE /api/v1/risk/rules/{id}
```

### Risk Events

#### Get Event by ID
```http
GET /api/v1/risk/events/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "990e8400-e29b-41d4-a716-446655440000",
    "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
    "user_id": "660e8400-e29b-41d4-a716-446655440000",
    "rule_id": "rule-001",
    "rule_type": "velocity",
    "risk_score": 85,
    "action": "block",
    "reason": "Exceeded 10 transactions in 5 minutes",
    "metadata": {
      "transaction_count": 12,
      "time_window": "5m"
    },
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Get Events by Transaction ID
```http
GET /api/v1/risk/transactions/{transactionId}/events
```

#### Get Events by User ID
```http
GET /api/v1/risk/users/{userId}/events
```

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "service": "risk",
  "status": "healthy",
  "version": "1.0.0"
}
```

## Rule Types

### Velocity Rule
Limits the number of transactions within a time window.

```json
{
  "rule_type": "velocity",
  "parameters": {
    "max_transactions": 10,
    "time_window_mins": 5,
    "per_user": true
  }
}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `max_transactions` | int | Maximum allowed transactions |
| `time_window_mins` | int | Time window in minutes |
| `per_user` | bool | Apply per user (true) or globally (false) |

### Daily Limit Rule
Limits total transaction amount per day.

```json
{
  "rule_type": "daily_limit",
  "parameters": {
    "max_amount": 10000000,
    "currency": "INR",
    "per_user": true
  }
}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `max_amount` | int64 | Maximum amount in paise |
| `currency` | string | Currency code |
| `per_user` | bool | Apply per user or globally |

### Threshold Rule
Flags or blocks transactions above/below certain amounts.

```json
{
  "rule_type": "threshold",
  "parameters": {
    "min_amount": 5000000,
    "max_amount": 0,
    "currency": "INR"
  }
}
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `min_amount` | int64 | Minimum amount to trigger (0 = no min) |
| `max_amount` | int64 | Maximum amount to trigger (0 = no max) |
| `currency` | string | Currency code |

## Risk Actions

| Action | Description | Effect |
|--------|-------------|--------|
| `allow` | Transaction allowed | Risk event logged only |
| `block` | Transaction blocked | Transaction rejected with error |
| `flag` | Transaction flagged | Transaction allowed but flagged for review |

## Risk Score

Risk scores range from 0-100:

| Score Range | Risk Level | Typical Action |
|-------------|------------|----------------|
| 0-20 | Low | Allow |
| 21-50 | Medium | Allow with monitoring |
| 51-80 | High | Flag for review |
| 81-100 | Critical | Block |

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+

### Environment Variables

Required:
- `SERVICE_PORT`: Server port (default: 8085)
- `DATABASE_PASSWORD`: PostgreSQL password
- `JWT_SECRET`: Secret for JWT validation

### Running the Service

```bash
cd services/risk
go run cmd/server/main.go
```

## Architecture

```
services/risk/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   │   ├── risk_handler.go
│   │   └── router.go
│   ├── service/         # Business logic
│   │   └── risk_service.go
│   ├── repository/      # Database operations
│   │   ├── risk_rule_repository.go
│   │   └── risk_event_repository.go
│   └── models/          # Domain models
│       ├── risk_rule.go
│       └── risk_event.go
├── Makefile
└── README.md
```

## Default Rules

The service comes with default rules for demonstration:

1. **Velocity Check**: Block if >10 transactions in 5 minutes
2. **Daily Limit**: Block if daily transfers exceed ₹1,00,000
3. **Large Transaction**: Flag transfers over ₹50,000

## Future Enhancements

- [ ] Machine learning-based risk scoring
- [ ] Device fingerprinting
- [ ] Geo-location based rules
- [ ] Real-time rule updates without restart
- [ ] Integration with external fraud detection services
- [ ] Watchlist/blacklist management
