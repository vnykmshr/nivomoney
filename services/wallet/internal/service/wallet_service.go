package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
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
	walletRepo     WalletRepositoryInterface
	eventPublisher *events.Publisher
	ledgerClient   *LedgerClient
}

// NewWalletService creates a new wallet service.
func NewWalletService(walletRepo WalletRepositoryInterface, eventPublisher *events.Publisher, ledgerClient *LedgerClient) *WalletService {
	return &WalletService{
		walletRepo:     walletRepo,
		eventPublisher: eventPublisher,
		ledgerClient:   ledgerClient,
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

	// If ledger_account_id is not provided, automatically create one
	ledgerAccountID := req.LedgerAccountID
	if ledgerAccountID == "" && s.ledgerClient != nil {
		// Create ledger account for this wallet
		ledgerReq := &CreateLedgerAccountRequest{
			Code:     fmt.Sprintf("WALLET-%s-%s", req.UserID[:8], req.Type),
			Name:     fmt.Sprintf("%s Wallet for User %s", req.Type, req.UserID[:8]),
			Type:     "asset", // Wallet accounts are assets
			Currency: string(req.Currency),
			Metadata: map[string]string{
				"wallet_type": string(req.Type),
				"user_id":     req.UserID,
			},
		}

		ledgerAccount, ledgerErr := s.ledgerClient.CreateAccount(ctx, ledgerReq)
		if ledgerErr != nil {
			return nil, errors.Internal(fmt.Sprintf("failed to create ledger account: %v", ledgerErr))
		}

		ledgerAccountID = ledgerAccount.ID
	}

	// Validate that we have a ledger account ID
	if ledgerAccountID == "" {
		return nil, errors.Internal("ledger account ID is required but could not be created")
	}

	// Create wallet (starts as inactive, needs KYC verification to become active)
	wallet := &models.Wallet{
		UserID:          req.UserID,
		Type:            req.Type,
		Currency:        req.Currency,
		Balance:         0, // Starts with zero balance
		Status:          models.WalletStatusInactive,
		LedgerAccountID: ledgerAccountID,
		Metadata:        metadata,
	}

	if createErr := s.walletRepo.Create(ctx, wallet); createErr != nil {
		return nil, createErr
	}

	// Publish wallet.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.created", wallet.ID, map[string]interface{}{
			"user_id":           wallet.UserID,
			"type":              string(wallet.Type),
			"currency":          string(wallet.Currency),
			"status":            string(wallet.Status),
			"balance":           wallet.Balance,
			"available_balance": wallet.AvailableBalance,
			"ledger_account_id": wallet.LedgerAccountID,
		})
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

	// Get updated wallet
	updatedWallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Publish wallet.status_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.status_changed", updatedWallet.ID, map[string]interface{}{
			"user_id":           updatedWallet.UserID,
			"currency":          string(updatedWallet.Currency),
			"old_status":        string(models.WalletStatusInactive),
			"new_status":        string(updatedWallet.Status),
			"balance":           updatedWallet.Balance,
			"available_balance": updatedWallet.AvailableBalance,
			"action":            "activated",
		})
	}

	return updatedWallet, nil
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

	// Get updated wallet
	updatedWallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Publish wallet.status_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.status_changed", updatedWallet.ID, map[string]interface{}{
			"user_id":           updatedWallet.UserID,
			"currency":          string(updatedWallet.Currency),
			"old_status":        string(models.WalletStatusActive),
			"new_status":        string(updatedWallet.Status),
			"balance":           updatedWallet.Balance,
			"available_balance": updatedWallet.AvailableBalance,
			"action":            "frozen",
			"reason":            reason,
		})
	}

	return updatedWallet, nil
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

	// Get updated wallet
	updatedWallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Publish wallet.status_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.status_changed", updatedWallet.ID, map[string]interface{}{
			"user_id":           updatedWallet.UserID,
			"currency":          string(updatedWallet.Currency),
			"old_status":        string(models.WalletStatusFrozen),
			"new_status":        string(updatedWallet.Status),
			"balance":           updatedWallet.Balance,
			"available_balance": updatedWallet.AvailableBalance,
			"action":            "unfrozen",
		})
	}

	return updatedWallet, nil
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

	// Store old status before closing
	oldStatus := wallet.Status

	// Close wallet
	if closeErr := s.walletRepo.Close(ctx, walletID, reason); closeErr != nil {
		return nil, closeErr
	}

	// Get updated wallet
	updatedWallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Publish wallet.status_changed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.status_changed", updatedWallet.ID, map[string]interface{}{
			"user_id":           updatedWallet.UserID,
			"currency":          string(updatedWallet.Currency),
			"old_status":        string(oldStatus),
			"new_status":        string(updatedWallet.Status),
			"balance":           updatedWallet.Balance,
			"available_balance": updatedWallet.AvailableBalance,
			"action":            "closed",
			"reason":            reason,
		})
	}

	return updatedWallet, nil
}

// GetWalletBalance retrieves the balance of a wallet.
func (s *WalletService) GetWalletBalance(ctx context.Context, walletID string) (*models.WalletBalance, *errors.Error) {
	return s.walletRepo.GetBalance(ctx, walletID)
}
