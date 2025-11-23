package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/vnykmshr/nivo/services/ledger/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// =====================================================================
// Mock Repositories for Testing
// =====================================================================

type mockAccountRepository struct {
	accounts       map[string]*models.Account
	getByIDFunc    func(ctx context.Context, id string) (*models.Account, *errors.Error)
	getByCodeFunc  func(ctx context.Context, code string) (*models.Account, *errors.Error)
	createFunc     func(ctx context.Context, account *models.Account) *errors.Error
	updateFunc     func(ctx context.Context, account *models.Account) *errors.Error
	getBalanceFunc func(ctx context.Context, accountID string) (int64, *errors.Error)
}

func (m *mockAccountRepository) GetByID(ctx context.Context, id string) (*models.Account, *errors.Error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	account, ok := m.accounts[id]
	if !ok {
		return nil, errors.NotFound("account not found")
	}
	return account, nil
}

func (m *mockAccountRepository) GetByCode(ctx context.Context, code string) (*models.Account, *errors.Error) {
	if m.getByCodeFunc != nil {
		return m.getByCodeFunc(ctx, code)
	}
	for _, account := range m.accounts {
		if account.Code == code {
			return account, nil
		}
	}
	return nil, errors.NotFound("account not found")
}

func (m *mockAccountRepository) Create(ctx context.Context, account *models.Account) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, account)
	}
	account.ID = uuid.New().String()
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) Update(ctx context.Context, account *models.Account) *errors.Error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, account)
	}
	if _, ok := m.accounts[account.ID]; !ok {
		return errors.NotFound("account not found")
	}
	m.accounts[account.ID] = account
	return nil
}

func (m *mockAccountRepository) List(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error) {
	return nil, nil
}

func (m *mockAccountRepository) GetBalance(ctx context.Context, accountID string) (int64, *errors.Error) {
	if m.getBalanceFunc != nil {
		return m.getBalanceFunc(ctx, accountID)
	}
	account, err := m.GetByID(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return account.Balance, nil
}

type mockJournalEntryRepository struct {
	entries     map[string]*models.JournalEntry
	getByIDFunc func(ctx context.Context, id string) (*models.JournalEntry, *errors.Error)
	createFunc  func(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error
	postFunc    func(ctx context.Context, entryID, postedBy string) *errors.Error
	voidFunc    func(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error
}

func (m *mockJournalEntryRepository) GetByID(ctx context.Context, id string) (*models.JournalEntry, *errors.Error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	entry, ok := m.entries[id]
	if !ok {
		return nil, errors.NotFound("journal entry not found")
	}
	return entry, nil
}

func (m *mockJournalEntryRepository) Create(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, entry, lines)
	}
	entry.ID = uuid.New().String()
	entry.EntryNumber = "JE-2025-00001"
	entry.Lines = lines
	// Set entry ID on lines
	for i := range entry.Lines {
		entry.Lines[i].EntryID = entry.ID
		entry.Lines[i].ID = uuid.New().String()
	}
	m.entries[entry.ID] = entry
	return nil
}

func (m *mockJournalEntryRepository) Post(ctx context.Context, entryID, postedBy string) *errors.Error {
	if m.postFunc != nil {
		return m.postFunc(ctx, entryID, postedBy)
	}
	entry, ok := m.entries[entryID]
	if !ok {
		return errors.NotFound("journal entry not found")
	}
	entry.Status = models.EntryStatusPosted
	entry.PostedBy = &postedBy
	return nil
}

func (m *mockJournalEntryRepository) Void(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error {
	if m.voidFunc != nil {
		return m.voidFunc(ctx, entryID, voidedBy, voidReason)
	}
	entry, ok := m.entries[entryID]
	if !ok {
		return errors.NotFound("journal entry not found")
	}
	entry.Status = models.EntryStatusVoided
	entry.VoidedBy = &voidedBy
	entry.VoidReason = &voidReason
	return nil
}

func (m *mockJournalEntryRepository) List(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error) {
	return nil, nil
}

// =====================================================================
// Test Helpers
// =====================================================================

// Compile-time interface checks
var _ AccountRepositoryInterface = (*mockAccountRepository)(nil)
var _ JournalEntryRepositoryInterface = (*mockJournalEntryRepository)(nil)

func setupTestService() (*LedgerService, *mockAccountRepository, *mockJournalEntryRepository) {
	accountRepo := &mockAccountRepository{
		accounts: make(map[string]*models.Account),
	}
	journalRepo := &mockJournalEntryRepository{
		entries: make(map[string]*models.JournalEntry),
	}
	service := NewLedgerService(accountRepo, journalRepo)
	return service, accountRepo, journalRepo
}

func createTestAccount(id, code, name string, accountType models.AccountType) *models.Account {
	return &models.Account{
		ID:       id,
		Code:     code,
		Name:     name,
		Type:     accountType,
		Currency: "INR",
		Status:   models.AccountStatusActive,
		Balance:  0,
	}
}

// =====================================================================
// CreateAccount Tests
// =====================================================================

func TestCreateAccount_Success(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	req := &models.CreateAccountRequest{
		Code:     "1000",
		Name:     "Cash",
		Type:     models.AccountTypeAsset,
		Currency: "INR",
	}

	account, err := service.CreateAccount(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if account.Code != req.Code {
		t.Errorf("expected code %s, got %s", req.Code, account.Code)
	}
	if account.Status != models.AccountStatusActive {
		t.Errorf("expected status active, got %s", account.Status)
	}
	if account.Balance != 0 {
		t.Errorf("expected balance 0, got %d", account.Balance)
	}
}

func TestCreateAccount_InvalidParent(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	parentID := uuid.New().String()
	req := &models.CreateAccountRequest{
		Code:     "1100",
		Name:     "Petty Cash",
		Type:     models.AccountTypeAsset,
		Currency: "INR",
		ParentID: &parentID, // Non-existent parent
	}

	_, err := service.CreateAccount(ctx, req)
	if err == nil {
		t.Fatal("expected error for non-existent parent, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// CreateJournalEntry Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestCreateJournalEntry_Success_BalancedEntry(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	// Create test accounts
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	revenueAccount := createTestAccount(uuid.New().String(), "4000", "Revenue", models.AccountTypeRevenue)
	accountRepo.accounts[cashAccount.ID] = cashAccount
	accountRepo.accounts[revenueAccount.ID] = revenueAccount

	req := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeStandard,
		Description:   "Test balanced entry",
		ReferenceType: "test",
		ReferenceID:   "test-001",
		Lines: []models.LedgerLineInput{
			{
				AccountID:    cashAccount.ID,
				DebitAmount:  10000, // ₹100.00 debit (increase asset)
				CreditAmount: 0,
				Description:  "Cash received",
			},
			{
				AccountID:    revenueAccount.ID,
				DebitAmount:  0,
				CreditAmount: 10000, // ₹100.00 credit (increase revenue)
				Description:  "Revenue earned",
			},
		},
	}

	entry, err := service.CreateJournalEntry(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if entry.Status != models.EntryStatusDraft {
		t.Errorf("expected draft status, got %s", entry.Status)
	}
	if len(entry.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(entry.Lines))
	}
}

func TestCreateJournalEntry_Error_UnbalancedEntry(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	// Create test accounts
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	revenueAccount := createTestAccount(uuid.New().String(), "4000", "Revenue", models.AccountTypeRevenue)
	accountRepo.accounts[cashAccount.ID] = cashAccount
	accountRepo.accounts[revenueAccount.ID] = revenueAccount

	req := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeStandard,
		Description:   "Test unbalanced entry",
		ReferenceType: "test",
		ReferenceID:   "test-002",
		Lines: []models.LedgerLineInput{
			{
				AccountID:    cashAccount.ID,
				DebitAmount:  10000, // ₹100.00 debit
				CreditAmount: 0,
				Description:  "Cash received",
			},
			{
				AccountID:    revenueAccount.ID,
				DebitAmount:  0,
				CreditAmount: 5000, // ₹50.00 credit - UNBALANCED!
				Description:  "Revenue earned",
			},
		},
	}

	_, err := service.CreateJournalEntry(ctx, req)
	if err == nil {
		t.Fatal("expected error for unbalanced entry, got nil")
	}
	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
	if err.Message != "entry not balanced: debits=10000, credits=5000" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestCreateJournalEntry_Error_InsufficientLines(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	// Create test account
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	accountRepo.accounts[cashAccount.ID] = cashAccount

	req := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeStandard,
		Description:   "Test single line entry",
		ReferenceType: "test",
		ReferenceID:   "test-003",
		Lines: []models.LedgerLineInput{
			{
				AccountID:    cashAccount.ID,
				DebitAmount:  10000,
				CreditAmount: 0,
				Description:  "Single line",
			},
		},
	}

	_, err := service.CreateJournalEntry(ctx, req)
	if err == nil {
		t.Fatal("expected error for insufficient lines, got nil")
	}
	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
}

func TestCreateJournalEntry_Error_InactiveAccount(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	// Create inactive account
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	cashAccount.Status = models.AccountStatusInactive
	revenueAccount := createTestAccount(uuid.New().String(), "4000", "Revenue", models.AccountTypeRevenue)
	accountRepo.accounts[cashAccount.ID] = cashAccount
	accountRepo.accounts[revenueAccount.ID] = revenueAccount

	req := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeStandard,
		Description:   "Test with inactive account",
		ReferenceType: "test",
		ReferenceID:   "test-004",
		Lines: []models.LedgerLineInput{
			{
				AccountID:    cashAccount.ID,
				DebitAmount:  10000,
				CreditAmount: 0,
				Description:  "Cash received",
			},
			{
				AccountID:    revenueAccount.ID,
				DebitAmount:  0,
				CreditAmount: 10000,
				Description:  "Revenue earned",
			},
		},
	}

	_, err := service.CreateJournalEntry(ctx, req)
	if err == nil {
		t.Fatal("expected error for inactive account, got nil")
	}
	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
}

func TestCreateJournalEntry_Error_NonExistentAccount(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	req := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeStandard,
		Description:   "Test with non-existent account",
		ReferenceType: "test",
		ReferenceID:   "test-005",
		Lines: []models.LedgerLineInput{
			{
				AccountID:    uuid.New().String(), // Non-existent account
				DebitAmount:  10000,
				CreditAmount: 0,
				Description:  "Invalid",
			},
			{
				AccountID:    uuid.New().String(), // Non-existent account
				DebitAmount:  0,
				CreditAmount: 10000,
				Description:  "Invalid",
			},
		},
	}

	_, err := service.CreateJournalEntry(ctx, req)
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
}

// =====================================================================
// PostJournalEntry Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestPostJournalEntry_Success(t *testing.T) {
	service, accountRepo, journalRepo := setupTestService()
	ctx := context.Background()

	// Create test accounts
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	revenueAccount := createTestAccount(uuid.New().String(), "4000", "Revenue", models.AccountTypeRevenue)
	accountRepo.accounts[cashAccount.ID] = cashAccount
	accountRepo.accounts[revenueAccount.ID] = revenueAccount

	// Create draft entry
	entry := &models.JournalEntry{
		ID:            uuid.New().String(),
		EntryNumber:   "JE-2025-00001",
		Type:          models.EntryTypeStandard,
		Status:        models.EntryStatusDraft,
		Description:   "Test entry",
		ReferenceType: "test",
		ReferenceID:   "test-001",
		Lines: []models.LedgerLine{
			{
				ID:           uuid.New().String(),
				AccountID:    cashAccount.ID,
				DebitAmount:  10000,
				CreditAmount: 0,
				Description:  "Cash received",
			},
			{
				ID:           uuid.New().String(),
				AccountID:    revenueAccount.ID,
				DebitAmount:  0,
				CreditAmount: 10000,
				Description:  "Revenue earned",
			},
		},
	}
	journalRepo.entries[entry.ID] = entry

	// Post the entry
	postedEntry, err := service.PostJournalEntry(ctx, entry.ID, "user-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if postedEntry.Status != models.EntryStatusPosted {
		t.Errorf("expected posted status, got %s", postedEntry.Status)
	}
	if postedEntry.PostedBy == nil || *postedEntry.PostedBy != "user-123" {
		t.Errorf("expected posted by user-123, got %v", postedEntry.PostedBy)
	}
}

func TestPostJournalEntry_Error_NotDraft(t *testing.T) {
	service, _, journalRepo := setupTestService()
	ctx := context.Background()

	// Create already-posted entry
	entry := &models.JournalEntry{
		ID:          uuid.New().String(),
		EntryNumber: "JE-2025-00001",
		Type:        models.EntryTypeStandard,
		Status:      models.EntryStatusPosted, // Already posted
		Description: "Test entry",
		Lines: []models.LedgerLine{
			{DebitAmount: 10000, CreditAmount: 0},
			{DebitAmount: 0, CreditAmount: 10000},
		},
	}
	journalRepo.entries[entry.ID] = entry

	_, err := service.PostJournalEntry(ctx, entry.ID, "user-123")
	if err == nil {
		t.Fatal("expected error for non-draft entry, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestPostJournalEntry_Error_EntryNotFound(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	_, err := service.PostJournalEntry(ctx, uuid.New().String(), "user-123")
	if err == nil {
		t.Fatal("expected error for non-existent entry, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// ReverseJournalEntry Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestReverseJournalEntry_Success(t *testing.T) {
	service, accountRepo, journalRepo := setupTestService()
	ctx := context.Background()

	// Create test accounts
	cashAccount := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	revenueAccount := createTestAccount(uuid.New().String(), "4000", "Revenue", models.AccountTypeRevenue)
	accountRepo.accounts[cashAccount.ID] = cashAccount
	accountRepo.accounts[revenueAccount.ID] = revenueAccount

	// Create posted original entry
	originalEntry := &models.JournalEntry{
		ID:            uuid.New().String(),
		EntryNumber:   "JE-2025-00001",
		Type:          models.EntryTypeStandard,
		Status:        models.EntryStatusPosted,
		Description:   "Original entry",
		ReferenceType: "sale",
		ReferenceID:   "sale-001",
		Lines: []models.LedgerLine{
			{
				ID:           uuid.New().String(),
				AccountID:    cashAccount.ID,
				DebitAmount:  10000, // Cash debit
				CreditAmount: 0,
				Description:  "Cash received",
			},
			{
				ID:           uuid.New().String(),
				AccountID:    revenueAccount.ID,
				DebitAmount:  0,
				CreditAmount: 10000, // Revenue credit
				Description:  "Revenue earned",
			},
		},
	}
	journalRepo.entries[originalEntry.ID] = originalEntry

	// Reverse the entry
	reversalEntry, err := service.ReverseJournalEntry(ctx, originalEntry.ID, "user-123", "correction needed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify reversal entry
	if reversalEntry.Type != models.EntryTypeReversing {
		t.Errorf("expected reversing type, got %s", reversalEntry.Type)
	}
	if reversalEntry.Status != models.EntryStatusPosted {
		t.Errorf("expected posted status, got %s", reversalEntry.Status)
	}
	if len(reversalEntry.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(reversalEntry.Lines))
	}

	// Verify amounts are swapped (debit becomes credit, credit becomes debit)
	for i, line := range reversalEntry.Lines {
		originalLine := originalEntry.Lines[i]
		if line.DebitAmount != originalLine.CreditAmount {
			t.Errorf("line %d: expected debit %d, got %d", i, originalLine.CreditAmount, line.DebitAmount)
		}
		if line.CreditAmount != originalLine.DebitAmount {
			t.Errorf("line %d: expected credit %d, got %d", i, originalLine.DebitAmount, line.CreditAmount)
		}
	}
}

func TestReverseJournalEntry_Error_NotPosted(t *testing.T) {
	service, _, journalRepo := setupTestService()
	ctx := context.Background()

	// Create draft entry (not posted)
	draftEntry := &models.JournalEntry{
		ID:          uuid.New().String(),
		EntryNumber: "JE-2025-00001",
		Type:        models.EntryTypeStandard,
		Status:      models.EntryStatusDraft, // Not posted
		Description: "Draft entry",
		Lines: []models.LedgerLine{
			{DebitAmount: 10000, CreditAmount: 0},
			{DebitAmount: 0, CreditAmount: 10000},
		},
	}
	journalRepo.entries[draftEntry.ID] = draftEntry

	_, err := service.ReverseJournalEntry(ctx, draftEntry.ID, "user-123", "test")
	if err == nil {
		t.Fatal("expected error for non-posted entry, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// =====================================================================
// VoidJournalEntry Tests
// =====================================================================

func TestVoidJournalEntry_Success(t *testing.T) {
	service, _, journalRepo := setupTestService()
	ctx := context.Background()

	// Create posted entry
	entry := &models.JournalEntry{
		ID:          uuid.New().String(),
		EntryNumber: "JE-2025-00001",
		Type:        models.EntryTypeStandard,
		Status:      models.EntryStatusPosted,
		Description: "Test entry",
		Lines: []models.LedgerLine{
			{DebitAmount: 10000, CreditAmount: 0},
			{DebitAmount: 0, CreditAmount: 10000},
		},
	}
	journalRepo.entries[entry.ID] = entry

	voidedEntry, err := service.VoidJournalEntry(ctx, entry.ID, "user-123", "error in entry")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if voidedEntry.Status != models.EntryStatusVoided {
		t.Errorf("expected voided status, got %s", voidedEntry.Status)
	}
	if voidedEntry.VoidedBy == nil || *voidedEntry.VoidedBy != "user-123" {
		t.Errorf("expected voided by user-123, got %v", voidedEntry.VoidedBy)
	}
	if voidedEntry.VoidReason == nil || *voidedEntry.VoidReason != "error in entry" {
		t.Errorf("expected void reason, got %v", voidedEntry.VoidReason)
	}
}

func TestVoidJournalEntry_Error_NotPosted(t *testing.T) {
	service, _, journalRepo := setupTestService()
	ctx := context.Background()

	// Create draft entry
	entry := &models.JournalEntry{
		ID:          uuid.New().String(),
		EntryNumber: "JE-2025-00001",
		Type:        models.EntryTypeStandard,
		Status:      models.EntryStatusDraft,
		Description: "Test entry",
		Lines: []models.LedgerLine{
			{DebitAmount: 10000, CreditAmount: 0},
			{DebitAmount: 0, CreditAmount: 10000},
		},
	}
	journalRepo.entries[entry.ID] = entry

	_, err := service.VoidJournalEntry(ctx, entry.ID, "user-123", "test")
	if err == nil {
		t.Fatal("expected error for non-posted entry, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// =====================================================================
// GetAccount Tests
// =====================================================================

func TestGetAccount_Success(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	account := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	accountRepo.accounts[account.ID] = account

	retrieved, err := service.GetAccount(ctx, account.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != account.ID {
		t.Errorf("expected ID %s, got %s", account.ID, retrieved.ID)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	_, err := service.GetAccount(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// GetAccountBalance Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestGetAccountBalance_Success(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	account := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	account.Balance = 50000 // ₹500.00
	accountRepo.accounts[account.ID] = account

	balance, err := service.GetAccountBalance(ctx, account.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if balance != 50000 {
		t.Errorf("expected balance 50000, got %d", balance)
	}
}

func TestGetAccountBalance_NotFound(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	_, err := service.GetAccountBalance(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// UpdateAccount Tests
// =====================================================================

func TestUpdateAccount_Success(t *testing.T) {
	service, accountRepo, _ := setupTestService()
	ctx := context.Background()

	account := createTestAccount(uuid.New().String(), "1000", "Cash", models.AccountTypeAsset)
	accountRepo.accounts[account.ID] = account

	req := &models.UpdateAccountRequest{
		Name:   "Updated Cash Account",
		Status: models.AccountStatusActive,
	}

	updated, err := service.UpdateAccount(ctx, account.ID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, updated.Name)
	}
}

func TestUpdateAccount_NotFound(t *testing.T) {
	service, _, _ := setupTestService()
	ctx := context.Background()

	req := &models.UpdateAccountRequest{
		Name:   "Updated",
		Status: models.AccountStatusActive,
	}

	_, err := service.UpdateAccount(ctx, uuid.New().String(), req)
	if err == nil {
		t.Fatal("expected error for non-existent account, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}
