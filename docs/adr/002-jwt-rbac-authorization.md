---
layout: default
title: "ADR-002: JWT + RBAC Authorization"
parent: Architecture Decision Records
nav_order: 2
permalink: /adr/002-jwt-rbac-authorization
---

# ADR-002: JWT Authentication with RBAC Authorization

**Status**: Accepted
**Date**: 2024-01-15
**Decision Makers**: Engineering Team

## Context

Nivo is a microservices-based platform requiring:

1. User authentication across multiple services
2. Fine-grained access control (admin vs user vs support)
3. Stateless request handling for horizontal scaling
4. Service-to-service authentication for internal calls

## Decision

Implement a two-layer security model:

1. **JWT (JSON Web Tokens)** for stateless authentication
2. **RBAC (Role-Based Access Control)** for authorization

### Authentication: JWT

```
┌─────────┐      1. Login         ┌──────────────┐
│  User   │ ──────────────────────│   Identity   │
│  App    │                       │   Service    │
└────┬────┘                       └──────┬───────┘
     │                                   │
     │      2. JWT Token                 │
     │ <─────────────────────────────────┘
     │
     │      3. API Request + JWT
     │ ──────────────────────────────────┐
     │                                   │
     │                            ┌──────┴───────┐
     │                            │   Gateway    │
     │                            │  (validates) │
     │                            └──────────────┘
```

**JWT Payload**:
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "roles": ["user"],
  "permissions": ["wallet:read", "transaction:create"],
  "iat": 1705312200,
  "exp": 1705398600
}
```

**Token Handling**:
- Issued on login with 24-hour expiry
- Contains embedded roles and permissions
- Validated by each service independently (no central auth call)
- Refresh token rotation for extended sessions

### Authorization: RBAC

**Role Hierarchy**:
```
user (base)
└── support (inherits user)
    ├── accountant (+ ledger access)
    └── compliance_officer (+ KYC access)
        └── admin (+ user management)
            └── super_admin (+ RBAC management)
```

**Permission Format**: `service:resource:action`
```
identity:kyc:verify      → Can verify KYC documents
wallet:wallet:freeze     → Can freeze user wallets
transaction:transfer:create → Can initiate transfers
ledger:journal:reverse   → Can reverse journal entries
```

**Permission Check Flow**:
```go
// Middleware checks permission before handler
func RequirePermission(permission string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := GetClaimsFromContext(r.Context())

            if !hasPermission(claims.Permissions, permission) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Service-to-Service Communication

Internal endpoints skip JWT validation but use path-based access control:

```
/api/v1/*        → Requires JWT authentication
/internal/v1/*   → No JWT, restricted to internal network
```

## Alternatives Considered

### 1. Session-Based Authentication
Store sessions in Redis/database, validate on each request.

**Rejected because:**
- Requires shared session store across services
- Additional network hop per request
- Single point of failure
- Harder to scale horizontally

### 2. OAuth2/OIDC with External Provider
Use Auth0, Okta, or Keycloak.

**Rejected because:**
- External dependency and cost
- Portfolio project should demonstrate auth implementation
- More complex setup for demo purposes
- Can integrate later for production

### 3. API Keys
Issue API keys per user/application.

**Rejected because:**
- No user context in token
- Harder to revoke (need database lookup)
- Better suited for service accounts, not users

### 4. Attribute-Based Access Control (ABAC)
Dynamic policies based on user/resource attributes.

**Rejected because:**
- More complex than needed for our use case
- RBAC covers our permission requirements
- Can evolve to ABAC if needed

## Consequences

### Positive

- **Stateless**: No shared session store needed
- **Scalable**: Any service can validate tokens independently
- **Flexible**: RBAC allows granular permission control
- **Auditable**: JWT contains user context for logging
- **Standard**: Wide library support for JWT

### Negative

- **Token Size**: JWT with permissions can be large
- **Revocation**: Can't instantly revoke tokens (must wait for expiry)
- **Secret Management**: JWT secret must be shared across services

### Mitigations

- **Token Size**: Keep minimal claims, fetch full permissions on demand for admin UI
- **Revocation**: Short expiry (24h) + refresh token rotation + database session tracking for logout
- **Secret Management**: Use environment variables, consider vault for production

## Implementation Details

### JWT Secret Configuration
```go
// All services share the same JWT secret
jwtSecret := os.Getenv("JWT_SECRET")
```

### Token Validation Middleware
```go
func Auth(config AuthConfig) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            claims, err := validateToken(token, config.JWTSecret)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), claimsKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Permission Assignment on Login
```go
// Identity service fetches permissions from RBAC service during login
permissions, err := rbacClient.GetUserPermissions(userID)
if err != nil {
    return nil, err
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "sub":         userID,
    "email":       user.Email,
    "roles":       roles,
    "permissions": permissions,
    "exp":         time.Now().Add(24 * time.Hour).Unix(),
})
```

## Security Considerations

- Passwords hashed with bcrypt (cost 12)
- JWT signed with HS256 (symmetric, shared secret)
- Token stored client-side (httpOnly cookie or localStorage)
- HTTPS required for all endpoints
- Rate limiting on login attempts

## Related Decisions

- Identity Service owns authentication logic
- RBAC Service owns role/permission definitions
- Gateway validates JWT on all /api/v1/* routes
- Internal endpoints (/internal/v1/*) trust network isolation

## References

- [RFC 7519: JSON Web Token](https://tools.ietf.org/html/rfc7519)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [NIST RBAC Model](https://csrc.nist.gov/projects/role-based-access-control)
