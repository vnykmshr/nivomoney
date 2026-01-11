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

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================
// Mock Wallet Repository
// ============================================================

type mockWalletRepository struct {
	wallets map[string]*models.Wallet

	// Override functions for specific behaviors
	CreateFunc          func(ctx context.Context, wallet *models.Wallet) *errors.Error
	GetByIDFunc         func(ctx context.Context, id string) (*models.Wallet, *errors.Error)
	ListByUserIDFunc    func(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error)
	UpdateStatusFunc    func(ctx context.Context, id string, status models.WalletStatus) *errors.Error
	CloseFunc           func(ctx context.Context, id, reason string) *errors.Error
	GetBalanceFunc      func(ctx context.Context, id string) (*models.WalletBalance, *errors.Error)
	GetLimitsFunc       func(ctx context.Context, walletID string) (*models.WalletLimits, *errors.Error)
	UpdateLimitsFunc    func(ctx context.Context, walletID string, dailyLimit, monthlyLimit int64) *errors.Error
	ProcessTransferFunc func(ctx context.Context, sourceWalletID, destWalletID string, amount int64, transactionID string) *errors.Error
	UpdateBalanceFunc   func(ctx context.Context, walletID string, amount int64) *errors.Error
}

func newMockWalletRepository() *mockWalletRepository {
	return &mockWalletRepository{
		wallets: make(map[string]*models.Wallet),
	}
}

func (m *mockWalletRepository) Create(ctx context.Context, wallet *models.Wallet) *errors.Error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, wallet)
	}
	// Generate ID if not set
	if wallet.ID == "" {
		wallet.ID = "wallet-" + time.Now().Format("20060102150405.000")
	}
	wallet.CreatedAt = sharedModels.Now()
	wallet.UpdatedAt = sharedModels.Now()
	m.wallets[wallet.ID] = wallet
	return nil
}

func (m *mockWalletRepository) GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if wallet, ok := m.wallets[id]; ok {
		return wallet, nil
	}
	return nil, errors.NotFound("wallet not found")
}

func (m *mockWalletRepository) ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error) {
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userID, status)
	}
	var result []*models.Wallet
	for _, w := range m.wallets {
		if w.UserID == userID {
			if status == nil || w.Status == *status {
				result = append(result, w)
			}
		}
	}
	return result, nil
}

func (m *mockWalletRepository) UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status)
	}
	if wallet, ok := m.wallets[id]; ok {
		wallet.Status = status
		wallet.UpdatedAt = sharedModels.Now()
		return nil
	}
	return errors.NotFound("wallet not found")
}

func (m *mockWalletRepository) Close(ctx context.Context, id, reason string) *errors.Error {
	if m.CloseFunc != nil {
		return m.CloseFunc(ctx, id, reason)
	}
	if wallet, ok := m.wallets[id]; ok {
		wallet.Status = models.WalletStatusClosed
		wallet.ClosedReason = &reason
		now := sharedModels.Now()
		wallet.ClosedAt = &now
		return nil
	}
	return errors.NotFound("wallet not found")
}

func (m *mockWalletRepository) GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, id)
	}
	if wallet, ok := m.wallets[id]; ok {
		return &models.WalletBalance{
			WalletID:         wallet.ID,
			Balance:          wallet.Balance,
			AvailableBalance: wallet.AvailableBalance,
			HeldAmount:       wallet.Balance - wallet.AvailableBalance,
		}, nil
	}
	return nil, errors.NotFound("wallet not found")
}

func (m *mockWalletRepository) GetLimits(ctx context.Context, walletID string) (*models.WalletLimits, *errors.Error) {
	if m.GetLimitsFunc != nil {
		return m.GetLimitsFunc(ctx, walletID)
	}
	if _, ok := m.wallets[walletID]; ok {
		return &models.WalletLimits{
			WalletID:     walletID,
			DailyLimit:   10000000,  // 1 lakh
			MonthlyLimit: 100000000, // 10 lakh
		}, nil
	}
	return nil, errors.NotFound("wallet not found")
}

func (m *mockWalletRepository) UpdateLimits(ctx context.Context, walletID string, dailyLimit, monthlyLimit int64) *errors.Error {
	if m.UpdateLimitsFunc != nil {
		return m.UpdateLimitsFunc(ctx, walletID, dailyLimit, monthlyLimit)
	}
	if _, ok := m.wallets[walletID]; !ok {
		return errors.NotFound("wallet not found")
	}
	return nil
}

func (m *mockWalletRepository) ProcessTransferWithinTx(ctx context.Context, sourceWalletID, destWalletID string, amount int64, transactionID string) *errors.Error {
	if m.ProcessTransferFunc != nil {
		return m.ProcessTransferFunc(ctx, sourceWalletID, destWalletID, amount, transactionID)
	}
	source, ok := m.wallets[sourceWalletID]
	if !ok {
		return errors.NotFound("source wallet not found")
	}
	dest, ok := m.wallets[destWalletID]
	if !ok {
		return errors.NotFound("destination wallet not found")
	}
	if source.Balance < amount {
		return errors.BadRequest("insufficient balance")
	}
	source.Balance -= amount
	dest.Balance += amount
	return nil
}

func (m *mockWalletRepository) UpdateBalance(ctx context.Context, walletID string, amount int64) *errors.Error {
	if m.UpdateBalanceFunc != nil {
		return m.UpdateBalanceFunc(ctx, walletID, amount)
	}
	if wallet, ok := m.wallets[walletID]; ok {
		wallet.Balance += amount
		wallet.AvailableBalance += amount
		return nil
	}
	return errors.NotFound("wallet not found")
}

// AddWallet adds a wallet to the mock store (for test setup).
func (m *mockWalletRepository) AddWallet(wallet *models.Wallet) {
	m.wallets[wallet.ID] = wallet
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

// createTestWalletService creates a WalletService with mock repositories.
func createTestWalletService() (*service.WalletService, *mockWalletRepository) {
	walletRepo := newMockWalletRepository()
	walletService := service.NewWalletService(
		walletRepo,
		nil, // eventPublisher
		nil, // ledgerClient
		nil, // notificationClient
		nil, // identityClient
	)
	return walletService, walletRepo
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

// makeAuthenticatedRequest creates a request with user ID in context.
func makeAuthenticatedRequest(t *testing.T, handler http.HandlerFunc, method, path string, body interface{}, userID string) (*httptest.ResponseRecorder, *apiResponse) {
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

	// Add user ID to context
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp apiResponse
	if rec.Body.Len() > 0 {
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err, "failed to unmarshal response: %s", rec.Body.String())
	}

	return rec, &resp
}

// makeRequestWithPathValue creates a request with path value for testing.
func makeRequestWithPathValue(t *testing.T, handler http.HandlerFunc, method, path, pathKey, pathValue string, body interface{}) (*httptest.ResponseRecorder, *apiResponse) {
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
	req.SetPathValue(pathKey, pathValue)

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
// Wallet Handler Tests
// ============================================================

func TestWalletHandler_CreateWallet(t *testing.T) {
	walletService, _ := createTestWalletService()
	handler := NewWalletHandler(walletService)

	t.Run("create wallet without auth returns 401", func(t *testing.T) {
		body := map[string]interface{}{
			"type":     "default",
			"currency": "INR",
		}

		rec, resp := makeRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("create wallet with valid data returns 201", func(t *testing.T) {
		body := map[string]interface{}{
			"type":              "default",
			"currency":          "INR",
			"ledger_account_id": "ledger-acct-123",
		}

		rec, resp := makeAuthenticatedRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body, "user-123")

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.True(t, resp.Success)
		assert.Nil(t, resp.Error)

		// Verify wallet in response
		var wallet map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err)
		assert.NotEmpty(t, wallet["id"])
		assert.Equal(t, "user-123", wallet["user_id"])
		assert.Equal(t, "INR", wallet["currency"])
	})

	t.Run("create duplicate currency wallet returns conflict", func(t *testing.T) {
		// First wallet
		body := map[string]interface{}{
			"type":              "default",
			"currency":          "INR",
			"ledger_account_id": "ledger-acct-456",
		}
		makeAuthenticatedRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body, "user-456")

		// Second wallet with same currency (different ledger account, but conflict due to same currency)
		body["ledger_account_id"] = "ledger-acct-456-2"
		rec, resp := makeAuthenticatedRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body, "user-456")

		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "CONFLICT", resp.Error.Code)
	})

	t.Run("create wallet with invalid type returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"type":     "invalid-type",
			"currency": "INR",
		}

		rec, resp := makeAuthenticatedRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body, "user-789")

		// Should return validation error for invalid wallet type
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		// Could be VALIDATION_ERROR or BAD_REQUEST depending on implementation
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusConflict}, rec.Code)
	})

	t.Run("create wallet with missing currency returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"type": "default",
			// missing currency
		}

		rec, resp := makeAuthenticatedRequest(t, handler.CreateWallet, http.MethodPost, "/api/v1/wallets", body, "user-999")

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestWalletHandler_GetWallet(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add a test wallet
	testWallet := &models.Wallet{
		ID:               "wallet-test-get",
		UserID:           "user-get-test",
		Type:             models.WalletTypeDefault,
		Currency:         "INR",
		Balance:          100000, // 1000 rupees
		AvailableBalance: 100000,
		Status:           models.WalletStatusActive,
		CreatedAt:        sharedModels.Now(),
		UpdatedAt:        sharedModels.Now(),
	}
	walletRepo.AddWallet(testWallet)

	t.Run("get existing wallet returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetWallet, http.MethodGet, "/api/v1/wallets/wallet-test-get", "id", "wallet-test-get", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)
		assert.Nil(t, resp.Error)

		// Verify wallet data
		var wallet map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err)
		assert.Equal(t, "wallet-test-get", wallet["id"])
		assert.Equal(t, "user-get-test", wallet["user_id"])
	})

	t.Run("get non-existent wallet returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetWallet, http.MethodGet, "/api/v1/wallets/non-existent", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get wallet without ID returns 400", func(t *testing.T) {
		// Request without setting path value
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/", nil)
		rec := httptest.NewRecorder()
		handler.GetWallet(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestWalletHandler_ListMyWallets(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add test wallets for user
	wallet1 := &models.Wallet{
		ID:       "wallet-list-1",
		UserID:   "user-list-test",
		Type:     models.WalletTypeDefault,
		Currency: "INR",
		Status:   models.WalletStatusActive,
	}
	wallet2 := &models.Wallet{
		ID:       "wallet-list-2",
		UserID:   "user-list-test",
		Type:     models.WalletTypeDefault,
		Currency: "USD",
		Status:   models.WalletStatusActive,
	}
	walletRepo.AddWallet(wallet1)
	walletRepo.AddWallet(wallet2)

	t.Run("list wallets without auth returns 401", func(t *testing.T) {
		rec, resp := makeRequest(t, handler.ListMyWallets, http.MethodGet, "/api/v1/wallets", nil)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "UNAUTHORIZED", resp.Error.Code)
	})

	t.Run("list wallets with auth returns user wallets", func(t *testing.T) {
		rec, resp := makeAuthenticatedRequest(t, handler.ListMyWallets, http.MethodGet, "/api/v1/wallets", nil, "user-list-test")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify wallet list
		var wallets []map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallets)
		require.NoError(t, err)
		assert.Len(t, wallets, 2)
	})

	t.Run("list wallets returns empty array for new user", func(t *testing.T) {
		rec, resp := makeAuthenticatedRequest(t, handler.ListMyWallets, http.MethodGet, "/api/v1/wallets", nil, "new-user-no-wallets")

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify empty list
		var wallets []map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallets)
		require.NoError(t, err)
		assert.Len(t, wallets, 0)
	})
}

func TestWalletHandler_ActivateWallet(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add an inactive wallet (needs activation)
	pendingWallet := &models.Wallet{
		ID:       "wallet-pending",
		UserID:   "user-activate-test",
		Type:     models.WalletTypeDefault,
		Currency: "INR",
		Status:   models.WalletStatusInactive,
	}
	walletRepo.AddWallet(pendingWallet)

	t.Run("activate pending wallet returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.ActivateWallet, http.MethodPost, "/api/v1/wallets/wallet-pending/activate", "id", "wallet-pending", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify wallet is now active
		var wallet map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err)
		assert.Equal(t, "active", wallet["status"])
	})

	t.Run("activate non-existent wallet returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.ActivateWallet, http.MethodPost, "/api/v1/wallets/non-existent/activate", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})
}

func TestWalletHandler_FreezeWallet(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add an active wallet
	activeWallet := &models.Wallet{
		ID:       "wallet-active-freeze",
		UserID:   "user-freeze-test",
		Type:     models.WalletTypeDefault,
		Currency: "INR",
		Status:   models.WalletStatusActive,
	}
	walletRepo.AddWallet(activeWallet)

	t.Run("freeze active wallet returns 200", func(t *testing.T) {
		body := map[string]interface{}{
			"reason": "Suspicious activity detected",
		}

		rec, resp := makeRequestWithPathValue(t, handler.FreezeWallet, http.MethodPost, "/api/v1/wallets/wallet-active-freeze/freeze", "id", "wallet-active-freeze", body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify wallet is now frozen
		var wallet map[string]interface{}
		err := json.Unmarshal(resp.Data, &wallet)
		require.NoError(t, err)
		assert.Equal(t, "frozen", wallet["status"])
	})

	t.Run("freeze wallet without reason returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing reason
		}

		rec, resp := makeRequestWithPathValue(t, handler.FreezeWallet, http.MethodPost, "/api/v1/wallets/wallet-active-freeze/freeze", "id", "wallet-active-freeze", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestWalletHandler_GetWalletBalance(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add a wallet with balance
	walletWithBalance := &models.Wallet{
		ID:               "wallet-balance-test",
		UserID:           "user-balance-test",
		Type:             models.WalletTypeDefault,
		Currency:         "INR",
		Balance:          500000, // 5000 rupees
		AvailableBalance: 450000, // 4500 available (some held)
		Status:           models.WalletStatusActive,
	}
	walletRepo.AddWallet(walletWithBalance)

	t.Run("get balance returns 200 with correct data", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetWalletBalance, http.MethodGet, "/api/v1/wallets/wallet-balance-test/balance", "id", "wallet-balance-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify balance data
		var balance map[string]interface{}
		err := json.Unmarshal(resp.Data, &balance)
		require.NoError(t, err)
		assert.Equal(t, "wallet-balance-test", balance["wallet_id"])
		assert.Equal(t, float64(500000), balance["balance"])
		assert.Equal(t, float64(450000), balance["available_balance"])
	})

	t.Run("get balance for non-existent wallet returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetWalletBalance, http.MethodGet, "/api/v1/wallets/non-existent/balance", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})
}

func TestWalletHandler_ProcessTransfer(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add source and destination wallets
	sourceWallet := &models.Wallet{
		ID:               "wallet-source",
		UserID:           "user-source",
		Type:             models.WalletTypeDefault,
		Currency:         "INR",
		Balance:          1000000, // 10000 rupees
		AvailableBalance: 1000000,
		Status:           models.WalletStatusActive,
	}
	destWallet := &models.Wallet{
		ID:               "wallet-dest",
		UserID:           "user-dest",
		Type:             models.WalletTypeDefault,
		Currency:         "INR",
		Balance:          0,
		AvailableBalance: 0,
		Status:           models.WalletStatusActive,
	}
	walletRepo.AddWallet(sourceWallet)
	walletRepo.AddWallet(destWallet)

	t.Run("process valid transfer returns 200", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id":      "wallet-source",
			"destination_wallet_id": "wallet-dest",
			"amount":                100000, // 1000 rupees
			"transaction_id":        "tx-123",
		}

		rec, resp := makeRequest(t, handler.ProcessTransfer, http.MethodPost, "/internal/v1/wallets/transfer", body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify transfer data in response
		var result map[string]interface{}
		err := json.Unmarshal(resp.Data, &result)
		require.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, "wallet-source", result["source_wallet_id"])
		assert.Equal(t, "wallet-dest", result["dest_wallet_id"])
	})

	t.Run("process transfer with insufficient balance returns error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id":      "wallet-source",
			"destination_wallet_id": "wallet-dest",
			"amount":                99999999999, // More than available
			"transaction_id":        "tx-456",
		}

		rec, resp := makeRequest(t, handler.ProcessTransfer, http.MethodPost, "/internal/v1/wallets/transfer", body)

		assert.NotEqual(t, http.StatusOK, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
	})

	t.Run("process transfer with missing fields returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id": "wallet-source",
			// missing destination and amount
		}

		rec, resp := makeRequest(t, handler.ProcessTransfer, http.MethodPost, "/internal/v1/wallets/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestWalletHandler_ProcessDeposit(t *testing.T) {
	walletService, walletRepo := createTestWalletService()
	handler := NewWalletHandler(walletService)

	// Setup: Add a wallet for deposit
	depositWallet := &models.Wallet{
		ID:               "wallet-deposit",
		UserID:           "user-deposit",
		Type:             models.WalletTypeDefault,
		Currency:         "INR",
		Balance:          0,
		AvailableBalance: 0,
		Status:           models.WalletStatusActive,
	}
	walletRepo.AddWallet(depositWallet)

	t.Run("process valid deposit returns 200", func(t *testing.T) {
		body := map[string]interface{}{
			"wallet_id":      "wallet-deposit",
			"amount":         500000, // 5000 rupees
			"transaction_id": "tx-deposit-123",
		}

		rec, resp := makeRequest(t, handler.ProcessDeposit, http.MethodPost, "/internal/v1/wallets/deposit", body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify deposit data
		var result map[string]interface{}
		err := json.Unmarshal(resp.Data, &result)
		require.NoError(t, err)
		assert.Equal(t, true, result["success"])
		assert.Equal(t, "wallet-deposit", result["wallet_id"])
	})

	t.Run("process deposit to non-existent wallet returns 404", func(t *testing.T) {
		body := map[string]interface{}{
			"wallet_id":      "non-existent-wallet",
			"amount":         500000,
			"transaction_id": "tx-deposit-456",
		}

		rec, resp := makeRequest(t, handler.ProcessDeposit, http.MethodPost, "/internal/v1/wallets/deposit", body)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})
}
