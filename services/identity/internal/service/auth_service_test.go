package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// =====================================================================
// Mock Repositories for Testing
// =====================================================================

type mockUserRepository struct {
	users          map[string]*models.User
	emailIndex     map[string]*models.User
	phoneIndex     map[string]*models.User
	createFunc     func(ctx context.Context, user *models.User) *errors.Error
	getByEmailFunc func(ctx context.Context, email string) (*models.User, *errors.Error)
	getByPhoneFunc func(ctx context.Context, phone string) (*models.User, *errors.Error)
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	// Check email uniqueness
	if _, exists := m.emailIndex[user.Email]; exists {
		return errors.Conflict("email already exists")
	}
	// Check phone uniqueness (skip for empty phones - User-Admin accounts)
	if user.Phone != "" {
		if _, exists := m.phoneIndex[user.Phone]; exists {
			return errors.Conflict("phone already exists")
		}
	}
	user.ID = uuid.New().String()
	user.CreatedAt = sharedModels.NewTimestamp(time.Now())
	user.UpdatedAt = user.CreatedAt
	m.users[user.ID] = user
	m.emailIndex[user.Email] = user
	if user.Phone != "" {
		m.phoneIndex[user.Phone] = user
	}
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, *errors.Error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	user, ok := m.emailIndex[email]
	if !ok {
		return nil, errors.NotFound("user")
	}
	return user, nil
}

func (m *mockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, *errors.Error) {
	if m.getByPhoneFunc != nil {
		return m.getByPhoneFunc(ctx, phone)
	}
	user, ok := m.phoneIndex[phone]
	if !ok {
		return nil, errors.NotFound("user")
	}
	return user, nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, *errors.Error) {
	user, ok := m.users[id]
	if !ok {
		return nil, errors.NotFound("user")
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) *errors.Error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.NotFound("user")
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) *errors.Error {
	user, ok := m.users[userID]
	if !ok {
		return errors.NotFound("user")
	}
	user.PasswordHash = passwordHash
	return nil
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) *errors.Error {
	user, ok := m.users[userID]
	if !ok {
		return errors.NotFound("user")
	}
	user.Status = status
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, userID string) *errors.Error {
	return m.UpdateStatus(ctx, userID, models.UserStatusClosed)
}

func (m *mockUserRepository) Count(ctx context.Context) (int, *errors.Error) {
	return len(m.users), nil
}

func (m *mockUserRepository) CountByStatus(ctx context.Context, status models.UserStatus) (int, *errors.Error) {
	count := 0
	for _, user := range m.users {
		if user.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*models.User, *errors.Error) {
	results := make([]*models.User, 0)
	queryLower := strings.ToLower(query)

	for _, user := range m.users {
		// Simple contains search for testing
		if strings.Contains(strings.ToLower(user.Email), queryLower) ||
			strings.Contains(strings.ToLower(user.Phone), queryLower) ||
			strings.Contains(strings.ToLower(user.FullName), queryLower) {
			results = append(results, user)
		}
	}

	// Apply pagination
	start := offset
	if start > len(results) {
		return []*models.User{}, nil
	}

	end := start + limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

func (m *mockUserRepository) SuspendUser(ctx context.Context, userID string, reason string, suspendedBy string) *errors.Error {
	user, exists := m.users[userID]
	if !exists {
		return errors.NotFoundWithID("user", userID)
	}
	if user.Status == models.UserStatusClosed {
		return errors.BadRequest("cannot suspend a closed account")
	}
	user.Status = models.UserStatusSuspended
	now := sharedModels.NewTimestamp(time.Now())
	user.SuspendedAt = &now
	user.SuspensionReason = &reason
	user.SuspendedBy = &suspendedBy
	return nil
}

func (m *mockUserRepository) UnsuspendUser(ctx context.Context, userID string) *errors.Error {
	user, exists := m.users[userID]
	if !exists {
		return errors.NotFoundWithID("user", userID)
	}
	if user.Status != models.UserStatusSuspended {
		return errors.BadRequest("user is not suspended")
	}
	user.Status = models.UserStatusActive
	user.SuspendedAt = nil
	user.SuspensionReason = nil
	user.SuspendedBy = nil
	return nil
}

type mockKYCRepository struct {
	kycData         map[string]*models.KYCInfo
	getByUserIDFunc func(ctx context.Context, userID string) (*models.KYCInfo, *errors.Error)
}

func (m *mockKYCRepository) GetByUserID(ctx context.Context, userID string) (*models.KYCInfo, *errors.Error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	kyc, ok := m.kycData[userID]
	if !ok {
		return nil, errors.NotFound("KYC data")
	}
	return kyc, nil
}

func (m *mockKYCRepository) Create(ctx context.Context, kyc *models.KYCInfo) *errors.Error {
	m.kycData[kyc.UserID] = kyc
	return nil
}

func (m *mockKYCRepository) UpdateStatus(ctx context.Context, userID string, status models.KYCStatus, reason string) *errors.Error {
	kyc, ok := m.kycData[userID]
	if !ok {
		return errors.NotFound("KYC data")
	}
	kyc.Status = status
	return nil
}

func (m *mockKYCRepository) ListPending(ctx context.Context, limit, offset int) ([]repository.KYCWithUser, *errors.Error) {
	// Not needed for current tests, return empty
	return []repository.KYCWithUser{}, nil
}

type mockSessionRepository struct {
	sessions           map[string]*models.Session
	tokenIndex         map[string]*models.Session
	createFunc         func(ctx context.Context, session *models.Session) *errors.Error
	getByTokenHashFunc func(ctx context.Context, tokenHash string) (*models.Session, *errors.Error)
}

func (m *mockSessionRepository) Create(ctx context.Context, session *models.Session) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, session)
	}
	session.ID = uuid.New().String()
	session.CreatedAt = sharedModels.NewTimestamp(time.Now())
	m.sessions[session.ID] = session
	m.tokenIndex[session.Token] = session
	return nil
}

func (m *mockSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, *errors.Error) {
	if m.getByTokenHashFunc != nil {
		return m.getByTokenHashFunc(ctx, tokenHash)
	}
	session, ok := m.tokenIndex[tokenHash]
	if !ok {
		return nil, errors.NotFound("session")
	}
	// Check if expired
	if time.Now().After(session.ExpiresAt.Time) {
		return nil, errors.Unauthorized("session expired")
	}
	return session, nil
}

func (m *mockSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) *errors.Error {
	session, ok := m.tokenIndex[tokenHash]
	if !ok {
		return errors.NotFound("session")
	}
	delete(m.sessions, session.ID)
	delete(m.tokenIndex, tokenHash)
	return nil
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) *errors.Error {
	// Delete all sessions for user
	for id, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, id)
			delete(m.tokenIndex, session.Token)
		}
	}
	return nil
}

type mockRBACClient struct {
	assignDefaultRoleFunc  func(ctx context.Context, userID string) error
	getUserPermissionsFunc func(ctx context.Context, userID string) (*UserPermissionsResponse, error)
}

func (m *mockRBACClient) AssignDefaultRole(ctx context.Context, userID string) error {
	if m.assignDefaultRoleFunc != nil {
		return m.assignDefaultRoleFunc(ctx, userID)
	}
	return nil
}

func (m *mockRBACClient) GetUserPermissions(ctx context.Context, userID string) (*UserPermissionsResponse, error) {
	if m.getUserPermissionsFunc != nil {
		return m.getUserPermissionsFunc(ctx, userID)
	}
	// Return default permissions
	return &UserPermissionsResponse{
		UserID: userID,
		Roles: []RoleInfo{
			{ID: "role-1", Name: "user"},
		},
		Permissions: []Permission{
			{ID: "perm-1", Name: "wallet:read"},
		},
	}, nil
}

func (m *mockRBACClient) AssignUserAdminRole(ctx context.Context, userID string) error {
	// Default: succeed silently
	return nil
}

type mockUserAdminRepository struct {
	pairings map[string]string // userID -> adminUserID and adminUserID -> userID
}

func (m *mockUserAdminRepository) CreatePairing(ctx context.Context, userID, adminUserID string) *errors.Error {
	if m.pairings == nil {
		m.pairings = make(map[string]string)
	}
	m.pairings[userID] = adminUserID
	m.pairings[adminUserID] = userID
	return nil
}

func (m *mockUserAdminRepository) GetPairedUserID(ctx context.Context, adminUserID string) (string, *errors.Error) {
	if userID, ok := m.pairings[adminUserID]; ok {
		// Check if this is actually an admin user (by checking if the stored value points back)
		if storedAdmin, exists := m.pairings[userID]; exists && storedAdmin == adminUserID {
			return userID, nil
		}
	}
	return "", errors.NotFound("no paired user found")
}

func (m *mockUserAdminRepository) GetAdminUserID(ctx context.Context, userID string) (string, *errors.Error) {
	if adminUserID, ok := m.pairings[userID]; ok {
		return adminUserID, nil
	}
	return "", errors.NotFound("no user-admin account found")
}

func (m *mockUserAdminRepository) IsUserAdmin(ctx context.Context, userID string) (bool, *errors.Error) {
	// Check if this userID is stored as an admin user ID
	_, ok := m.pairings[userID]
	return ok, nil
}

func (m *mockUserAdminRepository) ValidatePairing(ctx context.Context, adminUserID, userID string) (bool, *errors.Error) {
	if storedAdminID, ok := m.pairings[userID]; ok {
		return storedAdminID == adminUserID, nil
	}
	return false, nil
}

// =====================================================================
// Test Helpers
// =====================================================================

// Compile-time interface checks
var _ UserRepositoryInterface = (*mockUserRepository)(nil)
var _ UserAdminRepositoryInterface = (*mockUserAdminRepository)(nil)
var _ KYCRepositoryInterface = (*mockKYCRepository)(nil)
var _ SessionRepositoryInterface = (*mockSessionRepository)(nil)
var _ RBACClientInterface = (*mockRBACClient)(nil)

func setupTestAuthService() (*AuthService, *mockUserRepository, *mockKYCRepository, *mockSessionRepository, *mockRBACClient) {
	userRepo := &mockUserRepository{
		users:      make(map[string]*models.User),
		emailIndex: make(map[string]*models.User),
		phoneIndex: make(map[string]*models.User),
	}
	userAdminRepo := &mockUserAdminRepository{
		pairings: make(map[string]string),
	}
	kycRepo := &mockKYCRepository{
		kycData: make(map[string]*models.KYCInfo),
	}
	sessionRepo := &mockSessionRepository{
		sessions:   make(map[string]*models.Session),
		tokenIndex: make(map[string]*models.Session),
	}
	rbacClient := &mockRBACClient{}

	service := NewAuthService(
		userRepo,
		userAdminRepo,
		kycRepo,
		sessionRepo,
		rbacClient,
		nil, // wallet client (nil for tests)
		nil, // notification client (nil for tests)
		"test-secret-key-for-jwt-signing",
		24*time.Hour, // 24 hour token expiry
		nil,          // event publisher (nil for tests)
	)

	return service, userRepo, kycRepo, sessionRepo, rbacClient
}

func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// =====================================================================
// Register Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestRegister_Success(t *testing.T) {
	service, _, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Password: "SecurePassword123!",
	}

	user, err := service.Register(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
	if user.Phone != req.Phone {
		t.Errorf("expected phone %s, got %s", req.Phone, user.Phone)
	}
	if user.FullName != req.FullName {
		t.Errorf("expected full name %s, got %s", req.FullName, user.FullName)
	}
	if user.Status != models.UserStatusPending {
		t.Errorf("expected pending status, got %s", user.Status)
	}
	// Password should be sanitized (not returned)
	if user.PasswordHash != "" {
		t.Errorf("password hash should be sanitized, got %s", user.PasswordHash)
	}
	if user.ID == "" {
		t.Error("expected user ID to be set")
	}
}

func TestRegister_Error_DuplicateEmail(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create existing user
	existingUser := &models.User{
		ID:           uuid.New().String(),
		Email:        "existing@example.com",
		PasswordHash: hashPassword("password"),
		Status:       models.UserStatusActive,
	}
	userRepo.users[existingUser.ID] = existingUser
	userRepo.emailIndex[existingUser.Email] = existingUser
	userRepo.phoneIndex[existingUser.Phone] = existingUser

	// Try to register with same email
	req := &models.CreateUserRequest{
		Email:    "existing@example.com", // Duplicate
		Phone:    "+919876543210",
		FullName: "New User",
		Password: "SecurePassword123!",
	}

	_, err := service.Register(ctx, req)
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if err.Code != errors.ErrCodeConflict {
		t.Errorf("expected conflict error, got %s", err.Code)
	}
}

func TestRegister_RBACAssignmentFailure_ShouldFail(t *testing.T) {
	service, _, _, _, rbacClient := setupTestAuthService()
	ctx := context.Background()

	// Make RBAC assignment fail
	rbacClient.assignDefaultRoleFunc = func(ctx context.Context, userID string) error {
		return errors.Internal("RBAC service unavailable")
	}

	req := &models.CreateUserRequest{
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Password: "SecurePassword123!",
	}

	// With paired User-Admin accounts, registration should fail if RBAC fails
	// This ensures both accounts have proper roles assigned
	_, err := service.Register(ctx, req)
	if err == nil {
		t.Fatal("expected registration to fail when RBAC fails")
	}
	if err.Code != errors.ErrCodeInternal {
		t.Errorf("expected internal error, got %s", err.Code)
	}
}

// =====================================================================
// Login Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestLogin_Success(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user with known password
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		Phone:        "+919876543210",
		FullName:     "Test User",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Login
	req := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}

	response, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Token == "" {
		t.Error("expected token to be generated")
	}
	if response.ExpiresAt == 0 {
		t.Error("expected expiry time to be set")
	}
	if response.User.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, response.User.Email)
	}
	// Password should be sanitized
	if response.User.PasswordHash != "" {
		t.Error("password hash should be sanitized")
	}
}

func TestLogin_Error_InvalidEmail(t *testing.T) {
	service, _, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	req := &models.LoginRequest{
		Identifier: "nonexistent@example.com",
		Password:   "password",
	}

	_, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
	if err.Code != errors.ErrCodeUnauthorized {
		t.Errorf("expected unauthorized error, got %s", err.Code)
	}
	// Should not reveal if user exists
	if err.Message != "invalid credentials" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestLogin_Error_InvalidPassword(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword("CorrectPassword123!"),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Try to login with wrong password
	req := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   "WrongPassword",
	}

	_, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error for invalid password, got nil")
	}
	if err.Code != errors.ErrCodeUnauthorized {
		t.Errorf("expected unauthorized error, got %s", err.Code)
	}
	if err.Message != "invalid credentials" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestLogin_Error_ClosedAccount(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusClosed, // Closed account
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	req := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}

	_, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error for closed account, got nil")
	}
	if err.Code != errors.ErrCodeForbidden {
		t.Errorf("expected forbidden error, got %s", err.Code)
	}
	if err.Message != "account is closed" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestLogin_Error_SuspendedAccount(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusSuspended, // Suspended account
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	req := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}

	_, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error for suspended account, got nil")
	}
	if err.Code != errors.ErrCodeForbidden {
		t.Errorf("expected forbidden error, got %s", err.Code)
	}
	if err.Message != "account is suspended" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestLogin_WithPhone_Success(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user with known password
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "phonetest@example.com",
		Phone:        "+919876543210",
		FullName:     "Phone Test User",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Login with phone number
	req := &models.LoginRequest{
		Identifier: "+919876543210",
		Password:   password,
	}

	response, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected no error for phone login, got %v", err)
	}

	if response.Token == "" {
		t.Error("expected token to be generated")
	}
	if response.User.Phone != user.Phone {
		t.Errorf("expected phone %s, got %s", user.Phone, response.User.Phone)
	}
	if response.User.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, response.User.Email)
	}
}

func TestLogin_RBACFailure_GracefulDegradation(t *testing.T) {
	service, userRepo, _, _, rbacClient := setupTestAuthService()
	ctx := context.Background()

	// Make RBAC fail
	rbacClient.getUserPermissionsFunc = func(ctx context.Context, userID string) (*UserPermissionsResponse, error) {
		return nil, errors.Internal("RBAC service down")
	}

	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	req := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}

	// Should succeed despite RBAC failure (graceful degradation)
	response, err := service.Login(ctx, req, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected login to succeed despite RBAC failure, got %v", err)
	}
	if response.Token == "" {
		t.Error("expected token to be generated")
	}
}

// =====================================================================
// Logout Tests
// =====================================================================

func TestLogout_Success(t *testing.T) {
	service, userRepo, _, sessionRepo, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user and login to get token
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	loginReq := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}
	loginResp, _ := service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")

	// Verify session exists
	if len(sessionRepo.sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessionRepo.sessions))
	}

	// Logout
	err := service.Logout(ctx, loginResp.Token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify session deleted
	if len(sessionRepo.sessions) != 0 {
		t.Errorf("expected 0 sessions after logout, got %d", len(sessionRepo.sessions))
	}
}

func TestLogout_Error_InvalidToken(t *testing.T) {
	service, _, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	err := service.Logout(ctx, "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// LogoutAll Tests
// =====================================================================

func TestLogoutAll_Success(t *testing.T) {
	service, userRepo, _, sessionRepo, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Login multiple times to create multiple sessions
	loginReq := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}
	_, _ = service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")
	_, _ = service.Login(ctx, loginReq, "192.168.1.2", "Chrome")
	_, _ = service.Login(ctx, loginReq, "192.168.1.3", "Safari")

	// Verify at least one session exists (actual count may vary based on token generation)
	initialCount := len(sessionRepo.sessions)
	if initialCount == 0 {
		t.Fatal("expected at least one session to exist before logout")
	}

	// Logout all
	err := service.LogoutAll(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify all sessions deleted
	if len(sessionRepo.sessions) != 0 {
		t.Errorf("expected 0 sessions after logout all, got %d", len(sessionRepo.sessions))
	}
}

// =====================================================================
// ValidateToken Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestValidateToken_Success(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user and login
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		Phone:        "+919876543210",
		FullName:     "Test User",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	loginReq := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}
	loginResp, _ := service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")

	// Validate token
	validatedUser, err := service.ValidateToken(ctx, loginResp.Token)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if validatedUser.ID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, validatedUser.ID)
	}
	if validatedUser.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, validatedUser.Email)
	}
}

func TestValidateToken_Error_InvalidToken(t *testing.T) {
	service, _, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	_, err := service.ValidateToken(ctx, "invalid.jwt.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if err.Code != errors.ErrCodeUnauthorized {
		t.Errorf("expected unauthorized error, got %s", err.Code)
	}
}

func TestValidateToken_Error_ExpiredSession(t *testing.T) {
	service, userRepo, _, sessionRepo, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user and login
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	loginReq := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}
	loginResp, _ := service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")

	// Manually expire the session
	tokenHash := service.hashToken(loginResp.Token)
	session := sessionRepo.tokenIndex[tokenHash]
	session.ExpiresAt = sharedModels.NewTimestamp(time.Now().Add(-1 * time.Hour)) // 1 hour ago

	// Try to validate
	_, err := service.ValidateToken(ctx, loginResp.Token)
	if err == nil {
		t.Fatal("expected error for expired session, got nil")
	}
	if err.Code != errors.ErrCodeUnauthorized {
		t.Errorf("expected unauthorized error, got %s", err.Code)
	}
}

func TestValidateToken_Error_DeletedSession(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create user and login
	password := "TestPassword123!"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        "test@example.com",
		PasswordHash: hashPassword(password),
		Status:       models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	loginReq := &models.LoginRequest{
		Identifier: "test@example.com",
		Password:   password,
	}
	loginResp, _ := service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")

	// Logout (delete session)
	_ = service.Logout(ctx, loginResp.Token)

	// Try to validate
	_, err := service.ValidateToken(ctx, loginResp.Token)
	if err == nil {
		t.Fatal("expected error for deleted session, got nil")
	}
	// Accept either Unauthorized or NotFound (both prevent authentication)
	if err.Code != errors.ErrCodeUnauthorized && err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected unauthorized or not found error, got %s", err.Code)
	}
}

// =====================================================================
// User Suspension Tests
// =====================================================================

func TestSuspendUser_Success(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create test user
	email := "test@example.com"
	password := "password123"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		Phone:        "+919876543210",
		FullName:     "Test User",
		Status:       models.UserStatusActive,
		PasswordHash: hashPassword(password),
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Create admin user
	adminUser := &models.User{
		ID:       uuid.New().String(),
		Email:    "admin@example.com",
		Phone:    "+919876543211",
		FullName: "Admin User",
		Status:   models.UserStatusActive,
	}
	userRepo.users[adminUser.ID] = adminUser
	userRepo.emailIndex[adminUser.Email] = adminUser
	userRepo.phoneIndex[adminUser.Phone] = adminUser

	// Suspend the user
	reason := "Suspicious activity detected"
	err := service.SuspendUser(ctx, user.ID, reason, adminUser.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify user is suspended
	suspendedUser, _ := userRepo.GetByID(ctx, user.ID)
	if suspendedUser.Status != models.UserStatusSuspended {
		t.Errorf("expected status suspended, got %s", suspendedUser.Status)
	}
	if suspendedUser.SuspensionReason == nil || *suspendedUser.SuspensionReason != reason {
		t.Errorf("expected suspension reason %s, got %v", reason, suspendedUser.SuspensionReason)
	}
	if suspendedUser.SuspendedBy == nil || *suspendedUser.SuspendedBy != adminUser.ID {
		t.Errorf("expected suspended_by %s, got %v", adminUser.ID, suspendedUser.SuspendedBy)
	}
	if suspendedUser.SuspendedAt == nil {
		t.Error("expected suspended_at to be set")
	}
}

func TestSuspendUser_AlreadySuspended(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create suspended user
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Status:   models.UserStatusSuspended,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Try to suspend again
	err := service.SuspendUser(ctx, user.ID, "Another reason", "admin-id")
	if err == nil {
		t.Fatal("expected error for already suspended user")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestSuspendUser_ClosedAccount(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create closed user
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Status:   models.UserStatusClosed,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Try to suspend
	err := service.SuspendUser(ctx, user.ID, "Reason", "admin-id")
	if err == nil {
		t.Fatal("expected error for closed account")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestSuspendUser_UserNotFound(t *testing.T) {
	service, _, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Try to suspend non-existent user
	err := service.SuspendUser(ctx, "non-existent-id", "Reason", "admin-id")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

func TestUnsuspendUser_Success(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create suspended user
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Status:   models.UserStatusSuspended,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Set suspension fields manually in mock
	now := sharedModels.NewTimestamp(time.Now())
	reason := "Suspended for testing"
	adminID := "admin-123"
	user.SuspendedAt = &now
	user.SuspensionReason = &reason
	user.SuspendedBy = &adminID

	// Unsuspend the user
	err := service.UnsuspendUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify user is active again
	unsuspendedUser, _ := userRepo.GetByID(ctx, user.ID)
	if unsuspendedUser.Status != models.UserStatusActive {
		t.Errorf("expected status active, got %s", unsuspendedUser.Status)
	}
	if unsuspendedUser.SuspendedAt != nil {
		t.Error("expected suspended_at to be nil")
	}
	if unsuspendedUser.SuspensionReason != nil {
		t.Error("expected suspension_reason to be nil")
	}
	if unsuspendedUser.SuspendedBy != nil {
		t.Error("expected suspended_by to be nil")
	}
}

func TestUnsuspendUser_NotSuspended(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create active user
	user := &models.User{
		ID:       uuid.New().String(),
		Email:    "test@example.com",
		Phone:    "+919876543210",
		FullName: "Test User",
		Status:   models.UserStatusActive,
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Try to unsuspend
	err := service.UnsuspendUser(ctx, user.ID)
	if err == nil {
		t.Fatal("expected error for non-suspended user")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestLogin_SuspendedUser(t *testing.T) {
	service, userRepo, _, _, _ := setupTestAuthService()
	ctx := context.Background()

	// Create suspended user
	email := "test@example.com"
	password := "password123"
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		Phone:        "+919876543210",
		FullName:     "Test User",
		Status:       models.UserStatusSuspended,
		PasswordHash: hashPassword(password),
	}
	userRepo.users[user.ID] = user
	userRepo.emailIndex[user.Email] = user
	userRepo.phoneIndex[user.Phone] = user

	// Try to login
	loginReq := &models.LoginRequest{
		Identifier: email,
		Password:   password,
	}
	_, err := service.Login(ctx, loginReq, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error for suspended user login")
	}
	if err.Code != errors.ErrCodeForbidden {
		t.Errorf("expected forbidden error, got %s", err.Code)
	}
	if err.Message != "account is suspended" {
		t.Errorf("expected 'account is suspended' message, got %s", err.Message)
	}
}
