package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
)

// BeneficiaryRepositoryInterface defines the interface for beneficiary repository operations.
type BeneficiaryRepositoryInterface interface {
	Create(ctx context.Context, beneficiary *models.Beneficiary) *errors.Error
	GetByID(ctx context.Context, id, ownerUserID string) (*models.Beneficiary, *errors.Error)
	ListByOwner(ctx context.Context, ownerUserID string) ([]*models.Beneficiary, *errors.Error)
	UpdateNickname(ctx context.Context, id, ownerUserID, nickname string) *errors.Error
	Delete(ctx context.Context, id, ownerUserID string) *errors.Error
	GetByBeneficiaryUser(ctx context.Context, ownerUserID, beneficiaryUserID string) (*models.Beneficiary, *errors.Error)
}

// UserLookupClient defines the interface for looking up users from the identity service.
type UserLookupClient interface {
	LookupUserByPhone(ctx context.Context, phone string) (*UserInfo, *errors.Error)
}

// UserInfo represents basic user information from identity service.
type UserInfo struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`
	PhoneNumber string `json:"phone_number"` // Alias for phone
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
}

// BeneficiaryService handles business logic for beneficiary operations.
type BeneficiaryService struct {
	beneficiaryRepo BeneficiaryRepositoryInterface
	walletRepo      WalletRepositoryInterface
	userClient      UserLookupClient
	eventPublisher  *events.Publisher
}

// NewBeneficiaryService creates a new beneficiary service.
func NewBeneficiaryService(
	beneficiaryRepo BeneficiaryRepositoryInterface,
	walletRepo WalletRepositoryInterface,
	userClient UserLookupClient,
	eventPublisher *events.Publisher,
) *BeneficiaryService {
	return &BeneficiaryService{
		beneficiaryRepo: beneficiaryRepo,
		walletRepo:      walletRepo,
		userClient:      userClient,
		eventPublisher:  eventPublisher,
	}
}

// AddBeneficiary adds a new beneficiary for a user.
func (s *BeneficiaryService) AddBeneficiary(ctx context.Context, ownerUserID string, req *models.AddBeneficiaryRequest) (*models.Beneficiary, *errors.Error) {
	// Lookup user by phone using identity service
	userInfo, err := s.userClient.LookupUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}

	// Prevent adding self as beneficiary
	if userInfo.ID == ownerUserID {
		return nil, errors.BadRequest("cannot add yourself as a beneficiary")
	}

	// Check if beneficiary already exists
	existing, _ := s.beneficiaryRepo.GetByBeneficiaryUser(ctx, ownerUserID, userInfo.ID)
	if existing != nil {
		return nil, errors.Conflict("this user is already in your beneficiaries")
	}

	// Get beneficiary's default wallet (INR wallet)
	wallets, listErr := s.walletRepo.ListByUserID(ctx, userInfo.ID, nil)
	if listErr != nil {
		return nil, listErr
	}

	if len(wallets) == 0 {
		return nil, errors.BadRequest("beneficiary does not have a wallet")
	}

	// Find default INR wallet
	var defaultWallet *models.Wallet
	for _, wallet := range wallets {
		if wallet.Type == models.WalletTypeDefault && wallet.Currency == "INR" {
			defaultWallet = wallet
			break
		}
	}

	if defaultWallet == nil {
		return nil, errors.BadRequest("beneficiary does not have a default INR wallet")
	}

	// Validate wallet is active or inactive (not frozen/closed)
	if defaultWallet.Status == models.WalletStatusClosed || defaultWallet.Status == models.WalletStatusFrozen {
		return nil, errors.BadRequest("beneficiary's wallet is not available for transfers")
	}

	// Create beneficiary
	beneficiary := &models.Beneficiary{
		OwnerUserID:         ownerUserID,
		BeneficiaryUserID:   userInfo.ID,
		BeneficiaryWalletID: defaultWallet.ID,
		Nickname:            req.Nickname,
		BeneficiaryPhone:    userInfo.Phone,
		Metadata:            make(map[string]string),
	}

	if createErr := s.beneficiaryRepo.Create(ctx, beneficiary); createErr != nil {
		return nil, createErr
	}

	// Publish beneficiary.added event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("beneficiary.added", beneficiary.ID, map[string]interface{}{
			"owner_user_id":         beneficiary.OwnerUserID,
			"beneficiary_user_id":   beneficiary.BeneficiaryUserID,
			"beneficiary_wallet_id": beneficiary.BeneficiaryWalletID,
			"nickname":              beneficiary.Nickname,
		})
	}

	return beneficiary, nil
}

// GetBeneficiary retrieves a beneficiary by ID.
func (s *BeneficiaryService) GetBeneficiary(ctx context.Context, ownerUserID, beneficiaryID string) (*models.Beneficiary, *errors.Error) {
	return s.beneficiaryRepo.GetByID(ctx, beneficiaryID, ownerUserID)
}

// ListBeneficiaries retrieves all beneficiaries for a user.
func (s *BeneficiaryService) ListBeneficiaries(ctx context.Context, ownerUserID string) ([]*models.Beneficiary, *errors.Error) {
	return s.beneficiaryRepo.ListByOwner(ctx, ownerUserID)
}

// UpdateBeneficiary updates a beneficiary's nickname.
func (s *BeneficiaryService) UpdateBeneficiary(ctx context.Context, ownerUserID, beneficiaryID string, req *models.UpdateBeneficiaryRequest) (*models.Beneficiary, *errors.Error) {
	// Verify beneficiary exists and belongs to owner
	existing, err := s.beneficiaryRepo.GetByID(ctx, beneficiaryID, ownerUserID)
	if err != nil {
		return nil, err
	}

	// Update nickname
	if updateErr := s.beneficiaryRepo.UpdateNickname(ctx, beneficiaryID, ownerUserID, req.Nickname); updateErr != nil {
		return nil, updateErr
	}

	// Get updated beneficiary
	updated, err := s.beneficiaryRepo.GetByID(ctx, beneficiaryID, ownerUserID)
	if err != nil {
		return nil, err
	}

	// Publish beneficiary.updated event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("beneficiary.updated", beneficiaryID, map[string]interface{}{
			"owner_user_id": ownerUserID,
			"old_nickname":  existing.Nickname,
			"new_nickname":  updated.Nickname,
		})
	}

	return updated, nil
}

// DeleteBeneficiary removes a beneficiary.
func (s *BeneficiaryService) DeleteBeneficiary(ctx context.Context, ownerUserID, beneficiaryID string) *errors.Error {
	// Verify beneficiary exists and belongs to owner
	beneficiary, err := s.beneficiaryRepo.GetByID(ctx, beneficiaryID, ownerUserID)
	if err != nil {
		return err
	}

	// Delete beneficiary
	if deleteErr := s.beneficiaryRepo.Delete(ctx, beneficiaryID, ownerUserID); deleteErr != nil {
		return deleteErr
	}

	// Publish beneficiary.deleted event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("beneficiary.deleted", beneficiaryID, map[string]interface{}{
			"owner_user_id":       ownerUserID,
			"beneficiary_user_id": beneficiary.BeneficiaryUserID,
			"nickname":            beneficiary.Nickname,
		})
	}

	return nil
}

// ValidateBeneficiaryForTransfer validates that a beneficiary is eligible for receiving transfers.
func (s *BeneficiaryService) ValidateBeneficiaryForTransfer(ctx context.Context, ownerUserID, beneficiaryID string) (*models.Beneficiary, *errors.Error) {
	// Get beneficiary
	beneficiary, err := s.beneficiaryRepo.GetByID(ctx, beneficiaryID, ownerUserID)
	if err != nil {
		return nil, err
	}

	// Verify wallet still exists and is active
	wallet, walletErr := s.walletRepo.GetByID(ctx, beneficiary.BeneficiaryWalletID)
	if walletErr != nil {
		return nil, errors.BadRequest(fmt.Sprintf("beneficiary wallet not found: %v", walletErr))
	}

	if wallet.Status != models.WalletStatusActive {
		return nil, errors.BadRequest("beneficiary's wallet is not active for transfers")
	}

	return beneficiary, nil
}
