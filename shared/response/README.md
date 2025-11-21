# Response Package

Standardized HTTP response formats for Nivo APIs, providing consistent success and error response envelopes.

## Features

- **Standardized Response Envelope**: Consistent structure for all API responses
- **Success Responses**: Helper functions for common success scenarios
- **Error Responses**: Integration with `shared/errors` for consistent error handling
- **Pagination Support**: Built-in pagination metadata
- **Metadata**: Request ID, timestamp, and custom metadata support
- **Type-Safe**: Structured types for responses, errors, and metadata
- **HTTP Status Codes**: Automatic status code mapping from errors

## Installation

```bash
go get github.com/vnykmshr/nivo/shared/response
```

## Response Structure

All responses follow a standardized envelope format:

```json
{
  "success": true|false,
  "data": { ... },           // Present on success
  "error": { ... },          // Present on error
  "meta": {                  // Optional metadata
    "request_id": "...",
    "timestamp": "...",
    "pagination": { ... }
  }
}
```

## Usage

### Success Responses

```go
import "github.com/vnykmshr/nivo/shared/response"

// 200 OK
func getUser(w http.ResponseWriter, r *http.Request) {
    user := map[string]interface{}{
        "id":    "123",
        "name":  "John Doe",
        "email": "john@example.com",
    }
    response.OK(w, user)
}

// 201 Created
func createUser(w http.ResponseWriter, r *http.Request) {
    newUser := map[string]interface{}{
        "id":   "456",
        "name": "Jane Doe",
    }
    response.Created(w, newUser)
}

// 204 No Content
func deleteUser(w http.ResponseWriter, r *http.Request) {
    // ... perform deletion
    response.NoContent(w)
}
```

### Error Responses

```go
// 400 Bad Request
func validateInput(w http.ResponseWriter, r *http.Request) {
    if err := validateRequest(r); err != nil {
        response.BadRequest(w, "invalid request format")
        return
    }
}

// 401 Unauthorized
func authenticate(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        response.Unauthorized(w, "invalid or expired token")
        return
    }
}

// 403 Forbidden
func authorize(w http.ResponseWriter, r *http.Request) {
    if !hasPermission(r) {
        response.Forbidden(w, "insufficient permissions")
        return
    }
}

// 404 Not Found
func getResource(w http.ResponseWriter, r *http.Request) {
    resource := findResource(id)
    if resource == nil {
        response.NotFound(w, "user")
        return
    }
    response.OK(w, resource)
}

// 409 Conflict
func createResource(w http.ResponseWriter, r *http.Request) {
    if exists(email) {
        response.Conflict(w, "email already exists")
        return
    }
}

// 500 Internal Server Error
func handleDatabaseError(w http.ResponseWriter, r *http.Request) {
    if err := db.Query(...); err != nil {
        response.InternalError(w, "database error")
        return
    }
}
```

### Custom Error Responses

Use with `shared/errors` for full error control:

```go
import (
    "github.com/vnykmshr/nivo/shared/errors"
    "github.com/vnykmshr/nivo/shared/response"
)

func transfer(w http.ResponseWriter, r *http.Request) {
    if balance < amount {
        err := errors.InsufficientFunds("insufficient balance")
        err.WithDetail("balance", balance).
            WithDetail("required", amount)
        response.Error(w, err)
        return
    }
}
```

### Validation Errors

```go
func validateUser(w http.ResponseWriter, r *http.Request) {
    validationErrors := map[string]interface{}{
        "email":    "invalid email format",
        "password": "must be at least 8 characters",
        "age":      "must be 18 or older",
    }

    response.ValidationError(w, validationErrors)
}
```

Response:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "email": "invalid email format",
      "password": "must be at least 8 characters",
      "age": "must be 18 or older"
    }
  }
}
```

### Pagination

```go
func listUsers(w http.ResponseWriter, r *http.Request) {
    page := getPageFromQuery(r)     // e.g., 2
    pageSize := getPageSizeFromQuery(r) // e.g., 20

    users, total := getUsersFromDB(page, pageSize)

    response.Paginated(w, users, page, pageSize, total)
}
```

Response:
```json
{
  "success": true,
  "data": [
    {"id": "1", "name": "User 1"},
    {"id": "2", "name": "User 2"}
  ],
  "meta": {
    "pagination": {
      "page": 2,
      "page_size": 20,
      "total_pages": 5,
      "total_items": 95,
      "has_next": true,
      "has_prev": true
    }
  }
}
```

### Metadata

Include custom metadata in responses:

```go
func getWithMetadata(w http.ResponseWriter, r *http.Request) {
    data := getUserData()

    meta := &response.Meta{
        RequestID: r.Context().Value(logger.RequestIDKey).(string),
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    response.SuccessWithMeta(w, http.StatusOK, data, meta)
}
```

Response:
```json
{
  "success": true,
  "data": { "id": "123", "name": "John" },
  "meta": {
    "request_id": "abc123",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

## API Reference

### Success Functions

- `OK(w, data)` - 200 OK response
- `Created(w, data)` - 201 Created response
- `NoContent(w)` - 204 No Content response
- `Success(w, statusCode, data)` - Custom success response
- `SuccessWithMeta(w, statusCode, data, meta)` - Success with metadata

### Error Functions

- `BadRequest(w, message)` - 400 Bad Request
- `Unauthorized(w, message)` - 401 Unauthorized
- `Forbidden(w, message)` - 403 Forbidden
- `NotFound(w, resource)` - 404 Not Found
- `Conflict(w, message)` - 409 Conflict
- `InternalError(w, message)` - 500 Internal Server Error
- `ValidationError(w, details)` - 400 with validation details
- `Error(w, err)` - Custom error from `errors.Error`
- `ErrorWithMeta(w, err, meta)` - Error with metadata

### Pagination

- `Paginated(w, data, page, pageSize, totalItems)` - Paginated response

### Low-Level

- `JSON(w, statusCode, data)` - Write raw JSON response

## Types

### Response

```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorData  `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}
```

### ErrorData

```go
type ErrorData struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### Meta

```go
type Meta struct {
    RequestID  string      `json:"request_id,omitempty"`
    Timestamp  string      `json:"timestamp,omitempty"`
    Pagination *Pagination `json:"pagination,omitempty"`
}
```

### Pagination

```go
type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
    TotalItems int64 `json:"total_items"`
    HasNext    bool  `json:"has_next"`
    HasPrev    bool  `json:"has_prev"`
}
```

## Examples

### Complete CRUD API

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/vnykmshr/nivo/shared/errors"
    "github.com/vnykmshr/nivo/shared/response"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// List users with pagination
func listUsers(w http.ResponseWriter, r *http.Request) {
    page := 1
    pageSize := 20

    users, total := db.GetUsers(page, pageSize)
    response.Paginated(w, users, page, pageSize, total)
}

// Get single user
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")

    user := db.FindUser(id)
    if user == nil {
        response.NotFound(w, "user")
        return
    }

    response.OK(w, user)
}

// Create user
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    // Validate
    if validationErrors := validate(user); len(validationErrors) > 0 {
        response.ValidationError(w, validationErrors)
        return
    }

    // Check for duplicates
    if db.UserExists(user.Email) {
        response.Conflict(w, "email already exists")
        return
    }

    // Create
    created, err := db.CreateUser(user)
    if err != nil {
        response.InternalError(w, "failed to create user")
        return
    }

    response.Created(w, created)
}

// Update user
func updateUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")

    var updates User
    if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
        response.BadRequest(w, "invalid request body")
        return
    }

    user := db.FindUser(id)
    if user == nil {
        response.NotFound(w, "user")
        return
    }

    updated, err := db.UpdateUser(id, updates)
    if err != nil {
        response.InternalError(w, "failed to update user")
        return
    }

    response.OK(w, updated)
}

// Delete user
func deleteUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")

    if !db.UserExists(id) {
        response.NotFound(w, "user")
        return
    }

    if err := db.DeleteUser(id); err != nil {
        response.InternalError(w, "failed to delete user")
        return
    }

    response.NoContent(w)
}
```

### Error Handling with Context

```go
func transfer(w http.ResponseWriter, r *http.Request) {
    var req TransferRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Get account
    account := db.GetAccount(req.FromAccountID)
    if account == nil {
        response.NotFound(w, "account")
        return
    }

    // Check balance
    if account.Balance < req.Amount {
        err := errors.InsufficientFunds("insufficient balance")
        err.WithDetail("balance", account.Balance).
            WithDetail("required", req.Amount).
            WithDetail("shortfall", req.Amount - account.Balance)
        response.Error(w, err)
        return
    }

    // Check if account is frozen
    if account.Status == "frozen" {
        err := errors.AccountFrozen("account is frozen")
        err.WithDetail("account_id", account.ID).
            WithDetail("frozen_at", account.FrozenAt)
        response.Error(w, err)
        return
    }

    // Perform transfer
    result, err := db.Transfer(req)
    if err != nil {
        response.InternalError(w, "transfer failed")
        return
    }

    response.OK(w, result)
}
```

## Best Practices

1. **Always use response helpers** - Don't write raw JSON responses
2. **Use semantic HTTP status codes** - Match status to the actual result
3. **Include validation details** - Help clients understand what went wrong
4. **Add request IDs** - Include in metadata for request tracing
5. **Paginate large collections** - Always paginate list endpoints
6. **Document error codes** - Make error codes discoverable
7. **Be consistent** - Use the same response format across all endpoints
8. **Handle errors gracefully** - Return appropriate status codes and messages

## Integration with Middleware

Works seamlessly with `shared/middleware`:

```go
import (
    "github.com/vnykmshr/nivo/shared/logger"
    "github.com/vnykmshr/nivo/shared/middleware"
    "github.com/vnykmshr/nivo/shared/response"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Request ID is available from middleware
    requestID := r.Context().Value(logger.RequestIDKey)

    data := getData()

    meta := &response.Meta{
        RequestID: requestID.(string),
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }

    response.SuccessWithMeta(w, http.StatusOK, data, meta)
}

func main() {
    log := logger.NewDefault("api")

    app := middleware.Chain(
        handler,
        middleware.Recovery(log),
        middleware.RequestID(),
        middleware.Logging(log),
    )

    http.ListenAndServe(":8080", app)
}
```

## Testing

All response functions have 100% test coverage:

```bash
cd shared/response
go test -v
go test -cover
```

## Performance

Response functions are lightweight:

- JSON encoding: Standard library `encoding/json`
- No reflection overhead in hot path
- Minimal allocations for common cases
- ~1-2µs per response (excluding JSON marshaling)

## Related Packages

- [`shared/errors`](../errors/README.md) - Error types and codes
- [`shared/middleware`](../middleware/README.md) - HTTP middleware
- [`shared/logger`](../logger/README.md) - Structured logging

## License

Copyright © 2025 Nivo
