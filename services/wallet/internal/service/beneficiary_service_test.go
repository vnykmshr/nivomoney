package service

import (
	"context"
	"testing"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// Mock implementations for testing

type mockBeneficiaryRepository struct {
	beneficiaries map[string]*models.Beneficiary
}

func newMockBeneficiaryRepository() *mockBeneficiaryRepository {
	return &mockBeneficiaryRepository{
		beneficiaries: make(map[string]*models.Beneficiary),
	}
}

func (m *mockBeneficiaryRepository) Create(ctx context.Context, beneficiary *models.Beneficiary) *errors.Error {
	// Check for duplicate beneficiary user
	for _, b := range m.beneficiaries {
		if b.OwnerUserID == beneficiary.OwnerUserID && b.BeneficiaryUserID == beneficiary.BeneficiaryUserID {
			return errors.Conflict("this user is already in your beneficiaries")
		}
		if b.OwnerUserID == beneficiary.OwnerUserID && b.Nickname == beneficiary.Nickname {
			return errors.Conflict("a beneficiary with this nickname already exists")
		}
	}

	beneficiary.ID = "ben-123"
	m.beneficiaries[beneficiary.ID] = beneficiary
	return nil
}

func (m *mockBeneficiaryRepository) GetByID(ctx context.Context, id, ownerUserID string) (*models.Beneficiary, *errors.Error) {
	b, ok := m.beneficiaries[id]
	if !ok || b.OwnerUserID != ownerUserID {
		return nil, errors.NotFoundWithID("beneficiary", id)
	}
	return b, nil
}

func (m *mockBeneficiaryRepository) ListByOwner(ctx context.Context, ownerUserID string) ([]*models.Beneficiary, *errors.Error) {
	result := make([]*models.Beneficiary, 0)
	for _, b := range m.beneficiaries {
		if b.OwnerUserID == ownerUserID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockBeneficiaryRepository) UpdateNickname(ctx context.Context, id, ownerUserID, nickname string) *errors.Error {
	b, ok := m.beneficiaries[id]
	if !ok || b.OwnerUserID != ownerUserID {
		return errors.NotFoundWithID("beneficiary", id)
	}
	b.Nickname = nickname
	return nil
}

func (m *mockBeneficiaryRepository) Delete(ctx context.Context, id, ownerUserID string) *errors.Error {
	b, ok := m.beneficiaries[id]
	if !ok || b.OwnerUserID != ownerUserID {
		return errors.NotFoundWithID("beneficiary", id)
	}
	delete(m.beneficiaries, id)
	return nil
}

func (m *mockBeneficiaryRepository) GetByBeneficiaryUser(ctx context.Context, ownerUserID, beneficiaryUserID string) (*models.Beneficiary, *errors.Error) {
	for _, b := range m.beneficiaries {
		if b.OwnerUserID == ownerUserID && b.BeneficiaryUserID == beneficiaryUserID {
			return b, nil
		}
	}
	return nil, errors.NotFound("beneficiary not found")
}

type mockUserClient struct {
	users map[string]*UserInfo
}

func newMockUserClient() *mockUserClient {
	return &mockUserClient{
		users: map[string]*UserInfo{
			"+919876543210": {
				ID:       "user-2",
				Phone:    "+919876543210",
				FullName: "John Doe",
				Email:    "john@example.com",
			},
		},
	}
}

func (m *mockUserClient) LookupUserByPhone(ctx context.Context, phone string) (*UserInfo, *errors.Error) {
	user, ok := m.users[phone]
	if !ok {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

type mockWalletRepoForBeneficiary struct {
	wallets map[string]*models.Wallet
}

func newMockWalletRepoForBeneficiary() *mockWalletRepoForBeneficiary {
	return &mockWalletRepoForBeneficiary{
		wallets: map[string]*models.Wallet{
			"wallet-2": {
				ID:       "wallet-2",
				UserID:   "user-2",
				Type:     models.WalletTypeDefault,
				Currency: "INR",
				Status:   models.WalletStatusActive,
			},
		},
	}
}

func (m *mockWalletRepoForBeneficiary) Create(ctx context.Context, wallet *models.Wallet) *errors.Error {
	return nil
}

func (m *mockWalletRepoForBeneficiary) GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error) {
	wallet, ok := m.wallets[id]
	if !ok {
		return nil, errors.NotFoundWithID("wallet", id)
	}
	return wallet, nil
}

func (m *mockWalletRepoForBeneficiary) ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error) {
	result := make([]*models.Wallet, 0)
	for _, w := range m.wallets {
		if w.UserID == userID {
			if status == nil || w.Status == *status {
				result = append(result, w)
			}
		}
	}
	return result, nil
}

func (m *mockWalletRepoForBeneficiary) UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error {
	return nil
}

func (m *mockWalletRepoForBeneficiary) Close(ctx context.Context, id, reason string) *errors.Error {
	return nil
}

func (m *mockWalletRepoForBeneficiary) GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error) {
	return nil, nil
}

// Test cases

func TestAddBeneficiary_Success(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	req := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "John",
	}

	beneficiary, err := service.AddBeneficiary(context.Background(), "user-1", req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if beneficiary.Nickname != "John" {
		t.Errorf("Expected nickname 'John', got '%s'", beneficiary.Nickname)
	}

	if beneficiary.BeneficiaryUserID != "user-2" {
		t.Errorf("Expected beneficiary user ID 'user-2', got '%s'", beneficiary.BeneficiaryUserID)
	}

	if beneficiary.BeneficiaryWalletID != "wallet-2" {
		t.Errorf("Expected wallet ID 'wallet-2', got '%s'", beneficiary.BeneficiaryWalletID)
	}
}

func TestAddBeneficiary_UserNotFound(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	req := &models.AddBeneficiaryRequest{
		Phone:    "+919999999999", // Non-existent phone
		Nickname: "Unknown",
	}

	_, err := service.AddBeneficiary(context.Background(), "user-1", req)

	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}

	if err.Code != "NOT_FOUND" {
		t.Errorf("Expected 'NOT_FOUND' error, got '%s'", err.Code)
	}
}

func TestAddBeneficiary_CannotAddSelf(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	req := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "Myself",
	}

	// Try to add self as beneficiary (user-2 trying to add user-2)
	_, err := service.AddBeneficiary(context.Background(), "user-2", req)

	if err == nil {
		t.Fatal("Expected error when adding self as beneficiary")
	}

	if err.Code != "BAD_REQUEST" {
		t.Errorf("Expected 'BAD_REQUEST' error, got '%s'", err.Code)
	}
}

func TestAddBeneficiary_DuplicateBeneficiary(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	req := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "John",
	}

	// Add beneficiary first time
	_, err := service.AddBeneficiary(context.Background(), "user-1", req)
	if err != nil {
		t.Fatalf("Expected no error on first add, got %v", err)
	}

	// Try to add same beneficiary again
	_, err = service.AddBeneficiary(context.Background(), "user-1", req)
	if err == nil {
		t.Fatal("Expected error for duplicate beneficiary")
	}

	if err.Code != "CONFLICT" {
		t.Errorf("Expected 'CONFLICT' error, got '%s'", err.Code)
	}
}

func TestListBeneficiaries_Success(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	// Add a beneficiary
	req := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "John",
	}
	_, _ = service.AddBeneficiary(context.Background(), "user-1", req)

	// List beneficiaries
	beneficiaries, err := service.ListBeneficiaries(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(beneficiaries) != 1 {
		t.Errorf("Expected 1 beneficiary, got %d", len(beneficiaries))
	}
}

func TestDeleteBeneficiary_Success(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	// Add a beneficiary
	req := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "John",
	}
	beneficiary, _ := service.AddBeneficiary(context.Background(), "user-1", req)

	// Delete beneficiary
	err := service.DeleteBeneficiary(context.Background(), "user-1", beneficiary.ID)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify deletion
	beneficiaries, _ := service.ListBeneficiaries(context.Background(), "user-1")
	if len(beneficiaries) != 0 {
		t.Errorf("Expected 0 beneficiaries after deletion, got %d", len(beneficiaries))
	}
}

func TestUpdateBeneficiary_Success(t *testing.T) {
	beneficiaryRepo := newMockBeneficiaryRepository()
	walletRepo := newMockWalletRepoForBeneficiary()
	userClient := newMockUserClient()

	service := NewBeneficiaryService(beneficiaryRepo, walletRepo, userClient, nil)

	// Add a beneficiary
	addReq := &models.AddBeneficiaryRequest{
		Phone:    "+919876543210",
		Nickname: "John",
	}
	beneficiary, _ := service.AddBeneficiary(context.Background(), "user-1", addReq)

	// Update nickname
	updateReq := &models.UpdateBeneficiaryRequest{
		Nickname: "Johnny",
	}
	updated, err := service.UpdateBeneficiary(context.Background(), "user-1", beneficiary.ID, updateReq)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if updated.Nickname != "Johnny" {
		t.Errorf("Expected nickname 'Johnny', got '%s'", updated.Nickname)
	}
}
