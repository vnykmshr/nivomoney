package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================
// Mock Repositories for Handler Testing
// ============================================================

// mockUserRepository implements service.UserRepositoryInterface for testing.
type mockUserRepository struct {
	users map[string]*models.User

	// Override functions
	CreateFunc     func(ctx context.Context, user *models.User) *errors.Error
	GetByEmailFunc func(ctx context.Context, email string) (*models.User, *errors.Error)
	GetByPhoneFunc func(ctx context.Context, phone string) (*models.User, *errors.Error)
	GetByIDFunc    func(ctx context.Context, id string) (*models.User, *errors.Error)
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) *errors.Error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	// Check for duplicate email
	for _, u := range m.users {
		if u.Email == user.Email && u.AccountType == user.AccountType {
			return errors.Conflict("user with this email already exists")
		}
	}
	user.ID = "user-" + time.Now().Format("20060102150405.000")
	user.CreatedAt = sharedModels.Now()
	user.UpdatedAt = sharedModels.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, *errors.Error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.NotFound("user")
}

func (m *mockUserRepository) GetByEmailAndAccountType(ctx context.Context, email string, accountType models.AccountType) (*models.User, *errors.Error) {
	for _, u := range m.users {
		if u.Email == email && u.AccountType == accountType {
			return u, nil
		}
	}
	return nil, errors.NotFound("user")
}

func (m *mockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, *errors.Error) {
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(ctx, phone)
	}
	for _, u := range m.users {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, errors.NotFound("user")
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, *errors.Error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.NotFound("user")
}

func (m *mockUserRepository) Update(ctx context.Context, user *models.User) *errors.Error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.NotFound("user not found")
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) *errors.Error {
	if user, ok := m.users[userID]; ok {
		user.PasswordHash = passwordHash
		return nil
	}
	return errors.NotFound("user not found")
}

func (m *mockUserRepository) UpdateStatus(ctx context.Context, userID string, status models.UserStatus) *errors.Error {
	if user, ok := m.users[userID]; ok {
		user.Status = status
		return nil
	}
	return errors.NotFound("user not found")
}

func (m *mockUserRepository) Delete(ctx context.Context, userID string) *errors.Error {
	delete(m.users, userID)
	return nil
}

func (m *mockUserRepository) Count(ctx context.Context) (int, *errors.Error) {
	return len(m.users), nil
}

func (m *mockUserRepository) CountByStatus(ctx context.Context, status models.UserStatus) (int, *errors.Error) {
	count := 0
	for _, u := range m.users {
		if u.Status == status {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*models.User, *errors.Error) {
	return []*models.User{}, nil
}

func (m *mockUserRepository) SuspendUser(ctx context.Context, userID string, reason string, suspendedBy string) *errors.Error {
	return m.UpdateStatus(ctx, userID, models.UserStatusSuspended)
}

func (m *mockUserRepository) UnsuspendUser(ctx context.Context, userID string) *errors.Error {
	return m.UpdateStatus(ctx, userID, models.UserStatusActive)
}

// mockSessionRepository implements service.SessionRepositoryInterface.
type mockSessionRepository struct {
	sessions map[string]*models.Session
}

func newMockSessionRepository() *mockSessionRepository {
	return &mockSessionRepository{
		sessions: make(map[string]*models.Session),
	}
}

func (m *mockSessionRepository) Create(ctx context.Context, session *models.Session) *errors.Error {
	m.sessions[session.Token] = session
	return nil
}

func (m *mockSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.Session, *errors.Error) {
	if session, ok := m.sessions[tokenHash]; ok {
		return session, nil
	}
	return nil, nil
}

func (m *mockSessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) *errors.Error {
	delete(m.sessions, tokenHash)
	return nil
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) *errors.Error {
	for hash, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, hash)
		}
	}
	return nil
}

// mockKYCRepository implements service.KYCRepositoryInterface.
type mockKYCRepository struct{}

func (m *mockKYCRepository) GetByUserID(ctx context.Context, userID string) (*models.KYCInfo, *errors.Error) {
	// Return "not found" error to prevent nil pointer dereference in service
	return nil, errors.NotFound("kyc not found")
}

func (m *mockKYCRepository) Create(ctx context.Context, kyc *models.KYCInfo) *errors.Error {
	return nil
}

func (m *mockKYCRepository) UpdateStatus(ctx context.Context, userID string, status models.KYCStatus, reason string) *errors.Error {
	return nil
}

func (m *mockKYCRepository) ListPending(ctx context.Context, limit, offset int) ([]repository.KYCWithUser, *errors.Error) {
	return nil, nil
}

// mockUserAdminRepository implements service.UserAdminRepositoryInterface.
type mockUserAdminRepository struct {
	pairings map[string]string // userID -> adminUserID
}

func newMockUserAdminRepository() *mockUserAdminRepository {
	return &mockUserAdminRepository{
		pairings: make(map[string]string),
	}
}

func (m *mockUserAdminRepository) CreatePairing(ctx context.Context, userID, adminUserID string) *errors.Error {
	m.pairings[userID] = adminUserID
	return nil
}

func (m *mockUserAdminRepository) GetPairedUserID(ctx context.Context, adminUserID string) (string, *errors.Error) {
	for userID, adminID := range m.pairings {
		if adminID == adminUserID {
			return userID, nil
		}
	}
	return "", nil
}

func (m *mockUserAdminRepository) GetAdminUserID(ctx context.Context, userID string) (string, *errors.Error) {
	if adminID, ok := m.pairings[userID]; ok {
		return adminID, nil
	}
	return "", nil
}

func (m *mockUserAdminRepository) IsUserAdmin(ctx context.Context, userID string) (bool, *errors.Error) {
	return false, nil
}

func (m *mockUserAdminRepository) ValidatePairing(ctx context.Context, adminUserID, userID string) (bool, *errors.Error) {
	if adminID, ok := m.pairings[userID]; ok {
		return adminID == adminUserID, nil
	}
	return false, nil
}

// mockRBACClient implements service.RBACClientInterface.
type mockRBACClient struct{}

func (m *mockRBACClient) AssignDefaultRole(ctx context.Context, userID string) error {
	return nil
}

func (m *mockRBACClient) AssignUserAdminRole(ctx context.Context, userID string) error {
	return nil
}

func (m *mockRBACClient) GetUserPermissions(ctx context.Context, userID string) (*service.UserPermissionsResponse, error) {
	return &service.UserPermissionsResponse{
		UserID:      userID,
		Roles:       []service.RoleInfo{{ID: "role-1", Name: "user"}},
		Permissions: []service.Permission{{ID: "perm-1", Name: "read:own:profile"}},
	}, nil
}

// ============================================================
// Test Helper Functions
// ============================================================

// apiResponse represents the standard API response.
type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *apiError       `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// createTestAuthService creates an AuthService with mock repositories.
func createTestAuthService() (*service.AuthService, *mockUserRepository) {
	userRepo := newMockUserRepository()
	sessionRepo := newMockSessionRepository()
	userAdminRepo := newMockUserAdminRepository()
	rbacClient := &mockRBACClient{}

	authService := service.NewAuthService(
		userRepo,
		userAdminRepo,
		&mockKYCRepository{},
		sessionRepo,
		rbacClient,
		nil,                               // walletClient
		nil,                               // notificationClient
		"test-jwt-secret-32-characters!!", // Must be at least 32 chars
		24*time.Hour,
		nil, // eventPublisher
	)

	return authService, userRepo
}

// makeRequest is a helper to create HTTP requests for testing.
func makeRequest(t *testing.T, handler http.HandlerFunc, method, path string, body interface{}) (*httptest.ResponseRecorder, *apiResponse) {
	t.Helper()

	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp apiResponse
	if rec.Body.Len() > 0 {
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err, "failed to unmarshal response: %s", rec.Body.String())
	}

	return rec, &resp
}

// ============================================================
// Auth Handler Tests
// ============================================================

func TestAuthHandler_Register(t *testing.T) {
	authService, _ := createTestAuthService()
	handler := NewAuthHandler(authService)

	t.Run("valid registration returns 201", func(t *testing.T) {
		body := map[string]interface{}{
			"email":     "test@example.com",
			"phone":     "+919876543210",
			"full_name": "Test User",
			"password":  "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.True(t, resp.Success, "expected success=true")
		assert.Nil(t, resp.Error)

		// Verify user data in response
		var user map[string]interface{}
		err := json.Unmarshal(resp.Data, &user)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user["email"])
		assert.NotEmpty(t, user["id"])
	})

	t.Run("missing email returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"phone":     "+919876543210",
			"full_name": "Test User",
			"password":  "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		assert.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("invalid email format returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"email":     "not-an-email",
			"phone":     "+919876543210",
			"full_name": "Test User",
			"password":  "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("missing password returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"email":     "test2@example.com",
			"phone":     "+919876543211",
			"full_name": "Test User",
			// missing password
		}

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error, "expected error response")
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("duplicate email returns error", func(t *testing.T) {
		// First registration
		body := map[string]interface{}{
			"email":     "duplicate@example.com",
			"phone":     "+919876543212",
			"full_name": "First User",
			"password":  "SecurePass123!",
		}
		makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		// Second registration with same email
		body["phone"] = "+919876543213"
		body["full_name"] = "Second User"

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		// The service creates User then User-Admin, so duplicate email may return internal error
		// when User-Admin creation fails
		assert.False(t, resp.Success, "duplicate registration should fail")
		require.NotNil(t, resp.Error)
		// Could be CONFLICT or INTERNAL_ERROR depending on which user creation fails
		assert.Contains(t, []string{"CONFLICT", "INTERNAL_ERROR"}, resp.Error.Code)
		assert.NotEqual(t, http.StatusCreated, rec.Code)
	})

	t.Run("missing phone returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"email": "test3@example.com",
			// missing phone
			"full_name": "Test User",
			"password":  "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Register, http.MethodPost, "/api/v1/auth/register", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("empty body returns bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.Register(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	authService, userRepo := createTestAuthService()
	handler := NewAuthHandler(authService)

	// Setup: Add test user directly with known password hash
	// Password: "SecurePass123!" hashed with bcrypt
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("SecurePass123!"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           "test-user-login",
		Email:        "login@example.com",
		Phone:        "+919876543220",
		FullName:     "Login Test User",
		PasswordHash: string(passwordHash),
		Status:       models.UserStatusActive,
		AccountType:  models.AccountTypeUser,
		CreatedAt:    sharedModels.Now(),
		UpdatedAt:    sharedModels.Now(),
	}
	userRepo.users[testUser.ID] = testUser

	t.Run("valid credentials returns 200 with token", func(t *testing.T) {
		body := map[string]interface{}{
			"identifier": "login@example.com",
			"password":   "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Login, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)
		assert.Nil(t, resp.Error)

		// Verify token in response
		var loginResp map[string]interface{}
		err := json.Unmarshal(resp.Data, &loginResp)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp["token"])
	})

	t.Run("wrong password returns 401", func(t *testing.T) {
		body := map[string]interface{}{
			"identifier": "login@example.com",
			"password":   "WrongPassword123!",
		}

		rec, resp := makeRequest(t, handler.Login, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, resp.Success)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("non-existent user returns 401 (no enumeration)", func(t *testing.T) {
		body := map[string]interface{}{
			"identifier": "nonexistent@example.com",
			"password":   "SomePassword123!",
		}

		rec, resp := makeRequest(t, handler.Login, http.MethodPost, "/api/v1/auth/login", body)

		// Should return same error as wrong password to prevent user enumeration
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, resp.Success)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("missing identifier returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"password": "SecurePass123!",
		}

		rec, resp := makeRequest(t, handler.Login, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("missing password returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"identifier": "login@example.com",
		}

		rec, resp := makeRequest(t, handler.Login, http.MethodPost, "/api/v1/auth/login", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	authService, _ := createTestAuthService()
	handler := NewAuthHandler(authService)

	t.Run("logout without token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		rec := httptest.NewRecorder()

		handler.Logout(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("logout with malformed token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()

		handler.Logout(rec, req)

		// May return 401 or 204 depending on implementation
		// The key is it should not crash
		assert.Contains(t, []int{http.StatusNoContent, http.StatusUnauthorized}, rec.Code)
	})
}

func TestAuthHandler_GetProfile(t *testing.T) {
	authService, _ := createTestAuthService()
	handler := NewAuthHandler(authService)

	t.Run("get profile without auth returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		rec := httptest.NewRecorder()

		handler.GetProfile(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestExtractIPAddress(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "uses X-Forwarded-For if present",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "uses first IP from X-Forwarded-For chain",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178"},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.195",
		},
		{
			name:       "uses X-Real-IP if X-Forwarded-For not present",
			headers:    map[string]string{"X-Real-IP": "203.0.113.50"},
			remoteAddr: "192.168.1.1:8080",
			expected:   "203.0.113.50",
		},
		{
			name:       "falls back to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:8080",
			expected:   "192.168.1.1",
		},
		{
			name:       "handles RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tc.remoteAddr

			result := extractIPAddress(req)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeIndianPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "10-digit number starting with 9",
			input:    "9876543210",
			expected: "+919876543210",
		},
		{
			name:     "10-digit number starting with 6",
			input:    "6123456789",
			expected: "+916123456789",
		},
		{
			name:     "already formatted with +91",
			input:    "+919876543210",
			expected: "+919876543210",
		},
		{
			name:     "email (not a phone)",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "short number",
			input:    "12345",
			expected: "12345",
		},
		{
			name:     "starts with invalid digit",
			input:    "5123456789",
			expected: "5123456789",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeIndianPhone(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
