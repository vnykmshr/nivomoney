# Nivo Money Frontend

Two separate React apps: one for users, one for admins.

## What's Here

```
frontend/
├── shared/          # Common code (types, API client, security)
├── user-app/        # Public app for end users
└── admin-app/       # Internal app for admins
```

## Running Locally

```bash
# Install everything
cd frontend/shared && npm install
cd ../user-app && npm install
cd ../admin-app && npm install

# Build shared package first (required)
cd ../shared && npm run build

# Run user app (port 3000)
cd ../user-app && npm run dev

# Run admin app (port 3001)
cd ../admin-app && npm run dev
```

Set `VITE_API_URL` in `.env` if your backend isn't on localhost:8000.

## Building for Production

```bash
# Build everything
cd shared && npm run build
cd ../user-app && npm run build
cd ../admin-app && npm run build

# Output is in each app's dist/ folder
```

## What Each App Does

**User App** (port 3000):
- Register account, login
- Submit KYC documents
- Create wallets (auto-created on signup now)
- Send/receive money
- View transaction history

**Admin App** (port 3001):
- Review pending KYC submissions
- Approve or reject KYC with reason
- View system stats
- Everything is audit logged

**Shared Package**:
- TypeScript types (User, Wallet, Transaction, etc.)
- Base API client with auth handling
- Security utilities (CSRF, audit logging)
- Form validation
- Date/currency formatters

## Security Notes

**CSRF Protection**: Enabled in production only. Tokens auto-added to POST/PUT/DELETE/PATCH requests.

**Audit Logging**: Always on for admin app. Logs to console in dev, sends to backend in prod.

**Session Timeouts**:
- Users: 24 hours
- Admins: 2 hours

**IP Whitelisting**: Admin app restricted to office/VPN IPs in production. Configure in nginx.

See DEPLOYMENT_SECURITY.md for nginx configs and production setup.

## User Flows

**New User**:
1. Register → Auto-creates wallet with ₹0 balance
2. Submit KYC (PAN, DOB, address)
3. Wait for admin approval
4. Once approved: deposit, withdraw, transfer

**Admin**:
1. Login → Dashboard shows pending KYCs
2. Click KYC → Review details
3. Approve (instant) or Reject (with reason)
4. User notified automatically

## Common Issues

**"Cannot find @nivo/shared"**: Build the shared package first.

**API errors**: Check VITE_API_URL points to your backend.

**CSRF 403 errors**: Only happens in production. Backend must validate X-CSRF-Token header.

**Session expired**: Normal for admin after 2 hours idle.

## Testing

Start both apps and backend, then:

```bash
# User flow
1. Register new account → Should auto-create wallet
2. Submit KYC → Check DB shows pending
3. Try to transfer → Should block until KYC approved

# Admin flow
1. Login to admin app
2. Check dashboard → Should see pending KYC
3. Approve it → Check browser console for audit log
4. Verify user can now transfer

# Security
1. Check browser dev tools → Network tab
2. POST requests should have X-CSRF-Token header (production only)
3. Check console for [AUDIT] logs (admin app)
```

## Next Steps

See USER_STORIES.md for planned features and user flows.

Development approach: Start from user/admin perspective, build frontend first, then backend endpoints.

## Tech Stack

- React 18 + TypeScript
- Vite for build/dev
- Zustand for state
- Axios for API calls
- Tailwind CSS (user app has custom styles)
- Monorepo with shared package

## File Structure

**User App Routes**:
- `/` - Landing page
- `/login` - Login
- `/register` - Registration
- `/dashboard` - User dashboard
- `/transfer` - Send money
- `/kyc` - Submit KYC

**Admin App Routes**:
- `/` - Admin dashboard (stats)
- `/login` - Admin login
- `/kyc` - KYC review panel

**Key Files**:
- `shared/src/types/index.ts` - All TypeScript types
- `shared/src/lib/apiClient.ts` - Base API client
- `shared/src/lib/security.ts` - CSRF, audit logging, security configs
- `user-app/src/lib/api.ts` - User API client
- `admin-app/src/lib/adminApi.ts` - Admin API client
- `admin-app/src/lib/auditLogger.ts` - Admin action logging

## Making Changes

**Add new API endpoint**:
1. Add method to ApiClient (user) or AdminApiClient (admin)
2. Add types to shared/src/types if needed
3. Rebuild shared: `cd shared && npm run build`

**Add new page**:
1. Create component in src/pages/
2. Add route in App.tsx
3. Use existing stores (authStore, etc.)

**Add admin action**:
1. Add method to adminApi
2. Add audit log function to auditLogger.ts
3. Call logAdminAction.yourAction() after success

## Questions?

Check:
- DEPLOYMENT_SECURITY.md - Production deployment and nginx setup
- SECURITY.md - Full security architecture details
- Backend API docs - Endpoint contracts

Frontend split is done. Both apps work independently. Security hardened. Ready for production.
