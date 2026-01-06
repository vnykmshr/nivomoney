---
layout: default
title: Development Guide
nav_order: 3
description: "Complete guide for developing on Nivo"
permalink: /development
---

# Development Guide

Complete guide for developing Nivo - the showcase neobank platform.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Development Tools](#development-tools)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Docker & Infrastructure](#docker--infrastructure)
- [Database Migrations](#database-migrations)
- [Debugging](#debugging)
- [Common Tasks](#common-tasks)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

- **Go**: 1.25 or later ([install](https://go.dev/dl/))
- **Docker**: Latest stable version ([install](https://docker.com))
- **Docker Compose**: v2.0+ (bundled with Docker Desktop)
- **Git**: Latest version
- **Make**: Standard make utility (pre-installed on macOS/Linux)

### Recommended Tools

- **golangci-lint**: For linting ([install](https://golangci-lint.run/usage/install/))
- **VS Code** with Go extension, or **GoLand**
- **Postman** or **Insomnia**: For API testing
- **TablePlus** or **pgAdmin**: For database inspection

---

## Initial Setup

### Option 1: Automated Setup (Recommended)

Run the automated setup script:

```bash
./scripts/dev-setup.sh
```

This script will:
- ✓ Check system prerequisites
- ✓ Install development tools (golangci-lint)
- ✓ Set up git hooks
- ✓ Create `.env` from template
- ✓ Start Docker containers
- ✓ Verify services are ready

### Option 2: Manual Setup

```bash
# 1. Clone repository
git clone <repository-url>
cd nivo

# 2. Install development tools
make install-tools

# 3. Set up git hooks
./scripts/install-hooks.sh

# 4. Create environment configuration
cp .env.example .env
# Edit .env with your configuration

# 5. Start infrastructure
make docker-up

# 6. Verify setup
make test
```

---

## Development Workflow

### Daily Workflow

```bash
# 1. Pull latest changes
git pull origin main

# 2. Start infrastructure (if not running)
make docker-up

# 3. Create feature branch
git checkout -b feature/your-feature-name

# 4. Make changes and test frequently
make test
make lint

# 5. Build and verify
make build

# 6. Commit changes (pre-commit hooks will run automatically)
git add .
git commit -m "feat: your descriptive message"

# 7. Push and create PR
git push origin feature/your-feature-name
```

### Branch Naming Convention

- `feature/` - New features
- `fix/` - Bug fixes
- `refactor/` - Code refactoring
- `docs/` - Documentation updates
- `test/` - Test additions/modifications
- `chore/` - Maintenance tasks

### Commit Message Convention

Follow conventional commits:

```
<type>: <description>

[optional body]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `docs`: Documentation
- `test`: Tests
- `chore`: Maintenance
- `perf`: Performance improvement

Examples:
```
feat: Add wallet creation endpoint
fix: Resolve race condition in ledger updates
refactor: Simplify transaction validation logic
docs: Update API documentation for identity service
```

---

## Project Structure

```
nivo/
├── services/              # Microservices
│   ├── identity/         # User authentication & profiles
│   ├── ledger/           # Double-entry bookkeeping
│   ├── wallet/           # Wallet management
│   ├── transaction/      # Payment flows
│   └── risk/             # Fraud detection
├── gateway/              # API Gateway
├── shared/               # Shared packages
│   ├── config/          # Configuration management
│   ├── logger/          # Structured logging
│   ├── database/        # DB utilities
│   ├── errors/          # Error handling
│   └── middleware/      # HTTP middleware
├── scripts/              # Automation scripts
│   ├── dev-setup.sh     # Development setup
│   ├── install-hooks.sh # Git hooks installer
│   └── hooks/           # Git hook scripts
├── docs/                 # Documentation
├── Makefile             # Development commands
├── docker-compose.yml   # Local infrastructure
└── .golangci.yml        # Linter configuration
```

---

## Development Tools

### Make Commands

View all available commands:
```bash
make help
```

Key commands:

**Build:**
```bash
make build              # Build all services
make build-all          # Build all services and gateway
```

**Test:**
```bash
make test               # Run all tests
make test-coverage      # Generate coverage report
make test-integration   # Run integration tests
```

**Code Quality:**
```bash
make lint               # Run all linters
make fmt                # Format code
make vet                # Run go vet
```

**Docker:**
```bash
make docker-up          # Start containers
make docker-down        # Stop containers
make docker-logs        # View container logs
```

**Cleanup:**
```bash
make clean              # Remove build artifacts
make clean-all          # Deep clean (including caches)
```

### Editor Configuration

#### VS Code

Recommended `settings.json`:
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "editor.formatOnSave": true,
  "go.useLanguageServer": true,
  "go.buildOnSave": "workspace"
}
```

Recommended extensions:
- Go (golang.go)
- Error Lens (usernamehw.errorlens)
- Docker (ms-azuretools.vscode-docker)

#### GoLand

- Enable "gofmt" on save
- Configure golangci-lint as external tool
- Enable "Optimize imports" on save

---

## Testing

### Test Organization

```
service_name/
├── handler_test.go       # HTTP handler tests
├── service_test.go       # Business logic tests
├── repository_test.go    # Database tests
└── integration_test.go   # Integration tests (tag: integration)
```

### Running Tests

```bash
# Unit tests only
make test

# With coverage
make test-coverage
open coverage.html

# Integration tests (requires Docker)
make test-integration

# Specific package
go test ./services/identity/...

# Specific test
go test -run TestCreateWallet ./services/wallet/...

# Verbose output
go test -v ./...

# Race detection
go test -race ./...
```

### Writing Tests

**Example unit test:**
```go
func TestCreateUser(t *testing.T) {
    // Arrange
    repo := &mockUserRepository{}
    service := NewUserService(repo)

    // Act
    user, err := service.CreateUser(ctx, "test@example.com")

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

**Example integration test:**
```go
//go:build integration
package integration_test

func TestCreateUserIntegration(t *testing.T) {
    // Use real database
    db := setupTestDB(t)
    defer db.Close()

    // Test against real dependencies
    // ...
}
```

---

## Code Quality

### Pre-commit Hooks

Git hooks automatically run before each commit:
- ✓ Go formatting check (gofmt)
- ✓ Go vet
- ✓ golangci-lint
- ✓ Secret detection
- ✓ Build check

To bypass hooks (use sparingly):
```bash
git commit --no-verify
```

### Linting

Configuration: `.golangci.yml`

```bash
# Run all linters
make lint

# Run specific linter
golangci-lint run --enable=errcheck ./...

# Auto-fix issues
golangci-lint run --fix ./...
```

### Code Formatting

```bash
# Format all code
make fmt

# Format specific package
gofmt -s -w ./services/identity/

# Check formatting
gofmt -l .
```

### Code Review Checklist

- [ ] Code follows Go best practices
- [ ] Tests added/updated
- [ ] Error handling is appropriate
- [ ] Logging is adequate
- [ ] Comments explain "why", not "what"
- [ ] No sensitive data in code
- [ ] Database queries are safe (no SQL injection)
- [ ] API endpoints are documented

---

## Docker & Infrastructure

### Services

```bash
# Start all services
make docker-up

# Start specific service
docker-compose up -d postgres

# Stop all services
make docker-down

# View logs
make docker-logs

# View specific service logs
docker logs -f nivo-postgres

# Restart service
docker-compose restart redis
```

### Service Access

| Service       | URL/Port           | Credentials             |
|---------------|-------------------|-------------------------|
| PostgreSQL    | localhost:5432    | nivo / nivo_dev_password |
| Redis         | localhost:6379    | nivo_redis_password     |
| NSQ Admin     | localhost:4171    | -                       |
| Prometheus    | localhost:9090    | -                       |

### Database Access

```bash
# Connect to PostgreSQL
docker exec -it nivo-postgres psql -U nivo -d nivo

# Connect to Redis
docker exec -it nivo-redis redis-cli -a nivo_redis_password

# View database
psql -h localhost -U nivo -d nivo
```

---

## Database Migrations

> **Status**: Migration system will be configured in Phase 3

Planned approach:
- Tool: golang-migrate or goose
- Migrations in: `services/{service}/migrations/`
- Commands: `make migrate-up`, `make migrate-down`

---

## Debugging

### Application Debugging

**VS Code launch.json:**
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch Gateway",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/gateway",
      "env": {
        "LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### Log Levels

Set in `.env`:
```
LOG_LEVEL=debug   # Verbose output
LOG_LEVEL=info    # Standard output
LOG_LEVEL=warn    # Warnings and errors only
LOG_LEVEL=error   # Errors only
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Live profiling (add pprof import to service)
# Access: http://localhost:6060/debug/pprof/
```

---

## Common Tasks

### Adding a New Service

```bash
# 1. Create service directory
mkdir -p services/myservice

# 2. Initialize with main.go
cat > services/myservice/main.go << 'EOF'
package main

func main() {
    // Service implementation
}
EOF

# 3. Add dependencies
cd services/myservice
go mod tidy

# 4. Build and test
make build
make test
```

### Adding a New Endpoint

1. Define handler in `handler.go`
2. Add route in `routes.go`
3. Write tests in `handler_test.go`
4. Update API documentation
5. Test manually with Postman/curl

### Running Load Tests

```bash
# Using lobster (to be implemented)
make load-test

# Using hey
hey -n 10000 -c 100 http://localhost:8080/health
```

---

## Troubleshooting

### Docker Issues

**Problem**: Containers won't start
```bash
# Check Docker is running
docker info

# View container logs
docker-compose logs

# Restart Docker daemon
# macOS: Restart Docker Desktop
# Linux: sudo systemctl restart docker

# Clean and restart
make docker-down
docker system prune -f
make docker-up
```

**Problem**: Port already in use
```bash
# Find process using port
lsof -i :5432

# Kill process
kill -9 <PID>

# Or change port in .env
```

### Build Issues

**Problem**: Build fails with dependency errors
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy

# Verify go.mod
go mod verify
```

**Problem**: Lint errors after pull
```bash
# Update linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run with fresh cache
golangci-lint cache clean
make lint
```

### Test Issues

**Problem**: Integration tests fail
```bash
# Ensure Docker is running
docker ps

# Restart test database
docker-compose restart postgres

# Clear test data
# (Run cleanup script when available)
```

### Git Hook Issues

**Problem**: Pre-commit hook fails
```bash
# Check hook is installed
ls -la .git/hooks/pre-commit

# Reinstall hooks
./scripts/install-hooks.sh

# Bypass temporarily (use sparingly)
git commit --no-verify
```

---

## Getting Help

- **Documentation**: Check `/docs` directory
- **Issues**: Open GitHub issue with reproduction steps
- **Code Review**: Tag team members in PR

---

## Related Documentation

- [README.md](../README.md) - Project overview
- [Architecture Documentation](ARCHITECTURE.md) - System design (to be created)
- [API Documentation](API.md) - API reference (to be created)

---

**Last Updated**: Phase 1 (Development Tooling Complete)
