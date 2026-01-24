# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN GOTOOLCHAIN=auto go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/simulation/ ./services/simulation/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo -o bin/simulation ./services/simulation/cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && adduser -u 1000 -G appuser -s /bin/sh -D appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/simulation .

# Set ownership and switch to non-root user
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8086

# Run the application
CMD ["./simulation"]
