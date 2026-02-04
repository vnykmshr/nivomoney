# =============================================================================
# Nivo - Makefile
# =============================================================================
#
# Development and deployment commands for Nivo banking platform.
#
# Usage:
#   make help      - Show all available commands
#   make dev       - Start development environment
#   make deploy    - Deploy to production
#
# =============================================================================

.PHONY: help dev build down logs deploy obs-up obs-down secrets-edit secrets-view \
        db-shell db-backup db-restore seed test clean clean-all ssl-init ssl-renew \
        run-all run-identity run-ledger fmt vet lint install-lint

# Default target
help:
	@echo "Nivo - Development & Deployment Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev              Start development stack (Docker + exposed ports)"
	@echo "  make build            Build all Docker images"
	@echo "  make down             Stop all containers"
	@echo "  make logs             Tail all container logs"
	@echo "  make logs-SERVICE     Tail specific service logs (e.g., make logs-gateway)"
	@echo ""
	@echo "Local Services (without Docker):"
	@echo "  make run-all          Run all services locally (requires docker-up first)"
	@echo "  make run-identity     Run Identity service locally"
	@echo "  make run-ledger       Run Ledger service locally"
	@echo ""
	@echo "Database:"
	@echo "  make db-shell         Open PostgreSQL shell"
	@echo "  make db-backup        Create database backup"
	@echo "  make db-restore       Restore from backup (BACKUP_FILE=path)"
	@echo "  make seed             Run database seed"
	@echo "  make seed-clean       Run clean seed (reset data)"
	@echo ""
	@echo "Observability:"
	@echo "  make obs-up           Start with Prometheus + Grafana"
	@echo "  make obs-down         Stop observability stack"
	@echo ""
	@echo "Secrets (SOPS + age):"
	@echo "  make secrets-edit     Edit encrypted secrets"
	@echo "  make secrets-view     View decrypted secrets"
	@echo "  make secrets-encrypt  Encrypt a file (SRC=path)"
	@echo ""
	@echo "SSL Certificates:"
	@echo "  make ssl-init         Initial SSL certificate setup"
	@echo "  make ssl-renew        Renew SSL certificates"
	@echo ""
	@echo "Deployment:"
	@echo "  make deploy           Deploy to production (1GB VPS optimized)"
	@echo "  make setup-server     Run server setup script (requires sudo)"
	@echo ""
	@echo "Code Quality:"
	@echo "  make test             Run all tests"
	@echo "  make test-SERVICE     Run tests for specific service"
	@echo "  make lint             Run linters (requires golangci-lint)"
	@echo "  make fmt              Format code with gofmt"
	@echo "  make vet              Run go vet"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean            Remove build artifacts"
	@echo "  make clean-all        Remove all generated files"

# =============================================================================
# Development
# =============================================================================

dev:
	@echo "Starting development environment..."
	@docker compose up -d
	@echo ""
	@echo "Services running:"
	@echo "  Gateway:  http://localhost:8000"
	@echo "  Postgres: localhost:5432"
	@echo "  Redis:    localhost:6379"
	@echo ""
	@echo "Frontend apps (run separately):"
	@echo "  cd frontend/user-app && npm run dev"
	@echo "  cd frontend/admin-app && npm run dev"
	@echo ""
	@docker compose ps

build:
	@echo "Building all Docker images..."
	@docker compose build

down:
	@docker compose down

logs:
	@docker compose logs -f

logs-%:
	@docker compose logs -f $*

# =============================================================================
# Local Service Development (without Docker for services)
# =============================================================================

run-all:
	@echo "Run services with: make run-identity, make run-ledger, etc."
	@echo "Requires: make dev (for postgres/redis)"

run-identity:
	@echo "Running Identity Service..."
	@go run ./services/identity/cmd/server/main.go

run-ledger:
	@echo "Running Ledger Service..."
	@go run ./services/ledger/cmd/server/main.go

# =============================================================================
# Database
# =============================================================================

db-shell:
	@docker compose exec postgres psql -U nivo nivo

db-backup:
	@mkdir -p backups
	@BACKUP_FILE="backups/nivo_$$(date +%Y%m%d_%H%M%S).sql"; \
	docker compose exec -T postgres pg_dump -U nivo nivo > "$$BACKUP_FILE"; \
	echo "Backup created: $$BACKUP_FILE"

db-restore:
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Usage: make db-restore BACKUP_FILE=path/to/backup.sql"; \
		exit 1; \
	fi
	@echo "Restoring from $(BACKUP_FILE)..."
	@docker compose exec -T postgres psql -U nivo nivo < $(BACKUP_FILE)
	@echo "Restore complete"

seed:
	@echo "Running database seed..."
	@go run services/seed/cmd/server/main.go

seed-clean:
	@echo "Running clean seed (reset data)..."
	@go run services/seed/cmd/server/main.go --clean

# =============================================================================
# Observability (included in base docker-compose.yml)
# =============================================================================

obs-up: dev
	@echo ""
	@echo "Monitoring available at:"
	@echo "  Grafana:    http://localhost:3000"
	@echo "  Prometheus: http://localhost:9090"

obs-down: down

# =============================================================================
# Secrets Management (SOPS + age)
# =============================================================================

secrets-edit:
	@echo "Editing encrypted secrets..."
	@sops .env.enc

secrets-view:
	@sops --decrypt --input-type dotenv --output-type dotenv .env.enc

secrets-encrypt:
	@if [ -z "$(SRC)" ]; then \
		echo "Usage: make secrets-encrypt SRC=path/to/.env.prod"; \
		exit 1; \
	fi
	@sops --encrypt --input-type dotenv --output-type dotenv --output .env.enc $(SRC)
	@echo "Encrypted to .env.enc"

# =============================================================================
# SSL Certificates
# =============================================================================

ssl-init:
	@echo "Initializing SSL certificates..."
	@mkdir -p certbot/conf certbot/www
	@docker compose run --rm certbot certonly --webroot \
		--webroot-path=/var/www/certbot \
		-d nivomoney.com \
		-d www.nivomoney.com \
		-d admin.nivomoney.com \
		-d api.nivomoney.com \
		-d grafana.nivomoney.com \
		--email admin@nivomoney.com \
		--agree-tos \
		--no-eff-email
	@echo "SSL certificates obtained. Restart nginx: docker compose restart frontend"

ssl-renew:
	@docker compose run --rm certbot renew
	@docker compose restart frontend

# =============================================================================
# Deployment
# =============================================================================

deploy:
	@./scripts/deploy.sh

setup-server:
	@echo "Running server setup script..."
	@sudo ./scripts/setup-server.sh

# =============================================================================
# Code Quality
# =============================================================================

test:
	@echo "Running all tests..."
	@go test -v -race ./...

test-%:
	@echo "Running tests for $*..."
	@go test -v -race ./services/$*/...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: make install-lint"; \
		exit 1; \
	fi

fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Code formatted"

vet:
	@echo "Running go vet..."
	@go vet ./...

install-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# =============================================================================
# Build (Go binaries)
# =============================================================================

build-go:
	@echo "Building Go binaries..."
	@mkdir -p bin
	@go build -o bin/identity-service ./services/identity/cmd/server
	@go build -o bin/ledger-service ./services/ledger/cmd/server
	@go build -o bin/rbac-service ./services/rbac/cmd/server
	@go build -o bin/wallet-service ./services/wallet/cmd/server
	@go build -o bin/transaction-service ./services/transaction/cmd/server
	@go build -o bin/risk-service ./services/risk/cmd/server
	@go build -o bin/notification-service ./services/notification/cmd/server
	@go build -o bin/simulation-service ./services/simulation/cmd/server
	@go build -o bin/gateway ./gateway/cmd/server
	@echo "Build complete: ./bin/"

# =============================================================================
# Cleanup
# =============================================================================

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

clean-all: clean
	@echo "Removing all generated files..."
	@rm -rf vendor/
	@go clean -cache -testcache
	@echo "Deep clean complete"

clean-docker:
	@echo "Removing Docker resources..."
	@docker compose down -v
	@docker system prune -f
	@echo "Docker cleanup complete"
