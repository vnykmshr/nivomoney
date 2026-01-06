---
layout: default
title: "ADR-003: Microservices Architecture"
parent: Architecture Decision Records
nav_order: 3
---

# ADR-003: Microservices Architecture with Domain-Driven Boundaries

**Status**: Accepted
**Date**: 2024-01-15
**Decision Makers**: Engineering Team

## Context

Nivo needs an architecture that:

1. Demonstrates real-world fintech engineering patterns
2. Allows independent development and deployment of features
3. Provides clear separation of concerns
4. Scales appropriately for different workloads
5. Serves as a portfolio showcase of microservices done right

## Decision

Adopt a **microservices architecture** with services organized around **domain-driven design (DDD) bounded contexts**.

### Service Boundaries

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Gateway                              │
│                         (Routing, SSE)                           │
└───────────────┬─────────────────────────────────┬───────────────┘
                │                                 │
    ┌───────────┴───────────┐       ┌─────────────┴─────────────┐
    │   User Domain         │       │   Financial Domain         │
    │                       │       │                            │
    │  ┌─────────────────┐  │       │  ┌────────────────────┐   │
    │  │    Identity     │  │       │  │      Wallet        │   │
    │  │  (Auth, KYC)    │  │       │  │    (Balances)      │   │
    │  └─────────────────┘  │       │  └────────────────────┘   │
    │                       │       │                            │
    │  ┌─────────────────┐  │       │  ┌────────────────────┐   │
    │  │      RBAC       │  │       │  │   Transaction      │   │
    │  │  (Permissions)  │  │       │  │    (Transfers)     │   │
    │  └─────────────────┘  │       │  └────────────────────┘   │
    │                       │       │                            │
    └───────────────────────┘       │  ┌────────────────────┐   │
                                    │  │      Ledger        │   │
    ┌───────────────────────┐       │  │   (Bookkeeping)    │   │
    │   Support Domain      │       │  └────────────────────┘   │
    │                       │       │                            │
    │  ┌─────────────────┐  │       │  ┌────────────────────┐   │
    │  │      Risk       │  │       │  │     Notification   │   │
    │  │    (Fraud)      │  │       │  │     (Alerts)       │   │
    │  └─────────────────┘  │       │  └────────────────────┘   │
    │                       │       │                            │
    │  ┌─────────────────┐  │       └────────────────────────────┘
    │  │   Simulation    │  │
    │  │   (Demo Data)   │  │
    │  └─────────────────┘  │
    │                       │
    └───────────────────────┘
```

### Domain Ownership

| Service | Domain | Owns |
|:--------|:-------|:-----|
| **Identity** | User | Users, KYC, Sessions |
| **RBAC** | User | Roles, Permissions, Assignments |
| **Wallet** | Financial | Wallets, Balances, Beneficiaries |
| **Transaction** | Financial | Transfers, Deposits, Withdrawals |
| **Ledger** | Financial | Accounts, Journal Entries |
| **Risk** | Support | Rules, Risk Events |
| **Notification** | Support | Notifications, Templates |

### Communication Patterns

**Synchronous (HTTP)**:
- Client → Gateway → Service
- Service → Service (internal endpoints)

**Choreography (planned)**:
- Events published on significant state changes
- Services react to events they care about
- NSQ message queue for async processing

### Shared Database (MVP Trade-off)

```
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL Database                      │
├─────────────┬─────────────┬─────────────┬─────────────┬─────┤
│   users     │   wallets   │  accounts   │    roles    │ ... │
│   user_kyc  │transactions │journal_entry│ permissions │     │
│   sessions  │beneficiaries│ledger_lines │ user_roles  │     │
└─────────────┴─────────────┴─────────────┴─────────────┴─────┘
```

**Why shared database for MVP:**
- Simpler deployment and operations
- Easier transactions across services (not recommended but pragmatic)
- Reduces infrastructure complexity for demo

**Future evolution:**
- Per-service databases when needed
- Event-driven eventual consistency
- Saga pattern for distributed transactions

## Alternatives Considered

### 1. Monolithic Architecture
Single deployable unit with modular internal structure.

**Rejected because:**
- Doesn't demonstrate microservices patterns
- Less impressive for portfolio purposes
- Harder to show domain expertise separation
- Scaling is all-or-nothing

### 2. Serverless Functions
Deploy each endpoint as AWS Lambda / Cloud Functions.

**Rejected because:**
- Cold start latency issues
- Harder to share code between functions
- More complex local development
- Less transferable to traditional enterprise setups

### 3. Modular Monolith
Single deployment with strict module boundaries.

**Considered as alternative:**
- Good stepping stone to microservices
- Could have worked for MVP
- Chose microservices for portfolio impact

### 4. Service Mesh (Istio/Linkerd)
Advanced service-to-service communication layer.

**Rejected because:**
- Overkill for demo scale
- Adds significant complexity
- Document as future enhancement

## Consequences

### Positive

- **Clear Boundaries**: Each service owns its domain completely
- **Independent Deployment**: Can update Identity without touching Wallet
- **Technology Freedom**: Could use different languages per service
- **Scalability**: Scale busy services (Transaction) independently
- **Portfolio Impact**: Demonstrates real microservices expertise

### Negative

- **Operational Complexity**: More services to deploy and monitor
- **Network Latency**: Service-to-service calls add latency
- **Data Consistency**: Harder to maintain across services
- **Debugging**: Distributed tracing needed

### Mitigations

- **Complexity**: Docker Compose for local dev, clear service structure
- **Latency**: Internal endpoints on same network, keep call chains short
- **Consistency**: Document eventual consistency as future enhancement
- **Debugging**: Correlation IDs, structured logging, health endpoints

## Implementation Guidelines

### Service Structure
```
services/{name}/
├── cmd/
│   └── server/
│       └── main.go         # Entry point
├── internal/
│   ├── handler/            # HTTP handlers
│   ├── service/            # Business logic
│   ├── repository/         # Database access
│   └── models/             # Domain models
├── migrations/             # SQL migrations
├── Makefile
└── README.md
```

### Inter-Service Communication
```go
// Internal calls use http client with service discovery
client := &http.Client{Timeout: 5 * time.Second}
resp, err := client.Get("http://wallet-service:8083/internal/v1/wallets/" + id)
```

### Configuration
```go
// All services use shared config package
cfg, err := config.Load()
// Returns: DatabaseURL, ServicePort, JWTSecret, etc.
```

## Future Enhancements

1. **Per-Service Databases**: When consistency requirements diverge
2. **Message Queue**: NSQ for async notification delivery
3. **Service Mesh**: For advanced traffic management
4. **API Gateway Enhancement**: Rate limiting, circuit breakers
5. **Distributed Tracing**: Jaeger for request flow visualization

## Related Decisions

- [ADR-001: Double-Entry Ledger](001-double-entry-ledger.md) - Ledger service design
- [ADR-002: JWT + RBAC](002-jwt-rbac-authorization.md) - Authentication strategy

## References

- [Martin Fowler: Microservices](https://martinfowler.com/articles/microservices.html)
- [Sam Newman: Building Microservices](https://samnewman.io/books/building_microservices_2nd_edition/)
- [Chris Richardson: Microservices Patterns](https://microservices.io/patterns/)
