#!/bin/bash
# Nivo Development Environment Setup
# Run this script to set up your local development environment

set -e

echo "🚀 Nivo - Development Environment Setup"
echo "========================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${GREEN}✓${NC} $1"
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

check_command() {
    if command -v $1 > /dev/null 2>&1; then
        info "$1 is installed"
        return 0
    else
        warn "$1 is not installed"
        return 1
    fi
}

echo "Step 1: Checking system prerequisites..."
echo "----------------------------------------"

# Check Go
if check_command go; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "   Version: $GO_VERSION"
else
    error "Go is required but not installed. Install from https://go.dev/dl/"
fi

# Check Docker
if check_command docker; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo "   Version: $DOCKER_VERSION"
else
    error "Docker is required but not installed. Install from https://docker.com"
fi

# Check Docker Compose
if check_command docker-compose; then
    COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | sed 's/,//')
    echo "   Version: $COMPOSE_VERSION"
else
    warn "docker-compose not found, checking for docker compose plugin..."
    if docker compose version > /dev/null 2>&1; then
        info "Docker Compose plugin is available"
    else
        error "Docker Compose is required. Install from https://docs.docker.com/compose/install/"
    fi
fi

# Check git
if check_command git; then
    GIT_VERSION=$(git --version | awk '{print $3}')
    echo "   Version: $GIT_VERSION"
else
    error "Git is required but not installed"
fi

echo ""
echo "Step 2: Installing development tools..."
echo "----------------------------------------"

# Install golangci-lint if not present
if ! check_command golangci-lint; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    info "golangci-lint installed"
else
    GOLANGCI_VERSION=$(golangci-lint --version | head -1)
    echo "   $GOLANGCI_VERSION"
fi

echo ""
echo "Step 3: Setting up git hooks..."
echo "----------------------------------------"

if [ -f "./scripts/install-hooks.sh" ]; then
    ./scripts/install-hooks.sh
else
    warn "Git hooks installation script not found, skipping..."
fi

echo ""
echo "Step 4: Setting up environment configuration..."
echo "----------------------------------------"

if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        cp .env.example .env
        info "Created .env from .env.example"
        warn "Please review and update .env with your configuration"
    else
        warn ".env.example not found, skipping .env creation"
    fi
else
    info ".env already exists"
fi

echo ""
echo "Step 5: Starting Docker containers..."
echo "----------------------------------------"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    error "Docker daemon is not running. Please start Docker and try again."
fi

echo "Starting infrastructure containers..."
if docker-compose up -d > /dev/null 2>&1 || docker compose up -d > /dev/null 2>&1; then
    info "Docker containers started"
    echo ""
    echo "   Running containers:"
    docker ps --format "   - {{.Names}} ({{.Status}})" | grep nivo || true
else
    error "Failed to start Docker containers"
fi

echo ""
echo "Step 6: Waiting for services to be ready..."
echo "----------------------------------------"

echo "Waiting for PostgreSQL..."
MAX_RETRIES=30
RETRY_COUNT=0
until docker exec nivo-postgres pg_isready -U nivo > /dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT+1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        error "PostgreSQL failed to start in time"
    fi
    echo -n "."
    sleep 1
done
info "PostgreSQL is ready"

echo "Waiting for Redis..."
RETRY_COUNT=0
until docker exec nivo-redis redis-cli -a nivo_redis_password ping > /dev/null 2>&1; do
    RETRY_COUNT=$((RETRY_COUNT+1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        error "Redis failed to start in time"
    fi
    echo -n "."
    sleep 1
done
info "Redis is ready"

echo ""
echo "Step 7: Running database migrations..."
echo "----------------------------------------"
warn "Migration system not yet implemented (to be added in Phase 3)"

echo ""
echo "Step 8: Verifying setup..."
echo "----------------------------------------"

# Run go mod download
echo "Downloading Go dependencies..."
if go mod download; then
    info "Go dependencies downloaded"
else
    warn "Failed to download Go dependencies"
fi

# Try to build
echo "Running test build..."
if go build ./...; then
    info "Test build successful"
    rm -rf bin/ # Clean up test build artifacts
else
    warn "Test build failed (expected if no services implemented yet)"
fi

echo ""
echo "=========================================="
echo "✅ Development environment setup complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "  1. Review .env and update configuration if needed"
echo "  2. Start working on services: cd services/"
echo "  3. Run tests: make test"
echo "  4. View running containers: docker ps"
echo "  5. View logs: make docker-logs"
echo ""
echo "Useful commands:"
echo "  make help              - Show all available commands"
echo "  make docker-up         - Start Docker containers"
echo "  make docker-down       - Stop Docker containers"
echo "  make build             - Build all services"
echo "  make test              - Run tests"
echo "  make lint              - Run linters"
echo ""
echo "Access points:"
echo "  PostgreSQL:   localhost:5432 (user: nivo, db: nivo)"
echo "  Redis:        localhost:6379 (password in .env)"
echo "  NSQ Admin:    http://localhost:4171"
echo "  Prometheus:   http://localhost:9090"
echo ""
echo "Happy coding! 🎉"
