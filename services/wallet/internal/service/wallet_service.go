package service

import (
	"context"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// WalletRepositoryInterface defines the interface for wallet repository operations.
type WalletRepositoryInterface interface {
	Create(ctx context.Context, wallet *models.Wallet) *errors.Error
	GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error)
	ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error)
	UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error
	Close(ctx context.Context, id, reason string) *errors.Error
	GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error)
}

// WalletService handles business logic for wallet operations.
type WalletService struct {
	walletRepo WalletRepositoryInterface
}

// NewWalletService creates a new wallet service.
func NewWalletService(walletRepo WalletRepositoryInterface) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

// CreateWallet creates a new wallet for a user.
func (s *WalletService) CreateWallet(ctx context.Context, req *models.CreateWalletRequest) (*models.Wallet, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Validate wallet type
	if req.Type != models.WalletTypeSavings && req.Type != models.WalletTypeCurrent && req.Type != models.WalletTypeFixed {
		return nil, errors.Validation("invalid wallet type")
	}

	// Create wallet (starts as inactive, needs KYC verification to become active)
	wallet := &models.Wallet{
		UserID:          req.UserID,
		Type:            req.Type,
		Currency:        req.Currency,
		Balance:         0, // Starts with zero balance
		Status:          models.WalletStatusInactive,
		LedgerAccountID: req.LedgerAccountID,
		Metadata:        metadata,
	}

	if createErr := s.walletRepo.Create(ctx, wallet); createErr != nil {
		return nil, createErr
	}

	return wallet, nil
}

// GetWallet retrieves a wallet by ID.
func (s *WalletService) GetWallet(ctx context.Context, walletID string) (*models.Wallet, *errors.Error) {
	return s.walletRepo.GetByID(ctx, walletID)
}

// ListUserWallets retrieves all wallets for a user.
func (s *WalletService) ListUserWallets(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error) {
	return s.walletRepo.ListByUserID(ctx, userID, status)
}

// ActivateWallet activates a wallet (after KYC verification).
func (s *WalletService) ActivateWallet(ctx context.Context, walletID string) (*models.Wallet, *errors.Error) {
	// Get wallet to verify it exists and is inactive
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if wallet.Status != models.WalletStatusInactive {
		return nil, errors.BadRequest("only inactive wallets can be activated")
	}

	// Update status
	if updateErr := s.walletRepo.UpdateStatus(ctx, walletID, models.WalletStatusActive); updateErr != nil {
		return nil, updateErr
	}

	// Return updated wallet
	return s.walletRepo.GetByID(ctx, walletID)
}

// FreezeWallet freezes a wallet (for compliance or security reasons).
func (s *WalletService) FreezeWallet(ctx context.Context, walletID, reason string) (*models.Wallet, *errors.Error) {
	// Get wallet to verify it exists
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if wallet.Status != models.WalletStatusActive {
		return nil, errors.BadRequest("only active wallets can be frozen")
	}

	// Update status
	if updateErr := s.walletRepo.UpdateStatus(ctx, walletID, models.WalletStatusFrozen); updateErr != nil {
		return nil, updateErr
	}

	// TODO: Log freeze action with reason

	// Return updated wallet
	return s.walletRepo.GetByID(ctx, walletID)
}

// UnfreezeWallet unfreezes a wallet.
func (s *WalletService) UnfreezeWallet(ctx context.Context, walletID string) (*models.Wallet, *errors.Error) {
	// Get wallet to verify it exists
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if wallet.Status != models.WalletStatusFrozen {
		return nil, errors.BadRequest("only frozen wallets can be unfrozen")
	}

	// Update status
	if updateErr := s.walletRepo.UpdateStatus(ctx, walletID, models.WalletStatusActive); updateErr != nil {
		return nil, updateErr
	}

	// Return updated wallet
	return s.walletRepo.GetByID(ctx, walletID)
}

// CloseWallet closes a wallet permanently.
func (s *WalletService) CloseWallet(ctx context.Context, walletID, reason string) (*models.Wallet, *errors.Error) {
	// Get wallet to verify it exists
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Validate wallet can be closed
	if wallet.Status == models.WalletStatusClosed {
		return nil, errors.BadRequest("wallet is already closed")
	}

	// Validate balance is zero
	if wallet.Balance > 0 {
		return nil, errors.BadRequest("cannot close wallet with non-zero balance")
	}

	// Close wallet
	if closeErr := s.walletRepo.Close(ctx, walletID, reason); closeErr != nil {
		return nil, closeErr
	}

	// Return updated wallet
	return s.walletRepo.GetByID(ctx, walletID)
}

// GetWalletBalance retrieves the balance of a wallet.
func (s *WalletService) GetWalletBalance(ctx context.Context, walletID string) (*models.WalletBalance, *errors.Error) {
	return s.walletRepo.GetBalance(ctx, walletID)
}
