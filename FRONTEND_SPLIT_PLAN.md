# Frontend Application Split - Implementation Plan
**Date**: 2025-11-26
**Objective**: Separate user banking app and admin operations into distinct applications
**Principle**: Security first, rock solid, complete not half-done

---

## Executive Summary

**Current State**: Single React app mixing user banking features with admin operations
**Target State**: Two separate, production-ready applications with shared utilities
**Approach**: Thoughtful separation with proper architecture, not just lift-and-shift

---

## Architecture Vision

### User Banking App (`/frontend/user-app`)
**Persona**: End consumers (bank customers)
**Purpose**: Personal finance management
**Access**: Public internet, mobile-first
**Security**: Standard user authentication

### Admin Operations App (`/frontend/admin-app`)
**Persona**: Internal staff (compliance officers, support, admins)
**Purpose**: Operations, compliance, user management
**Access**: Internal network / VPN (production)
**Security**: Role-based access, audit logging, stricter controls

### Shared Package (`/frontend/shared`)
**Purpose**: DRY principle - common utilities
**Scope**: Types, API utilities, validators, constants
**Not Shared**: Components, routes, business logic

---

## Detailed Comparison

| Aspect | User App | Admin App |
|--------|----------|-----------|
| **URL** | app.nivomoney.com | admin.nivomoney.com |
| **Port (Dev)** | 3000 | 3001 |
| **Build Output** | /dist/user | /dist/admin |
| **Bundle Size** | Small (~200KB) | Larger OK (~500KB) |
| **Target Devices** | Mobile + Desktop | Desktop only |
| **Auth Required** | Optional (landing page) | All routes |
| **Roles** | `user` | `admin`, `compliance`, `support` |
| **Session Timeout** | 24 hours | 2 hours |
| **Rate Limiting** | Standard | Strict |
| **Features** | Banking operations | Admin operations |
| **Deployment** | CDN, global | Single region, VPN |

---

## Security Architecture

### User App Security

**Authentication**:
- Public routes: `/`, `/login`, `/register`
- Protected routes: `/dashboard`, `/send`, `/deposit`, `/withdraw`, `/kyc`
- JWT with 24h expiry
- Remember me option
- Password reset flow

**Authorization**:
- Role: `user` (automatically assigned on registration)
- Permissions: `wallet:create`, `transaction:create`, etc.
- No admin permissions

**Security Controls**:
- Rate limiting: 100 req/min per user
- CSRF protection
- XSS sanitization
- Input validation (triple layer)
- HTTPS only (production)

**Data Protection**:
- No PII display without masking
- Transaction amounts masked in lists
- Sensitive fields encrypted in transit

### Admin App Security

**Authentication**:
- NO public routes (all require authentication)
- Redirect to login if not authenticated
- Admin credentials only
- JWT with 2h expiry (stricter)
- No "remember me"
- MFA required (future)

**Authorization**:
- Roles: `admin`, `compliance`, `support`
- Fine-grained permissions:
  - `admin:users:read`, `admin:users:update`
  - `admin:kyc:verify`, `admin:kyc:reject`
  - `admin:transactions:read`, `admin:reports:generate`
- Permission check on every route

**Security Controls**:
- Rate limiting: 50 req/min per admin
- IP whitelisting (production)
- Audit logging: All admin actions logged
- Session timeout: Auto-logout after 2h
- HTTPS + VPN only (production)

**Data Protection**:
- Full PII access (with audit trail)
- Sensitive data encrypted at rest
- Audit log: Who accessed what, when
- Data export controls

---

## Application Structure

### User App (`/frontend/user-app`)

```
user-app/
├── public/
│   ├── favicon.ico
│   └── index.html
├── src/
│   ├── assets/
│   │   └── logo.svg
│   ├── components/
│   │   ├── WalletCard.tsx
│   │   ├── TransactionList.tsx
│   │   └── ProtectedRoute.tsx
│   ├── hooks/
│   │   ├── useSSE.ts
│   │   └── useAuth.ts
│   ├── lib/
│   │   ├── api.ts           # User API client
│   │   └── utils.ts
│   ├── pages/
│   │   ├── LandingPage.tsx
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Dashboard.tsx
│   │   ├── SendMoney.tsx
│   │   ├── Deposit.tsx
│   │   ├── Withdraw.tsx
│   │   └── KYC.tsx
│   ├── stores/
│   │   ├── authStore.ts
│   │   └── walletStore.ts
│   ├── types/
│   │   └── index.ts         # User-specific types
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── README.md
```

**Key Features**:
- Landing page with value proposition
- User registration/login
- KYC submission form
- Wallet creation and management
- Send money, deposit, withdraw
- Transaction history
- Real-time updates (SSE)

**NOT Included**:
- ❌ Admin routes
- ❌ Admin components
- ❌ User management features
- ❌ KYC review features
- ❌ Admin statistics

### Admin App (`/frontend/admin-app`)

```
admin-app/
├── public/
│   ├── favicon-admin.ico
│   └── index.html
├── src/
│   ├── assets/
│   │   └── logo-admin.svg
│   ├── components/
│   │   ├── AdminLayout.tsx
│   │   ├── Sidebar.tsx
│   │   ├── NotificationPanel.tsx
│   │   ├── StatCard.tsx
│   │   ├── DataTable.tsx       # Reusable table
│   │   └── AdminRoute.tsx      # Admin-only route guard
│   ├── hooks/
│   │   ├── useAdminAuth.ts
│   │   └── useAuditLog.ts
│   ├── lib/
│   │   ├── adminApi.ts         # Admin API client
│   │   ├── permissions.ts
│   │   └── utils.ts
│   ├── pages/
│   │   ├── AdminLogin.tsx      # Admin-specific login
│   │   ├── Dashboard/
│   │   │   ├── index.tsx
│   │   │   ├── NotificationsTab.tsx
│   │   │   ├── UsersTab.tsx
│   │   │   ├── TransactionsTab.tsx
│   │   │   └── StatsTab.tsx
│   │   ├── KYC/
│   │   │   ├── KYCReview.tsx
│   │   │   ├── KYCList.tsx
│   │   │   └── KYCDetails.tsx
│   │   ├── Users/
│   │   │   ├── UserList.tsx
│   │   │   ├── UserDetails.tsx
│   │   │   └── UserEdit.tsx
│   │   ├── Transactions/
│   │   │   ├── TransactionList.tsx
│   │   │   ├── TransactionDetails.tsx
│   │   │   └── TransactionSearch.tsx
│   │   ├── Reports/
│   │   │   ├── ComplianceReports.tsx
│   │   │   └── FinancialReports.tsx
│   │   └── Settings/
│   │       ├── AdminUsers.tsx
│   │       ├── Permissions.tsx
│   │       └── AuditLog.tsx
│   ├── stores/
│   │   ├── adminAuthStore.ts
│   │   └── adminDataStore.ts
│   ├── types/
│   │   └── admin.ts            # Admin-specific types
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── package.json
├── vite.config.ts
├── tailwind.config.js
└── README.md
```

**Key Features**:
- Admin dashboard with notifications
- KYC review and approval
- User management (search, view, edit, suspend)
- Transaction monitoring and search
- Compliance reports
- System statistics
- Audit log viewer
- Admin user management

**NOT Included**:
- ❌ User banking features
- ❌ Public landing page
- ❌ User registration
- ❌ Money transfer features

### Shared Package (`/frontend/shared`)

```
shared/
├── src/
│   ├── types/
│   │   ├── user.ts           # User, KYCInfo
│   │   ├── wallet.ts         # Wallet
│   │   ├── transaction.ts    # Transaction
│   │   ├── api.ts            # API response types
│   │   └── index.ts
│   ├── lib/
│   │   ├── apiClient.ts      # Base API client class
│   │   ├── validation.ts     # Shared validators
│   │   ├── formatters.ts     # Currency, date formatters
│   │   └── constants.ts      # API URLs, etc.
│   ├── utils/
│   │   ├── phone.ts          # Phone normalization
│   │   ├── currency.ts       # Currency utils
│   │   └── date.ts           # Date utils
│   └── index.ts
├── package.json
├── tsconfig.json
└── README.md
```

**What's Shared**:
- ✅ Type definitions (User, Wallet, Transaction)
- ✅ API client base class
- ✅ Validation utilities
- ✅ Formatters (currency, date, phone)
- ✅ Constants (API URLs)

**NOT Shared**:
- ❌ React components (different UX)
- ❌ Stores (different state management)
- ❌ Routes (different navigation)
- ❌ Business logic (different flows)

---

## Implementation Phases

### Phase 1: Create Shared Package ✅
**Objective**: Extract common utilities to avoid duplication

**Tasks**:
1. Create `/frontend/shared` directory
2. Extract type definitions from user-app
3. Create base API client class
4. Extract validation utilities
5. Extract formatters and constants
6. Set up as npm workspace
7. Build and test

**Deliverables**:
- `@nivo/shared` package
- Types, API client, utils
- README with usage guide

**Estimated Time**: 1 hour

### Phase 2: Create Admin App Structure ✅
**Objective**: Set up new admin app with proper foundation

**Tasks**:
1. Create `/frontend/admin-app` directory
2. Initialize Vite + React + TypeScript
3. Set up Tailwind CSS
4. Configure routing (React Router v6)
5. Set up state management (Zustand)
6. Create admin layout component
7. Set up API client (extends shared base)
8. Configure build pipeline

**Deliverables**:
- Working admin app skeleton
- Build/dev scripts
- Proper configuration

**Estimated Time**: 30 minutes

### Phase 3: Implement Admin Authentication ✅
**Objective**: Secure admin app with role-based access

**Tasks**:
1. Create AdminAuthStore (with role checking)
2. Create AdminLogin page
3. Create AdminRoute guard (checks admin role)
4. Add role detection to JWT claims
5. Add session timeout (2h)
6. Add auto-logout on timeout
7. Add permission checking utilities

**Deliverables**:
- Secure admin authentication
- Role-based route protection
- Session management

**Estimated Time**: 45 minutes

### Phase 4: Move Admin Features ✅
**Objective**: Migrate admin features from user app

**Tasks**:
1. Move AdminDashboard component
2. Move AdminKYC component
3. Refactor to use shared types
4. Update API calls to use admin API client
5. Add proper error handling
6. Add loading states
7. Test all features

**Deliverables**:
- Admin dashboard working
- KYC review working
- All features tested

**Estimated Time**: 30 minutes

### Phase 5: Enhance Admin App ✅
**Objective**: Complete admin features for production

**Tasks**:
1. Add User Management pages
2. Add Transaction Monitoring pages
3. Add Audit Log viewer
4. Add Reports section
5. Add Admin Settings
6. Add comprehensive error handling
7. Add loading skeletons

**Deliverables**:
- Complete admin feature set
- Production-ready UI
- Comprehensive error handling

**Estimated Time**: 2 hours

### Phase 6: Clean Up User App ✅
**Objective**: Remove admin code and optimize

**Tasks**:
1. Remove admin routes from App.tsx
2. Remove AdminDashboard component
3. Remove AdminKYC component
4. Update API client (user-specific only)
5. Optimize bundle (remove unused code)
6. Test all user features
7. Update navigation (no admin links)

**Deliverables**:
- Clean user app (no admin code)
- Smaller bundle size
- All user features working

**Estimated Time**: 30 minutes

### Phase 7: Security Hardening ✅
**Objective**: Implement security best practices

**Tasks**:
1. **User App**:
   - Add CSRF protection
   - Verify rate limiting
   - Add input sanitization
   - Configure CSP headers
   - Test authentication flow

2. **Admin App**:
   - Add role-based access control
   - Add audit logging
   - Add session timeout
   - Configure stricter rate limiting
   - Add IP whitelisting (production config)
   - Test permission checks

**Deliverables**:
- Security hardening complete
- Audit logging functional
- RBAC working

**Estimated Time**: 1 hour

### Phase 8: Documentation & Testing ✅
**Objective**: Complete, production-ready documentation

**Tasks**:
1. Write README for user-app
2. Write README for admin-app
3. Write README for shared package
4. Create development guide
5. Create deployment guide
6. Document environment variables
7. Write testing guide
8. Manual testing checklist
9. Update main README

**Deliverables**:
- Comprehensive documentation
- Deployment guides
- Testing procedures

**Estimated Time**: 45 minutes

---

## Total Estimated Time: 6-7 hours

---

## Success Criteria

### User App
- ✅ No admin code in bundle
- ✅ All user features working
- ✅ Bundle size < 300KB (gzipped)
- ✅ Mobile-responsive
- ✅ Fast load time (<2s)
- ✅ Secure authentication
- ✅ Real-time updates working

### Admin App
- ✅ Role-based access control working
- ✅ All admin features functional
- ✅ Audit logging implemented
- ✅ Desktop-optimized
- ✅ Secure authentication
- ✅ Permission checks on all routes
- ✅ Production-ready

### Shared Package
- ✅ Zero duplication of common code
- ✅ TypeScript types exported
- ✅ Build process working
- ✅ Imported correctly in both apps
- ✅ Well-documented

### Security
- ✅ No cross-app access (user can't access admin)
- ✅ Admin role required for admin app
- ✅ Audit logging for admin actions
- ✅ Rate limiting configured
- ✅ Session management working
- ✅ HTTPS enforced (production)

### Documentation
- ✅ README for each app
- ✅ Development setup guide
- ✅ Deployment guide
- ✅ Architecture documentation
- ✅ Security documentation

---

## Deployment Architecture

### Development
```
User App:     http://localhost:3000
Admin App:    http://localhost:3001
API Gateway:  http://localhost:8000
```

### Production
```
User App:     https://app.nivomoney.com     (CDN, global)
Admin App:    https://admin.nivomoney.com   (VPN-only, single region)
API Gateway:  https://api.nivomoney.com
```

### Build Process
```bash
# User App
cd frontend/user-app
npm run build
# Output: dist/

# Admin App
cd frontend/admin-app
npm run build
# Output: dist/

# Shared Package (auto-built by workspaces)
cd frontend/shared
npm run build
```

---

## Risk Mitigation

### Risk: Code Duplication
**Mitigation**: Shared package for all common code
**Monitoring**: Regular code reviews

### Risk: Shared Package Changes Break Apps
**Mitigation**: Semantic versioning, changelog
**Testing**: Integration tests

### Risk: Admin Access by Regular Users
**Mitigation**: Role-based route guards, backend permission checks
**Testing**: Security testing, penetration testing

### Risk: Deployment Complexity
**Mitigation**: Clear documentation, CI/CD pipelines
**Automation**: Deploy scripts

---

## Next Steps After Implementation

1. **Set up CI/CD pipelines**:
   - Separate pipelines for user-app and admin-app
   - Automated testing
   - Automated deployment

2. **Add E2E Testing**:
   - Cypress for user flows
   - Admin workflow testing

3. **Performance Monitoring**:
   - Bundle size monitoring
   - Load time tracking
   - Error tracking (Sentry)

4. **Security Audits**:
   - Penetration testing
   - Code security scanning
   - Dependency auditing

---

## Approval & Sign-off

**Prepared by**: AI Engineering Assistant
**Date**: 2025-11-26
**Status**: Ready for implementation
**Estimated Total Time**: 6-7 hours
**Approach**: Comprehensive, security-first, production-ready

**Proceed with implementation?** ✅

---

**End of Plan**
