package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// UserRepositoryInterface defines the interface for user repository operations.
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) *errors.Error
	GetByEmail(ctx context.Context, email string) (*models.User, *errors.Error)
	GetByPhone(ctx context.Context, phone string) (*models.User, *errors.Error)
	GetByID(ctx context.Context, id string) (*models.User, *errors.Error)
	Update(ctx context.Context, user *models.User) *errors.Error
	UpdatePassword(ctx context.Context, userID string, passwordHash string) *errors.Error
	UpdateStatus(ctx context.Context, userID string, status models.UserStatus) *errors.Error
	Delete(ctx context.Context, userID string) *errors.Error
	Count(ctx context.Context) (int, *errors.Error)
	CountByStatus(ctx context.Context, status models.UserStatus) (int, *errors.Error)
}

// KYCRepositoryInterface defines the interface for KYC repository operations.
type KYCRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID string) (*models.KYCInfo, *errors.Error)
	Create(ctx context.Context, kyc *models.KYCInfo) *errors.Error
	UpdateStatus(ctx context.Context, userID string, status models.KYCStatus, reason string) *errors.Error
	ListPending(ctx context.Context, limit, offset int) ([]repository.KYCWithUser, *errors.Error)
}

// SessionRepositoryInterface defines the interface for session repository operations.
type SessionRepositoryInterface interface {
	Create(ctx context.Context, session *models.Session) *errors.Error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, *errors.Error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) *errors.Error
	DeleteByUserID(ctx context.Context, userID string) *errors.Error
}

// RBACClientInterface defines the interface for RBAC client operations.
type RBACClientInterface interface {
	AssignDefaultRole(ctx context.Context, userID string) error
	GetUserPermissions(ctx context.Context, userID string) (*UserPermissionsResponse, error)
}

// AuthService handles authentication and authorization.
type AuthService struct {
	userRepo           UserRepositoryInterface
	kycRepo            KYCRepositoryInterface
	sessionRepo        SessionRepositoryInterface
	rbacClient         RBACClientInterface
	walletClient       *WalletClient
	notificationClient *clients.NotificationClient
	jwtSecret          string
	jwtExpiry          time.Duration
	eventPublisher     *events.Publisher
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo UserRepositoryInterface,
	kycRepo KYCRepositoryInterface,
	sessionRepo SessionRepositoryInterface,
	rbacClient RBACClientInterface,
	walletClient *WalletClient,
	notificationClient *clients.NotificationClient,
	jwtSecret string,
	jwtExpiry time.Duration,
	eventPublisher *events.Publisher,
) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		kycRepo:            kycRepo,
		sessionRepo:        sessionRepo,
		rbacClient:         rbacClient,
		walletClient:       walletClient,
		notificationClient: notificationClient,
		jwtSecret:          jwtSecret,
		jwtExpiry:          jwtExpiry,
		eventPublisher:     eventPublisher,
	}
}

// JWTClaims represents the JWT token claims with RBAC support.
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Status      string   `json:"status"`
	Roles       []string `json:"roles,omitempty"`       // User's role names
	Permissions []string `json:"permissions,omitempty"` // Flattened permission list
	jwt.RegisteredClaims
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.User, *errors.Error) {
	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, errors.Internal("failed to process password")
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		PasswordHash: hashedPassword,
		Status:       models.UserStatusPending, // Pending until KYC
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Assign default "user" role in RBAC service
	// This is now required - registration fails if role assignment fails
	if err := s.rbacClient.AssignDefaultRole(ctx, user.ID); err != nil {
		// Delete the user since role assignment failed (cleanup partial state)
		_ = s.userRepo.Delete(ctx, user.ID)
		fmt.Printf("[identity] Error: Failed to assign default role to user %s: %v\n", user.ID, err)
		return nil, errors.Internal("failed to complete user registration")
	}

	// Create default INR wallet for the user
	// This is optional - registration succeeds even if wallet creation fails
	if s.walletClient != nil {
		wallet, walletErr := s.walletClient.CreateDefaultWallet(ctx, user.ID)
		if walletErr != nil {
			// Log error but continue (wallet can be created later manually)
			fmt.Printf("[identity] Warning: Failed to create default wallet for user %s: %v\n", user.ID, walletErr)
		} else {
			fmt.Printf("[identity] Created default wallet %s for user %s\n", wallet.ID, user.ID)
		}
	}

	// Publish user.registered event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.registered", user.ID, map[string]interface{}{
			"email":     user.Email,
			"phone":     user.Phone,
			"full_name": user.FullName,
			"status":    string(user.Status),
		})
	}

	// Send welcome notification
	if s.notificationClient != nil {
		// Get template IDs by querying templates (for now using hardcoded names)
		// In production, these would be cached or configured
		emailTemplateID := "welcome_email"
		smsTemplateID := "welcome_sms"

		// Send welcome email
		if user.Email != "" {
			emailReq := &clients.SendNotificationRequest{
				UserID:        &user.ID,
				Recipient:     user.Email,
				Channel:       clients.NotificationChannelEmail,
				Type:          clients.NotificationTypeWelcome,
				Priority:      clients.NotificationPriorityNormal,
				TemplateID:    emailTemplateID,
				Variables:     map[string]interface{}{"user_name": user.FullName, "full_name": user.FullName},
				CorrelationID: &user.ID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(emailReq, "identity")
		}

		// Send welcome SMS if phone provided
		if user.Phone != "" {
			smsReq := &clients.SendNotificationRequest{
				UserID:        &user.ID,
				Recipient:     user.Phone,
				Channel:       clients.NotificationChannelSMS,
				Type:          clients.NotificationTypeWelcome,
				Priority:      clients.NotificationPriorityNormal,
				TemplateID:    smsTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName},
				CorrelationID: &user.ID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(smsReq, "identity")
		}
	}

	// Sanitize before returning
	user.Sanitize()

	return user, nil
}

// Login authenticates a user and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, *errors.Error) {
	// Determine if identifier is email or phone
	// Phone numbers start with +91 for India
	var user *models.User
	var err *errors.Error

	if len(req.Identifier) > 0 && req.Identifier[0] == '+' {
		// Identifier is a phone number
		user, err = s.userRepo.GetByPhone(ctx, req.Identifier)
	} else {
		// Identifier is an email
		user, err = s.userRepo.GetByEmail(ctx, req.Identifier)
	}

	if err != nil {
		// Don't reveal if user exists
		return nil, errors.Unauthorized("invalid credentials")
	}

	// Verify password
	if !s.verifyPassword(req.Password, user.PasswordHash) {
		return nil, errors.Unauthorized("invalid credentials")
	}

	// Check if account is active
	if user.Status == models.UserStatusClosed {
		return nil, errors.Forbidden("account is closed")
	}

	if user.Status == models.UserStatusSuspended {
		return nil, errors.Forbidden("account is suspended")
	}

	// Fetch user permissions from RBAC service
	var roles []string
	var permissions []string

	userPerms, rbacErr := s.rbacClient.GetUserPermissions(ctx, user.ID)
	if rbacErr != nil {
		// Log warning but continue with empty permissions (graceful degradation)
		fmt.Printf("[identity] Warning: Failed to fetch permissions for user %s: %v\n", user.ID, rbacErr)
	} else {
		// Extract role names
		for _, role := range userPerms.Roles {
			roles = append(roles, role.Name)
		}
		// Extract permission names
		for _, perm := range userPerms.Permissions {
			permissions = append(permissions, perm.Name)
		}
	}

	// Generate JWT token with roles and permissions
	token, expiresAt, genErr := s.generateToken(user, roles, permissions)
	if genErr != nil {
		return nil, errors.Internal("failed to generate token")
	}

	// Create session
	tokenHash := s.hashToken(token)
	session := &models.Session{
		UserID:    user.ID,
		Token:     tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: sharedModels.NewTimestamp(time.Unix(expiresAt, 0)),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Load KYC info if available
	kyc, err := s.kycRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		user.KYC = *kyc
	}

	// Sanitize user
	user.Sanitize()

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// Logout invalidates a user's session.
func (s *AuthService) Logout(ctx context.Context, token string) *errors.Error {
	tokenHash := s.hashToken(token)
	return s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
}

// LogoutAll invalidates all sessions for a user.
func (s *AuthService) LogoutAll(ctx context.Context, userID string) *errors.Error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// ValidateToken validates a JWT token and returns the user.
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*models.User, *errors.Error) {
	// Parse and validate JWT
	claims := &JWTClaims{}
	token, parseErr := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if parseErr != nil || !token.Valid {
		return nil, errors.Unauthorized("invalid token")
	}

	// Check if session exists and is not expired
	tokenHash := s.hashToken(tokenString)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err // Returns unauthorized error
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Check user status
	if user.Status == models.UserStatusClosed || user.Status == models.UserStatusSuspended {
		return nil, errors.Forbidden("account is not active")
	}

	// Load KYC info
	kyc, err := s.kycRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		user.KYC = *kyc
	}

	user.Sanitize()

	return user, nil
}

// GetUserByID retrieves a user by ID.
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*models.User, *errors.Error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Load KYC info
	kyc, err := s.kycRepo.GetByUserID(ctx, userID)
	if err == nil {
		user.KYC = *kyc
	}

	user.Sanitize()

	return user, nil
}

// LookupUserByPhone finds a user by phone number (for recipient lookup in transfers).
func (s *AuthService) LookupUserByPhone(ctx context.Context, phone string) (*models.User, *errors.Error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	// Load KYC info
	kyc, err := s.kycRepo.GetByUserID(ctx, user.ID)
	if err == nil {
		user.KYC = *kyc
	}

	user.Sanitize()

	return user, nil
}

// UpdateKYC updates or creates KYC information for a user.
func (s *AuthService) UpdateKYC(ctx context.Context, userID string, req *models.UpdateKYCRequest) (*models.KYCInfo, *errors.Error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create KYC info
	kyc := &models.KYCInfo{
		UserID:      userID,
		Status:      models.KYCStatusPending,
		PAN:         req.PAN,
		Aadhaar:     req.Aadhaar,
		DateOfBirth: req.DateOfBirth,
		Address:     req.Address,
	}

	if err := s.kycRepo.Create(ctx, kyc); err != nil {
		return nil, err
	}

	// Publish user.kyc_updated event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.kyc_updated", userID, map[string]interface{}{
			"kyc_status":    string(kyc.Status),
			"date_of_birth": kyc.DateOfBirth,
			"address":       kyc.Address,
		})
	}

	// Send notification to admins for KYC review
	// Following the admin workflow pattern: User Action → Notification → Admin Validation
	if s.notificationClient != nil {
		// Get user info for notification
		user, userErr := s.userRepo.GetByID(ctx, userID)
		if userErr == nil {
			correlationID := fmt.Sprintf("kyc-review-%s", userID)
			notifReq := &clients.SendNotificationRequest{
				Recipient:  "admin@nivomoney.com", // In production, send to admin role/group
				Channel:    clients.NotificationChannelEmail,
				Type:       clients.NotificationTypeKYCStatus,
				Priority:   clients.NotificationPriorityHigh,
				TemplateID: "admin_kyc_review_required",
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

	return kyc, nil
}

// VerifyKYC approves a user's KYC (admin operation).
func (s *AuthService) VerifyKYC(ctx context.Context, userID string) *errors.Error {
	// Get user info for notifications
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Update KYC status
	if err := s.kycRepo.UpdateStatus(ctx, userID, models.KYCStatusVerified, ""); err != nil {
		return err
	}

	// Publish user.kyc_updated event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.kyc_updated", userID, map[string]interface{}{
			"kyc_status": string(models.KYCStatusVerified),
		})
	}

	// Update user status to active
	if err := s.userRepo.UpdateStatus(ctx, userID, models.UserStatusActive); err != nil {
		return err
	}

	// Publish user.status_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.status_changed", userID, map[string]interface{}{
			"status":     string(models.UserStatusActive),
			"previous":   string(models.UserStatusPending),
			"kyc_status": string(models.KYCStatusVerified),
		})
	}

	// Send KYC approved notification
	if s.notificationClient != nil {
		emailTemplateID := "kyc_approved_email"
		smsTemplateID := "kyc_approved_sms"
		correlationID := fmt.Sprintf("kyc-approved-%s", userID)

		// Send email notification
		if user.Email != "" {
			emailReq := &clients.SendNotificationRequest{
				UserID:        &userID,
				Recipient:     user.Email,
				Channel:       clients.NotificationChannelEmail,
				Type:          clients.NotificationTypeKYCStatus,
				Priority:      clients.NotificationPriorityHigh,
				TemplateID:    emailTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName, "status": "approved"},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(emailReq, "identity")
		}

		// Send SMS notification if phone provided
		if user.Phone != "" {
			smsReq := &clients.SendNotificationRequest{
				UserID:        &userID,
				Recipient:     user.Phone,
				Channel:       clients.NotificationChannelSMS,
				Type:          clients.NotificationTypeKYCStatus,
				Priority:      clients.NotificationPriorityHigh,
				TemplateID:    smsTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(smsReq, "identity")
		}
	}

	return nil
}

// RejectKYC rejects a user's KYC (admin operation).
func (s *AuthService) RejectKYC(ctx context.Context, userID string, reason string) *errors.Error {
	// Get user info for notifications
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.kycRepo.UpdateStatus(ctx, userID, models.KYCStatusRejected, reason); err != nil {
		return err
	}

	// Publish user.kyc_updated event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.kyc_updated", userID, map[string]interface{}{
			"kyc_status":       string(models.KYCStatusRejected),
			"rejection_reason": reason,
		})
	}

	// Send KYC rejected notification
	if s.notificationClient != nil {
		emailTemplateID := "kyc_rejected_email"
		smsTemplateID := "kyc_rejected_sms"
		correlationID := fmt.Sprintf("kyc-rejected-%s", userID)

		// Send email notification
		if user.Email != "" {
			emailReq := &clients.SendNotificationRequest{
				UserID:        &userID,
				Recipient:     user.Email,
				Channel:       clients.NotificationChannelEmail,
				Type:          clients.NotificationTypeKYCStatus,
				Priority:      clients.NotificationPriorityHigh,
				TemplateID:    emailTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName, "status": "rejected", "reason": reason},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(emailReq, "identity")
		}

		// Send SMS notification if phone provided
		if user.Phone != "" {
			smsReq := &clients.SendNotificationRequest{
				UserID:        &userID,
				Recipient:     user.Phone,
				Channel:       clients.NotificationChannelSMS,
				Type:          clients.NotificationTypeKYCStatus,
				Priority:      clients.NotificationPriorityHigh,
				TemplateID:    smsTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName, "reason": reason},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(smsReq, "identity")
		}
	}

	return nil
}

// UpdateProfile updates a user's profile information.
func (s *AuthService) UpdateProfile(ctx context.Context, userID string, req *models.UpdateProfileRequest) (*models.User, *errors.Error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Track changed fields for event
	changes := make(map[string]interface{})
	emailChanged := user.Email != req.Email
	phoneChanged := user.Phone != req.Phone

	// Check if email is being changed and if it's already taken
	if emailChanged {
		existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.Conflict("email already in use")
		}
		changes["email"] = map[string]string{"old": user.Email, "new": req.Email}
	}

	// Check if phone is being changed and if it's already taken
	if phoneChanged {
		existingUser, _ := s.userRepo.GetByPhone(ctx, req.Phone)
		if existingUser != nil && existingUser.ID != userID {
			return nil, errors.Conflict("phone number already in use")
		}
		changes["phone"] = map[string]string{"old": user.Phone, "new": req.Phone}
	}

	// Track full name change
	if user.FullName != req.FullName {
		changes["full_name"] = map[string]string{"old": user.FullName, "new": req.FullName}
	}

	// Update user fields
	user.FullName = req.FullName
	user.Email = req.Email
	user.Phone = req.Phone

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Publish user.profile_updated event with change details
	if s.eventPublisher != nil && len(changes) > 0 {
		s.eventPublisher.PublishUserEvent("user.profile_updated", userID, changes)
	}

	// Send notification if email or phone changed (security notification)
	if s.notificationClient != nil && (emailChanged || phoneChanged) {
		correlationID := fmt.Sprintf("profile-updated-%s", userID)

		// Send notification to old email if email was changed
		if emailChanged {
			emailReq := &clients.SendNotificationRequest{
				UserID:     &userID,
				Recipient:  changes["email"].(map[string]string)["old"],
				Channel:    clients.NotificationChannelEmail,
				Type:       "profile_change",
				Priority:   clients.NotificationPriorityHigh,
				TemplateID: "profile_email_changed",
				Variables: map[string]interface{}{
					"full_name": user.FullName,
					"new_email": req.Email,
				},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(emailReq, "identity")
		}

		// Send SMS if phone was changed
		if phoneChanged {
			smsReq := &clients.SendNotificationRequest{
				UserID:     &userID,
				Recipient:  changes["phone"].(map[string]string)["old"],
				Channel:    clients.NotificationChannelSMS,
				Type:       "profile_change",
				Priority:   clients.NotificationPriorityHigh,
				TemplateID: "profile_phone_changed",
				Variables: map[string]interface{}{
					"full_name": user.FullName,
					"new_phone": req.Phone,
				},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(smsReq, "identity")
		}
	}

	// Sanitize before returning
	user.Sanitize()

	return user, nil
}

// ChangePassword changes a user's password after verifying the current password.
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) *errors.Error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !s.verifyPassword(req.CurrentPassword, user.PasswordHash) {
		return errors.Unauthorized("current password is incorrect")
	}

	// Verify new password is different from current password
	// Check both plaintext equality and hash equality for security
	if req.CurrentPassword == req.NewPassword {
		return errors.BadRequest("new password must be different from current password")
	}

	// Additional check: verify new password doesn't match current hash
	if s.verifyPassword(req.NewPassword, user.PasswordHash) {
		return errors.BadRequest("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, hashErr := s.hashPassword(req.NewPassword)
	if hashErr != nil {
		return errors.Internal("failed to hash password")
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	// Publish user.password_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishUserEvent("user.password_changed", userID, map[string]interface{}{
			"changed_at": time.Now().Unix(),
		})
	}

	// Send notification about password change
	if s.notificationClient != nil {
		emailTemplateID := "password_changed_email"
		correlationID := fmt.Sprintf("password-changed-%s", userID)

		// Send email notification
		if user.Email != "" {
			emailReq := &clients.SendNotificationRequest{
				UserID:        &userID,
				Recipient:     user.Email,
				Channel:       clients.NotificationChannelEmail,
				Type:          "password_change",
				Priority:      clients.NotificationPriorityHigh,
				TemplateID:    emailTemplateID,
				Variables:     map[string]interface{}{"full_name": user.FullName},
				CorrelationID: &correlationID,
				SourceService: "identity",
			}
			s.notificationClient.SendNotificationAsync(emailReq, "identity")
		}
	}

	return nil
}

// ListPendingKYCs retrieves all pending KYC submissions for admin review.
func (s *AuthService) ListPendingKYCs(ctx context.Context, limit, offset int) ([]repository.KYCWithUser, *errors.Error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.kycRepo.ListPending(ctx, limit, offset)
}

// AdminStats represents admin dashboard statistics.
type AdminStats struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	PendingKYC        int `json:"pending_kyc"`
	TotalWallets      int `json:"total_wallets"`
	TotalTransactions int `json:"total_transactions"`
}

// GetAdminStats retrieves statistics for admin dashboard.
func (s *AuthService) GetAdminStats(ctx context.Context) (*AdminStats, *errors.Error) {
	// Get total users count
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	// Get active users count (users with active status)
	activeUsers, err := s.userRepo.CountByStatus(ctx, models.UserStatusActive)
	if err != nil {
		// If method doesn't exist, return 0 for now
		activeUsers = 0
	}

	// Get pending KYC count
	pendingKYCs, err := s.kycRepo.ListPending(ctx, 1000, 0) // Get all pending
	if err != nil {
		return nil, err
	}

	stats := &AdminStats{
		TotalUsers:        totalUsers,
		ActiveUsers:       activeUsers,
		PendingKYC:        len(pendingKYCs),
		TotalWallets:      0, // TODO: Add wallet service call
		TotalTransactions: 0, // TODO: Add transaction service call
	}

	return stats, nil
}

// hashPassword hashes a password using bcrypt.
func (s *AuthService) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a hash.
func (s *AuthService) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken generates a JWT token for a user.
func (s *AuthService) generateToken(user *models.User, roles []string, permissions []string) (string, int64, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	claims := &JWTClaims{
		UserID:      user.ID,
		Email:       user.Email,
		Status:      string(user.Status),
		Roles:       roles,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nivo-identity",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// hashToken creates a SHA-256 hash of a token for storage.
func (s *AuthService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
