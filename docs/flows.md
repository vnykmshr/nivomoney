---
layout: default
title: End-to-End Flows
nav_order: 5
description: "Complete user journeys through the Nivo platform"
permalink: /end-to-end-flows
---

# End-to-End Flows

This document describes the complete user journeys through the Nivo platform, illustrating how different microservices work together through the API Gateway.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Flow 1: User Onboarding](#flow-1-user-onboarding)
- [Flow 2: Wallet Creation and Activation](#flow-2-wallet-creation-and-activation)
- [Flow 3: Deposit Funds](#flow-3-deposit-funds)
- [Flow 4: Transfer Money](#flow-4-transfer-money)
- [Flow 5: Withdrawal](#flow-5-withdrawal)
- [System Interactions](#system-interactions)

---

## Architecture Overview

### Microservices
- **Gateway** (Port 8000) - Unified API entry point
- **Identity Service** (Port 8080) - User authentication & KYC
- **Wallet Service** (Port 8083) - Wallet management
- **Transaction Service** (Port 8084) - Money transfers
- **Ledger Service** (Port 8081) - Double-entry bookkeeping
- **RBAC Service** (Port 8082) - Role-based access control

### Request Flow
```
Client → Gateway (8000) → Backend Service (8080-8084)
                ↓
        Path Transformation
        /api/v1/{service}/{endpoint}
                ↓
        /api/v1/{endpoint}
```

### Data Stores
- **PostgreSQL** - Primary data store (all services)
- **Redis** - Caching and sessions
- **NSQ** - Message queue for async operations

---

## Flow 1: User Onboarding

### Complete New User Registration

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant I as Identity Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/identity/auth/register
    Note over C,G: {email, password, full_name, phone}

    G->>I: POST /api/v1/auth/register
    Note over G,I: Gateway strips "identity" from path

    I->>I: Validate email format
    I->>I: Validate phone format (+91XXXXXXXXXX)
    I->>I: Hash password (bcrypt)

    I->>DB: INSERT INTO users
    DB-->>I: User created

    I->>DB: INSERT INTO user_kyc (status: pending)
    DB-->>I: KYC record created

    I-->>G: 201 Created
    Note over I,G: {id, email, phone, status: "pending"}

    G-->>C: 201 Created
    Note over G,C: User account created
```

### Step-by-Step

1. **Client sends registration request** to Gateway
   ```bash
   POST http://localhost:8000/api/v1/identity/auth/register
   Content-Type: application/json

   {
     "email": "user@example.com",
     "password": "SecurePass123",
     "full_name": "John Doe",
     "phone": "+919876543210"
   }
   ```

2. **Gateway routes to Identity Service**
   - Strips `identity` from path: `/api/v1/auth/register`
   - Forwards to: `http://identity-service:8080/api/v1/auth/register`
   - Adds `X-Forwarded-*` headers

3. **Identity Service validates and creates user**
   - Validates email format (regex)
   - Validates phone format (must be +91XXXXXXXXXX)
   - Checks for duplicate email/phone
   - Hashes password with bcrypt
   - Creates user record (status: `pending`)
   - Creates KYC record (status: `pending`)

4. **Response**
   ```json
   {
     "success": true,
     "data": {
       "id": "f10f76f8-1c42-4f32-8254-45cd0c62ee68",
       "email": "user@example.com",
       "phone": "+919876543210",
       "full_name": "John Doe",
       "status": "pending",
       "created_at": "2025-11-24T01:33:52Z",
       "kyc": {
         "status": "pending"
       }
     }
   }
   ```

### User Login

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant I as Identity Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/identity/auth/login
    G->>I: POST /api/v1/auth/login

    I->>DB: SELECT user WHERE email = ?
    DB-->>I: User record

    I->>I: Compare password hash
    I->>I: Generate JWT token

    I->>DB: INSERT INTO sessions
    DB-->>I: Session created

    I-->>G: 200 OK
    G-->>C: 200 OK
    Note over G,C: {token, expires_at, user}
```

**Request:**
```bash
POST http://localhost:8000/api/v1/identity/auth/login

{
  "email": "user@example.com",
  "password": "SecurePass123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1764034442,
    "user": {
      "id": "f10f76f8-1c42-4f32-8254-45cd0c62ee68",
      "email": "user@example.com",
      "status": "pending"
    }
  }
}
```

---

## Flow 2: Wallet Creation and Activation

### Create Wallet

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant W as Wallet Service
    participant L as Ledger Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/wallet/wallets
    Note over C,G: Authorization: Bearer {token}

    G->>W: POST /api/v1/wallets
    Note over G,W: + X-Forwarded-Host<br/>+ X-Real-IP

    W->>W: Validate JWT token
    W->>W: Extract user_id from token

    W->>L: POST /api/v1/accounts/create
    Note over W,L: Create ledger account
    L->>DB: INSERT INTO accounts
    DB-->>L: Account created
    L-->>W: Account ID

    W->>DB: INSERT INTO wallets
    DB-->>W: Wallet created

    W-->>G: 201 Created
    G-->>C: 201 Created
```

**Request:**
```bash
POST http://localhost:8000/api/v1/wallet/wallets
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "currency": "INR"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "wallet_uuid",
    "user_id": "user_uuid",
    "currency": "INR",
    "balance": "0",
    "status": "pending",
    "created_at": "2025-11-24T01:35:00Z"
  }
}
```

### Activate Wallet

**Request:**
```bash
POST http://localhost:8000/api/v1/wallet/wallets/{wallet_id}/activate
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "wallet_uuid",
    "status": "active"
  }
}
```

---

## Flow 3: Deposit Funds

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant T as Transaction Service
    participant W as Wallet Service
    participant L as Ledger Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/transaction/transactions/deposit
    G->>T: POST /api/v1/transactions/deposit

    T->>T: Validate JWT & extract user
    T->>T: Generate transaction ID
    T->>T: Generate idempotency key

    T->>W: GET /api/v1/wallets/{wallet_id}
    W->>DB: SELECT wallet
    DB-->>W: Wallet data
    W-->>T: Wallet (verify ownership)

    T->>T: Verify wallet status = active
    T->>T: Verify wallet belongs to user

    T->>L: POST /api/v1/entries/create
    Note over T,L: Create double-entry<br/>Debit: System<br/>Credit: User Wallet
    L->>DB: BEGIN TRANSACTION
    L->>DB: INSERT INTO journal_entries
    L->>DB: INSERT INTO ledger_lines (debit)
    L->>DB: INSERT INTO ledger_lines (credit)
    L->>DB: COMMIT
    DB-->>L: Entry created
    L-->>T: Journal entry ID

    T->>W: POST /api/v1/wallets/{id}/credit
    W->>DB: UPDATE wallets<br/>SET balance = balance + amount
    DB-->>W: Balance updated
    W-->>T: New balance

    T->>DB: INSERT INTO transactions
    DB-->>T: Transaction saved

    T-->>G: 201 Created
    G-->>C: 201 Created
```

**Request:**
```bash
POST http://localhost:8000/api/v1/transaction/transactions/deposit
Authorization: Bearer {token}

{
  "wallet_id": "wallet_uuid",
  "amount": "5000.00",
  "currency": "INR",
  "description": "Initial deposit"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "tx_uuid",
    "type": "deposit",
    "to_wallet_id": "wallet_uuid",
    "amount": "5000.00",
    "currency": "INR",
    "status": "completed",
    "description": "Initial deposit",
    "created_at": "2025-11-24T01:40:00Z"
  }
}
```

---

## Flow 4: Transfer Money

### Between User Wallets

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant T as Transaction Service
    participant W as Wallet Service
    participant L as Ledger Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/transaction/transactions/transfer
    G->>T: POST /api/v1/transactions/transfer

    T->>W: GET /api/v1/wallets/{from_wallet_id}
    W-->>T: Source wallet (verify ownership)

    T->>W: GET /api/v1/wallets/{to_wallet_id}
    W-->>T: Destination wallet

    T->>T: Verify source wallet balance >= amount
    T->>T: Verify both wallets active
    T->>T: Verify same currency

    T->>DB: BEGIN TRANSACTION

    T->>W: POST /api/v1/wallets/{from}/debit
    W->>DB: UPDATE wallets<br/>SET balance = balance - amount<br/>WHERE id = from_wallet_id
    W-->>T: Debited

    T->>W: POST /api/v1/wallets/{to}/credit
    W->>DB: UPDATE wallets<br/>SET balance = balance + amount<br/>WHERE id = to_wallet_id
    W-->>T: Credited

    T->>L: POST /api/v1/entries/create
    Note over T,L: Double-entry:<br/>Debit: From Wallet<br/>Credit: To Wallet
    L->>DB: INSERT journal entries
    L-->>T: Entry ID

    T->>DB: INSERT INTO transactions
    T->>DB: COMMIT

    T-->>G: 201 Created
    G-->>C: 201 Created
```

**Request:**
```bash
POST http://localhost:8000/api/v1/transaction/transactions/transfer
Authorization: Bearer {token}

{
  "from_wallet_id": "wallet1_uuid",
  "to_wallet_id": "wallet2_uuid",
  "amount": "1500.00",
  "currency": "INR",
  "description": "Payment for services"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "tx_uuid",
    "type": "transfer",
    "from_wallet_id": "wallet1_uuid",
    "to_wallet_id": "wallet2_uuid",
    "amount": "1500.00",
    "currency": "INR",
    "status": "completed",
    "created_at": "2025-11-24T01:45:00Z"
  }
}
```

---

## Flow 5: Withdrawal

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant T as Transaction Service
    participant W as Wallet Service
    participant L as Ledger Service

    C->>G: POST /api/v1/transaction/transactions/withdrawal
    G->>T: POST /api/v1/transactions/withdrawal

    T->>W: GET /api/v1/wallets/{wallet_id}
    W-->>T: Wallet data

    T->>T: Verify balance >= amount
    alt Insufficient funds
        T-->>G: 400 Bad Request
        G-->>C: "Insufficient funds"
    else Sufficient funds
        T->>W: POST /api/v1/wallets/{id}/debit
        W-->>T: Balance updated

        T->>L: POST /api/v1/entries/create
        L-->>T: Entry created

        T-->>G: 201 Created
        G-->>C: Withdrawal successful
    end
```

**Request:**
```bash
POST http://localhost:8000/api/v1/transaction/transactions/withdrawal
Authorization: Bearer {token}

{
  "wallet_id": "wallet_uuid",
  "amount": "2000.00",
  "currency": "INR",
  "description": "Cash withdrawal"
}
```

**Error Response (Insufficient Funds):**
```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Wallet balance insufficient for withdrawal"
  }
}
```

---

## System Interactions

### Gateway Routing

The gateway performs path transformation for all requests:

**Client Request:**
```
POST http://localhost:8000/api/v1/identity/auth/register
```

**Gateway Processing:**
1. Extract service name: `identity`
2. Lookup service URL: `http://identity-service:8080`
3. Strip service name from path: `/api/v1/auth/register`
4. Add headers:
   - `X-Forwarded-Host`: Original host
   - `X-Forwarded-Proto`: http/https
   - `X-Real-IP`: Client IP
   - `X-Request-ID`: Request tracking ID

**Forwarded Request:**
```
POST http://identity-service:8080/api/v1/auth/register
```

### Authentication Flow

```
1. User logs in → Receives JWT token
2. Client includes token in header: "Authorization: Bearer {token}"
3. Gateway forwards token to backend service
4. Backend service validates token (shared JWT secret)
5. Service extracts user_id from token claims
6. Service processes request for that user
```

### Error Handling

All errors follow a consistent format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

Common error codes:
- `UNAUTHORIZED` (401) - Missing or invalid auth token
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource doesn't exist
- `CONFLICT` (409) - Duplicate resource
- `VALIDATION_ERROR` (400) - Invalid request data
- `INSUFFICIENT_FUNDS` (400) - Not enough balance
- `INTERNAL_ERROR` (500) - Server error

### Data Consistency

**Ledger Double-Entry:**
Every financial transaction creates two ledger entries:
- Debit entry (source account)
- Credit entry (destination account)
- Total debits = Total credits (always balanced)

**Database Transactions:**
Multi-step operations use database transactions:
```sql
BEGIN TRANSACTION;
  -- Debit source wallet
  UPDATE wallets SET balance = balance - 1500 WHERE id = 'wallet1';

  -- Credit destination wallet
  UPDATE wallets SET balance = balance + 1500 WHERE id = 'wallet2';

  -- Create transaction record
  INSERT INTO transactions (...);

  -- Create ledger entries
  INSERT INTO journal_entries (...);
COMMIT;
```

If any step fails, entire transaction rolls back.

---

## Complete User Journey Example

```bash
# 1. Register user
curl -X POST http://localhost:8000/api/v1/identity/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123",
    "full_name": "John Doe",
    "phone": "+919876543210"
  }'

# 2. Login
TOKEN=$(curl -X POST http://localhost:8000/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"SecurePass123"}' \
  | jq -r '.data.token')

# 3. Create wallet
WALLET_ID=$(curl -X POST http://localhost:8000/api/v1/wallet/wallets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"currency":"INR"}' \
  | jq -r '.data.id')

# 4. Activate wallet
curl -X POST http://localhost:8000/api/v1/wallet/wallets/$WALLET_ID/activate \
  -H "Authorization: Bearer $TOKEN"

# 5. Deposit funds
curl -X POST http://localhost:8000/api/v1/transaction/transactions/deposit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"wallet_id\": \"$WALLET_ID\",
    \"amount\": \"10000.00\",
    \"currency\": \"INR\",
    \"description\": \"Initial deposit\"
  }"

# 6. Check balance
curl -X GET http://localhost:8000/api/v1/wallet/wallets/$WALLET_ID/balance \
  -H "Authorization: Bearer $TOKEN"
```

---

## Next Steps

- Implement KYC verification flow
- Add RBAC permission checks
- Implement transaction reversals
- Add webhook notifications
- Implement rate limiting per user
- Add audit logging for all operations
