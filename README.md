# Nivo

> A showcase neobank platform demonstrating production-ready fintech architecture

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

> 🚧 **Status**: Under active development

### Prerequisites
- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### Setup
```bash
# Clone the repository
git clone https://github.com/vnykmshr/nivo.git
cd nivo

# Start infrastructure
make docker-up

# Run database migrations
make migrate-up

# Start services
make run

# Run tests
make test
```

Detailed setup instructions will be available as the project evolves.

## Project Status

**Phase 0: Foundation** ✅
- [x] Repository setup
- [x] Directory structure
- [x] Go module initialization
- [ ] Development tooling
- [ ] CI/CD pipeline

**Phase 1: Core Services** 🚧
- [ ] Identity service
- [ ] Ledger service
- [ ] Wallet service
- [ ] Transaction service

**Phase 2: Advanced Features** 📋
- [ ] Risk engine
- [ ] Scheduled payments
- [ ] Observability stack
- [ ] Web dashboard

See [todos/todos.md](todos/todos.md) for complete execution plan (dev-only, not in git).

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
