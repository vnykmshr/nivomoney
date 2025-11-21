# Errors Package

Custom error types and error handling utilities for Nivo services.

## Overview

The `errors` package provides structured error handling with error codes, HTTP status code mapping, and support for error wrapping and details. It ensures consistent error responses across all services.

## Features

- **Error Codes**: Predefined codes for common error scenarios
- **HTTP Mapping**: Automatic HTTP status code mapping
- **Error Wrapping**: Support for Go 1.13+ error wrapping
- **Details**: Attach structured metadata to errors
- **Type Checking**: Helper functions for error type identification

## Error Codes

### Client Errors (4xx)
- `NOT_FOUND` - Resource not found
- `BAD_REQUEST` - Invalid request
- `VALIDATION_ERROR` - Input validation failed
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Permission denied
- `CONFLICT` - Resource conflict
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `PRECONDITION_FAILED` - Precondition not met

### Server Errors (5xx)
- `INTERNAL_ERROR` - Internal server error
- `SERVICE_UNAVAILABLE` - Service unavailable
- `TIMEOUT` - Request timeout
- `DATABASE_ERROR` - Database operation failed

### Domain-Specific Errors
- `INSUFFICIENT_FUNDS` - Not enough balance
- `ACCOUNT_FROZEN` - Account is frozen
- `TRANSACTION_FAILED` - Transaction failed
- `DUPLICATE_IDEMPOTENCY_KEY` - Idempotency key already used
- `INVALID_AMOUNT` - Invalid transaction amount
- `INVALID_CURRENCY` - Invalid or unsupported currency

## Usage

### Creating Errors

```go
import "github.com/vnykmshr/nivo/shared/errors"

// Using predefined constructors
err := errors.NotFound("user")
err := errors.BadRequest("missing required field: email")
err := errors.Unauthorized("invalid token")
err := errors.Internal("unexpected error occurred")

// With resource ID
err := errors.NotFoundWithID("wallet", "wallet-123")
// Output: "wallet with id wallet-123 not found"
```

### Custom Errors

```go
// Create custom error
err := errors.New(errors.ErrCodeValidation, "email format is invalid")

// With error code
err := errors.New(errors.ErrCodeInsufficientFunds, "balance too low")
```

### Adding Details

```go
// Add structured details
err := errors.Validation("invalid request").
    AddDetail("field", "email").
    AddDetail("constraint", "must be valid email format")

// Add multiple details at once
details := map[string]interface{}{
    "field": "amount",
    "min":   0.01,
    "max":   10000.00,
}
err := errors.Validation("amount out of range").WithDetails(details)
```

### Validation with Field Errors

```go
fieldErrors := map[string]string{
    "email":    "must be valid email",
    "password": "must be at least 8 characters",
    "age":      "must be 18 or older",
}
err := errors.ValidationWithFields("validation failed", fieldErrors)
```

### Wrapping Errors

```go
// Wrap existing error
dbErr := db.Query(...)
if dbErr != nil {
    return errors.DatabaseWrap(dbErr, "failed to fetch user")
}

// Wrap with formatted message
err := errors.Wrapf(dbErr, errors.ErrCodeDatabaseError,
    "failed to query table %s", tableName)

// Wrap as internal error
err := errors.InternalWrap(panicErr, "recovered from panic")
```

### HTTP Status Code Mapping

```go
err := errors.NotFound("user")
statusCode := err.HTTPStatusCode()  // Returns 404

// Or use the helper function
statusCode := errors.GetHTTPStatus(err)
```

### Error Type Checking

```go
err := someOperation()

// Check specific error types
if errors.IsNotFound(err) {
    // Handle not found
}

if errors.IsValidation(err) {
    // Handle validation error
}

if errors.IsUnauthorized(err) {
    // Handle auth error
}

if errors.IsInternal(err) {
    // Log and return generic message
}

// Get error code
code := errors.GetErrorCode(err)
switch code {
case errors.ErrCodeNotFound:
    // Handle not found
case errors.ErrCodeValidation:
    // Handle validation
}
```

### Error Unwrapping

```go
// Standard Go error unwrapping works
underlying := errors.Unwrap(err)

// Check if error wraps another
if errors.Is(err, sql.ErrNoRows) {
    // Handle no rows
}

// Extract specific error type
var dbError *DatabaseError
if errors.As(err, &dbError) {
    // Handle database error
}
```

## HTTP Handler Integration

```go
func handleError(w http.ResponseWriter, err error) {
    // Extract error details
    var customErr *errors.Error
    if errors.As(err, &customErr) {
        statusCode := customErr.HTTPStatusCode()
        response := map[string]interface{}{
            "error": map[string]interface{}{
                "code":    customErr.Code,
                "message": customErr.Message,
                "details": customErr.Details,
            },
        }
        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(response)
        return
    }

    // Fallback for non-custom errors
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "internal server error",
    })
}
```

## Complete Example

```go
package main

import (
    "database/sql"
    "github.com/vnykmshr/nivo/shared/errors"
)

func GetUser(id string) (*User, error) {
    // Validate input
    if id == "" {
        return nil, errors.BadRequest("user id is required")
    }

    // Query database
    user, err := db.QueryUser(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NotFoundWithID("user", id)
        }
        return nil, errors.DatabaseWrap(err, "failed to query user")
    }

    // Check business rules
    if user.Status == "frozen" {
        return nil, errors.AccountFrozen("user account is frozen")
    }

    return user, nil
}

func TransferMoney(from, to string, amount float64) error {
    // Validate amount
    if amount <= 0 {
        return errors.Validation("amount must be positive").
            AddDetail("amount", amount)
    }

    // Check balance
    balance := getBalance(from)
    if balance < amount {
        return errors.InsufficientFunds("insufficient balance").
            AddDetail("balance", balance).
            AddDetail("required", amount)
    }

    // Perform transfer
    if err := doTransfer(from, to, amount); err != nil {
        return errors.TransactionFailed("transfer failed").
            AddDetail("from", from).
            AddDetail("to", to).
            AddDetail("amount", amount)
    }

    return nil
}
```

## JSON Response Format

When serialized to JSON, errors have the following structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid input",
    "details": {
      "field": "email",
      "constraint": "must be valid email format"
    }
  }
}
```

## Best Practices

1. **Use Specific Error Codes**: Choose the most specific error code
2. **Add Context**: Use `WithDetails` to add helpful debugging information
3. **Wrap Underlying Errors**: Preserve error chains with `Wrap`
4. **Don't Expose Internal Details**: Be careful with error messages in production
5. **Log Internal Errors**: Always log full error details server-side
6. **Return Generic Messages**: Return generic messages to clients for internal errors

## Testing

```bash
go test ./shared/errors/...
go test -cover ./shared/errors/...
```

## Related Packages

- [shared/logger](../logger/README.md) - For logging errors
- [shared/response](../response/README.md) - For formatting error responses (to be created)
