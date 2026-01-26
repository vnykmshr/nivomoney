---
layout: default
title: Demo Walkthrough
nav_order: 2
description: "Try Nivo with pre-configured demo accounts"
permalink: /demo
---

# Demo Walkthrough
{: .no_toc }

Experience Nivo's features with pre-configured demo accounts.
{: .fs-6 .fw-300 }

---

## Table of Contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Demo Accounts

All demo accounts have verified KYC and active wallets ready to use:

| Persona | Email | Password | Balance | Description |
|:--------|:------|:---------|:--------|:------------|
| **Recommended** | raj.kumar@gmail.com | raj123 | ₹50,000 | Regular user - best for exploring |
| Business Owner | priya.electronics@business.com | priya123 | ₹1,50,000 | Higher balance, business profile |
| Freelancer | arjun.design@freelance.com | arjun123 | ₹75,000 | Freelance designer profile |
| Student | neha.singh@student.com | neha123 | ₹25,000 | Lower balance account |
| Corporate | vikram.m@corporate.com | vikram123 | ₹2,00,000 | Highest balance account |
| Admin | admin@nivo.local | admin123 | ₹1,00,000 | Admin dashboard access |

{: .note }
> All data is synthetic. These are development-only credentials with dummy PII. No real money is involved.

---

## Guided Walkthrough

### Step 1: Login

1. Go to [nivomoney.com](https://nivomoney.com)
2. Click **Login** or **Get Started**
3. Enter credentials: `raj.kumar@gmail.com` / `raj123`
4. You'll land on the dashboard

**What to notice:**
- JWT authentication (check Network tab for token)
- Role-based permissions embedded in token

---

### Step 2: Explore the Dashboard

After login, the dashboard shows:

- **Current Balance**: ₹50,000.00 (from pre-seeded data)
- **Recent Transactions**: Initial balance deposit from seed
- **Quick Actions**: Send Money, Add Funds, View History

**What to notice:**
- Balance fetched from Wallet Service
- Real-time updates via SSE (Server-Sent Events)

---

### Step 3: Send Money

Try a peer-to-peer transfer:

1. Click **Send Money**
2. Enter recipient: `priya.electronics@business.com`
3. Enter amount: `1000` (₹1,000)
4. Add a note: "Test transfer"
5. Confirm the transfer

**What to notice:**
- Beneficiary lookup before transfer
- Idempotency key generated for safe retries
- Double-entry ledger creates balanced entries:
  ```
  Debit:  Raj's Wallet    ₹1,000
  Credit: Priya's Wallet  ₹1,000
  ```

---

### Step 4: View Transaction History

1. Click **Transactions** in the navigation
2. See the transfer you just made
3. Click on a transaction for details

**What to notice:**
- Transaction status progression: `pending` → `completed`
- Reference numbers for each transaction
- Audit trail with timestamps

---

### Step 5: Check Profile & KYC

1. Click **Profile** or the user menu
2. View your KYC status (pre-verified)
3. See account details

**What to notice:**
- KYC fields: PAN, Aadhaar (masked), DOB, Address
- Account status: `active`
- India-specific validation (PAN format, Aadhaar masking)

---

### Step 6: Try Another Account

Logout and login as a different persona to see varied balances:

1. Click **Logout**
2. Login as `priya.electronics@business.com` / `priya123`
3. Check the dashboard - you'll see ₹1,51,000 (original + received transfer)

---

## Verify Portal

The Verify Portal is for trusted verifiers (family members, guardians) who help paired users approve transactions:

1. Go to [verify.nivomoney.com](https://verify.nivomoney.com)
2. Login as `priya.electronics@business.com` / `priya123`
3. Explore:
   - Pending verifications dashboard
   - OTP codes for transaction approval
   - Verification history

**How it works:**
1. A paired user initiates a transaction requiring verification
2. The OTP code appears in the Verify Portal
3. Share the code with the paired user to complete the transaction
4. Codes expire after 5 minutes

This accessibility feature allows shared account management for users who need assistance.

---

## Admin Dashboard

For admin features, login to the admin app:

1. Go to [admin.nivomoney.com](https://admin.nivomoney.com)
2. Login as `admin@nivo.local` / `admin123`
3. Explore:
   - User management
   - KYC verification queue
   - Transaction monitoring
   - System health

---

## Technical Features Demonstrated

| Feature | Where to See It |
|:--------|:----------------|
| **JWT Authentication** | Login flow, network requests |
| **RBAC Permissions** | Admin vs User capabilities |
| **Double-Entry Ledger** | Transaction details |
| **Idempotency** | Send money (retry safe) |
| **Real-time Updates** | Dashboard balance after transfer |
| **KYC Workflow** | Profile page status |
| **Beneficiary Management** | Send money flow |

---

## API Exploration

Want to explore the APIs directly? Use the demo credentials with:

```bash
# Login and get JWT token
curl -X POST https://api.nivomoney.com/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "raj.kumar@gmail.com", "password": "raj123"}'

# Use the token for authenticated requests
curl https://api.nivomoney.com/api/v1/wallet/wallets \
  -H "Authorization: Bearer <your-token>"
```

---

## Running Locally

To run the full stack locally for development:

```bash
# Start infrastructure
docker-compose up -d

# Seed the database
go run services/seed/cmd/server/main.go

# Start services
make run-all

# Start frontend
cd frontend/user-app && npm run dev
```

See [Quick Start](/quickstart) for detailed setup instructions.

---

## Troubleshooting

### "Invalid credentials"
Make sure you're using the exact email and password from the table above.

### Balance not updating
Wait a moment for SSE event propagation, or refresh the page.

### Transfer fails
Check that:
- Recipient email exists (use another demo account)
- Amount doesn't exceed your balance
- Amount is positive

### Session expired
JWT tokens expire after 24 hours. Login again to get a fresh token.
