package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// =====================================================================
// Mock Repositories for Testing
// =====================================================================

type mockTransactionRepository struct {
	transactions     map[string]*models.Transaction
	createFunc       func(ctx context.Context, transaction *models.Transaction) *errors.Error
	getByIDFunc      func(ctx context.Context, id string) (*models.Transaction, *errors.Error)
	listByWalletFunc func(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
}

func (m *mockTransactionRepository) Create(ctx context.Context, transaction *models.Transaction) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, transaction)
	}
	transaction.ID = uuid.New().String()
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *mockTransactionRepository) GetByID(ctx context.Context, id string) (*models.Transaction, *errors.Error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	tx, ok := m.transactions[id]
	if !ok {
		return nil, errors.NotFound("transaction")
	}
	return tx, nil
}

func (m *mockTransactionRepository) ListByWallet(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	if m.listByWalletFunc != nil {
		return m.listByWalletFunc(ctx, walletID, filter)
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

// =====================================================================
// Test Helpers
// =====================================================================

// Compile-time interface check
var _ TransactionRepositoryInterface = (*mockTransactionRepository)(nil)

func setupTestService() (*TransactionService, *mockTransactionRepository) {
	repo := &mockTransactionRepository{
		transactions: make(map[string]*models.Transaction),
	}
	service := NewTransactionService(repo)
	return service, repo
}

// =====================================================================
// CreateTransfer Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestCreateTransfer_Success(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	sourceWalletID := uuid.New().String()
	destWalletID := uuid.New().String()

	req := &models.CreateTransferRequest{
		SourceWalletID:      sourceWalletID,
		DestinationWalletID: destWalletID,
		Amount:              50000, // ₹500.00
		Currency:            sharedModels.INR,
		Description:         "Test transfer",
		Reference:           "REF-001",
	}

	tx, err := service.CreateTransfer(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Type != models.TransactionTypeTransfer {
		t.Errorf("expected transfer type, got %s", tx.Type)
	}
	if tx.Status != models.TransactionStatusPending {
		t.Errorf("expected pending status, got %s", tx.Status)
	}
	if tx.SourceWalletID == nil || *tx.SourceWalletID != sourceWalletID {
		t.Errorf("expected source wallet %s, got %v", sourceWalletID, tx.SourceWalletID)
	}
	if tx.DestinationWalletID == nil || *tx.DestinationWalletID != destWalletID {
		t.Errorf("expected dest wallet %s, got %v", destWalletID, tx.DestinationWalletID)
	}
	if tx.Amount != 50000 {
		t.Errorf("expected amount 50000, got %d", tx.Amount)
	}
	if tx.Currency != sharedModels.INR {
		t.Errorf("expected INR currency, got %s", tx.Currency)
	}
}

func TestCreateTransfer_Error_SameWallet(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	walletID := uuid.New().String()

	req := &models.CreateTransferRequest{
		SourceWalletID:      walletID,
		DestinationWalletID: walletID, // Same as source!
		Amount:              50000,
		Currency:            sharedModels.INR,
		Description:         "Test transfer",
	}

	_, err := service.CreateTransfer(ctx, req)
	if err == nil {
		t.Fatal("expected error for same source and destination wallet, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
	if err.Message != "source and destination wallets must be different" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestCreateTransfer_Success_WithoutReference(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	req := &models.CreateTransferRequest{
		SourceWalletID:      uuid.New().String(),
		DestinationWalletID: uuid.New().String(),
		Amount:              10000,
		Currency:            sharedModels.INR,
		Description:         "Test transfer without reference",
		Reference:           "", // Empty reference
	}

	tx, err := service.CreateTransfer(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Reference != nil {
		t.Errorf("expected nil reference, got %v", tx.Reference)
	}
}

// =====================================================================
// CreateDeposit Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestCreateDeposit_Success(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	walletID := uuid.New().String()

	req := &models.CreateDepositRequest{
		WalletID:    walletID,
		Amount:      100000, // ₹1000.00
		Currency:    sharedModels.INR,
		Description: "Test deposit",
		Reference:   "DEPOSIT-001",
	}

	tx, err := service.CreateDeposit(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Type != models.TransactionTypeDeposit {
		t.Errorf("expected deposit type, got %s", tx.Type)
	}
	if tx.Status != models.TransactionStatusPending {
		t.Errorf("expected pending status, got %s", tx.Status)
	}
	if tx.DestinationWalletID == nil || *tx.DestinationWalletID != walletID {
		t.Errorf("expected dest wallet %s, got %v", walletID, tx.DestinationWalletID)
	}
	if tx.SourceWalletID != nil {
		t.Errorf("expected nil source wallet for deposit, got %v", tx.SourceWalletID)
	}
	if tx.Amount != 100000 {
		t.Errorf("expected amount 100000, got %d", tx.Amount)
	}
}

func TestCreateDeposit_Success_WithoutReference(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	req := &models.CreateDepositRequest{
		WalletID:    uuid.New().String(),
		Amount:      50000,
		Currency:    sharedModels.INR,
		Description: "Test deposit without reference",
		Reference:   "",
	}

	tx, err := service.CreateDeposit(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Reference != nil {
		t.Errorf("expected nil reference, got %v", tx.Reference)
	}
}

// =====================================================================
// CreateWithdrawal Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestCreateWithdrawal_Success(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	walletID := uuid.New().String()

	req := &models.CreateWithdrawalRequest{
		WalletID:    walletID,
		Amount:      75000, // ₹750.00
		Currency:    sharedModels.INR,
		Description: "Test withdrawal",
		Reference:   "WITHDRAWAL-001",
	}

	tx, err := service.CreateWithdrawal(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Type != models.TransactionTypeWithdrawal {
		t.Errorf("expected withdrawal type, got %s", tx.Type)
	}
	if tx.Status != models.TransactionStatusPending {
		t.Errorf("expected pending status, got %s", tx.Status)
	}
	if tx.SourceWalletID == nil || *tx.SourceWalletID != walletID {
		t.Errorf("expected source wallet %s, got %v", walletID, tx.SourceWalletID)
	}
	if tx.DestinationWalletID != nil {
		t.Errorf("expected nil dest wallet for withdrawal, got %v", tx.DestinationWalletID)
	}
	if tx.Amount != 75000 {
		t.Errorf("expected amount 75000, got %d", tx.Amount)
	}
}

func TestCreateWithdrawal_Success_WithoutReference(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	req := &models.CreateWithdrawalRequest{
		WalletID:    uuid.New().String(),
		Amount:      25000,
		Currency:    sharedModels.INR,
		Description: "Test withdrawal without reference",
		Reference:   "",
	}

	tx, err := service.CreateWithdrawal(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx.Reference != nil {
		t.Errorf("expected nil reference, got %v", tx.Reference)
	}
}

// =====================================================================
// ReverseTransaction Tests - CRITICAL PATH (100% coverage needed)
// =====================================================================

func TestReverseTransaction_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	// Create a completed transfer transaction
	sourceWalletID := uuid.New().String()
	destWalletID := uuid.New().String()
	originalTx := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusCompleted, // Completed
		SourceWalletID:      &sourceWalletID,
		DestinationWalletID: &destWalletID,
		Amount:              50000,
		Currency:            sharedModels.INR,
		Description:         "Original transfer",
	}
	repo.transactions[originalTx.ID] = originalTx

	// Reverse the transaction
	reversalTx, err := service.ReverseTransaction(ctx, originalTx.ID, "correction needed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify reversal transaction
	if reversalTx.Type != models.TransactionTypeReversal {
		t.Errorf("expected reversal type, got %s", reversalTx.Type)
	}
	if reversalTx.Status != models.TransactionStatusPending {
		t.Errorf("expected pending status, got %s", reversalTx.Status)
	}
	if reversalTx.Amount != originalTx.Amount {
		t.Errorf("expected amount %d, got %d", originalTx.Amount, reversalTx.Amount)
	}

	// Verify direction is reversed (source becomes dest, dest becomes source)
	if reversalTx.SourceWalletID == nil || *reversalTx.SourceWalletID != *originalTx.DestinationWalletID {
		t.Errorf("expected reversal source wallet to be original dest wallet")
	}
	if reversalTx.DestinationWalletID == nil || *reversalTx.DestinationWalletID != *originalTx.SourceWalletID {
		t.Errorf("expected reversal dest wallet to be original source wallet")
	}

	// Verify parent transaction ID is set
	if reversalTx.ParentTransactionID == nil || *reversalTx.ParentTransactionID != originalTx.ID {
		t.Errorf("expected parent transaction ID %s, got %v", originalTx.ID, reversalTx.ParentTransactionID)
	}

	// Verify metadata contains reversal reason
	if reason, ok := reversalTx.Metadata["reversal_reason"]; !ok || reason != "correction needed" {
		t.Errorf("expected reversal reason in metadata, got %v", reversalTx.Metadata)
	}
}

func TestReverseTransaction_Error_NotCompleted(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	// Create a pending transaction (not completed)
	pendingTx := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusPending, // Not completed
		SourceWalletID:      ptrString(uuid.New().String()),
		DestinationWalletID: ptrString(uuid.New().String()),
		Amount:              50000,
		Currency:            sharedModels.INR,
		Description:         "Pending transfer",
	}
	repo.transactions[pendingTx.ID] = pendingTx

	_, err := service.ReverseTransaction(ctx, pendingTx.ID, "test")
	if err == nil {
		t.Fatal("expected error for non-completed transaction, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
	if err.Message != "only completed transactions can be reversed" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestReverseTransaction_Error_CannotReverseReversal(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	// Create a reversal transaction (completed)
	reversalTx := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeReversal, // Already a reversal
		Status:              models.TransactionStatusCompleted,
		SourceWalletID:      ptrString(uuid.New().String()),
		DestinationWalletID: ptrString(uuid.New().String()),
		Amount:              50000,
		Currency:            sharedModels.INR,
		Description:         "Reversal",
	}
	repo.transactions[reversalTx.ID] = reversalTx

	_, err := service.ReverseTransaction(ctx, reversalTx.ID, "test")
	if err == nil {
		t.Fatal("expected error for reversing a reversal, got nil")
	}
	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
	if err.Message != "cannot reverse a reversal transaction" {
		t.Errorf("unexpected error message: %s", err.Message)
	}
}

func TestReverseTransaction_Error_NotFound(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	_, err := service.ReverseTransaction(ctx, uuid.New().String(), "test")
	if err == nil {
		t.Fatal("expected error for non-existent transaction, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

func TestReverseTransaction_Success_DepositReversal(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	// Create a completed deposit transaction (only has destination wallet)
	walletID := uuid.New().String()
	depositTx := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusCompleted,
		DestinationWalletID: &walletID,
		Amount:              100000,
		Currency:            sharedModels.INR,
		Description:         "Deposit",
	}
	repo.transactions[depositTx.ID] = depositTx

	// Reverse the deposit
	reversalTx, err := service.ReverseTransaction(ctx, depositTx.ID, "refund")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify reversal reverses direction (deposit becomes withdrawal)
	if reversalTx.SourceWalletID == nil || *reversalTx.SourceWalletID != walletID {
		t.Errorf("expected source wallet to be original dest wallet")
	}
	if reversalTx.DestinationWalletID != nil {
		t.Errorf("expected nil destination wallet for deposit reversal, got %v", reversalTx.DestinationWalletID)
	}
}

func TestReverseTransaction_Success_WithdrawalReversal(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	// Create a completed withdrawal transaction (only has source wallet)
	walletID := uuid.New().String()
	withdrawalTx := &models.Transaction{
		ID:             uuid.New().String(),
		Type:           models.TransactionTypeWithdrawal,
		Status:         models.TransactionStatusCompleted,
		SourceWalletID: &walletID,
		Amount:         50000,
		Currency:       sharedModels.INR,
		Description:    "Withdrawal",
	}
	repo.transactions[withdrawalTx.ID] = withdrawalTx

	// Reverse the withdrawal
	reversalTx, err := service.ReverseTransaction(ctx, withdrawalTx.ID, "cancelled")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify reversal reverses direction (withdrawal becomes deposit)
	if reversalTx.DestinationWalletID == nil || *reversalTx.DestinationWalletID != walletID {
		t.Errorf("expected dest wallet to be original source wallet")
	}
	if reversalTx.SourceWalletID != nil {
		t.Errorf("expected nil source wallet for withdrawal reversal, got %v", reversalTx.SourceWalletID)
	}
}

// =====================================================================
// GetTransaction Tests
// =====================================================================

func TestGetTransaction_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	tx := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusCompleted,
		DestinationWalletID: ptrString(uuid.New().String()),
		Amount:              10000,
		Currency:            sharedModels.INR,
		Description:         "Test transaction",
	}
	repo.transactions[tx.ID] = tx

	retrieved, err := service.GetTransaction(ctx, tx.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != tx.ID {
		t.Errorf("expected ID %s, got %s", tx.ID, retrieved.ID)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	service, _ := setupTestService()
	ctx := context.Background()

	_, err := service.GetTransaction(ctx, uuid.New().String())
	if err == nil {
		t.Fatal("expected error for non-existent transaction, got nil")
	}
	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// =====================================================================
// ListWalletTransactions Tests
// =====================================================================

func TestListWalletTransactions_Success(t *testing.T) {
	service, repo := setupTestService()
	ctx := context.Background()

	walletID := uuid.New().String()

	// Create transactions for this wallet
	tx1 := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusCompleted,
		DestinationWalletID: &walletID,
		Amount:              10000,
		Currency:            sharedModels.INR,
		Description:         "Deposit",
	}
	tx2 := &models.Transaction{
		ID:             uuid.New().String(),
		Type:           models.TransactionTypeWithdrawal,
		Status:         models.TransactionStatusCompleted,
		SourceWalletID: &walletID,
		Amount:         5000,
		Currency:       sharedModels.INR,
		Description:    "Withdrawal",
	}
	tx3 := &models.Transaction{
		ID:                  uuid.New().String(),
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusCompleted,
		DestinationWalletID: ptrString(uuid.New().String()), // Different wallet
		Amount:              20000,
		Currency:            sharedModels.INR,
		Description:         "Other wallet deposit",
	}

	repo.transactions[tx1.ID] = tx1
	repo.transactions[tx2.ID] = tx2
	repo.transactions[tx3.ID] = tx3

	transactions, err := service.ListWalletTransactions(ctx, walletID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(transactions))
	}
}

// =====================================================================
// Helper Functions
// =====================================================================

func ptrString(s string) *string {
	return &s
}
