---
layout: default
title: End-to-End Flows
nav_order: 5
description: "Complete user journeys through the Nivo platform"
permalink: /flows
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

## Flow 6: OTP Verification (User-Admin Self-Service)

The User-Admin self-service model provides secure OTP verification for sensitive operations without relying on external SMS/email providers.

### Architecture

```
User App → Regular User Account → Request sensitive operation
                    ↓
        System creates verification record + OTP
                    ↓
User Admin Portal → User-Admin Account → View pending OTPs
                    ↓
        User reads OTP, enters in User App
                    ↓
User App → Verify OTP → Complete operation
```

### Verification Flow

```mermaid
sequenceDiagram
    participant UA as User App
    participant G as Gateway
    participant I as Identity Service
    participant UAP as User-Admin Portal
    participant DB as PostgreSQL

    Note over UA,UAP: 1. User initiates sensitive operation
    UA->>G: POST /api/v1/identity/auth/password/change
    G->>I: POST /api/v1/auth/password/change
    I->>DB: Check verification token
    I-->>G: 202 Accepted + verification_required
    G-->>UA: Need verification

    Note over UA,UAP: 2. User opens separate admin portal (different browser/device)
    UAP->>G: POST /api/v1/identity/auth/login
    Note over UAP,G: Login with same email, account_type=user_admin
    G->>I: POST /api/v1/auth/login
    I-->>G: User-Admin JWT token
    G-->>UAP: Admin portal access granted

    Note over UA,UAP: 3. User-Admin views pending verifications
    UAP->>G: GET /api/v1/identity/verifications/pending
    G->>I: GET /api/v1/verifications/pending
    I->>DB: SELECT verifications WHERE status=pending
    DB-->>I: Pending verifications with OTPs
    I-->>G: {verifications: [{id, type, otp: "847291", expires_at}]}
    G-->>UAP: Display OTPs

    Note over UA,UAP: 4. User enters OTP in original app
    UA->>G: POST /api/v1/identity/verifications/{id}/verify
    G->>I: POST /api/v1/verifications/{id}/verify
    I->>DB: Validate OTP
    I->>DB: UPDATE verification SET status=verified
    I-->>G: {verified: true, token: "ver_token"}
    G-->>UA: Verification successful

    Note over UA,UAP: 5. Complete original operation with token
    UA->>G: POST /api/v1/identity/auth/password/change
    Note over UA,G: Include verification_token in body
    G->>I: POST /api/v1/auth/password/change
    I->>DB: Verify token, update password
    I-->>G: Password changed
    G-->>UA: Success
```

### Operations Requiring Verification

| Operation | Endpoint | Verification Type |
|-----------|----------|-------------------|
| Password Change | `POST /auth/password/change` | `password_change` |
| Add Beneficiary | `POST /beneficiaries` | `beneficiary_add` |
| High-Value Transfer (>₹50,000) | `POST /transactions/transfer` | `high_value_transfer` |

### API Examples

**1. Initiate Password Change (returns verification required):**
```bash
POST http://localhost:8000/api/v1/identity/auth/password/change
Authorization: Bearer {user_token}

{
  "current_password": "OldPass123",
  "new_password": "NewSecurePass456"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verification_required": true,
    "verification_id": "ver_abc123def456",
    "message": "Please complete OTP verification via User-Admin portal"
  }
}
```

**2. Login to User-Admin Portal:**
```bash
POST http://localhost:8000/api/v1/identity/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "OldPass123",
  "account_type": "user_admin"
}
```

**3. View Pending Verifications:**
```bash
GET http://localhost:8000/api/v1/identity/verifications/pending
Authorization: Bearer {user_admin_token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verifications": [{
      "id": "ver_abc123def456",
      "operation_type": "password_change",
      "otp": "847291",
      "status": "pending",
      "expires_at": "2025-01-07T15:30:00Z",
      "created_at": "2025-01-07T15:25:00Z"
    }]
  }
}
```

**4. Verify OTP:**
```bash
POST http://localhost:8000/api/v1/identity/verifications/ver_abc123def456/verify
Authorization: Bearer {user_token}

{
  "otp": "847291"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "verified": true,
    "token": {
      "token": "ver_token_xyz789"
    }
  }
}
```

**5. Complete Password Change:**
```bash
POST http://localhost:8000/api/v1/identity/auth/password/change
Authorization: Bearer {user_token}

{
  "current_password": "OldPass123",
  "new_password": "NewSecurePass456",
  "verification_token": "ver_token_xyz789"
}
```

---

## Flow 7: UPI Deposit (Simulation)

UPI deposits simulate the flow of adding funds via UPI payment. In demo mode, payments auto-complete after a delay.

```mermaid
sequenceDiagram
    participant C as Client
    participant G as Gateway
    participant W as Wallet Service
    participant S as Simulation Engine
    participant T as Transaction Service
    participant DB as PostgreSQL

    C->>G: POST /api/v1/wallet/wallets/{id}/deposit/upi
    G->>W: POST /api/v1/wallets/{id}/deposit/upi
    Note over G,W: {amount, upi_id: "user@paytm"}

    W->>DB: INSERT INTO upi_deposits (status: pending)
    W->>S: Queue deposit completion
    W-->>G: 201 Created (pending)
    G-->>C: Deposit initiated

    Note over S,DB: After 3-10 seconds (simulation delay)
    S->>T: POST /api/v1/transactions/deposit
    T->>DB: Complete deposit, update balance
    S->>DB: UPDATE upi_deposits SET status=completed
```

**Request:**
```bash
POST http://localhost:8000/api/v1/wallet/wallets/{walletId}/deposit/upi
Authorization: Bearer {token}

{
  "amount": 5000,
  "upi_id": "user@paytm"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "deposit_id": "dep_uuid",
    "amount": 5000,
    "upi_id": "user@paytm",
    "status": "pending",
    "message": "Simulating UPI payment - will complete in 3-10 seconds"
  }
}
```

---

## Flow 8: Virtual Card Operations

Users can create virtual debit cards linked to their wallets for online payments.

### Create Virtual Card

```bash
POST http://localhost:8000/api/v1/wallet/wallets/{walletId}/cards
Authorization: Bearer {token}

{
  "card_holder_name": "John Doe"
}
```

**Response (includes CVV - only shown once):**
```json
{
  "success": true,
  "data": {
    "card": {
      "id": "card_uuid",
      "card_number_masked": "4000 **** **** 1234",
      "card_holder_name": "John Doe",
      "expiry_month": 12,
      "expiry_year": 2028,
      "status": "active",
      "daily_limit": 50000,
      "monthly_limit": 500000
    },
    "details": {
      "card_number": "4000123456781234",
      "expiry_month": 12,
      "expiry_year": 2028,
      "cvv": "123"
    },
    "message": "Save your card details securely. CVV will not be shown again."
  }
}
```

### Card Management

| Operation | Method | Endpoint |
|-----------|--------|----------|
| List Cards | GET | `/wallets/{walletId}/cards` |
| Get Card | GET | `/cards/{cardId}` |
| Freeze Card | POST | `/cards/{cardId}/freeze` |
| Unfreeze Card | POST | `/cards/{cardId}/unfreeze` |
| Update Limits | PATCH | `/cards/{cardId}/limits` |
| Cancel Card | DELETE | `/cards/{cardId}` |

---

## Flow 9: Spending Categories and Statement Export

### Update Transaction Category

```bash
PATCH http://localhost:8000/api/v1/transaction/transactions/{id}/category
Authorization: Bearer {token}

{
  "category": "food"
}
```

### Get Spending Summary

```bash
GET http://localhost:8000/api/v1/transaction/wallets/{walletId}/spending-summary?period=monthly
Authorization: Bearer {token}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "period": "2025-01",
    "total_spent": 25000,
    "categories": {
      "food": 8500,
      "transport": 3200,
      "shopping": 7800,
      "utilities": 3500,
      "entertainment": 2000
    }
  }
}
```

### Export Statement

**CSV Export:**
```bash
GET http://localhost:8000/api/v1/transaction/wallets/{walletId}/statements/csv?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}
```

**PDF Export:**
```bash
GET http://localhost:8000/api/v1/transaction/wallets/{walletId}/statements/pdf?start_date=2025-01-01&end_date=2025-01-31
Authorization: Bearer {token}
```

---

## Error Code Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | Resource not found |
| `BAD_REQUEST` | 400 | Invalid request data |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `CONFLICT` | 409 | Resource conflict |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INSUFFICIENT_FUNDS` | 412 | Not enough balance |
| `ACCOUNT_FROZEN` | 412 | Account is frozen |
| `LIMIT_EXCEEDED` | 412 | Transaction limit exceeded |
| `VERIFICATION_REQUIRED` | 202 | OTP verification needed |
| `VERIFICATION_EXPIRED` | 410 | Verification timed out |
| `INVALID_OTP` | 400 | Wrong OTP code |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Next Steps

- Add webhook notifications for transaction events
- Implement scheduled payments
- Add support for recurring transfers
- Implement P2P payment requests
