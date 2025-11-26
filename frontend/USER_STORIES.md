# User Stories - Next Iteration

Drive development from user perspective. Build frontend first, backend follows.

## Regular User Stories

### Story 1: View Transaction Details
**As a user, I want to click a transaction and see full details**

What the user sees:
- Transaction list shows basic info (date, amount, status)
- Click transaction → Modal/page with full details
- Details: timestamp, from/to accounts, fee, reference ID, notes, status

Frontend needs:
- Transaction detail component
- Modal or separate page
- Format transaction types nicely (deposit, withdrawal, transfer)
- Show pending/completed/failed status clearly

Backend needs:
- GET /api/v1/transaction/transactions/:id already exists
- Maybe add more fields to response (fee breakdown, etc.)

Priority: Medium

---

### Story 2: Search My Transactions
**As a user, I want to search/filter my transaction history**

What the user sees:
- Search bar above transaction list
- Filters: date range, type (deposit/withdrawal/transfer), amount range
- Search by reference ID or recipient
- Results update as I type

Frontend needs:
- Search input component
- Filter dropdowns (transaction type, date picker)
- Local filtering or API call (depends on volume)
- Clear filters button

Backend needs:
- GET /api/v1/transaction/transactions/wallet/:walletId with query params
- ?search=, ?type=, ?from_date=, ?to_date=, ?min_amount=, ?max_amount=

Priority: High (users will need this)

---

### Story 3: Download Transaction Statement
**As a user, I want to download my transaction history as PDF/CSV**

What the user sees:
- "Download Statement" button on dashboard
- Choose date range, format (PDF or CSV)
- Downloads immediately

Frontend needs:
- Download button component
- Date range picker
- Format selector (PDF/CSV toggle)
- Trigger download via API or generate client-side

Backend needs:
- GET /api/v1/transaction/transactions/wallet/:walletId/export?format=pdf&from=&to=
- Return PDF or CSV file
- Include wallet balance, transaction summary

Priority: Medium

---

### Story 4: Set Transfer Limits
**As a user, I want to set daily/monthly transfer limits for safety**

What the user sees:
- Settings page → Transfer Limits section
- Set daily limit (₹10,000 default)
- Set monthly limit (₹100,000 default)
- Confirm with password

Frontend needs:
- Settings page (new)
- Limit input fields with validation
- Password confirmation modal
- Show current limits and remaining amount today/this month

Backend needs:
- GET /api/v1/wallet/wallets/:id/limits
- PUT /api/v1/wallet/wallets/:id/limits
- Validate password, update limits
- Track usage (sum of transfers today/this month)

Priority: High (security feature)

---

### Story 5: Add Beneficiary/Favorite Recipients
**As a user, I want to save frequent recipients for quick transfers**

What the user sees:
- Transfer page → "Saved Recipients" section
- Add beneficiary: phone/email, nickname
- Quick transfer: select from list, enter amount, done
- Edit/delete beneficiaries

Frontend needs:
- Beneficiary list component
- Add beneficiary form (phone/email, nickname)
- Quick transfer from saved list
- Manage beneficiaries (edit/delete)

Backend needs:
- GET /api/v1/wallet/beneficiaries
- POST /api/v1/wallet/beneficiaries (phone/email, nickname)
- DELETE /api/v1/wallet/beneficiaries/:id
- Validate beneficiary exists (has wallet)

Priority: High (UX improvement)

---

### Story 6: Transaction Notifications
**As a user, I want to see in-app notifications for transactions**

What the user sees:
- Bell icon in header with notification count
- Click → Notification panel
- Shows: "You received ₹500 from John Doe" (2 mins ago)
- Mark as read, clear all

Frontend needs:
- Notification icon with badge count
- Notification panel/dropdown
- Notification list component
- Mark as read API call

Backend needs:
- WebSocket for real-time notifications OR polling
- GET /api/v1/notifications (unread count, list)
- PUT /api/v1/notifications/:id/read
- Create notification on transaction events

Priority: High (user engagement)

---

### Story 7: Update Profile Information
**As a user, I want to update my name, phone, email**

What the user sees:
- Profile page (new)
- Edit name, phone, email
- Verify email/phone change with OTP
- Update password (separate form)

Frontend needs:
- Profile page with edit form
- OTP verification modal (if changing phone/email)
- Password change form (old password + new password)
- Validation and error handling

Backend needs:
- GET /api/v1/identity/users/me already exists
- PUT /api/v1/identity/users/me (name only, or other safe fields)
- PUT /api/v1/identity/users/me/phone (requires OTP)
- PUT /api/v1/identity/users/me/email (requires OTP)
- PUT /api/v1/identity/users/me/password (requires old password)

Priority: Medium

---

### Story 8: View KYC Status and Resubmit
**As a user, I want to see why my KYC was rejected and resubmit**

What the user sees:
- KYC page shows status: Pending / Approved / Rejected
- If rejected: shows reason from admin
- "Resubmit KYC" button
- Update PAN, DOB, address, submit again

Frontend needs:
- KYC status display component
- Show rejection reason clearly
- Resubmit flow (same as initial submission)
- Confirmation message

Backend needs:
- GET /api/v1/identity/auth/kyc already returns status and rejection_reason
- PUT /api/v1/identity/auth/kyc already exists (resubmit)
- Reset status to pending on resubmit

Priority: High (critical user flow)

---

### Story 9: Request Account Statement via Email
**As a user, I want to receive my statement via email automatically**

What the user sees:
- Settings → Email Preferences
- Toggle: "Monthly statement email" (on/off)
- Sent on 1st of each month

Frontend needs:
- Settings page
- Email preferences toggle
- Confirmation message

Backend needs:
- PUT /api/v1/identity/users/me/preferences
- Store preference (monthly_statement_email: true/false)
- Cron job to send statements on 1st of month

Priority: Low (nice to have)

---

### Story 10: Dark Mode Toggle
**As a user, I want to switch between light and dark themes**

What the user sees:
- Toggle in header or settings
- Switches immediately
- Preference saved across sessions

Frontend needs:
- Theme context/state
- Dark mode CSS classes
- Toggle component
- Store preference in localStorage

Backend needs:
- None (purely frontend)

Priority: Low

---

## Admin User Stories

### Story 11: View All Users
**As an admin, I want to see a list of all users**

What the admin sees:
- "Users" page in admin app
- Searchable, paginated user list
- Shows: name, email, phone, KYC status, wallet balance
- Click user → User detail page

Frontend needs:
- Users list page (new)
- Search input
- User table with sorting
- Pagination

Backend needs:
- GET /api/v1/admin/users?search=&limit=&offset=&kyc_status=
- Return user list with wallet info

Priority: High

---

### Story 12: View User Details and Transaction History
**As an admin, I want to see a user's full profile and transactions**

What the admin sees:
- User detail page
- Shows: profile, KYC info, wallet(s), recent transactions
- Can view all transactions (same as user sees)
- Can see audit trail (admin actions on this user)

Frontend needs:
- User detail page (new)
- Tabs: Profile, KYC, Wallets, Transactions, Audit Log
- Reuse transaction list component

Backend needs:
- GET /api/v1/admin/users/:id (full profile)
- GET /api/v1/admin/users/:id/transactions
- GET /api/v1/admin/users/:id/audit-log

Priority: High

---

### Story 13: Suspend/Unsuspend User Account
**As an admin, I want to suspend users for violations**

What the admin sees:
- User detail page → "Suspend Account" button
- Enter reason (required)
- Confirm suspension
- Suspended users can't login or transfer
- "Unsuspend" button to restore access

Frontend needs:
- Suspend button with reason modal
- Confirmation dialog
- Unsuspend button
- Show suspension status clearly

Backend needs:
- POST /api/v1/admin/users/:id/suspend (reason)
- POST /api/v1/admin/users/:id/unsuspend
- Block login and transactions for suspended users
- Audit log both actions

Priority: High

---

### Story 14: Manual Transaction Reversal
**As an admin, I want to reverse a transaction (refund)**

What the admin sees:
- Transaction detail page (admin view)
- "Reverse Transaction" button (only for completed)
- Enter reason
- Confirm reversal
- Creates opposite transaction

Frontend needs:
- Transaction detail page in admin app
- Reverse button (conditional on status)
- Reason input modal
- Confirmation dialog

Backend needs:
- POST /api/v1/admin/transactions/:id/reverse (reason)
- Create reverse transaction (credit sender, debit recipient)
- Update original transaction status to "reversed"
- Audit log action

Priority: Medium (fraud/error handling)

---

### Story 15: View System Audit Logs
**As an admin, I want to search and view all admin actions**

What the admin sees:
- "Audit Log" page in admin app
- Search by: admin user, action type, date range, target user
- Shows: timestamp, admin, action, target, details
- Export as CSV

Frontend needs:
- Audit log page (new)
- Search/filter form
- Audit log table with sorting
- Export button

Backend needs:
- GET /api/v1/admin/audit-logs?admin_id=&action=&from_date=&to_date=&target_user_id=
- Store audit logs in DB (already sending from frontend in production)
- Export endpoint

Priority: Medium (compliance)

---

### Story 16: Bulk KYC Approval
**As an admin, I want to approve multiple KYCs at once**

What the admin sees:
- KYC review page → Checkboxes on each KYC
- "Approve Selected" button (bulk action)
- Confirm count: "Approve 5 KYCs?"
- All approved, users notified

Frontend needs:
- Checkbox selection in KYC list
- Bulk action buttons
- Multi-select state management
- Confirmation dialog

Backend needs:
- POST /api/v1/admin/kyc/bulk-approve
- Body: { user_ids: ["id1", "id2", ...] }
- Process all, return results
- Audit log each approval

Priority: Low (efficiency, only if volume high)

---

### Story 17: Generate System Reports
**As an admin, I want to generate reports (users, transactions, revenue)**

What the admin sees:
- "Reports" page in admin app
- Report types: User Growth, Transaction Volume, Revenue, KYC Stats
- Select date range
- Generate and download PDF/CSV

Frontend needs:
- Reports page (new)
- Report type selector
- Date range picker
- Generate button → Download

Backend needs:
- GET /api/v1/admin/reports/:type?from_date=&to_date=&format=
- Types: user-growth, transaction-volume, revenue, kyc-stats
- Return PDF or CSV

Priority: Low

---

### Story 18: Set System-Wide Limits
**As an admin, I want to set global transfer limits**

What the admin sees:
- Settings page (admin) → System Limits
- Set: max transfer amount per transaction, max daily transfers per user
- Apply to all users (existing limits can be lower)

Frontend needs:
- Admin settings page (new)
- Limit configuration form
- Save button with confirmation

Backend needs:
- GET /api/v1/admin/settings
- PUT /api/v1/admin/settings/limits
- Store in config table or environment
- Enforce in transaction validation

Priority: Low

---

### Story 19: View Real-Time Transaction Feed
**As an admin, I want to see live transactions happening**

What the admin sees:
- Dashboard → Live Feed section
- Shows transactions as they happen
- Scroll to see recent (last 100)
- Click to view details

Frontend needs:
- Live feed component
- WebSocket connection OR polling
- Auto-scroll on new transaction
- Click to view details

Backend needs:
- WebSocket endpoint for admin: /ws/admin/transactions
- Or polling endpoint with last_id param
- Push transaction events to admin clients

Priority: Low (monitoring)

---

### Story 20: Manage Admin Users (Super Admin)
**As a super admin, I want to create/remove admin accounts**

What the admin sees:
- "Admin Users" page
- List of all admin accounts
- Add admin: email, name, permissions (future: roles)
- Remove admin with confirmation

Frontend needs:
- Admin users page (new)
- Add admin form
- Remove button with confirmation
- Permissions/roles selector (future)

Backend needs:
- GET /api/v1/admin/admin-users
- POST /api/v1/admin/admin-users (create)
- DELETE /api/v1/admin/admin-users/:id
- RBAC: Only super admin can access
- Audit log admin user changes

Priority: Medium (multi-admin setup)

---

## Implementation Approach

**For each story:**

1. **Frontend First**
   - Design the UI/UX
   - Build components with mock data
   - Get user feedback early

2. **Define API Contract**
   - Agree on request/response format
   - Document in API spec

3. **Backend Implementation**
   - Build endpoint to match contract
   - Add validation, error handling
   - Write tests

4. **Integration**
   - Connect frontend to real API
   - End-to-end testing
   - Fix issues

5. **Deploy**
   - Frontend to CDN/web server
   - Backend to production
   - Monitor for issues

## Prioritization

**This Sprint (High Priority):**
- Story 2: Search transactions
- Story 4: Transfer limits
- Story 5: Saved beneficiaries
- Story 6: Notifications
- Story 8: KYC resubmission
- Story 11: Admin user list
- Story 12: Admin user details
- Story 13: Suspend/unsuspend users

**Next Sprint (Medium Priority):**
- Story 1: Transaction details
- Story 3: Download statement
- Story 7: Update profile
- Story 14: Transaction reversal
- Story 15: View audit logs
- Story 20: Manage admins

**Later (Low Priority):**
- Story 9: Email statements
- Story 10: Dark mode
- Story 16: Bulk KYC approval
- Story 17: System reports
- Story 18: System-wide limits
- Story 19: Live transaction feed

## Notes

- Start with high-value user features (search, beneficiaries, notifications)
- Build admin tools as needed (user management, suspension)
- Focus on completing flows end-to-end rather than half-implementing everything
- Get user feedback after each story
- Iterate based on real usage patterns

Frontend is ready. Let's build features users actually need.
