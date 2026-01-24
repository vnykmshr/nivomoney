# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN GOTOOLCHAIN=auto go mod download

# Copy shared code
COPY shared/ ./shared/

# Copy transaction service code
COPY services/transaction/ ./services/transaction/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo -o bin/transaction-service ./services/transaction/cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && adduser -u 1000 -G appuser -s /bin/sh -D appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/transaction-service .

# Copy migrations
COPY --from=builder /app/services/transaction/migrations ./migrations

# Set ownership and switch to non-root user
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8084

# Run the service
CMD ["./transaction-service"]
