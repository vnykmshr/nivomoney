package service

import (
	"context"
	"testing"
	"time"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================================
// Mock Repository
// ============================================================================

type mockWalletRepository struct {
	wallets map[string]*models.Wallet

	// Function hooks for error injection
	createFunc       func(ctx context.Context, wallet *models.Wallet) *errors.Error
	getByIDFunc      func(ctx context.Context, id string) (*models.Wallet, *errors.Error)
	updateStatusFunc func(ctx context.Context, id string, status models.WalletStatus) *errors.Error
	closeFunc        func(ctx context.Context, id, reason string) *errors.Error
}

func newMockWalletRepository() *mockWalletRepository {
	return &mockWalletRepository{
		wallets: make(map[string]*models.Wallet),
	}
}

func (m *mockWalletRepository) Create(ctx context.Context, wallet *models.Wallet) *errors.Error {
	if m.createFunc != nil {
		return m.createFunc(ctx, wallet)
	}

	// Generate ID
	wallet.ID = "wallet_" + wallet.UserID + "_" + string(wallet.Type)
	wallet.CreatedAt = sharedModels.NewTimestamp(time.Now())
	wallet.UpdatedAt = sharedModels.NewTimestamp(time.Now())

	m.wallets[wallet.ID] = wallet

	return nil
}

func (m *mockWalletRepository) GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}

	wallet, exists := m.wallets[id]
	if !exists {
		return nil, errors.NotFound("wallet not found")
	}

	// Return a copy
	walletCopy := *wallet
	return &walletCopy, nil
}

func (m *mockWalletRepository) ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error) {
	var wallets []*models.Wallet

	for _, wallet := range m.wallets {
		if wallet.UserID == userID {
			// Filter by status if provided
			if status != nil && wallet.Status != *status {
				continue
			}
			walletCopy := *wallet
			wallets = append(wallets, &walletCopy)
		}
	}

	return wallets, nil
}

func (m *mockWalletRepository) UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, id, status)
	}

	wallet, exists := m.wallets[id]
	if !exists {
		return errors.NotFound("wallet not found")
	}

	wallet.Status = status
	wallet.UpdatedAt = sharedModels.NewTimestamp(time.Now())

	return nil
}

func (m *mockWalletRepository) Close(ctx context.Context, id, reason string) *errors.Error {
	if m.closeFunc != nil {
		return m.closeFunc(ctx, id, reason)
	}

	wallet, exists := m.wallets[id]
	if !exists {
		return errors.NotFound("wallet not found")
	}

	wallet.Status = models.WalletStatusClosed
	closedAt := sharedModels.NewTimestamp(time.Now())
	wallet.ClosedAt = &closedAt
	wallet.UpdatedAt = sharedModels.NewTimestamp(time.Now())

	return nil
}

func (m *mockWalletRepository) GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error) {
	wallet, exists := m.wallets[id]
	if !exists {
		return nil, errors.NotFound("wallet not found")
	}

	return &models.WalletBalance{
		WalletID:         wallet.ID,
		Balance:          wallet.Balance,
		AvailableBalance: wallet.AvailableBalance,
		HeldAmount:       0,
	}, nil
}

// ============================================================================
// Tests: Wallet Creation
// ============================================================================

func TestCreateWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	req := &models.CreateWalletRequest{
		UserID:          "user_123",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_savings_001",
		MetadataRaw:     []byte(`{"purpose": "personal savings"}`),
	}

	wallet, err := service.CreateWallet(ctx, req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if wallet.UserID != "user_123" {
		t.Errorf("expected user ID 'user_123', got %s", wallet.UserID)
	}

	if wallet.Type != models.WalletTypeSavings {
		t.Errorf("expected type SAVINGS, got %s", wallet.Type)
	}

	if wallet.Currency != "INR" {
		t.Errorf("expected currency INR, got %s", wallet.Currency)
	}

	if wallet.Balance != 0 {
		t.Errorf("expected zero balance, got %d", wallet.Balance)
	}

	if wallet.Status != models.WalletStatusInactive {
		t.Errorf("expected status INACTIVE, got %s", wallet.Status)
	}
}

func TestCreateWallet_Error_InvalidType(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	req := &models.CreateWalletRequest{
		UserID:          "user_123",
		Type:            "INVALID_TYPE",
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}

	_, err := service.CreateWallet(ctx, req)

	if err == nil {
		t.Fatal("expected error for invalid wallet type")
	}

	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
}

func TestCreateWallet_Error_InvalidMetadata(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	req := &models.CreateWalletRequest{
		UserID:          "user_123",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
		MetadataRaw:     []byte(`{invalid json}`),
	}

	_, err := service.CreateWallet(ctx, req)

	if err == nil {
		t.Fatal("expected error for invalid metadata")
	}

	if err.Code != errors.ErrCodeValidation {
		t.Errorf("expected validation error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Wallet Retrieval
// ============================================================================

func TestGetWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_456",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_current_001",
	}
	created, _ := service.CreateWallet(ctx, req)

	// Get wallet
	wallet, err := service.GetWallet(ctx, created.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if wallet.ID != created.ID {
		t.Errorf("expected wallet ID %s, got %s", created.ID, wallet.ID)
	}

	if wallet.Type != models.WalletTypeCurrent {
		t.Errorf("expected type CURRENT, got %s", wallet.Type)
	}
}

func TestGetWallet_Error_NotFound(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	_, err := service.GetWallet(ctx, "non_existent_wallet")

	if err == nil {
		t.Fatal("expected error for non-existent wallet")
	}

	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

func TestListUserWallets_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create multiple wallets for same user
	req1 := &models.CreateWalletRequest{
		UserID:          "user_789",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_savings_001",
	}
	service.CreateWallet(ctx, req1)

	req2 := &models.CreateWalletRequest{
		UserID:          "user_789",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_current_001",
	}
	service.CreateWallet(ctx, req2)

	// Create wallet for different user
	req3 := &models.CreateWalletRequest{
		UserID:          "user_999",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_savings_002",
	}
	service.CreateWallet(ctx, req3)

	// List wallets for user_789
	wallets, err := service.ListUserWallets(ctx, "user_789", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(wallets) != 2 {
		t.Fatalf("expected 2 wallets, got %d", len(wallets))
	}
}

func TestListUserWallets_FilterByStatus(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create wallets
	req1 := &models.CreateWalletRequest{
		UserID:          "user_filter",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet1, _ := service.CreateWallet(ctx, req1)

	req2 := &models.CreateWalletRequest{
		UserID:          "user_filter",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_002",
	}
	wallet2, _ := service.CreateWallet(ctx, req2)

	// Activate one wallet
	service.ActivateWallet(ctx, wallet1.ID)

	// List only active wallets
	activeStatus := models.WalletStatusActive
	wallets, err := service.ListUserWallets(ctx, "user_filter", &activeStatus)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(wallets) != 1 {
		t.Fatalf("expected 1 active wallet, got %d", len(wallets))
	}

	if wallets[0].ID != wallet1.ID {
		t.Error("expected wallet1 to be active")
	}

	// List only inactive wallets
	inactiveStatus := models.WalletStatusInactive
	wallets, err = service.ListUserWallets(ctx, "user_filter", &inactiveStatus)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(wallets) != 1 {
		t.Fatalf("expected 1 inactive wallet, got %d", len(wallets))
	}

	if wallets[0].ID != wallet2.ID {
		t.Error("expected wallet2 to be inactive")
	}
}

// ============================================================================
// Tests: Wallet Activation
// ============================================================================

func TestActivateWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create inactive wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_activate",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)

	// Activate wallet
	activated, err := service.ActivateWallet(ctx, wallet.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if activated.Status != models.WalletStatusActive {
		t.Errorf("expected status ACTIVE, got %s", activated.Status)
	}
}

func TestActivateWallet_Error_NotInactive(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create and activate wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_double_activate",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)

	// Try to activate again
	_, err := service.ActivateWallet(ctx, wallet.ID)

	if err == nil {
		t.Fatal("expected error when activating already active wallet")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Wallet Freeze/Unfreeze
// ============================================================================

func TestFreezeWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create and activate wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_freeze",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)

	// Freeze wallet
	frozen, err := service.FreezeWallet(ctx, wallet.ID, "suspicious activity")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if frozen.Status != models.WalletStatusFrozen {
		t.Errorf("expected status FROZEN, got %s", frozen.Status)
	}
}

func TestFreezeWallet_Error_NotActive(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create inactive wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_freeze_inactive",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)

	// Try to freeze inactive wallet
	_, err := service.FreezeWallet(ctx, wallet.ID, "test reason")

	if err == nil {
		t.Fatal("expected error when freezing non-active wallet")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestUnfreezeWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create, activate, and freeze wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_unfreeze",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)
	service.FreezeWallet(ctx, wallet.ID, "test freeze")

	// Unfreeze wallet
	unfrozen, err := service.UnfreezeWallet(ctx, wallet.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if unfrozen.Status != models.WalletStatusActive {
		t.Errorf("expected status ACTIVE after unfreeze, got %s", unfrozen.Status)
	}
}

func TestUnfreezeWallet_Error_NotFrozen(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create and activate wallet (not frozen)
	req := &models.CreateWalletRequest{
		UserID:          "user_unfreeze_active",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)

	// Try to unfreeze non-frozen wallet
	_, err := service.UnfreezeWallet(ctx, wallet.ID)

	if err == nil {
		t.Fatal("expected error when unfreezing non-frozen wallet")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Wallet Closure
// ============================================================================

func TestCloseWallet_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create active wallet with zero balance
	req := &models.CreateWalletRequest{
		UserID:          "user_close",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)

	// Close wallet
	closed, err := service.CloseWallet(ctx, wallet.ID, "user requested closure")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if closed.Status != models.WalletStatusClosed {
		t.Errorf("expected status CLOSED, got %s", closed.Status)
	}

	if closed.ClosedAt == nil {
		t.Error("expected ClosedAt to be set")
	}
}

func TestCloseWallet_Error_AlreadyClosed(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create and close wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_double_close",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)
	service.CloseWallet(ctx, wallet.ID, "first closure")

	// Try to close again
	_, err := service.CloseWallet(ctx, wallet.ID, "second closure")

	if err == nil {
		t.Fatal("expected error when closing already closed wallet")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

func TestCloseWallet_Error_NonZeroBalance(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create wallet with non-zero balance
	req := &models.CreateWalletRequest{
		UserID:          "user_close_balance",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)
	service.ActivateWallet(ctx, wallet.ID)

	// Manually set non-zero balance
	repo.wallets[wallet.ID].Balance = 1000

	// Try to close wallet with balance
	_, err := service.CloseWallet(ctx, wallet.ID, "closure attempt")

	if err == nil {
		t.Fatal("expected error when closing wallet with non-zero balance")
	}

	if err.Code != errors.ErrCodeBadRequest {
		t.Errorf("expected bad request error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Wallet Balance
// ============================================================================

func TestGetWalletBalance_Success(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// Create wallet
	req := &models.CreateWalletRequest{
		UserID:          "user_balance",
		Type:            models.WalletTypeSavings,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)

	// Get balance
	balance, err := service.GetWalletBalance(ctx, wallet.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance.WalletID != wallet.ID {
		t.Errorf("expected wallet ID %s, got %s", wallet.ID, balance.WalletID)
	}

	if balance.Balance != 0 {
		t.Errorf("expected zero balance, got %d", balance.Balance)
	}

	if balance.AvailableBalance != 0 {
		t.Errorf("expected zero available balance, got %d", balance.AvailableBalance)
	}
}

func TestGetWalletBalance_Error_NotFound(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	_, err := service.GetWalletBalance(ctx, "non_existent_wallet")

	if err == nil {
		t.Fatal("expected error for non-existent wallet")
	}

	if err.Code != errors.ErrCodeNotFound {
		t.Errorf("expected not found error, got %s", err.Code)
	}
}

// ============================================================================
// Tests: Wallet Status Transitions
// ============================================================================

func TestWalletStatusTransitions_FullLifecycle(t *testing.T) {
	repo := newMockWalletRepository()
	service := NewWalletService(repo)
	ctx := context.Background()

	// 1. Create wallet (INACTIVE)
	req := &models.CreateWalletRequest{
		UserID:          "user_lifecycle",
		Type:            models.WalletTypeCurrent,
		Currency:        "INR",
		LedgerAccountID: "acc_001",
	}
	wallet, _ := service.CreateWallet(ctx, req)

	if wallet.Status != models.WalletStatusInactive {
		t.Errorf("expected initial status INACTIVE, got %s", wallet.Status)
	}

	// 2. Activate wallet (INACTIVE -> ACTIVE)
	wallet, _ = service.ActivateWallet(ctx, wallet.ID)

	if wallet.Status != models.WalletStatusActive {
		t.Errorf("expected status ACTIVE after activation, got %s", wallet.Status)
	}

	// 3. Freeze wallet (ACTIVE -> FROZEN)
	wallet, _ = service.FreezeWallet(ctx, wallet.ID, "compliance check")

	if wallet.Status != models.WalletStatusFrozen {
		t.Errorf("expected status FROZEN after freeze, got %s", wallet.Status)
	}

	// 4. Unfreeze wallet (FROZEN -> ACTIVE)
	wallet, _ = service.UnfreezeWallet(ctx, wallet.ID)

	if wallet.Status != models.WalletStatusActive {
		t.Errorf("expected status ACTIVE after unfreeze, got %s", wallet.Status)
	}

	// 5. Close wallet (ACTIVE -> CLOSED)
	wallet, _ = service.CloseWallet(ctx, wallet.ID, "user closed account")

	if wallet.Status != models.WalletStatusClosed {
		t.Errorf("expected status CLOSED after closure, got %s", wallet.Status)
	}
}
