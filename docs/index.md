---
layout: default
title: Home
nav_order: 1
description: "Nivo - A production-ready neobank platform showcasing fintech engineering excellence"
permalink: /
---

# Nivo Documentation
{: .fs-9 }

A portfolio-grade neobank platform demonstrating production-ready microservices architecture with fintech domain expertise.
{: .fs-6 .fw-300 }

[Live Demo](https://nivomoney.com){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 }

---

## What is Nivo?

Nivo is a **showcase neobank platform** built to demonstrate engineering excellence in the fintech domain. It's not a production bank — it's a **portfolio project** that shows how a real neobank would be architected.

### Key Highlights

- **9 Microservices** — Identity, Ledger, Wallet, Transaction, Risk, RBAC, Notification, Simulation, Gateway
- **Double-Entry Ledger** — Proper accounting with balanced journal entries
- **JWT + RBAC** — Role-based access control with secure authentication
- **India-Centric** — INR currency, PAN/Aadhaar validation, IST timezone

---

## Quick Links

### Getting Started
{: .text-delta }

| Guide | Description |
|:------|:------------|
| [Why Nivo?](why-nivo) | The story and architecture philosophy |
| [Demo Walkthrough](demo) | Try Nivo with pre-configured demo accounts |
| [Quick Start](quickstart) | Get the platform running locally |
| [Development Guide](development) | Full development setup and workflow |
| [End-to-End Flows](flows) | User journeys and API sequences |

### Architecture
{: .text-delta }

| Document | Description |
|:---------|:------------|
| [System Architecture](architecture) | High-level system design and service overview |
| [Design System](design-system) | Frontend design patterns and components |
| [SSE Integration](sse) | Real-time updates with Server-Sent Events |
| [Observability](observability) | Prometheus metrics and Grafana dashboards |

### API Reference
{: .text-delta }

| Service | Port | Description |
|:--------|:-----|:------------|
| Gateway | 8000 | API entry point, routing, auth verification |
| Identity | 8080 | User registration, login, KYC, profiles |
| Ledger | 8081 | Double-entry accounting, journal entries |
| RBAC | 8082 | Roles, permissions, access control |
| Wallet | 8083 | User wallets, balances |
| Transaction | 8084 | Payments, transfers, UPI |
| Risk | 8085 | Fraud detection, limits |
| Notification | 8087 | Alerts, emails, SMS |
| Simulation | 8086 | Test data, demo scenarios |

---

## Tech Stack

### Backend

| Technology | Purpose |
|:-----------|:--------|
| **Go 1.24** | Microservices runtime |
| **PostgreSQL 15** | Primary database |
| **Chi Router** | HTTP routing |
| **JWT** | Authentication tokens |
| **golang-migrate** | Database migrations |

### Frontend

| Technology | Purpose |
|:-----------|:--------|
| **React 19** | UI framework |
| **TypeScript** | Type safety |
| **Vite** | Build tool |
| **TailwindCSS v4** | Styling |
| **React Router** | Navigation |

### Infrastructure

| Technology | Purpose |
|:-----------|:--------|
| **Docker Compose** | Local development |
| **Prometheus** | Metrics collection |
| **Grafana** | Dashboards |

---

## Demo Access

Try the live demo with pre-seeded accounts:

| App | URL | Description |
|:----|:----|:------------|
| **User App** | [nivomoney.com](https://nivomoney.com) | Customer banking experience |
| **Verify Portal** | [verify.nivomoney.com](https://verify.nivomoney.com) | OTP verification for paired users |
| **Admin Dashboard** | [admin.nivomoney.com](https://admin.nivomoney.com) | Operations & KYC management |

**User App** — [nivomoney.com](https://nivomoney.com)
```
Email: raj.kumar@gmail.com
Password: raj123
Balance: ₹50,000
```

**Verify Portal** — [verify.nivomoney.com](https://verify.nivomoney.com)
```
Email: priya.electronics@business.com
Password: priya123
```

**Admin Dashboard** — [admin.nivomoney.com](https://admin.nivomoney.com)
```
Email: admin@nivo.local
Password: admin123
```

See [Demo Walkthrough](/demo) for a guided tour of all features.

{: .note }
> This is a **portfolio demo**. No real money is involved. All data is synthetic.

---

## Project Structure

```
nivo/
├── services/           # Go microservices
│   ├── identity/      # Auth, users, KYC
│   ├── ledger/        # Double-entry accounting
│   ├── wallet/        # Wallet management
│   ├── transaction/   # Payment processing
│   ├── rbac/          # Access control
│   ├── risk/          # Fraud detection
│   ├── notification/  # Alerts & messaging
│   └── simulation/    # Test data generation
├── gateway/           # API Gateway
├── shared/            # Common packages
├── frontend/
│   ├── user-app/        # Customer-facing React app
│   ├── user-admin-app/  # Verification portal (OTP approval)
│   ├── admin-app/       # Admin dashboard
│   └── shared/          # Shared components
└── docs/             # This documentation
```

---

## What Makes This Impressive

This project demonstrates:

1. **Real microservices** — Not just split code, but proper domain boundaries
2. **Fintech expertise** — Double-entry ledger, KYC, RBAC, risk management
3. **Production patterns** — Idempotency, graceful shutdown, structured logging
4. **Clean architecture** — Consistent code style across all services
5. **Working demo** — End-to-end flows that actually work

---

## About

Nivo is a portfolio project showcasing fintech engineering capabilities including microservices architecture, double-entry ledger systems, and modern frontend development.

{: .fs-2 }
Nivo &copy; 2026 | [Live Demo](https://nivomoney.com) | [Why Nivo?](/why-nivo)
