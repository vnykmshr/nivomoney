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

	"github.com/vnykmshr/nivo/services/ledger/internal/models"
	"github.com/vnykmshr/nivo/services/ledger/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================
// Mock Account Repository
// ============================================================

type mockAccountRepository struct {
	accounts map[string]*models.Account

	CreateFunc     func(ctx context.Context, account *models.Account) *errors.Error
	GetByIDFunc    func(ctx context.Context, id string) (*models.Account, *errors.Error)
	GetByCodeFunc  func(ctx context.Context, code string) (*models.Account, *errors.Error)
	ListFunc       func(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error)
	UpdateFunc     func(ctx context.Context, account *models.Account) *errors.Error
	GetBalanceFunc func(ctx context.Context, accountID string) (int64, *errors.Error)
}

func newMockAccountRepository() *mockAccountRepository {
	return &mockAccountRepository{
		accounts: make(map[string]*models.Account),
	}
}

func (m *mockAccountRepository) Create(ctx context.Context, account *models.Account) *errors.Error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, account)
	}
	if account.ID == "" {
		account.ID = "acct-" + time.Now().Format("20060102150405.000")
	}
	account.CreatedAt = sharedModels.Now()
	account.UpdatedAt = sharedModels.Now()
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) GetByID(ctx context.Context, id string) (*models.Account, *errors.Error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if acct, ok := m.accounts[id]; ok {
		return acct, nil
	}
	return nil, errors.NotFound("account not found")
}

func (m *mockAccountRepository) GetByCode(ctx context.Context, code string) (*models.Account, *errors.Error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	for _, acct := range m.accounts {
		if acct.Code == code {
			return acct, nil
		}
	}
	return nil, errors.NotFound("account not found")
}

func (m *mockAccountRepository) List(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, accountType, status, limit, offset)
	}
	var result []*models.Account
	for _, acct := range m.accounts {
		if accountType != nil && acct.Type != *accountType {
			continue
		}
		if status != nil && acct.Status != *status {
			continue
		}
		result = append(result, acct)
	}
	return result, nil
}

func (m *mockAccountRepository) Update(ctx context.Context, account *models.Account) *errors.Error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, account)
	}
	if _, ok := m.accounts[account.ID]; !ok {
		return errors.NotFound("account not found")
	}
	account.UpdatedAt = sharedModels.Now()
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) GetBalance(ctx context.Context, accountID string) (int64, *errors.Error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, accountID)
	}
	if acct, ok := m.accounts[accountID]; ok {
		return acct.Balance, nil
	}
	return 0, errors.NotFound("account not found")
}

func (m *mockAccountRepository) AddAccount(acct *models.Account) {
	m.accounts[acct.ID] = acct
}

// ============================================================
// Mock Journal Entry Repository
// ============================================================

type mockJournalEntryRepository struct {
	entries map[string]*models.JournalEntry

	CreateFunc  func(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error
	GetByIDFunc func(ctx context.Context, id string) (*models.JournalEntry, *errors.Error)
	ListFunc    func(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error)
	PostFunc    func(ctx context.Context, entryID, postedBy string) *errors.Error
	VoidFunc    func(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error
}

func newMockJournalEntryRepository() *mockJournalEntryRepository {
	return &mockJournalEntryRepository{
		entries: make(map[string]*models.JournalEntry),
	}
}

func (m *mockJournalEntryRepository) Create(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entry, lines)
	}
	if entry.ID == "" {
		entry.ID = "je-" + time.Now().Format("20060102150405.000")
	}
	entry.CreatedAt = sharedModels.Now()
	entry.UpdatedAt = sharedModels.Now()
	m.entries[entry.ID] = entry
	return nil
}

func (m *mockJournalEntryRepository) GetByID(ctx context.Context, id string) (*models.JournalEntry, *errors.Error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	if entry, ok := m.entries[id]; ok {
		return entry, nil
	}
	return nil, errors.NotFound("journal entry not found")
}

func (m *mockJournalEntryRepository) List(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, status, limit, offset)
	}
	var result []*models.JournalEntry
	for _, entry := range m.entries {
		if status != nil && entry.Status != *status {
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}

func (m *mockJournalEntryRepository) Post(ctx context.Context, entryID, postedBy string) *errors.Error {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, entryID, postedBy)
	}
	if entry, ok := m.entries[entryID]; ok {
		entry.Status = models.EntryStatusPosted
		now := sharedModels.Now()
		entry.PostedAt = &now
		entry.PostedBy = &postedBy
		return nil
	}
	return errors.NotFound("journal entry not found")
}

func (m *mockJournalEntryRepository) Void(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error {
	if m.VoidFunc != nil {
		return m.VoidFunc(ctx, entryID, voidedBy, voidReason)
	}
	if entry, ok := m.entries[entryID]; ok {
		entry.Status = models.EntryStatusVoided
		now := sharedModels.Now()
		entry.VoidedAt = &now
		entry.VoidedBy = &voidedBy
		entry.VoidReason = &voidReason
		return nil
	}
	return errors.NotFound("journal entry not found")
}

func (m *mockJournalEntryRepository) AddEntry(entry *models.JournalEntry) {
	m.entries[entry.ID] = entry
}

// ============================================================
// Test Helper Functions
// ============================================================

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *apiError       `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func createTestLedgerService() (*service.LedgerService, *mockAccountRepository, *mockJournalEntryRepository) {
	accountRepo := newMockAccountRepository()
	journalRepo := newMockJournalEntryRepository()
	ledgerService := service.NewLedgerService(accountRepo, journalRepo)
	return ledgerService, accountRepo, journalRepo
}

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
// Account Handler Tests
// ============================================================

func TestLedgerHandler_CreateAccount(t *testing.T) {
	ledgerService, _, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	t.Run("create account with valid data returns 201", func(t *testing.T) {
		body := map[string]interface{}{
			"code":     "1001",
			"name":     "Cash Account",
			"type":     "asset",
			"currency": "INR",
		}

		rec, resp := makeRequest(t, handler.CreateAccount, http.MethodPost, "/api/v1/accounts", body)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.True(t, resp.Success)

		var acct map[string]interface{}
		err := json.Unmarshal(resp.Data, &acct)
		require.NoError(t, err)
		assert.NotEmpty(t, acct["id"])
		assert.Equal(t, "1001", acct["code"])
		assert.Equal(t, "Cash Account", acct["name"])
	})

	t.Run("create account with missing code returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing code
			"name":     "Test Account",
			"type":     "asset",
			"currency": "INR",
		}

		rec, resp := makeRequest(t, handler.CreateAccount, http.MethodPost, "/api/v1/accounts", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create account with missing name returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"code": "1002",
			// missing name
			"type":     "asset",
			"currency": "INR",
		}

		rec, resp := makeRequest(t, handler.CreateAccount, http.MethodPost, "/api/v1/accounts", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create account with missing type returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"code":     "1003",
			"name":     "Test Account",
			"currency": "INR",
			// missing type
		}

		rec, resp := makeRequest(t, handler.CreateAccount, http.MethodPost, "/api/v1/accounts", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})

	t.Run("create account with missing currency returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"code": "1004",
			"name": "Test Account",
			"type": "asset",
			// missing currency
		}

		rec, resp := makeRequest(t, handler.CreateAccount, http.MethodPost, "/api/v1/accounts", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestLedgerHandler_GetAccount(t *testing.T) {
	ledgerService, accountRepo, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a test account
	testAccount := &models.Account{
		ID:       "acct-get-test",
		Code:     "1100",
		Name:     "Test Asset Account",
		Type:     models.AccountTypeAsset,
		Currency: "INR",
		Balance:  100000,
		Status:   models.AccountStatusActive,
	}
	accountRepo.AddAccount(testAccount)

	t.Run("get existing account returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccount, http.MethodGet, "/api/v1/accounts/acct-get-test", "id", "acct-get-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var acct map[string]interface{}
		err := json.Unmarshal(resp.Data, &acct)
		require.NoError(t, err)
		assert.Equal(t, "acct-get-test", acct["id"])
		assert.Equal(t, "1100", acct["code"])
	})

	t.Run("get non-existent account returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccount, http.MethodGet, "/api/v1/accounts/non-existent", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get account without ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/", nil)
		rec := httptest.NewRecorder()
		handler.GetAccount(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLedgerHandler_GetAccountByCode(t *testing.T) {
	ledgerService, accountRepo, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a test account
	testAccount := &models.Account{
		ID:       "acct-code-test",
		Code:     "2100",
		Name:     "Test Liability Account",
		Type:     models.AccountTypeLiability,
		Currency: "INR",
		Status:   models.AccountStatusActive,
	}
	accountRepo.AddAccount(testAccount)

	t.Run("get account by code returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccountByCode, http.MethodGet, "/api/v1/accounts/code/2100", "code", "2100", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var acct map[string]interface{}
		err := json.Unmarshal(resp.Data, &acct)
		require.NoError(t, err)
		assert.Equal(t, "2100", acct["code"])
	})

	t.Run("get account by non-existent code returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccountByCode, http.MethodGet, "/api/v1/accounts/code/9999", "code", "9999", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get account without code returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/code/", nil)
		rec := httptest.NewRecorder()
		handler.GetAccountByCode(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLedgerHandler_ListAccounts(t *testing.T) {
	ledgerService, accountRepo, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add test accounts
	acct1 := &models.Account{
		ID:       "acct-list-1",
		Code:     "1000",
		Name:     "Asset 1",
		Type:     models.AccountTypeAsset,
		Currency: "INR",
		Status:   models.AccountStatusActive,
	}
	acct2 := &models.Account{
		ID:       "acct-list-2",
		Code:     "2000",
		Name:     "Liability 1",
		Type:     models.AccountTypeLiability,
		Currency: "INR",
		Status:   models.AccountStatusActive,
	}
	accountRepo.AddAccount(acct1)
	accountRepo.AddAccount(acct2)

	t.Run("list all accounts returns 200", func(t *testing.T) {
		rec, resp := makeRequest(t, handler.ListAccounts, http.MethodGet, "/api/v1/accounts", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var accounts []map[string]interface{}
		err := json.Unmarshal(resp.Data, &accounts)
		require.NoError(t, err)
		assert.Len(t, accounts, 2)
	})

	t.Run("list accounts with type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts?type=asset", nil)
		rec := httptest.NewRecorder()
		handler.ListAccounts(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("list accounts with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts?status=active", nil)
		rec := httptest.NewRecorder()
		handler.ListAccounts(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestLedgerHandler_UpdateAccount(t *testing.T) {
	ledgerService, accountRepo, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a test account
	testAccount := &models.Account{
		ID:       "acct-update-test",
		Code:     "3100",
		Name:     "Original Name",
		Type:     models.AccountTypeEquity,
		Currency: "INR",
		Status:   models.AccountStatusActive,
	}
	accountRepo.AddAccount(testAccount)

	t.Run("update account with valid data returns 200", func(t *testing.T) {
		body := map[string]interface{}{
			"name":   "Updated Name",
			"status": "active",
		}

		rec, resp := makeRequestWithPathValue(t, handler.UpdateAccount, http.MethodPut, "/api/v1/accounts/acct-update-test", "id", "acct-update-test", body)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var acct map[string]interface{}
		err := json.Unmarshal(resp.Data, &acct)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", acct["name"])
	})

	t.Run("update account without ID returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"name":   "Test",
			"status": "active",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/accounts/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.UpdateAccount(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("update account with missing name returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			// missing name
			"status": "active",
		}

		rec, resp := makeRequestWithPathValue(t, handler.UpdateAccount, http.MethodPut, "/api/v1/accounts/acct-update-test", "id", "acct-update-test", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestLedgerHandler_GetAccountBalance(t *testing.T) {
	ledgerService, accountRepo, _ := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a test account with balance
	testAccount := &models.Account{
		ID:       "acct-balance-test",
		Code:     "1500",
		Name:     "Cash Account",
		Type:     models.AccountTypeAsset,
		Currency: "INR",
		Balance:  500000, // 5000 rupees
		Status:   models.AccountStatusActive,
	}
	accountRepo.AddAccount(testAccount)

	t.Run("get account balance returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccountBalance, http.MethodGet, "/api/v1/accounts/acct-balance-test/balance", "id", "acct-balance-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var balance map[string]interface{}
		err := json.Unmarshal(resp.Data, &balance)
		require.NoError(t, err)
		assert.Equal(t, "acct-balance-test", balance["account_id"])
		assert.Equal(t, float64(500000), balance["balance"])
	})

	t.Run("get balance for non-existent account returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetAccountBalance, http.MethodGet, "/api/v1/accounts/non-existent/balance", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get balance without account ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts//balance", nil)
		rec := httptest.NewRecorder()
		handler.GetAccountBalance(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// ============================================================
// Journal Entry Handler Tests
// ============================================================

func TestLedgerHandler_GetJournalEntry(t *testing.T) {
	ledgerService, _, journalRepo := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a test journal entry
	testEntry := &models.JournalEntry{
		ID:          "je-get-test",
		EntryNumber: "JE-001",
		Type:        models.EntryTypeStandard,
		Description: "Test journal entry",
		Status:      models.EntryStatusDraft,
	}
	journalRepo.AddEntry(testEntry)

	t.Run("get existing journal entry returns 200", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetJournalEntry, http.MethodGet, "/api/v1/journal-entries/je-get-test", "id", "je-get-test", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var entry map[string]interface{}
		err := json.Unmarshal(resp.Data, &entry)
		require.NoError(t, err)
		assert.Equal(t, "je-get-test", entry["id"])
	})

	t.Run("get non-existent journal entry returns 404", func(t *testing.T) {
		rec, resp := makeRequestWithPathValue(t, handler.GetJournalEntry, http.MethodGet, "/api/v1/journal-entries/non-existent", "id", "non-existent", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "NOT_FOUND", resp.Error.Code)
	})

	t.Run("get journal entry without ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/journal-entries/", nil)
		rec := httptest.NewRecorder()
		handler.GetJournalEntry(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestLedgerHandler_ListJournalEntries(t *testing.T) {
	ledgerService, _, journalRepo := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add test entries
	entry1 := &models.JournalEntry{
		ID:          "je-list-1",
		EntryNumber: "JE-LIST-001",
		Type:        models.EntryTypeStandard,
		Description: "Entry 1",
		Status:      models.EntryStatusDraft,
	}
	entry2 := &models.JournalEntry{
		ID:          "je-list-2",
		EntryNumber: "JE-LIST-002",
		Type:        models.EntryTypeStandard,
		Description: "Entry 2",
		Status:      models.EntryStatusPosted,
	}
	journalRepo.AddEntry(entry1)
	journalRepo.AddEntry(entry2)

	t.Run("list all journal entries returns 200", func(t *testing.T) {
		rec, resp := makeRequest(t, handler.ListJournalEntries, http.MethodGet, "/api/v1/journal-entries", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, resp.Success)

		var entries []map[string]interface{}
		err := json.Unmarshal(resp.Data, &entries)
		require.NoError(t, err)
		assert.Len(t, entries, 2)
	})

	t.Run("list journal entries with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/journal-entries?status=posted", nil)
		rec := httptest.NewRecorder()
		handler.ListJournalEntries(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestLedgerHandler_PostJournalEntry(t *testing.T) {
	ledgerService, _, journalRepo := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a draft entry
	draftEntry := &models.JournalEntry{
		ID:          "je-post-test",
		EntryNumber: "JE-POST-001",
		Type:        models.EntryTypeStandard,
		Description: "Draft entry to post",
		Status:      models.EntryStatusDraft,
	}
	journalRepo.AddEntry(draftEntry)

	t.Run("post journal entry without ID returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"entry_id":  "je-post-test",
			"posted_by": "user-123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/journal-entries//post", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.PostJournalEntry(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("post journal entry with ID mismatch returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"entry_id":  "different-id",
			"posted_by": "user-123",
		}

		rec, resp := makeRequestWithPathValue(t, handler.PostJournalEntry, http.MethodPost, "/api/v1/journal-entries/je-post-test/post", "id", "je-post-test", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
	})

	t.Run("post journal entry with missing posted_by returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"entry_id": "je-post-test",
			// missing posted_by
		}

		rec, resp := makeRequestWithPathValue(t, handler.PostJournalEntry, http.MethodPost, "/api/v1/journal-entries/je-post-test/post", "id", "je-post-test", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestLedgerHandler_VoidJournalEntry(t *testing.T) {
	ledgerService, _, journalRepo := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a posted entry
	postedEntry := &models.JournalEntry{
		ID:          "je-void-test",
		EntryNumber: "JE-VOID-001",
		Type:        models.EntryTypeStandard,
		Description: "Posted entry to void",
		Status:      models.EntryStatusPosted,
	}
	journalRepo.AddEntry(postedEntry)

	t.Run("void journal entry without ID returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"entry_id":    "je-void-test",
			"voided_by":   "user-123",
			"void_reason": "Test void reason",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/journal-entries//void", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.VoidJournalEntry(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("void journal entry with missing voided_by returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"entry_id": "je-void-test",
			// missing voided_by
			"void_reason": "Test void reason",
		}

		rec, resp := makeRequestWithPathValue(t, handler.VoidJournalEntry, http.MethodPost, "/api/v1/journal-entries/je-void-test/void", "id", "je-void-test", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}

func TestLedgerHandler_ReverseJournalEntry(t *testing.T) {
	ledgerService, _, journalRepo := createTestLedgerService()
	handler := NewLedgerHandler(ledgerService)

	// Setup: Add a posted entry
	postedEntry := &models.JournalEntry{
		ID:          "je-reverse-test",
		EntryNumber: "JE-REVERSE-001",
		Type:        models.EntryTypeStandard,
		Description: "Posted entry to reverse",
		Status:      models.EntryStatusPosted,
	}
	journalRepo.AddEntry(postedEntry)

	t.Run("reverse journal entry without ID returns 400", func(t *testing.T) {
		body := map[string]interface{}{
			"reversed_by": "user-123",
			"reason":      "Test reversal reason for journal entry",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/journal-entries//reverse", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ReverseJournalEntry(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("reverse journal entry with short reason returns validation error", func(t *testing.T) {
		body := map[string]interface{}{
			"reversed_by": "user-123",
			"reason":      "Short", // less than 10 characters
		}

		rec, resp := makeRequestWithPathValue(t, handler.ReverseJournalEntry, http.MethodPost, "/api/v1/journal-entries/je-reverse-test/reverse", "id", "je-reverse-test", body)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
	})
}
