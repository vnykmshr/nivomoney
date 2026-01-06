# RBAC Service

The RBAC (Role-Based Access Control) Service manages roles, permissions, and user assignments for Nivo. It implements hierarchical roles where child roles inherit all permissions from parent roles.

## Features

- **Hierarchical Roles**: Roles inherit permissions from parent roles
- **Fine-Grained Permissions**: Format `service:resource:action` (e.g., `identity:kyc:verify`)
- **User-Role Assignments**: Assign roles to users with optional expiry
- **Permission Checking**: Real-time single and batch permission validation
- **System Roles**: Pre-seeded roles (user, support, accountant, compliance_officer, admin, super_admin)

## API Endpoints

### Protected Endpoints (Requires Authentication)

All protected endpoints require an `Authorization` header with a Bearer token:
```http
Authorization: Bearer <jwt_token>
```

### Role Management

#### Create Role (Admin Only)
```http
POST /api/v1/roles
Content-Type: application/json

{
  "name": "analyst",
  "description": "Data analyst with read-only access to reports",
  "parent_role_id": "00000000-0000-0000-0000-000000000002"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "analyst",
    "description": "Data analyst with read-only access to reports",
    "parent_role_id": "00000000-0000-0000-0000-000000000002",
    "is_system": false,
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### List Roles
```http
GET /api/v1/roles
GET /api/v1/roles?active=true
```

#### Get Role
```http
GET /api/v1/roles/{id}
```

#### Get Role Hierarchy
Returns the role with its full parent chain.
```http
GET /api/v1/roles/{id}/hierarchy
```

#### Update Role (Admin Only)
```http
PUT /api/v1/roles/{id}
Content-Type: application/json

{
  "name": "senior_analyst",
  "description": "Updated description",
  "is_active": true
}
```

#### Delete Role (Admin Only)
```http
DELETE /api/v1/roles/{id}
```

### Permission Management

#### Create Permission (Admin Only)
```http
POST /api/v1/permissions
Content-Type: application/json

{
  "name": "reports:dashboard:view",
  "service": "reports",
  "resource": "dashboard",
  "action": "view",
  "description": "View analytics dashboard"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "name": "reports:dashboard:view",
    "service": "reports",
    "resource": "dashboard",
    "action": "view",
    "description": "View analytics dashboard",
    "is_system": false,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### List Permissions
```http
GET /api/v1/permissions
GET /api/v1/permissions?service=identity
```

#### Get Permission
```http
GET /api/v1/permissions/{id}
```

### Role-Permission Assignment

#### Assign Permission to Role (Admin Only)
```http
POST /api/v1/roles/{id}/permissions
Content-Type: application/json

{
  "permission_id": "660e8400-e29b-41d4-a716-446655440000"
}
```

#### Get Role Permissions
```http
GET /api/v1/roles/{id}/permissions
GET /api/v1/roles/{id}/permissions?inherited=true
```

The `inherited=true` parameter includes permissions from parent roles.

#### Remove Permission from Role (Admin Only)
```http
DELETE /api/v1/roles/{roleId}/permissions/{permissionId}
```

### User-Role Assignment

#### Assign Role to User (Admin Only)
```http
POST /api/v1/users/{userId}/roles
Content-Type: application/json

{
  "role_id": "00000000-0000-0000-0000-000000000003",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "770e8400-e29b-41d4-a716-446655440000",
    "role_id": "00000000-0000-0000-0000-000000000003",
    "assigned_at": "2024-01-15T10:30:00Z",
    "expires_at": "2025-12-31T23:59:59Z",
    "is_active": true,
    "role": {
      "id": "00000000-0000-0000-0000-000000000003",
      "name": "accountant",
      "description": "Financial operations and reporting"
    }
  }
}
```

#### Get User Roles
```http
GET /api/v1/users/{userId}/roles
```

#### Get User Permissions
Returns all permissions for a user (includes inherited via role hierarchy).
```http
GET /api/v1/users/{userId}/permissions
```

#### Remove Role from User (Admin Only)
```http
DELETE /api/v1/users/{userId}/roles/{roleId}
```

### Permission Checking

#### Check Single Permission
```http
POST /api/v1/check-permission
Content-Type: application/json

{
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "permission": "identity:kyc:verify"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "allowed": true,
    "roles": ["compliance_officer"],
    "reason": "Permission granted via role: compliance_officer"
  }
}
```

#### Check Multiple Permissions (Batch)
```http
POST /api/v1/check-permissions
Content-Type: application/json

{
  "user_id": "770e8400-e29b-41d4-a716-446655440000",
  "permissions": [
    "identity:kyc:verify",
    "wallet:wallet:freeze",
    "rbac:role:create"
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "results": {
      "identity:kyc:verify": true,
      "wallet:wallet:freeze": false,
      "rbac:role:create": false
    },
    "roles": ["compliance_officer", "support", "user"]
  }
}
```

### Internal Endpoints (Service-to-Service)

No authentication required. Used by other services internally.

#### Assign Default User Role
Called by Identity Service during user registration.
```http
POST /internal/v1/users/{userId}/assign-default-role
```

#### Get User Permissions (Internal)
Called by Identity Service during login/token generation.
```http
GET /internal/v1/users/{userId}/permissions
```

### Health Check
```http
GET /health
```

## Role Hierarchy

The system comes with pre-seeded roles in a hierarchy:

```
user (base)
└── support (inherits user)
    ├── accountant (inherits support)
    └── compliance_officer (inherits support)
        └── admin (inherits compliance_officer)
            └── super_admin (inherits admin)
```

| Role | ID | Description |
|------|----|-------------|
| user | `00000000-...-000001` | Regular user with basic permissions |
| support | `00000000-...-000002` | Customer support with read-only access |
| accountant | `00000000-...-000003` | Financial operations and reporting |
| compliance_officer | `00000000-...-000004` | KYC/AML verification and compliance |
| admin | `00000000-...-000005` | System administrator with elevated permissions |
| super_admin | `00000000-...-000006` | Super administrator with all permissions |

## Permission Format

Permissions follow the format: `service:resource:action`

Examples:
- `identity:kyc:verify` - Verify KYC documents
- `wallet:wallet:freeze` - Freeze a wallet
- `transaction:transfer:create` - Create transfers
- `ledger:journal:reverse` - Reverse journal entries

## Setup

### Prerequisites

- Go 1.23+
- PostgreSQL 14+

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_PORT` | Server port | 8082 |
| `DATABASE_PASSWORD` | PostgreSQL password | (required) |
| `JWT_SECRET` | JWT validation secret | (required) |
| `MIGRATIONS_DIR` | Migrations directory | ./migrations |

### Running the Service

```bash
cd services/rbac

# Run migrations and start
go run cmd/server/main.go
```

## Architecture

```
services/rbac/
├── cmd/
│   └── server/           # Server entry point
├── internal/
│   ├── handler/          # HTTP handlers
│   │   ├── rbac_handler.go
│   │   └── routes.go
│   ├── service/          # Business logic
│   │   └── rbac_service.go
│   ├── repository/       # Database operations
│   │   └── rbac_repository.go
│   └── models/           # Domain models
│       └── role.go
├── migrations/           # SQL migrations
└── README.md
```

## Access Control

| Endpoint Category | Required Role |
|-------------------|---------------|
| Role CRUD | admin, super_admin |
| Permission CRUD | admin, super_admin |
| User-Role Assignment | admin, super_admin |
| List/Read Operations | Any authenticated user |
| Permission Checks | Any authenticated user |

## Future Enhancements

- [ ] Permission caching with Redis
- [ ] Wildcard permissions (e.g., `wallet:*:read`)
- [ ] Role templates for quick setup
- [ ] Audit logging for role changes
- [ ] Time-based permission grants
