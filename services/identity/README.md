# Identity Service

The Identity Service handles user authentication, authorization, and KYC (Know Your Customer) verification for Nivo, an India-centric neobank platform.

## Features

- **User Registration**: Create new user accounts with India-specific validation (email, phone, PAN, Aadhaar)
- **Authentication**: JWT-based authentication with session management
- **KYC Management**: Submit, verify, and track India-specific KYC documents (PAN, Aadhaar)
- **Session Tracking**: Monitor active sessions with IP address and user agent
- **Security**: Bcrypt password hashing, SHA-256 token hashing, secure session storage

## API Endpoints

### Public Endpoints (No Authentication)

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "phone": "+919876543210",
  "full_name": "John Doe",
  "password": "secure_password_123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "phone": "+919876543210",
    "full_name": "John Doe",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "secure_password_123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1705320600,
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "phone": "+919876543210",
      "full_name": "John Doe",
      "status": "pending"
    }
  }
}
```

### Protected Endpoints (Requires Authentication)

All protected endpoints require an `Authorization` header with a Bearer token:
```http
Authorization: Bearer <jwt_token>
```

#### Get Profile
```http
GET /api/v1/auth/me
```

#### Logout
```http
POST /api/v1/auth/logout
```

#### Logout All Devices
```http
POST /api/v1/auth/logout-all
```

#### Get KYC Status
```http
GET /api/v1/auth/kyc
```

#### Submit/Update KYC
```http
PUT /api/v1/auth/kyc
Content-Type: application/json

{
  "pan": "ABCDE1234F",
  "aadhaar": "123456789012",
  "date_of_birth": "1990-01-15",
  "address": {
    "street": "123 Main Street",
    "city": "Mumbai",
    "state": "Maharashtra",
    "pin": "400001",
    "country": "IN"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "KYC information updated successfully",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "pending",
    "pan": "ABCDE1234F",
    "date_of_birth": "1990-01-15",
    "address": {
      "street": "123 Main Street",
      "city": "Mumbai",
      "state": "Maharashtra",
      "pin": "400001",
      "country": "IN"
    },
    "created_at": "2024-01-15T10:35:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

### Admin Endpoints (Requires Admin Status)

#### Verify KYC
```http
POST /api/v1/admin/kyc/verify
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### Reject KYC
```http
POST /api/v1/admin/kyc/reject
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "reason": "Invalid PAN card details"
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
  "service": "identity"
}
```

## Setup

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 14 or higher
- Make (optional, for using Makefile commands)

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Required environment variables:
- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection URL
- `JWT_SECRET`: Secret key for JWT signing (change in production!)
- `ENVIRONMENT`: Environment (development, staging, production)

### Database Setup

1. Create PostgreSQL database:
```sql
CREATE DATABASE nivo;
CREATE USER nivo WITH PASSWORD 'nivo_dev_password';
GRANT ALL PRIVILEGES ON DATABASE nivo TO nivo;
```

2. Run migrations:
```bash
make migrate-up
```

Or manually:
```bash
psql -U nivo -d nivo -f migrations/001_create_users.up.sql
psql -U nivo -d nivo -f migrations/002_create_kyc.up.sql
psql -U nivo -d nivo -f migrations/003_create_sessions.up.sql
```

### Running the Service

#### Using Make
```bash
make run
```

#### Using Go directly
```bash
go run cmd/server/main.go
```

#### Using Docker
```bash
make docker-build
make docker-run
```

## India-Specific Validations

The Identity Service implements India-centric validations:

### Phone Numbers
- Format: `+91` followed by 10 digits
- First digit must be 6-9
- Example: `+919876543210`

### PAN (Permanent Account Number)
- Format: 5 letters + 4 digits + 1 letter
- Must be uppercase
- Example: `ABCDE1234F`

### Aadhaar
- Format: 12 digits
- Cannot start with 0 or 1
- Example: `234567890123`

### PIN Code
- Format: 6 digits
- Cannot start with 0
- Example: `400001`

### Address
- Required fields: street, city, state, pin, country
- PIN code must be valid Indian PIN
- Country defaults to "IN" (India)

## User Status Workflow

1. **Pending**: User registered but KYC not submitted
2. **Active**: KYC verified, full access to platform
3. **Suspended**: Temporarily blocked
4. **Closed**: Account permanently closed

## KYC Status Workflow

1. **Pending**: KYC submitted, awaiting verification
2. **Verified**: KYC approved by admin
3. **Rejected**: KYC rejected with reason
4. **Expired**: KYC documents expired (periodic re-verification)

## Security Features

- **Password Hashing**: Bcrypt with DefaultCost (10)
- **JWT Tokens**: HS256 signing with configurable expiry
- **Token Storage**: SHA-256 hashed tokens in database
- **Session Tracking**: IP address and user agent logging
- **PII Protection**: Aadhaar never exposed in API responses
- **CORS**: Configurable CORS middleware

## Testing

Run tests:
```bash
make test
```

Or:
```bash
go test -v -cover ./...
```

## Architecture

```
services/identity/
├── cmd/
│   └── server/          # Server entry point
├── internal/
│   ├── handler/         # HTTP handlers and middleware
│   ├── service/         # Business logic
│   ├── repository/      # Database operations
│   └── models/          # Domain models
├── migrations/          # SQL migration files
├── proto/              # gRPC definitions (future)
├── Makefile            # Build and run commands
└── README.md           # This file
```

## Dependencies

- `github.com/golang-jwt/jwt/v5`: JWT token generation and validation
- `golang.org/x/crypto/bcrypt`: Password hashing
- `github.com/lib/pq`: PostgreSQL driver
- `github.com/vnykmshr/gopantic`: Request validation
- `github.com/vnykmshr/nivo/shared`: Shared utilities

## Future Enhancements

- [ ] OAuth2 integration (Google, Facebook)
- [ ] Two-factor authentication (2FA)
- [ ] Rate limiting per user
- [ ] Email verification
- [ ] SMS OTP for phone verification
- [ ] Admin dashboard
- [ ] Audit logging
- [ ] gRPC API for inter-service communication
