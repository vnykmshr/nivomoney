---
layout: default
title: Architecture Decision Records
nav_order: 10
has_children: true
permalink: /adr
---

# Architecture Decision Records

Architecture Decision Records (ADRs) document the key architectural decisions made during the development of Nivo, along with their context and consequences.

## What is an ADR?

An ADR captures a significant architectural decision along with:
- **Context**: The situation and requirements that led to the decision
- **Decision**: What was decided and why
- **Alternatives**: Other options that were considered
- **Consequences**: The trade-offs and implications

## ADR Index

| ID | Title | Status | Summary |
|:---|:------|:-------|:--------|
| [001](001-double-entry-ledger.md) | Double-Entry Ledger | Accepted | Implement double-entry bookkeeping for financial accuracy and audit trails |
| [002](002-jwt-rbac-authorization.md) | JWT + RBAC Authorization | Accepted | Stateless JWT authentication with role-based access control |
| [003](003-microservices-architecture.md) | Microservices Architecture | Accepted | Domain-driven service boundaries with shared database for MVP |

## ADR Template

For future decisions, use this template:

```markdown
# ADR-XXX: [Title]

**Status**: [Proposed | Accepted | Deprecated | Superseded]
**Date**: YYYY-MM-DD
**Decision Makers**: [Names/Roles]

## Context
[What is the issue that we're seeing that motivates this decision?]

## Decision
[What is the change that we're proposing and/or doing?]

## Alternatives Considered
[What other options were evaluated?]

## Consequences
[What becomes easier or harder as a result of this decision?]
```

## References

- [Michael Nygard's ADR Article](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR GitHub Organization](https://adr.github.io/)
