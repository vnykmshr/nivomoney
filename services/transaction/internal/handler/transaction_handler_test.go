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

	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/services/transaction/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================
// Mock Transaction Repository
// ============================================================

type mockTransactionRepository struct {
	transactions map[string]*models.Transaction

	// Override functions for specific behaviors
	CreateFunc              func(ctx context.Context, transaction *models.Transaction) *errors.Error
	GetByIDFunc             func(ctx context.Context, id string) (*models.Transaction, *errors.Error)
	ListByWalletFunc        func(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
	SearchAllFunc           func(ctx context.Context, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
	UpdateMetadataFunc      func(ctx context.Context, id string, metadata map[string]string) *errors.Error
	CompleteFunc            func(ctx context.Context, id string, metadata map[string]string) *errors.Error
	UpdateStatusFunc        func(ctx context.Context, id string, status models.TransactionStatus, failureReason *string) *errors.Error
	UpdateCategoryFunc      func(ctx context.Context, id string, category models.SpendingCategory) *errors.Error
	GetCategoryPatternsFunc func(ctx context.Context) ([]*models.CategoryPattern, *errors.Error)
	GetCategorySummaryFunc  func(ctx context.Context, walletID string, startDate, endDate string) ([]models.CategorySummary, *errors.Error)
}

func newMockTransactionRepository() *mockTransactionRepository {
	return &mockTransactionRepository{
		transactions: make(map[string]*models.Transaction),
	}
}

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *models.Transaction) *errors.Error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, transaction)
	}
	// Generate ID if not set
	if transaction.ID == "" {
		transaction.ID = "tx-" + time.Now().Format("20060102150405.000")
	}
	transaction.CreatedAt = sharedModels.Now()
	transaction.UpdatedAt = sharedModels.Now()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *mockTransactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, *errors.Error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if tx, ok := m.transactions[id]; ok {
		return tx, nil
	}
	return nil, errors.NotFound("transaction not found")
}

func (m *mockTransactionRepository) ListByWallet(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	if m.ListByWalletFunc != nil {
		return m.ListByWalletFunc(ctx, walletID, filter)
	}
	var result []*models.Transaction
	for _, tx := range m.transactions {
		if (tx.SourceWalletID != nil && *tx.SourceWalletID == walletID) ||
			(tx.DestinationWalletID != nil && *tx.DestinationWalletID == walletID) {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *mockTransactionRepository) SearchAll(ctx context.Context, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	if m.SearchAllFunc != nil {
		return m.SearchAllFunc(ctx, filter)
	}
	var result []*models.Transaction
	for _, tx := range m.transactions {
		result = append(result, tx)
	}
	return result, nil
}

func (m *mockTransactionRepository) UpdateMetadata(ctx context.Context, id string, metadata map[string]string) *errors.Error {
	if m.UpdateMetadataFunc != nil {
		return m.UpdateMetadataFunc(ctx, id, metadata)
	}
	if tx, ok := m.transactions[id]; ok {
		tx.Metadata = metadata
		return nil
	}
	return errors.NotFound("transaction not found")
}

func (m *mockTransactionRepository) CompleteWithMetadata(ctx context.Context, id string, metadata map[string]string) *errors.Error {
	if m.CompleteFunc != nil {
		return m.CompleteFunc(ctx, id, metadata)
	}
	if tx, ok := m.transactions[id]; ok {
		tx.Status = models.TransactionStatusCompleted
		tx.Metadata = metadata
		now := sharedModels.Now()
		tx.CompletedAt = &now
		return nil
	}
	return errors.NotFound("transaction not found")
}

func (m *mockTransactionRepository) UpdateStatus(ctx context.Context, id string, status models.TransactionStatus, failureReason *string) *errors.Error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, id, status, failureReason)
	}
	if tx, ok := m.transactions[id]; ok {
		tx.Status = status
		tx.FailureReason = failureReason
		return nil
	}
	return errors.NotFound("transaction not found")
}

func (m *mockTransactionRepository) UpdateCategory(ctx context.Context, id string, category models.SpendingCategory) *errors.Error {
	if m.UpdateCategoryFunc != nil {
		return m.UpdateCategoryFunc(ctx, id, category)
	}
	if tx, ok := m.transactions[id]; ok {
		tx.Category = category
		return nil
	}
	return errors.NotFound("transaction not found")
}

func (m *mockTransactionRepository) GetCategoryPatterns(ctx context.Context) ([]*models.CategoryPattern, *errors.Error) {
	if m.GetCategoryPatternsFunc != nil {
		return m.GetCategoryPatternsFunc(ctx)
	}
	return []*models.CategoryPattern{}, nil
}

func (m *mockTransactionRepository) GetCategorySummary(ctx context.Context, walletID string, startDate, endDate string) ([]models.CategorySummary, *errors.Error) {
	if m.GetCategorySummaryFunc != nil {
		return m.GetCategorySummaryFunc(ctx, walletID, startDate, endDate)
	}
	return []models.CategorySummary{}, nil
}

// AddTransaction adds a transaction to the mock store (for test setup).
func (m *mockTransactionRepository) AddTransaction(tx *models.Transaction) {
	m.transactions[tx.ID] = tx
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

// createTestTransactionService creates a TransactionService with mock repositories.
func createTestTransactionService() (*service.TransactionService, *mockTransactionRepository) {
	txRepo := newMockTransactionRepository()
	txService := service.NewTransactionService(
		txRepo,
		nil, // riskClient
		nil, // walletClient
		nil, // ledgerClient
		nil, // eventPublisher
	)
	return txService, txRepo
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
// Transaction Handler Tests
// ============================================================

func TestTransactionHandler_GetTransaction(t *testing.T) {
	txService, txRepo := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	// Setup: Add a test transaction
	sourceWallet := "wallet-source-1"
	destWallet := "wallet-dest-1"
	testTx := &models.Transaction{
		ID:                  "tx-get-test",
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusCompleted,
		SourceWalletID:      &sourceWallet,
		DestinationWalletID: &destWallet,
		Amount:              100000,
		Currency:            "INR",
		Description:         "Test transfer",
		CreatedAt:           sharedModels.Now(),
		UpdatedAt:           sharedModels.Now(),
	}
	txRepo.AddTransaction(testTx)

	t.Run("get existing transaction returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetTransaction, http.MethodGet, "/api/v1/transactions/tx-get-test", "id", "tx-get-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)
		assert.Nil(t, resp.Error)

		// Verify transaction data
		var tx map[string]interface{}
		err := json.Unmarshal(resp.Data, &tx)
		require.NoError(t, err)
		assert.Equal(t, "tx-get-test", tx["id"])
		assert.Equal(t, "transfer", tx["type"])
		assert.Equal(t, "completed", tx["status"])
	})

	t.Run("get non-existent transaction returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetTransaction, http.MethodGet, "/api/v1/transactions/non-existent", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get transaction without ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/", nil)
		rec := httptest.NewRecorder()
		handler.GetTransaction(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestTransactionHandler_ListWalletTransactions(t *testing.T) {
	txService, txRepo := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	// Setup: Add test transactions for a wallet
	walletID := "wallet-list-test"
	tx1 := &models.Transaction{
		ID:                  "tx-list-1",
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusCompleted,
		DestinationWalletID: &walletID,
		Amount:              500000,
		Currency:            "INR",
		Description:         "Test deposit 1",
	}
	tx2 := &models.Transaction{
		ID:             "tx-list-2",
		Type:           models.TransactionTypeWithdrawal,
		Status:         models.TransactionStatusCompleted,
		SourceWalletID: &walletID,
		Amount:         100000,
		Currency:       "INR",
		Description:    "Test withdrawal",
	}
	txRepo.AddTransaction(tx1)
	txRepo.AddTransaction(tx2)

	t.Run("list wallet transactions returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.ListWalletTransactions, http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions", "walletId", "wallet-list-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		// Verify transaction list
		var transactions []map[string]interface{}
		err := json.Unmarshal(resp.Data, &transactions)
		require.NoError(t, err)
		assert.Len(t, transactions, 2)
	})

	t.Run("list wallet transactions without wallet ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets//transactions", nil)
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("list wallet transactions with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?status=completed", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("list wallet transactions with invalid status returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?status=invalid", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("list wallet transactions with type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?type=deposit", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("list wallet transactions with amount range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?min_amount=10000&max_amount=1000000", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("list wallet transactions with invalid amount range returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?min_amount=1000000&max_amount=10000", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("list wallet transactions with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/wallet-list-test/transactions?limit=10&offset=0", nil)
		req.SetPathValue("walletId", "wallet-list-test")
		rec := httptest.NewRecorder()
		handler.ListWalletTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestTransactionHandler_ProcessTransfer(t *testing.T) {
	txService, txRepo := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	// Setup: Add a pending transfer transaction
	sourceWallet := "wallet-source"
	destWallet := "wallet-dest"
	pendingTx := &models.Transaction{
		ID:                  "tx-pending-transfer",
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusPending,
		SourceWalletID:      &sourceWallet,
		DestinationWalletID: &destWallet,
		Amount:              50000,
		Currency:            "INR",
		Description:         "Pending transfer",
	}
	txRepo.AddTransaction(pendingTx)

	t.Run("process transfer without ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/internal/v1/transactions//process", nil)
		rec := httptest.NewRecorder()
		handler.ProcessTransfer(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("process non-existent transfer returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.ProcessTransfer, http.MethodPost, "/internal/v1/transactions/non-existent/process", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})
}

func TestTransactionHandler_ReverseTransaction(t *testing.T) {
	txService, txRepo := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	// Setup: Add a completed transaction
	sourceWallet := "wallet-reverse-source"
	destWallet := "wallet-reverse-dest"
	completedTx := &models.Transaction{
		ID:                  "tx-to-reverse",
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusCompleted,
		SourceWalletID:      &sourceWallet,
		DestinationWalletID: &destWallet,
		Amount:              75000,
		Currency:            "INR",
		Description:         "Transfer to reverse",
	}
	txRepo.AddTransaction(completedTx)

	t.Run("reverse transaction without ID returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"reason": "Customer requested reversal",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions//reverse", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ReverseTransaction(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("reverse transaction without reason returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing reason
		}

		rec, resp := makeRequestWithPathValue(t, handler.ReverseTransaction, http.MethodPost, "/api/v1/transactions/tx-to-reverse/reverse", "id", "tx-to-reverse", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("reverse transaction with short reason returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"reason": "Short", // less than 10 characters
		}

		rec, resp := makeRequestWithPathValue(t, handler.ReverseTransaction, http.MethodPost, "/api/v1/transactions/tx-to-reverse/reverse", "id", "tx-to-reverse", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestTransactionHandler_CreateTransfer(t *testing.T) {
	txService, _ := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	t.Run("create transfer with missing source wallet returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing source_wallet_id
			"destination_wallet_id": "wallet-dest",
			"amount":                100000,
			"currency":              "INR",
			"description":           "Test transfer",
		}

		rec, resp := makeRequest(t, handler.CreateTransfer, http.MethodPost, "/api/v1/transactions/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create transfer with missing destination returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id": "wallet-source",
			// missing destination_wallet_id
			"amount":      100000,
			"currency":    "INR",
			"description": "Test transfer",
		}

		rec, resp := makeRequest(t, handler.CreateTransfer, http.MethodPost, "/api/v1/transactions/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create transfer with zero amount returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id":      "wallet-source",
			"destination_wallet_id": "wallet-dest",
			"amount":                0,
			"currency":              "INR",
			"description":           "Test transfer",
		}

		rec, resp := makeRequest(t, handler.CreateTransfer, http.MethodPost, "/api/v1/transactions/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create transfer with missing currency returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id":      "wallet-source",
			"destination_wallet_id": "wallet-dest",
			"amount":                100000,
			// missing currency
			"description": "Test transfer",
		}

		rec, resp := makeRequest(t, handler.CreateTransfer, http.MethodPost, "/api/v1/transactions/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create transfer with short description returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"source_wallet_id":      "wallet-source",
			"destination_wallet_id": "wallet-dest",
			"amount":                100000,
			"currency":              "INR",
			"description":           "Hi", // less than 3 characters
		}

		rec, resp := makeRequest(t, handler.CreateTransfer, http.MethodPost, "/api/v1/transactions/transfer", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestTransactionHandler_CreateDeposit(t *testing.T) {
	txService, _ := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	t.Run("create deposit with missing wallet returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing wallet_id
			"amount":      100000,
			"currency":    "INR",
			"description": "Test deposit",
		}

		rec, resp := makeRequest(t, handler.CreateDeposit, http.MethodPost, "/api/v1/transactions/deposit", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create deposit with zero amount returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"wallet_id":   "wallet-deposit",
			"amount":      0,
			"currency":    "INR",
			"description": "Test deposit",
		}

		rec, resp := makeRequest(t, handler.CreateDeposit, http.MethodPost, "/api/v1/transactions/deposit", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create deposit with missing currency returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"wallet_id": "wallet-deposit",
			"amount":    100000,
			// missing currency
			"description": "Test deposit",
		}

		rec, resp := makeRequest(t, handler.CreateDeposit, http.MethodPost, "/api/v1/transactions/deposit", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestTransactionHandler_CreateWithdrawal(t *testing.T) {
	txService, _ := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	t.Run("create withdrawal with missing wallet returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing wallet_id
			"amount":      100000,
			"currency":    "INR",
			"description": "Test withdrawal",
		}

		rec, resp := makeRequest(t, handler.CreateWithdrawal, http.MethodPost, "/api/v1/transactions/withdrawal", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create withdrawal with zero amount returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"wallet_id":   "wallet-withdraw",
			"amount":      0,
			"currency":    "INR",
			"description": "Test withdrawal",
		}

		rec, resp := makeRequest(t, handler.CreateWithdrawal, http.MethodPost, "/api/v1/transactions/withdrawal", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestTransactionHandler_SearchAllTransactions(t *testing.T) {
	txService, txRepo := createTestTransactionService()
	handler := NewTransactionHandler(txService, nil)

	// Setup: Add some test transactions
	wallet1 := "wallet-search-1"
	wallet2 := "wallet-search-2"
	tx1 := &models.Transaction{
		ID:                  "tx-search-1",
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusCompleted,
		SourceWalletID:      &wallet1,
		DestinationWalletID: &wallet2,
		Amount:              100000,
		Currency:            "INR",
		Description:         "Transfer for testing",
	}
	tx2 := &models.Transaction{
		ID:                  "tx-search-2",
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusPending,
		DestinationWalletID: &wallet1,
		Amount:              200000,
		Currency:            "INR",
		Description:         "Deposit for testing",
	}
	txRepo.AddTransaction(tx1)
	txRepo.AddTransaction(tx2)

	t.Run("search all transactions returns 200", func(t *testing.T) {
		rec, resp := makeRequest(t, handler.SearchAllTransactions, http.MethodGet, "/api/v1/admin/transactions/search", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var transactions []map[string]interface{}
		err := json.Unmarshal(resp.Data, &transactions)
		require.NoError(t, err)
		assert.Len(t, transactions, 2)
	})

	t.Run("search with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transactions/search?status=completed", nil)
		rec := httptest.NewRecorder()
		handler.SearchAllTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("search with type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transactions/search?type=deposit", nil)
		rec := httptest.NewRecorder()
		handler.SearchAllTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("search with too short query returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transactions/search?search=a", nil)
		rec := httptest.NewRecorder()
		handler.SearchAllTransactions(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("search with amount range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/transactions/search?min_amount=50000&max_amount=150000", nil)
		rec := httptest.NewRecorder()
		handler.SearchAllTransactions(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestValidateDateRange(t *testing.T) {
	t.Run("valid date range returns nil", func(t *testing.T) {
		err := validateDateRange("2025-01-01", "2025-01-31")
		assert.Nil(t, err)
	})

	t.Run("invalid start date format returns error", func(t *testing.T) {
		err := validateDateRange("2025/01/01", "2025-01-31")
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "start_date")
	})

	t.Run("invalid end date format returns error", func(t *testing.T) {
		err := validateDateRange("2025-01-01", "2025/01/31")
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "end_date")
	})

	t.Run("start after end returns error", func(t *testing.T) {
		err := validateDateRange("2025-02-01", "2025-01-01")
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "cannot be after")
	})

	t.Run("date range exceeding 1 year returns error", func(t *testing.T) {
		err := validateDateRange("2024-01-01", "2025-12-31")
		assert.NotNil(t, err)
		assert.Contains(t, err.Message, "exceed 1 year")
	})
}
