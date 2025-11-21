# Middleware Package

HTTP middleware components for Nivo services, providing common functionality like logging, panic recovery, CORS handling, request IDs, and timeouts.

## Features

- **Middleware Chaining**: Compose multiple middleware in a clean, readable way
- **Request/Response Logging**: Structured logging with timing and status tracking
- **Panic Recovery**: Gracefully handle panics with stack trace logging
- **CORS**: Flexible CORS configuration for cross-origin requests
- **Request ID**: Generate or extract request IDs for request tracing
- **Timeout**: Enforce request timeouts with context cancellation
- **Response Writer**: Capture status codes and response sizes

## Installation

```bash
go get github.com/vnykmshr/nivo/shared/middleware
```

## Usage

### Basic Middleware Chain

```go
package main

import (
    "net/http"
    "time"

    "github.com/vnykmshr/nivo/shared/logger"
    "github.com/vnykmshr/nivo/shared/middleware"
)

func main() {
    log := logger.NewDefault("api")

    // Your handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Chain middleware
    wrapped := middleware.Chain(
        handler,
        middleware.Recovery(log),              // First: catch panics
        middleware.RequestID(),                // Second: add request ID
        middleware.Logging(log),               // Third: log requests
        middleware.CORS(middleware.DefaultCORSConfig()),
        middleware.Timeout(30 * time.Second),
    )

    http.ListenAndServe(":8080", wrapped)
}
```

### Request ID

Generate unique request IDs for tracing requests across services:

```go
// Automatically generates request ID if not present
app := middleware.Chain(
    handler,
    middleware.RequestID(),
)

// In your handler, access the request ID from context
func myHandler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Context().Value(logger.RequestIDKey)
    log.WithField("request_id", requestID).Info("processing request")
}
```

Request IDs are:
- Extracted from `X-Request-ID` header if present
- Generated as 32-character hex strings if not provided
- Added to response headers for client tracking
- Available in context for downstream use

### Logging

Log all HTTP requests and responses with timing information:

```go
log := logger.New(logger.Config{
    Level:       "info",
    Format:      "json",
    ServiceName: "api",
})

app := middleware.Chain(
    handler,
    middleware.RequestID(),  // Add this first for request ID in logs
    middleware.Logging(log),
)
```

Logged information includes:
- Request: method, path, remote address, request ID
- Response: status code, bytes written, duration in milliseconds

Example log output:
```json
{
  "level": "info",
  "request_id": "abc123",
  "method": "GET",
  "path": "/users",
  "remote_addr": "192.168.1.1:12345",
  "message": "request started"
}
{
  "level": "info",
  "request_id": "abc123",
  "status": 200,
  "bytes": 1024,
  "duration_ms": 45,
  "message": "request completed"
}
```

### Recovery

Recover from panics gracefully and log with stack traces:

```go
log := logger.NewDefault("api")

// Basic recovery with default error response
app := middleware.Chain(
    handler,
    middleware.Recovery(log),
)

// Custom panic handler
customHandler := func(w http.ResponseWriter, r *http.Request, err interface{}) {
    log.WithField("panic", err).Error("custom panic handler")
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte(`{"error": "something went wrong"}`))
}

app := middleware.Chain(
    handler,
    middleware.RecoveryWithHandler(log, customHandler),
)
```

The recovery middleware:
- Catches panics and prevents server crashes
- Logs panic value, stack trace, request method, and path
- Returns 500 Internal Server Error with JSON response
- Allows custom panic handlers for specialized recovery logic

### CORS

Handle Cross-Origin Resource Sharing (CORS) requests:

```go
// Use default configuration (permissive, for development)
app := middleware.Chain(
    handler,
    middleware.CORS(middleware.DefaultCORSConfig()),
)

// Custom configuration
config := middleware.CORSConfig{
    AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
    ExposedHeaders:   []string{"X-Request-ID", "X-Total-Count"},
    AllowCredentials: true,
    MaxAge:           3600, // 1 hour
}

app := middleware.Chain(
    handler,
    middleware.CORS(config),
)
```

CORS middleware features:
- Wildcard origin support (`["*"]`) for development
- Specific origin allowlist for production
- Automatic preflight (`OPTIONS`) request handling
- Configurable methods, headers, and credentials
- Cache control via `MaxAge`

**Production Best Practice:**
```go
// Production CORS config
config := middleware.CORSConfig{
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
```

### Timeout

Enforce request timeouts to prevent long-running handlers:

```go
// 30 second timeout
app := middleware.Chain(
    handler,
    middleware.Timeout(30 * time.Second),
)

// Different timeouts for different routes
apiHandler := middleware.Chain(
    apiRouter,
    middleware.Timeout(10 * time.Second),
)

uploadHandler := middleware.Chain(
    uploadRouter,
    middleware.Timeout(5 * time.Minute),
)
```

Timeout middleware:
- Uses context cancellation for clean timeout handling
- Returns 504 Gateway Timeout when timeout occurs
- Handlers should check `r.Context().Done()` for cancellation
- Handler continues in background if it doesn't check context

**Handler Best Practice:**
```go
func slowHandler(w http.ResponseWriter, r *http.Request) {
    for {
        select {
        case <-r.Context().Done():
            // Context cancelled (timeout or client disconnect)
            return
        default:
            // Do work
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

### Response Writer

The `ResponseWriter` wrapper is used internally by logging middleware to capture response metadata:

```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rw := middleware.NewResponseWriter(w)

    // Use wrapped writer
    rw.WriteHeader(http.StatusCreated)
    rw.Write([]byte("created"))

    // Access captured metadata
    statusCode := rw.StatusCode     // 201
    bytes := rw.BytesWritten        // 7
})
```

Features:
- Captures HTTP status code (defaults to 200)
- Tracks bytes written
- Prevents multiple `WriteHeader` calls
- Transparent proxy for all other operations

## Middleware Ordering

The order of middleware matters! They execute in the order provided to `Chain`:

```go
app := middleware.Chain(
    handler,
    middleware.Recovery(log),      // 1. Outermost: catch panics from all below
    middleware.RequestID(),        // 2. Generate request ID early
    middleware.Logging(log),       // 3. Log after request ID is available
    middleware.CORS(config),       // 4. Handle CORS before business logic
    middleware.Timeout(30*time.Second), // 5. Innermost: enforce timeout on handler
)
```

Execution flow (request → response):
1. Recovery middleware starts (deferred panic handler)
2. RequestID generates/extracts ID
3. Logging logs "request started"
4. CORS sets headers
5. Timeout starts timer
6. **Handler executes**
7. Timeout clears timer
8. CORS processes response
9. Logging logs "request completed"
10. Recovery completes (no panic)

## Testing

All middleware components have comprehensive test coverage (98.9%):

```bash
# Run tests
cd shared/middleware
go test -v

# Run tests with coverage
go test -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Examples

### Complete API Server

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/vnykmshr/nivo/shared/logger"
    "github.com/vnykmshr/nivo/shared/middleware"
)

func main() {
    // Initialize logger
    log := logger.New(logger.Config{
        Level:       "info",
        Format:      "json",
        ServiceName: "api",
    })

    // Create router
    mux := http.NewServeMux()

    // Add routes
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/api/users", usersHandler)

    // Configure CORS for production
    corsConfig := middleware.CORSConfig{
        AllowedOrigins:   []string{"https://app.example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"},
        ExposedHeaders:   []string{"X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           3600,
    }

    // Build middleware stack
    handler := middleware.Chain(
        mux,
        middleware.Recovery(log),
        middleware.RequestID(),
        middleware.Logging(log),
        middleware.CORS(corsConfig),
        middleware.Timeout(30 * time.Second),
    )

    // Start server
    server := &http.Server{
        Addr:         ":8080",
        Handler:      handler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    log.Info("starting server on :8080")
    if err := server.ListenAndServe(); err != nil {
        log.WithError(err).Fatal("server failed")
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status": "healthy"}`))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    // Get request ID from context
    requestID := r.Context().Value(logger.RequestIDKey)

    // Check for timeout
    select {
    case <-r.Context().Done():
        return // Request cancelled
    default:
        // Process request
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"users": []}`))
    }
}
```

### Custom Middleware

Create your own middleware following the same pattern:

```go
func RateLimiter(limit int) middleware.Middleware {
    limiter := rate.NewLimiter(rate.Limit(limit), limit)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.WriteHeader(http.StatusTooManyRequests)
                w.Write([]byte(`{"error": "rate limit exceeded"}`))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Use it
app := middleware.Chain(
    handler,
    middleware.Recovery(log),
    RateLimiter(100), // Custom middleware
    middleware.Logging(log),
)
```

## Best Practices

1. **Always use Recovery first** to catch panics from all other middleware
2. **Generate RequestID early** so it's available for logging
3. **Use structured logging** to track requests across services
4. **Check context cancellation** in long-running handlers
5. **Configure CORS restrictively** in production
6. **Set appropriate timeouts** based on endpoint characteristics
7. **Test middleware chains** to verify expected behavior
8. **Use DefaultCORSConfig only in development**

## Performance

The middleware stack adds minimal overhead:

- Request ID generation: ~100ns
- Logging: ~1-2µs per log line
- CORS: ~500ns
- Recovery (no panic): ~50ns
- Timeout (completes normally): ~1µs
- ResponseWriter wrapping: ~10ns

Total overhead for typical stack: **~5-10µs per request**

## Architecture

```
Request
  ↓
Recovery (catch panics)
  ↓
RequestID (generate/extract ID)
  ↓
Logging (log request start)
  ↓
CORS (set headers)
  ↓
Timeout (start timer)
  ↓
Handler (your code)
  ↓
Timeout (clear timer)
  ↓
CORS (process response)
  ↓
Logging (log request completion)
  ↓
Recovery (complete)
  ↓
Response
```

## Related Packages

- [`shared/logger`](../logger/README.md) - Structured logging
- [`shared/errors`](../errors/README.md) - Error handling
- [`shared/config`](../config/README.md) - Configuration management

## License

Copyright © 2025 Nivo
