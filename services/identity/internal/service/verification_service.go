package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/shared/crypto"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// VerificationService handles verification request business logic.
type VerificationService struct {
	repo          *repository.VerificationRepository
	userAdminRepo *repository.UserAdminRepository
}

// NewVerificationService creates a new verification service.
func NewVerificationService(
	repo *repository.VerificationRepository,
	userAdminRepo *repository.UserAdminRepository,
) *VerificationService {
	return &VerificationService{
		repo:          repo,
		userAdminRepo: userAdminRepo,
	}
}

// CreateVerification creates a new verification request with OTP.
func (s *VerificationService) CreateVerification(
	ctx context.Context,
	userID string,
	operationType models.OperationType,
	metadata map[string]interface{},
) (*models.VerificationRequest, *errors.Error) {
	// Validate operation type
	if !models.IsValidOperationType(operationType) {
		return nil, errors.BadRequest("invalid operation type")
	}

	// Check for existing pending verification of same type (rate limiting)
	hasRecent, err := s.repo.HasRecentVerification(ctx, userID, operationType, repository.VerificationRateLimit)
	if err != nil {
		return nil, err
	}
	if hasRecent {
		return nil, errors.TooManyRequests("please wait before requesting another verification")
	}

	// Generate OTP
	otpCode, otpErr := crypto.GenerateOTP6()
	if otpErr != nil {
		log.Printf("[identity] Failed to generate OTP: %v", otpErr)
		return nil, errors.Internal("failed to generate verification code")
	}

	// Create verification request
	now := time.Now()
	req := &models.VerificationRequest{
		ID:            "ver_" + uuid.New().String()[:8],
		UserID:        userID,
		OperationType: operationType,
		OTPCode:       otpCode,
		Status:        models.VerificationStatusPending,
		Metadata:      metadata,
		ExpiresAt:     sharedModels.NewTimestamp(now.Add(repository.VerificationTTL)),
		CreatedAt:     sharedModels.NewTimestamp(now),
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	log.Printf("[identity] Verification request created: %s for user %s, operation: %s", req.ID, userID, operationType)

	// Return sanitized version (no OTP) for regular user
	return req.SanitizeForUser(), nil
}

// GetPendingForUserAdmin retrieves pending verifications for User-Admin view.
// Includes OTP codes since this is the User-Admin portal.
func (s *VerificationService) GetPendingForUserAdmin(
	ctx context.Context,
	adminUserID string,
) ([]*models.VerificationRequest, *errors.Error) {
	// Get paired user ID
	pairedUserID, err := s.userAdminRepo.GetPairedUserID(ctx, adminUserID)
	if err != nil {
		return nil, err
	}

	// Get pending verifications (includes OTP)
	return s.repo.GetPendingByUserID(ctx, pairedUserID)
}

// GetUserVerifications retrieves verifications for a user (sanitized, no OTP).
func (s *VerificationService) GetUserVerifications(
	ctx context.Context,
	userID string,
	status string,
) ([]*models.VerificationRequest, *errors.Error) {
	return s.repo.GetByUserID(ctx, userID, status, 50) // Limit to 50 results
}

// GetByID retrieves a verification request by ID.
func (s *VerificationService) GetByID(ctx context.Context, id string) (*models.VerificationRequest, *errors.Error) {
	return s.repo.GetByID(ctx, id)
}

// VerifyOTP validates OTP and returns verification token.
func (s *VerificationService) VerifyOTP(
	ctx context.Context,
	verificationID string,
	userID string,
	otp string,
) (*models.VerificationToken, *errors.Error) {
	// Validate OTP format
	if !crypto.ValidateOTPFormat(otp, 6) {
		return nil, errors.BadRequest("OTP must be 6 digits")
	}

	// Get verification request
	req, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return nil, err
	}

	// Validate ownership
	if req.UserID != userID {
		log.Printf("[identity] User %s tried to verify request %s belonging to %s", userID, verificationID, req.UserID)
		return nil, errors.Forbidden("not authorized to verify this request")
	}

	// Check status
	if !req.IsPending() {
		return nil, errors.BadRequest("verification request is not pending")
	}

	// Check expiry
	if req.IsExpired() {
		_ = s.repo.UpdateStatus(ctx, verificationID, models.VerificationStatusExpired)
		return nil, errors.BadRequest("verification request has expired")
	}

	// Increment attempts
	attempts, err := s.repo.IncrementAttempts(ctx, verificationID)
	if err != nil {
		return nil, err
	}

	// Check max attempts
	if attempts > repository.MaxVerificationAttempts {
		_ = s.repo.UpdateStatus(ctx, verificationID, models.VerificationStatusCancelled)
		return nil, errors.TooManyRequests("too many verification attempts, request cancelled")
	}

	// Validate OTP (constant time comparison to prevent timing attacks)
	if !crypto.SecureCompare(req.OTPCode, otp) {
		remaining := repository.MaxVerificationAttempts - attempts
		if remaining <= 0 {
			_ = s.repo.UpdateStatus(ctx, verificationID, models.VerificationStatusCancelled)
			return nil, errors.BadRequest("invalid OTP code, verification cancelled")
		}
		return nil, errors.BadRequest(fmt.Sprintf("invalid OTP code (%d attempts remaining)", remaining))
	}

	// Mark as verified
	if err := s.repo.UpdateStatus(ctx, verificationID, models.VerificationStatusVerified); err != nil {
		return nil, err
	}

	log.Printf("[identity] Verification %s verified successfully for user %s", verificationID, userID)

	// Generate verification token (short-lived JWT)
	token, tokenErr := s.generateVerificationToken(req)
	if tokenErr != nil {
		return nil, tokenErr
	}

	return &models.VerificationToken{
		Token:         token,
		OperationType: req.OperationType,
		Metadata:      req.Metadata,
		ExpiresAt:     time.Now().Add(5 * time.Minute), // Token valid for 5 min
	}, nil
}

// generateVerificationToken creates a short-lived JWT for the verified operation.
func (s *VerificationService) generateVerificationToken(req *models.VerificationRequest) (string, *errors.Error) {
	claims := &models.VerificationClaims{
		VerificationID: req.ID,
		UserID:         req.UserID,
		OperationType:  req.OperationType,
		Metadata:       req.Metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   req.UserID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}

	tokenString, jwtErr := token.SignedString([]byte(secret))
	if jwtErr != nil {
		log.Printf("[identity] Failed to generate verification token: %v", jwtErr)
		return "", errors.Internal("failed to generate verification token")
	}

	return tokenString, nil
}

// ValidateVerificationToken validates a verification token for an operation.
func (s *VerificationService) ValidateVerificationToken(
	ctx context.Context,
	tokenString string,
	expectedOperation models.OperationType,
	expectedUserID string,
) (*models.VerificationClaims, *errors.Error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}

	token, parseErr := jwt.ParseWithClaims(tokenString, &models.VerificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if parseErr != nil || !token.Valid {
		return nil, errors.Unauthorized("invalid or expired verification token")
	}

	claims, ok := token.Claims.(*models.VerificationClaims)
	if !ok {
		return nil, errors.Unauthorized("invalid verification token claims")
	}

	// Validate operation type
	if claims.OperationType != expectedOperation {
		return nil, errors.Forbidden("verification token is for different operation")
	}

	// Validate user
	if claims.UserID != expectedUserID {
		return nil, errors.Forbidden("verification token belongs to different user")
	}

	return claims, nil
}

// CancelVerification cancels a pending verification.
func (s *VerificationService) CancelVerification(
	ctx context.Context,
	verificationID string,
	userID string,
) *errors.Error {
	req, err := s.repo.GetByID(ctx, verificationID)
	if err != nil {
		return err
	}

	if req.UserID != userID {
		return errors.Forbidden("not authorized to cancel this verification")
	}

	if !req.IsPending() {
		return errors.BadRequest("verification is not pending")
	}

	log.Printf("[identity] Verification %s cancelled by user %s", verificationID, userID)
	return s.repo.UpdateStatus(ctx, verificationID, models.VerificationStatusCancelled)
}

// CancelAllPendingForUser cancels all pending verifications for a user.
func (s *VerificationService) CancelAllPendingForUser(ctx context.Context, userID string) *errors.Error {
	return s.repo.CancelPendingForUser(ctx, userID)
}

// CountPendingForUser returns the count of pending verifications for a user.
func (s *VerificationService) CountPendingForUser(ctx context.Context, userID string) (int, *errors.Error) {
	return s.repo.CountPendingByUserID(ctx, userID)
}

// CanUserAdminAccessVerification checks if a User-Admin can access a specific verification.
// Returns true only if the User-Admin is paired with the verification's owner.
func (s *VerificationService) CanUserAdminAccessVerification(
	ctx context.Context,
	adminUserID string,
	verification *models.VerificationRequest,
) (bool, *errors.Error) {
	// Get the paired user ID for this User-Admin
	pairedUserID, err := s.userAdminRepo.GetPairedUserID(ctx, adminUserID)
	if err != nil {
		return false, err
	}

	// Check if the verification belongs to the paired user
	return verification.UserID == pairedUserID, nil
}
