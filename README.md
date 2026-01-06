# Nivo

> A showcase neobank platform demonstrating production-ready fintech architecture

[![Documentation](https://img.shields.io/badge/docs-docs.nivomoney.com-blue)](https://docs.nivomoney.com)
[![Live Demo](https://img.shields.io/badge/demo-nivomoney.com-green)](https://nivomoney.com)

## Overview

**Nivo** is a portfolio-grade digital banking system built to demonstrate strong backend engineering, system design depth, and fintech domain understanding. It's a self-contained, simulated banking environment where users can open accounts, perform transfers, manage wallets, and interact with realistic fintech workflows—all in a safe, synthetic environment with no real money.

This is not a toy project. It's a carefully crafted demonstration of how modern digital banking systems can be architected, implemented, and explained.

## Vision

Create a lean, practical fintech system that showcases:
- **Real-world architecture** with Go microservices
- **Clean domain modeling** and system boundaries
- **Reliability and auditability** through event-sourcing and ledger patterns
- **Production-ready patterns** including CQRS, circuit breakers, and observability
- **Fintech domain expertise** with double-entry ledger, idempotency, and risk controls

## Key Features

### Core Banking
- 🏦 **Account Management** - Multiple account types, virtual balances, freeze/limits
- 💰 **Wallet & Ledger** - Double-entry bookkeeping with atomic balance updates
- 💸 **Transactions** - Internal transfers, scheduled payments, reversals
- 🛡️ **Risk & Controls** - Velocity checks, daily limits, fraud simulation

### Engineering Excellence
- 📊 **Observability** - Metrics, tracing, and monitoring with Prometheus/Grafana
- 🔄 **Event-Driven** - Async processing with durable job queues
- 🚦 **Resilience** - Circuit breakers, backpressure, dead-letter queues
- ✅ **Quality** - Comprehensive testing, load testing, CI/CD pipeline

## Tech Stack

- **Language**: Go
- **Database**: PostgreSQL
- **Cache**: Redis
- **Messaging**: NSQ/NATS with ledgerq for durable queuing
- **Workers**: goflow for pipeline orchestration
- **Resilience**: autobreaker for circuit breaking
- **Validation**: gopantic for input validation
- **Containers**: Docker + Docker Compose
- **Observability**: Prometheus + Grafana

## Architecture

Nivo follows a **modular monolith** approach evolving toward microservices:

```
├── services/          # Core banking services
│   ├── identity/      # User auth & profiles
│   ├── ledger/        # Double-entry engine
│   ├── wallet/        # Balance management
│   ├── transaction/   # Payment flows
│   └── risk/          # Fraud detection
├── gateway/           # API Gateway (unified access point)
├── shared/            # Common packages & utilities
├── scripts/           # Automation & deployment
└── docs/              # Documentation
```

**Design Patterns**: Event-driven communication, CQRS, idempotency, compensating transactions

## Quick Start

> ✅ **Status**: MVP READY - Production-ready for launch

### Prerequisites
- Go 1.23+
- Docker & Docker Compose
- Node.js 18+ (for frontend)

### Setup
```bash
# Clone the repository
git clone <repository-url>
cd nivo

# Start infrastructure (PostgreSQL, Redis, NSQ, Prometheus, Grafana)
docker-compose up -d

# Run database migrations and seed data
./scripts/seed-data.sh

# Start all services
make run-all

# Start frontend apps (in separate terminals)
cd frontend/user-app && npm install && npm run dev
cd frontend/admin-app && npm install && npm run dev
```

Detailed setup instructions: [quickstart.md](docs/quickstart.md)

### Documentation

Full documentation available at: **[docs.nivomoney.com](https://docs.nivomoney.com)**

- [Quick Start](https://docs.nivomoney.com/quickstart)
- [Development Guide](https://docs.nivomoney.com/development)
- [System Architecture](https://docs.nivomoney.com/architecture)
- [End-to-End Flows](https://docs.nivomoney.com/flows)

## Project Status

**MVP Release: READY** ✅

**Phase 1: Core Services** ✅ **COMPLETE**
- [x] Identity service (auth, KYC, user management)
- [x] Ledger service (double-entry bookkeeping)
- [x] Wallet service (balance management, limits)
- [x] Transaction service (transfer, deposit, withdrawal)
- [x] RBAC service (roles & permissions)
- [x] Notification service (email/SMS templates)
- [x] Risk service (fraud detection)
- [x] Gateway service (API gateway + SSE events)

**Phase 2: Frontend Apps** ✅ **COMPLETE**
- [x] User web app (12 pages, React + TypeScript)
- [x] Admin dashboard (5 pages, React + TypeScript)

**Phase 3: Production Features** ✅ **COMPLETE**
- [x] Database migrations (23+ migration files)
- [x] Seed data for development
- [x] Docker Compose infrastructure
- [x] Prometheus + Grafana monitoring
- [x] Comprehensive test coverage (23+ test files)
- [x] Security (JWT, RBAC, rate limiting, audit trails)

**Total Implementation:**
- **8 microservices** (77+ endpoints)
- **2 frontend apps** (17 pages)
- **23 database migrations**
- **12 shared libraries**
- **End-to-end integration tested**

See `todos/MVP_READINESS_REPORT.md` for detailed status and launch checklist.

## Key Engineering Concepts

This project demonstrates:
- **Idempotency keys** for safe retries
- **Atomicity guarantees** in distributed transactions
- **Compensating transactions** for reversal flows
- **Circuit breakers** to prevent cascading failures
- **Event sourcing** for audit trails
- **Rate limiting** and backpressure strategies

## Target Audience

- Engineering hiring managers
- Staff / Principal Engineers
- System design reviewers
- Open-source community
- Fintech engineers and enthusiasts

## Contributing

This is primarily a portfolio project, but suggestions and discussions are welcome! Feel free to open issues for questions or feedback.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Built with focus. Delivered with quality.**
