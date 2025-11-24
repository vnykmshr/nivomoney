# Integration Tests

This directory contains integration tests for the Nivo platform that validate end-to-end flows through the API Gateway.

## Overview

These tests verify that all microservices work correctly together through the unified API Gateway at `http://localhost:8000`.

## Prerequisites

Before running integration tests, ensure:

1. **All services are running**:
   ```bash
   docker compose up -d
   ```

2. **Verify gateway is healthy**:
   ```bash
   curl http://localhost:8000/health
   ```

3. **Testify dependency is installed**:
   ```bash
   go get github.com/stretchr/testify
   go mod tidy
   ```

## Running Tests

### Run all integration tests:
```bash
make test-integration
```

Or directly with go:
```bash
go test -v -tags=integration ./tests/integration/... -timeout 5m
```

### Run specific tests:
```bash
go test -v -tags=integration ./tests/integration/... -run TestUserRegistrationAndLogin
```

### Run without verbose output:
```bash
go test -tags=integration ./tests/integration/...
```

### Skip long-running tests:
```bash
go test -v -tags=integration -short ./tests/integration/...
```

## Test Coverage

### TestGatewayHealthCheck
- Verifies gateway is running and responding
- Endpoint: `GET /health`

### TestUserRegistrationAndLogin ✅
Complete authentication flow through the gateway:

1. **User Registration**
   - Endpoint: `POST /api/v1/identity/auth/register`
   - Validates user creation
   - Verifies email, phone, and status fields

2. **User Login**
   - Endpoint: `POST /api/v1/identity/auth/login`
   - Validates JWT token generation
   - Stores auth token for subsequent requests

3. **Login with Wrong Password**
   - Validates 401 Unauthorized response
   - Confirms error handling

4. **Duplicate User Registration**
   - Validates 409 Conflict response
   - Confirms uniqueness constraints

### TestWalletCreationAndManagement
Wallet operations through the gateway:

1. **Create Wallet**
   - Endpoint: `POST /api/v1/wallet/wallets`
   - Requires authentication
   - Validates wallet creation with zero balance

2. **Get Wallet Details**
   - Endpoint: `GET /api/v1/wallet/wallets/:id`
   - Validates wallet retrieval

3. **Activate Wallet**
   - Endpoint: `POST /api/v1/wallet/wallets/:id/activate`
   - Validates status transition

4. **Authentication Required**
   - Validates 401 response without auth token

### TestTransactionFlow
Transaction operations through the gateway:

1. **Deposit Transaction**
   - Endpoint: `POST /api/v1/transaction/transactions/deposit`
   - Validates deposit creation
   - Verifies transaction metadata

2. **Verify Balance**
   - Endpoint: `GET /api/v1/wallet/wallets/:id/balance`
   - Validates balance update after deposit

3. **Insufficient Funds**
   - Validates withdrawal failure with insufficient balance
   - Confirms proper error handling

### TestEndToEndUserJourney
Complete user journey simulating real-world usage:

1. **User Registration** → New user account
2. **User Login** → JWT authentication
3. **Create Primary Wallet** → INR wallet
4. **Activate Wallet** → Status: pending → active
5. **Deposit Funds** → 5000.00 INR
6. **Create Second Wallet** → For transfers
7. **Transfer Between Wallets** → 1500.00 INR

This test validates the complete flow a real user would follow.

## Architecture

### Test Client Helper
The `TestClient` struct provides helper methods for making authenticated HTTP requests:

```go
client := NewTestClient(t)
client.SetAuthToken(token)

// Make authenticated POST request
resp, statusCode := client.Post("/api/v1/wallet/wallets", body, true)

// Make unauthenticated GET request
resp, statusCode := client.Get("/health", false)
```

### API Response Structure
Tests expect responses in this format:
```json
{
  "success": true|false,
  "data": { ... },      // On success
  "error": {            // On failure
    "code": "ERROR_CODE",
    "message": "Error message"
  }
}
```

## Test Data Generation

Tests generate unique test data using timestamps to avoid conflicts:
```go
timestamp := time.Now().Unix()
email := fmt.Sprintf("test-%d@nivo.com", timestamp)
phone := fmt.Sprintf("+919%09d", timestamp%1000000000)
```

## Known Issues

### Rate Limiting
Some tests may fail with "rate limit exceeded" if run in rapid succession. This is expected behavior to protect the API. Solutions:

1. Add delays between test executions
2. Run tests with `-p 1` flag to disable parallelization
3. Clear rate limit cache between runs

### Health Check Format
The gateway health endpoint returns a non-standard format:
```json
{
  "service": "gateway",
  "status": "healthy",
  "version": "1.0.0"
}
```

This differs from other API responses and requires special handling.

## Test Results

Last run results:
- ✅ `TestUserRegistrationAndLogin` - PASSED (all 4 subtests)
- ⚠️  `TestGatewayHealthCheck` - Format mismatch
- ⚠️  `TestWalletCreationAndManagement` - Rate limiting
- ⚠️  `TestTransactionFlow` - Rate limiting
- ⚠️  `TestEndToEndUserJourney` - Rate limiting

## Debugging Tests

### View detailed request/response:
Modify test code to log request/response data:
```go
t.Logf("Request: %+v", req)
t.Logf("Response: %s", string(bodyBytes))
```

### Check service logs:
```bash
docker compose logs -f gateway identity-service wallet-service transaction-service
```

### Verify service connectivity:
```bash
# Check all services are healthy
docker compose ps

# Test direct service access
curl http://localhost:8080/health  # Identity
curl http://localhost:8083/health  # Wallet
curl http://localhost:8084/health  # Transaction
```

## Best Practices

1. **Unique Test Data**: Always generate unique emails/phones to avoid conflicts
2. **Authentication**: Store and reuse JWT tokens for authenticated requests
3. **Cleanup**: Tests create real data - consider cleanup scripts
4. **Timeouts**: Use appropriate timeouts for long-running operations
5. **Idempotency**: Design tests to be repeatable

## Contributing

When adding new integration tests:

1. Follow the existing test structure
2. Use descriptive test names: `Test<Feature><Scenario>`
3. Use subtests with `t.Run()` for logical groupings
4. Document expected behavior in test comments
5. Add test coverage to this README

## Future Enhancements

- [ ] Add RBAC integration tests
- [ ] Add ledger integration tests
- [ ] Test concurrent transactions
- [ ] Test wallet freezing/unfreezing
- [ ] Test transaction reversals
- [ ] Add performance benchmarks
- [ ] Integrate with CI/CD pipeline
