# Nivo

A production-grade neobank platform demonstrating fintech architecture with Go microservices.

[![Documentation](https://img.shields.io/badge/docs-docs.nivomoney.com-blue)](https://docs.nivomoney.com)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## Overview

Nivo is a portfolio project implementing a complete digital banking system. It demonstrates microservices architecture, double-entry accounting, and fintech domain patterns in a working, deployable application.

**What it includes:**
- 9 Go microservices with domain-driven boundaries
- Double-entry ledger with balanced journal entries
- JWT authentication with role-based access control
- React frontends for users and admins
- Full observability stack (Prometheus, Grafana)

## Live Demo

Try it at [nivomoney.com](https://nivomoney.com):

| Email | Password | Balance |
|-------|----------|---------|
| raj.kumar@gmail.com | raj123 | ₹50,000 |
| priya.electronics@business.com | priya123 | ₹1,50,000 |

Admin dashboard: [admin.nivomoney.com](https://admin.nivomoney.com) (`admin@nivo.local` / `admin123`)

All data is synthetic. No real money.

## Architecture

```
services/
├── identity/       # Auth, users, KYC
├── ledger/         # Double-entry accounting
├── wallet/         # Balance management
├── transaction/    # Transfers, payments
├── rbac/           # Roles & permissions
├── risk/           # Fraud detection
├── notification/   # Alerts, messaging
├── simulation/     # Test data generation
└── seed/           # Database seeding

gateway/            # API Gateway with SSE
frontend/
├── user-app/       # Customer React app
└── admin-app/      # Admin dashboard
```

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Node.js 18+

### Setup

```bash
git clone https://github.com/vnykmshr/nivomoney.git
cd nivomoney

# Start infrastructure
docker-compose up -d

# Seed database
./scripts/seed-data.sh

# Start services
make run-all

# Start frontend (separate terminal)
cd frontend/user-app && npm install && npm run dev
```

Open http://localhost:5173 and login with demo credentials.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Services | Go 1.24, Chi router |
| Database | PostgreSQL 15 |
| Cache | Redis |
| Auth | JWT, bcrypt |
| Frontend | React 18, TypeScript, Vite, TailwindCSS |
| Infrastructure | Docker Compose |
| Monitoring | Prometheus, Grafana |

## Key Patterns

- **Double-entry ledger** - Every transaction creates balanced debit/credit entries
- **Idempotency keys** - Safe retry handling for financial operations
- **RBAC** - Granular permissions with role hierarchies
- **Circuit breakers** - Fault isolation between services
- **Event-driven** - Async processing with durable queues

## Documentation

Full documentation: [docs.nivomoney.com](https://docs.nivomoney.com)

- [Quick Start](https://docs.nivomoney.com/quickstart)
- [Architecture](https://docs.nivomoney.com/architecture)
- [API Flows](https://docs.nivomoney.com/flows)
- [ADRs](https://docs.nivomoney.com/adr)

## Project Scope

| Category | Count |
|----------|-------|
| Microservices | 9 |
| API Endpoints | 77+ |
| Frontend Pages | 17 |
| Database Migrations | 23 |

This is a portfolio demonstration, not a production bank. It shows how a neobank *would* be built.

## License

MIT - see [LICENSE](LICENSE)
