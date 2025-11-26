# Frontend Security Architecture

This document details the security measures implemented in the Nivo Money frontend applications.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Security Features](#security-features)
4. [CSRF Protection](#csrf-protection)
5. [Audit Logging](#audit-logging)
6. [Session Management](#session-management)
7. [Content Security Policy](#content-security-policy)
8. [API Security](#api-security)
9. [Best Practices](#best-practices)
10. [Testing Security](#testing-security)

---

## Overview

The Nivo Money frontend consists of two separate applications with different security requirements:

- **User App** (`user-app`): Public-facing application for end users
- **Admin App** (`admin-app`): Internal application for administrators with stricter security controls

Both applications share security utilities from the `@nivo/shared` package, ensuring consistent security implementation.

## Architecture

```
frontend/
├── shared/                 # Shared security utilities
│   └── src/lib/
│       ├── security.ts    # Security configurations and utilities
│       └── apiClient.ts   # Base API client with CSRF protection
├── user-app/              # Public user application
│   ├── src/lib/
│   │   └── api.ts         # User API client with CSRF
│   └── src/stores/
│       └── authStore.ts   # User authentication with CSRF management
└── admin-app/             # Admin application (stricter security)
    ├── src/lib/
    │   ├── adminApi.ts    # Admin API client with CSRF
    │   └── auditLogger.ts # Admin audit logging
    └── src/stores/
        └── adminAuthStore.ts # Admin auth with CSRF and audit logging
```

## Security Features

### Environment-Based Configuration

Security settings vary by environment (development vs production):

**User App - Development:**
- Session timeout: 24 hours
- CSRF protection: Disabled
- Audit logging: Disabled
- IP whitelisting: Disabled

**User App - Production:**
- Session timeout: 24 hours
- CSRF protection: **Enabled**
- Audit logging: Enabled
- IP whitelisting: Disabled (public app)

**Admin App - Development:**
- Session timeout: 2 hours
- CSRF protection: Disabled
- Audit logging: **Always enabled**
- IP whitelisting: Disabled

**Admin App - Production:**
- Session timeout: 2 hours
- CSRF protection: **Enabled**
- Audit logging: **Always enabled**
- IP whitelisting: **Enabled** (VPN/office only)

### Configuration Access

```typescript
import { getSecurityConfig } from '@nivo/shared';

// Get user app config
const userConfig = getSecurityConfig(false, 'production');

// Get admin app config (stricter)
const adminConfig = getSecurityConfig(true, 'production');
```

---

## CSRF Protection

### What is CSRF?

Cross-Site Request Forgery (CSRF) is an attack that forces users to execute unwanted actions on a web application where they're authenticated.

### Implementation

CSRF protection is implemented at multiple layers:

#### 1. CSRFProtection Class

Located in `shared/src/lib/security.ts`:

```typescript
import { CSRFProtection } from '@nivo/shared';

const csrf = new CSRFProtection('X-CSRF-Token');

// Initialize token (call on app start)
csrf.initialize();

// Get token
const token = csrf.getToken();

// Regenerate token (call on login)
const newToken = csrf.generateToken();
csrf.setToken(newToken);

// Clear token (call on logout)
csrf.clear();
```

#### 2. API Client Integration

CSRF tokens are automatically added to all non-GET requests:

**BaseApiClient** (used by admin-app):

```typescript
// In shared/src/lib/apiClient.ts
class BaseApiClient {
  constructor(config: ApiClientConfig) {
    // CSRF enabled based on config
    if (config.csrfEnabled) {
      this.csrf = new CSRFProtection();
      this.csrf.initialize();
    }
  }

  // Token automatically added to POST, PUT, DELETE, PATCH requests
  private setupInterceptors() {
    this.client.interceptors.request.use((config) => {
      if (this.csrf && !['get', 'head', 'options'].includes(config.method)) {
        config.headers['X-CSRF-Token'] = this.csrf.getToken();
      }
      return config;
    });
  }
}
```

**User App ApiClient**:

```typescript
// In user-app/src/lib/api.ts
class ApiClient {
  constructor() {
    // Initialize CSRF in production
    if (import.meta.env.MODE === 'production') {
      this.csrf = new CSRFProtection();
      this.csrf.initialize();
    }

    // Token added to all non-GET requests
    this.client.interceptors.request.use((config) => {
      if (this.csrf && !['get', 'head', 'options'].includes(config.method)) {
        config.headers['X-CSRF-Token'] = this.csrf.getToken();
      }
      return config;
    });
  }
}
```

#### 3. Auth Store Integration

CSRF tokens are managed during authentication lifecycle:

**User App**:

```typescript
// Login - regenerate token
login: async (identifier, password) => {
  const response = await api.login({ identifier, password });
  api.regenerateCSRF(); // Fresh token on login
  // ... set state
}

// Logout - clear token
logout: () => {
  api.clearCSRF(); // Remove token
  // ... clear state
}
```

**Admin App**:

```typescript
// Same pattern with adminApi
login: async (identifier, password) => {
  const response = await adminApi.login({ identifier, password });
  adminApi.regenerateCSRF();
  logAdminAction.login(response.user.id, response.user.full_name);
  // ... set state
}

logout: async () => {
  logAdminAction.logout(user.id, user.full_name);
  await adminApi.logout();
  adminApi.clearCSRF();
  // ... clear state
}
```

### Backend Requirements

The backend must:

1. **Accept CSRF token header**: `X-CSRF-Token`
2. **Validate token**: For POST, PUT, DELETE, PATCH requests
3. **Return 403 Forbidden**: When token is missing or invalid
4. **Implement token storage**: Use session or signed cookies

Example middleware (pseudo-code):

```go
func CSRFMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Skip GET, HEAD, OPTIONS
    if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
      next.ServeHTTP(w, r)
      return
    }

    // Get token from header
    token := r.Header.Get("X-CSRF-Token")

    // Validate token
    if !validateCSRFToken(r.Context(), token) {
      http.Error(w, "Invalid CSRF token", http.StatusForbidden)
      return
    }

    next.ServeHTTP(w, r)
  })
}
```

---

## Audit Logging

### Purpose

Audit logging tracks all administrative actions for:
- Compliance requirements
- Security investigations
- User accountability
- Debugging

### Implementation

#### AuditLogger Class

Located in `shared/src/lib/security.ts`:

```typescript
import { AuditLogger } from '@nivo/shared';

const logger = new AuditLogger(true, 'production');

logger.log({
  userId: 'admin-123',
  userName: 'Admin User',
  action: 'VERIFY_KYC',
  resource: 'user:user-456',
  details: {
    targetUserId: 'user-456',
    decision: 'approved',
  },
});
```

#### Admin-Specific Logger

Located in `admin-app/src/lib/auditLogger.ts`:

```typescript
import { logAdminAction } from '../lib/auditLogger';

// Login
logAdminAction.login(userId, userName);

// Logout
logAdminAction.logout(userId, userName);

// KYC Verification
logAdminAction.verifyKYC(adminId, adminName, targetUserId, targetUserName);

// KYC Rejection
logAdminAction.rejectKYC(adminId, adminName, targetUserId, targetUserName, reason);

// View Dashboard
logAdminAction.viewDashboard(adminId, adminName);

// Search User
logAdminAction.searchUser(adminId, adminName, searchQuery);

// View User Details
logAdminAction.viewUserDetails(adminId, adminName, targetUserId);

// Suspend User
logAdminAction.suspendUser(adminId, adminName, targetUserId, reason);

// View Transaction
logAdminAction.viewTransaction(adminId, adminName, transactionId);
```

### Log Format

```typescript
interface AuditLog {
  timestamp: string;        // ISO 8601 format
  userId: string;           // Admin user ID
  userName: string;         // Admin user name
  action: string;           // Action performed (e.g., 'VERIFY_KYC')
  resource: string;         // Resource affected (e.g., 'user:123')
  details?: Record<string, unknown>; // Additional context
  ipAddress?: string;       // Client IP (if available)
  userAgent?: string;       // Client user agent (if available)
}
```

### Behavior by Environment

**Development:**
- Logs to browser console: `[AUDIT] { ... }`
- Stored in memory (last 1000 logs)
- Does not send to backend

**Production:**
- Logs to browser console
- Stored in memory (last 1000 logs)
- **Sends to backend** via API (TODO: implement endpoint)

### Viewing Logs

```typescript
import { auditLogger } from '../lib/auditLogger';

// Get all logs
const logs = auditLogger.getLogs();

// Clear logs (use sparingly)
auditLogger.clearLogs();
```

### Backend Requirements

The backend should implement:

```
POST /api/v1/admin/audit-log
Content-Type: application/json
Authorization: Bearer <admin-token>

{
  "timestamp": "2025-11-26T10:30:00.000Z",
  "userId": "admin-123",
  "userName": "Admin User",
  "action": "VERIFY_KYC",
  "resource": "user:user-456",
  "details": {
    "targetUserId": "user-456",
    "targetUserName": "John Doe",
    "decision": "approved"
  }
}
```

---

## Session Management

### User App

- **Timeout**: 24 hours of inactivity
- **Check interval**: Every 5 minutes
- **Storage**: `localStorage` (auth_token)
- **Activity tracking**: Mouse, keyboard, click, scroll events

### Admin App

- **Timeout**: 2 hours of inactivity (**stricter**)
- **Check interval**: Every 1 minute
- **Storage**: `localStorage` (admin_token)
- **Activity tracking**: Mouse, keyboard, click, scroll events
- **Forced logout**: Alert shown, redirect to login

### Implementation

Located in `admin-app/src/stores/adminAuthStore.ts`:

```typescript
// Session timeout configuration
const ADMIN_SESSION_TIMEOUT = 2 * 60 * 60 * 1000; // 2 hours

// Activity tracking
window.addEventListener('mousemove', updateActivity);
window.addEventListener('keypress', updateActivity);
window.addEventListener('click', updateActivity);
window.addEventListener('scroll', updateActivity);

// Periodic check (every minute)
setInterval(() => {
  const store = useAdminAuthStore.getState();
  if (store.isAuthenticated && store.checkSessionTimeout()) {
    store.logout();
    alert('Session expired due to inactivity. Please login again.');
    window.location.href = '/login';
  }
}, 60000);
```

---

## Content Security Policy

### Purpose

Content Security Policy (CSP) prevents:
- Cross-Site Scripting (XSS) attacks
- Data injection attacks
- Clickjacking
- Mixed content vulnerabilities

### Configuration

CSP is configured in:
1. Vite development/preview servers
2. Production nginx configuration

### Development CSP

Located in `vite.config.ts`:

```typescript
server: {
  headers: {
    'X-Frame-Options': 'DENY',
    'X-Content-Type-Options': 'nosniff',
    'Referrer-Policy': 'strict-origin-when-cross-origin',
  },
}
```

### Preview CSP

Located in `vite.config.ts`:

```typescript
preview: {
  headers: {
    'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
    'X-Frame-Options': 'DENY',
    'X-Content-Type-Options': 'nosniff',
    'Referrer-Policy': 'strict-origin-when-cross-origin',
    'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' http://localhost:* ws://localhost:*",
    'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
  },
}
```

### Production CSP (Nginx)

See `DEPLOYMENT_SECURITY.md` for full nginx configuration.

**User App:**
```
Content-Security-Policy: default-src 'self';
  script-src 'self' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data: https:;
  font-src 'self' data:;
  connect-src 'self' https://api.nivomoney.com
```

**Admin App (Stricter):**
```
Content-Security-Policy: default-src 'self';
  script-src 'self';
  style-src 'self' 'unsafe-inline';
  img-src 'self' data:;
  font-src 'self' data:;
  connect-src 'self' https://api.nivomoney.com;
  frame-ancestors 'none'
```

### CSP Violation Reporting

To monitor CSP violations:

```nginx
add_header Content-Security-Policy "default-src 'self'; ...; report-uri https://api.nivomoney.com/csp-report" always;
```

Backend endpoint to receive reports:

```
POST /api/v1/csp-report
Content-Type: application/csp-report

{
  "csp-report": {
    "document-uri": "https://app.nivomoney.com/",
    "violated-directive": "script-src",
    "blocked-uri": "https://evil.com/malicious.js",
    ...
  }
}
```

---

## API Security

### Authentication

Both apps use JWT Bearer tokens:

```typescript
// User App
Authorization: Bearer <auth_token>

// Admin App
Authorization: Bearer <admin_token>
```

### Token Storage

- **User App**: `localStorage.getItem('auth_token')`
- **Admin App**: `localStorage.getItem('admin_token')`

### Automatic Token Handling

Axios interceptors automatically:
1. Add `Authorization` header to all requests
2. Handle 401 Unauthorized responses
3. Clear tokens and redirect to login

### HTTPS Only

All API requests should use HTTPS in production:

```typescript
const API_BASE_URL = import.meta.env.VITE_API_URL || 'https://api.nivomoney.com';
```

### Rate Limiting

Client-side rate limiting is available:

```typescript
import { ClientRateLimiter } from '@nivo/shared';

const limiter = new ClientRateLimiter();

// Check if request is allowed
if (!limiter.isAllowed('login', 5, 60000)) {
  // Max 5 login attempts per minute
  throw new Error('Too many requests. Please try again later.');
}

// Make API call
await api.login(identifier, password);
```

---

## Best Practices

### For Developers

1. **Never disable security features in production**
   ```typescript
   // ❌ BAD
   csrfEnabled: false,

   // ✅ GOOD
   csrfEnabled: import.meta.env.MODE === 'production',
   ```

2. **Always use the shared security utilities**
   ```typescript
   // ✅ GOOD
   import { CSRFProtection, AuditLogger } from '@nivo/shared';
   ```

3. **Log all admin actions**
   ```typescript
   // ✅ GOOD
   await adminApi.verifyKYC(userId);
   logAdminAction.verifyKYC(adminId, adminName, userId, userName);
   ```

4. **Clear sensitive data on logout**
   ```typescript
   // ✅ GOOD
   logout: () => {
     api.clearCSRF();
     localStorage.removeItem('auth_token');
     // Clear any other sensitive state
   }
   ```

5. **Validate and sanitize user input**
   ```typescript
   import { sanitizeHTML, sanitizeURL } from '@nivo/shared';

   const safeHTML = sanitizeHTML(userInput);
   const safeURL = sanitizeURL(userProvidedURL);
   ```

6. **Use environment variables for sensitive config**
   ```typescript
   // ✅ GOOD
   const API_URL = import.meta.env.VITE_API_URL;

   // ❌ BAD - hardcoded URLs
   const API_URL = 'https://api.nivomoney.com';
   ```

### For Admins

1. **Always access admin app via VPN in production**
2. **Use strong, unique passwords**
3. **Enable 2FA when available (future enhancement)**
4. **Review audit logs regularly**
5. **Report suspicious activity immediately**
6. **Never share admin credentials**

---

## Testing Security

### Manual Testing

#### 1. CSRF Protection

Test that CSRF tokens are required:

```bash
# Without CSRF token (should fail)
curl -X POST https://api.nivomoney.com/api/v1/admin/kyc/verify \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123"}'

# Expected: 403 Forbidden

# With CSRF token (should succeed)
curl -X POST https://api.nivomoney.com/api/v1/admin/kyc/verify \
  -H "Authorization: Bearer <token>" \
  -H "X-CSRF-Token: <csrf-token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "123"}'

# Expected: 200 OK
```

#### 2. Session Timeout

Test admin session timeout:

1. Login to admin app
2. Wait 2 hours without activity
3. Try to perform an action
4. **Expected**: Session expired, forced logout

#### 3. IP Whitelisting

Test admin IP restrictions:

1. Access admin app from allowed IP
   - **Expected**: Access granted
2. Access admin app from blocked IP
   - **Expected**: 403 Forbidden

#### 4. Security Headers

Test that security headers are present:

```bash
curl -I https://admin.nivomoney.com

# Expected headers:
# Strict-Transport-Security: max-age=31536000; includeSubDomains
# X-Frame-Options: DENY
# X-Content-Type-Options: nosniff
# Content-Security-Policy: ...
```

#### 5. Audit Logging

Test that admin actions are logged:

1. Login to admin app (check browser console)
   - **Expected**: `[AUDIT] { action: 'LOGIN', ... }`
2. Approve a KYC
   - **Expected**: `[AUDIT] { action: 'VERIFY_KYC', ... }`
3. Reject a KYC
   - **Expected**: `[AUDIT] { action: 'REJECT_KYC', ... }`

### Automated Testing

#### Unit Tests

Test CSRF token generation:

```typescript
import { CSRFProtection } from '@nivo/shared';

describe('CSRFProtection', () => {
  it('should generate unique tokens', () => {
    const csrf = new CSRFProtection();
    const token1 = csrf.generateToken();
    const token2 = csrf.generateToken();

    expect(token1).not.toBe(token2);
    expect(token1).toHaveLength(64);
  });

  it('should store and retrieve token', () => {
    const csrf = new CSRFProtection();
    const token = csrf.generateToken();
    csrf.setToken(token);

    expect(csrf.getToken()).toBe(token);
  });

  it('should clear token', () => {
    const csrf = new CSRFProtection();
    csrf.setToken('test-token');
    csrf.clear();

    expect(csrf.getToken()).toBeNull();
  });
});
```

#### Integration Tests

Test API client CSRF integration:

```typescript
import { api } from './lib/api';

describe('API Client CSRF', () => {
  it('should add CSRF token to POST requests', async () => {
    // Mock axios
    const mockPost = jest.spyOn(api, 'post');

    // Make POST request
    await api.login({ identifier: 'test', password: 'test' });

    // Verify CSRF header was added
    expect(mockPost).toHaveBeenCalledWith(
      expect.any(String),
      expect.any(Object),
      expect.objectContaining({
        headers: expect.objectContaining({
          'X-CSRF-Token': expect.any(String),
        }),
      })
    );
  });
});
```

### Security Scanning

Use automated security scanners:

1. **OWASP ZAP**: Penetration testing
2. **Snyk**: Dependency vulnerability scanning
3. **Lighthouse**: Security audit in Chrome DevTools
4. **npm audit**: Check for vulnerable dependencies

```bash
# Run npm audit
npm audit

# Fix vulnerabilities
npm audit fix
```

---

## Incident Response

### Security Breach Detected

If you detect a security breach:

1. **Immediately**: Disable affected accounts
2. **Revoke**: All active sessions
3. **Notify**: Users and administrators
4. **Review**: Audit logs for suspicious activity
5. **Patch**: Any identified vulnerabilities
6. **Document**: Incident details and response

### Suspicious Activity

If audit logs show suspicious activity:

1. **Investigate**: Review full audit trail
2. **Contact**: Affected user/admin
3. **Verify**: Legitimate vs. malicious activity
4. **Act**: Suspend account if confirmed malicious
5. **Report**: To security team

### CSRF Token Issues

If CSRF tokens are causing issues:

1. **Check**: Environment (production only)
2. **Verify**: Token is being sent in requests
3. **Confirm**: Backend is validating correctly
4. **Review**: Browser console for errors
5. **Clear**: Session storage and retry

---

## Future Enhancements

Planned security improvements:

1. **Two-Factor Authentication (2FA)**
   - TOTP-based 2FA for admin accounts
   - SMS/Email backup codes

2. **Backend Audit Log Storage**
   - Implement `/api/v1/admin/audit-log` endpoint
   - Store logs in database with retention policy
   - Searchable audit log viewer in admin app

3. **Advanced Rate Limiting**
   - Backend rate limiting (per IP, per user)
   - Distributed rate limiting (Redis)
   - Adaptive rate limiting based on behavior

4. **Enhanced CSP**
   - Nonce-based CSP (remove `'unsafe-inline'`)
   - Subresource Integrity (SRI) for CDN assets
   - Stricter CSP in admin app

5. **Security Monitoring**
   - Real-time alerting on suspicious activity
   - Anomaly detection in audit logs
   - Automated threat response

6. **Certificate Pinning**
   - Pin API certificate in mobile apps
   - Prevent MITM attacks

---

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [Content Security Policy Reference](https://content-security-policy.com/)
- [MDN Web Security](https://developer.mozilla.org/en-US/docs/Web/Security)

---

**Last Updated**: Phase 7 - Security Hardening
**Maintained By**: Frontend Team
**Questions**: See `DEPLOYMENT_SECURITY.md` for deployment-specific security
