# Changes for Admin Workflow Pattern

## Date: 2025-11-26

## Overview
Aligned KYC flow with the standardized admin workflow pattern: **User Action → Notification → Admin Validation**

## Key Principle
**NO FILE UPLOADS** - All document verification uses text-based data (document numbers, addresses) as placeholders/dummies. Admins validate based on data correctness, not actual document review.

## Changes Made

### 1. Documentation
**Created: `/ADMIN_WORKFLOW_PATTERN.md`**
- Comprehensive documentation of the admin approval pattern
- Guidelines for implementing user actions requiring admin approval
- Examples: KYC, withdrawal requests, account closure
- Benefits and rationale for no-file-upload approach

### 2. Backend - Admin Notifications
**Modified: `services/identity/internal/service/auth_service.go`**
- Added notification to admins when user submits KYC (line 379-404)
- Notification includes:
  - User details (name, email, ID)
  - PAN number
  - Action URL to review KYC
- Uses notification service client asynchronously
- Template: `admin_kyc_review_required`
- Priority: High
- Channel: Email (to admin@nivomoney.com placeholder)

**Code snippet:**
```go
// Send notification to admins for KYC review
// Following the admin workflow pattern: User Action → Notification → Admin Validation
if s.notificationClient != nil {
    user, userErr := s.userRepo.GetByID(ctx, userID)
    if userErr == nil {
        correlationID := fmt.Sprintf("kyc-review-%s", userID)
        notifReq := &clients.SendNotificationRequest{
            Recipient:     "admin@nivomoney.com",
            Channel:       clients.NotificationChannelEmail,
            Type:          clients.NotificationTypeKYCStatus,
            Priority:      clients.NotificationPriorityHigh,
            TemplateID:    "admin_kyc_review_required",
            Variables: map[string]interface{}{
                "user_name":  user.FullName,
                "user_id":    userID,
                "user_email": user.Email,
                "pan":        kyc.PAN,
                "action_url": fmt.Sprintf("/admin/kyc?user_id=%s", userID),
            },
            CorrelationID: &correlationID,
            SourceService: "identity",
        }
        s.notificationClient.SendNotificationAsync(notifReq, "identity")
    }
}
```

### 3. Frontend Documentation
**Modified: `frontend/user-app/src/pages/KYC.tsx`**
- Added comprehensive JSDoc comment explaining:
  - No file upload requirement
  - Text-only data collection
  - Admin workflow steps
  - Reference to pattern documentation

**Modified: `frontend/user-app/src/pages/AdminKYC.tsx`**
- Added JSDoc comment clarifying:
  - This is a specialized view
  - PRIMARY workflow is through notification panel
  - Can be accessed directly or via notification link
  - Reference to pattern documentation

## Current KYC Flow

### User Side
1. User navigates to `/kyc`
2. Fills form with:
   - PAN card number (text) - e.g., ABCDE1234F
   - Aadhaar number (text) - e.g., 1234 5678 9012
   - Date of birth (date picker)
   - Address (street, city, state, PIN, country)
3. Frontend validates:
   - PAN format (5 letters, 4 digits, 1 letter)
   - Aadhaar format (12 digits)
   - Age requirement (18+)
   - Required fields
4. Submits to backend
5. Sees success message
6. Dashboard shows "KYC Pending Review" status

### Backend Processing
1. Receives KYC submission
2. Validates data (gopantic validation)
3. Creates `user_kyc` record with status="pending"
4. Publishes `user.kyc_updated` event (for SSE)
5. **NEW:** Sends email notification to admins
6. Returns success to user

### Admin Side
1. **PRIMARY:** Admin sees notification in email (or will see in notification panel when dashboard is built)
2. Admin clicks notification link → goes to `/admin/kyc?user_id=xxx`
3. **ALTERNATE:** Admin can navigate directly to `/admin/kyc` to see all pending KYCs
4. Admin reviews:
   - User information
   - PAN number
   - Date of birth
   - Address details
   - Submission date
5. Admin takes action:
   - **Approve:** One-click approval → User status becomes "active"
   - **Reject:** Must provide reason (min 10 chars) → User can resubmit
6. User receives notification of decision
7. User's dashboard updates to show new status

## What We Don't Do (By Design)

### ❌ No File Uploads
- No document upload component
- No file storage/management
- No file format validation
- No virus scanning
- No document preview

### ❌ No OCR/Document Parsing
- No automated data extraction from documents
- No image processing
- No ML-based verification

### ❌ No Third-Party API Integration (Yet)
- No real PAN verification via government APIs
- No Aadhaar verification
- No address validation services

## Benefits of This Approach

1. **Simplicity**: Faster development, easier testing
2. **Focus**: Build workflow first, add complexity later
3. **Flexibility**: Easy to modify validation rules
4. **Security**: No file security concerns
5. **Performance**: No storage/bandwidth issues
6. **Testing**: Easy to seed test data

## Future Enhancements (When Needed)

1. **Admin Dashboard Notification Panel**
   - Centralized view of all pending actions
   - Badge counts, priority sorting
   - Quick access to review pages

2. **Real Verification**
   - Integrate with government APIs (DigiLocker, etc.)
   - Auto-verify PAN/Aadhaar
   - Address validation

3. **Optional Document Upload**
   - If compliance requires actual documents
   - S3/blob storage integration
   - Document expiry tracking

4. **Audit Trail**
   - Who approved/rejected and when
   - Comments/notes from admins
   - History of resubmissions

## Testing the Flow

### Manual Test
```bash
# 1. Start services
docker-compose up -d

# 2. Register new user
curl -X POST http://localhost:8000/api/v1/identity/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Test User",
    "email": "test@example.com",
    "phone": "9876543210",
    "password": "Test1234"
  }'

# 3. Login
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier": "test@example.com", "password": "Test1234"}' \
  | jq -r '.data.token')

# 4. Submit KYC
curl -X PUT http://localhost:8000/api/v1/identity/auth/kyc \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pan": "ABCDE1234F",
    "aadhaar": "123456789012",
    "date_of_birth": "1995-01-01",
    "address": {
      "street": "123 Test Street",
      "city": "Mumbai",
      "state": "Maharashtra",
      "pin": "400001",
      "country": "IN"
    }
  }'

# 5. Check notification logs (should see admin notification)
# Look for: "admin_kyc_review_required" notification

# 6. Admin lists pending KYCs
curl http://localhost:8000/api/v1/identity/admin/kyc/pending \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 7. Admin approves KYC
curl -X POST http://localhost:8000/api/v1/identity/admin/kyc/verify \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "USER_ID_HERE"}'
```

## Files Modified
- `/ADMIN_WORKFLOW_PATTERN.md` (created)
- `/CHANGES_ADMIN_WORKFLOW.md` (this file, created)
- `services/identity/internal/service/auth_service.go` (modified)
- `frontend/user-app/src/pages/KYC.tsx` (documentation added)
- `frontend/user-app/src/pages/AdminKYC.tsx` (documentation added)

## Next Steps
1. Build Admin Dashboard with Notification Panel (next user story)
2. Add notification template: `admin_kyc_review_required`
3. Test end-to-end KYC flow with notification
4. Apply this pattern to other approval workflows (withdrawals, closures)
