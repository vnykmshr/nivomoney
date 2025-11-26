# Admin Workflow Pattern

## Overview
This document defines the standard pattern for user actions that require admin approval/validation in the Nivo Money platform.

## Core Pattern

### User Action → Notification → Admin Validation

1. **User submits action** (e.g., KYC, withdrawal request, account closure)
   - Frontend form collects necessary data
   - NO file uploads from frontend - use dummy/placeholder data
   - Submission creates a record in database with "pending" status

2. **System generates notification**
   - Backend creates notification for admin users
   - Notification appears in Admin Dashboard notification panel
   - Notification contains summary and link to review details

3. **Admin reviews and approves/rejects**
   - Admin sees notification in Admin Dashboard
   - Clicks notification to see details
   - Takes action (approve/reject) with optional reason
   - System updates status and notifies user

## Implementation Guidelines

### Frontend - User Forms
- ✅ Collect text/structured data only
- ❌ NO file upload components
- ❌ NO file selection/preview UI
- ✅ Use text fields for document numbers (PAN, Aadhaar, etc.)
- ✅ Show clear success message after submission
- ✅ Redirect to dashboard or status page

### Backend - Action Processing
- ✅ Validate input data
- ✅ Create record with "pending" status
- ✅ Send notification to admin users (via notification service)
- ✅ Return success response to user
- ✅ Include notification ID in response if needed

### Backend - Notification Creation
```go
// Example: After creating pending record
if s.notificationClient != nil {
    notifReq := &clients.SendNotificationRequest{
        Recipient:     adminEmail, // or admin role
        Channel:       clients.NotificationChannelEmail,
        Type:          clients.NotificationTypeAdminAction,
        Priority:      clients.NotificationPriorityHigh,
        TemplateID:    "admin_kyc_review",
        Variables:     map[string]interface{}{
            "user_name": user.FullName,
            "user_id": user.ID,
            "action": "kyc_submission",
        },
    }
    s.notificationClient.SendNotificationAsync(notifReq, "identity")
}
```

### Admin Dashboard
- ✅ Primary interface: Notification panel
- ✅ Notifications grouped by type/priority
- ✅ Click notification → Detail view
- ✅ Specialized pages (like `/admin/kyc`) are supplementary
- ✅ Notifications should be the main entry point

## Examples

### KYC Verification Flow
1. User fills KYC form (PAN, Aadhaar, DOB, Address) - NO document upload
2. Submission creates `user_kyc` record with status="pending"
3. Notification sent to admins: "New KYC submission from [User Name]"
4. Admin clicks notification → sees KYC details
5. Admin approves/rejects
6. User receives notification of decision
7. User status updated accordingly

### Withdrawal Request Flow (Future)
1. User requests withdrawal (amount, account details) - NO bank statement upload
2. Creates `withdrawal_request` with status="pending"
3. Notification to admins: "Withdrawal request: ₹[amount] from [User]"
4. Admin reviews balance, history
5. Admin approves/rejects with reason
6. User notified, funds processed if approved

### Account Closure Flow (Future)
1. User requests account closure (reason, confirmation)
2. Creates `account_closure_request` with status="pending"
3. Notification to admins
4. Admin reviews account activity, outstanding balances
5. Admin approves/rejects
6. If approved, account status → closed

## File Handling

### Current Approach (NO FILE UPLOAD)
- Document numbers (PAN: ABCDE1234F) stored as text
- Aadhaar number stored as text (encrypted)
- Address stored as JSON/structured data
- No actual document files stored or uploaded

### Rationale
- Simplifies frontend and backend
- Faster submission process
- Admin validation based on data accuracy, not document verification
- In production, can integrate with government APIs for verification
- Placeholder approach allows us to build the full workflow

## Admin Dashboard Structure (Future)

```
Admin Dashboard
├── Notifications Panel (PRIMARY)
│   ├── KYC Submissions (badge count)
│   ├── Withdrawal Requests (badge count)
│   ├── Account Closures (badge count)
│   └── Other Actions (badge count)
├── KYC Management (specialized view)
├── Transaction Monitoring
├── User Management
└── Reports
```

## Notification Service Integration

All services that require admin approval must:
1. Have access to notification service client
2. Send notifications on pending actions
3. Include action link in notification
4. Update notification status when action is taken

## Benefits of This Pattern

1. **Centralized workflow**: All admin actions through one interface
2. **Audit trail**: Notifications create paper trail
3. **Prioritization**: Urgent items can be flagged
4. **Scalability**: Easy to add new approval workflows
5. **User experience**: Clear feedback on submission status
6. **No file management complexity**: Avoid storage, security, format issues

## Status
- ✅ Pattern documented
- ✅ KYC flow aligned (no file upload)
- ⏳ Admin Dashboard notifications panel (to be built)
- ⏳ Other approval workflows (withdrawal, closure, etc.)
