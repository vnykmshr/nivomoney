# RBAC Service

Role-Based Access Control (RBAC) service for Nivo platform with hierarchical role support.

## Features

- **Hierarchical Roles**: Roles can inherit permissions from parent roles
- **Fine-Grained Permissions**: Format `service:resource:action` (e.g., `identity:kyc:verify`)
- **User-Role Assignments**: Assign roles to users with optional expiry
- **Permission Checking**: Real-time and batch permission validation
- **System Roles**: Pre-seeded roles (user, support, accountant, compliance_officer, admin, super_admin)

## API Endpoints

### Roles
- `POST /api/v1/roles` - Create role
- `GET /api/v1/roles` - List all roles
- `GET /api/v1/roles/{id}` - Get role details
- `PUT /api/v1/roles/{id}` - Update role
- `DELETE /api/v1/roles/{id}` - Delete role
- `GET /api/v1/roles/{id}/hierarchy` - Get role with parent chain

### Permissions
- `POST /api/v1/permissions` - Create permission
- `GET /api/v1/permissions` - List all permissions
- `GET /api/v1/permissions/{id}` - Get permission details

### Role-Permission Assignment
- `POST /api/v1/roles/{id}/permissions` - Assign permission to role
- `GET /api/v1/roles/{id}/permissions` - Get role permissions
- `DELETE /api/v1/roles/{roleId}/permissions/{permissionId}` - Remove permission

### User-Role Assignment
- `POST /api/v1/users/{userId}/roles` - Assign role to user
- `GET /api/v1/users/{userId}/roles` - Get user roles
- `GET /api/v1/users/{userId}/permissions` - Get all user permissions (includes hierarchy)
- `DELETE /api/v1/users/{userId}/roles/{roleId}` - Remove role from user

### Permission Checks
- `POST /api/v1/check-permission` - Check single permission
- `POST /api/v1/check-permissions` - Batch check permissions

## Role Hierarchy

```
user (base)
├── support (inherits user)
    ├── accountant (inherits support + user)
    ├── compliance_officer (inherits support + user)
        ├── admin (inherits compliance_officer + support + user)
            ├── super_admin (inherits all)
```

## Environment Variables

- `SERVICE_PORT` - Server port (default: 8082)
- `DATABASE_URL` - PostgreSQL connection string
- `MIGRATIONS_DIR` - Migrations directory path
- `ENVIRONMENT` - Environment (development/production)

## Local Development

```bash
# Run migrations
make migrate-up

# Build
make build

# Run
make run

# Test
make test
```

## Docker

```bash
# Build image
docker build -t nivo-rbac-service -f services/rbac/Dockerfile .

# Run container
docker run -p 8082:8082 nivo-rbac-service
```
