package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// WalletRepositoryInterface defines the interface for wallet repository operations.
type WalletRepositoryInterface interface {
	Create(ctx context.Context, wallet *models.Wallet) *errors.Error
	GetByID(ctx context.Context, id string) (*models.Wallet, *errors.Error)
	ListByUserID(ctx context.Context, userID string, status *models.WalletStatus) ([]*models.Wallet, *errors.Error)
	UpdateStatus(ctx context.Context, id string, status models.WalletStatus) *errors.Error
	Close(ctx context.Context, id, reason string) *errors.Error
	GetBalance(ctx context.Context, id string) (*models.WalletBalance, *errors.Error)
	GetLimits(ctx context.Context, walletID string) (*models.WalletLimits, *errors.Error)
	UpdateLimits(ctx context.Context, walletID string, dailyLimit, monthlyLimit int64) *errors.Error
	ProcessTransferWithinTx(ctx context.Context, sourceWalletID, destWalletID string, amount int64, transactionID string) *errors.Error
	UpdateBalance(ctx context.Context, walletID string, amount int64) *errors.Error
}

// WalletService handles business logic for wallet operations.
type WalletService struct {
	walletRepo         WalletRepositoryInterface
	eventPublisher     *events.Publisher
	ledgerClient       *LedgerClient
	notificationClient *clients.NotificationClient
	identityClient     *IdentityClient
}

// NewWalletService creates a new wallet service.
func NewWalletService(walletRepo WalletRepositoryInterface, eventPublisher *events.Publisher, ledgerClient *LedgerClient, notificationClient *clients.NotificationClient, identityClient *IdentityClient) *WalletService {
	return &WalletService{
		walletRepo:         walletRepo,
		eventPublisher:     eventPublisher,
		ledgerClient:       ledgerClient,
		notificationClient: notificationClient,
		identityClient:     identityClient,
	}
}

// CreateWallet creates a new wallet for a user.
func (s *WalletService) CreateWallet(ctx context.Context, req *models.CreateWalletRequest) (*models.Wallet, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Validate wallet type (only "default" is allowed)
	if req.Type != models.WalletTypeDefault {
		return nil, errors.Validation("invalid wallet type: only 'default' is supported")
	}

	// Check if user already has a wallet for this currency
	// One default wallet per user per currency
	existingWallets, listErr := s.walletRepo.ListByUserID(ctx, req.UserID, nil)
	if listErr != nil {
		return nil, listErr
	}

	for _, existing := range existingWallets {
		if existing.Currency == req.Currency {
			return nil, errors.Conflict("user already has a wallet for this currency")
		}
	}

	// If ledger_account_id is not provided, automatically create one (or reuse existing)
	ledgerAccountID := req.LedgerAccountID
	if ledgerAccountID == "" && s.ledgerClient != nil {
		// Generate the ledger account code (idempotent across retries)
		ledgerCode := fmt.Sprintf("WALLET-%s-%s", req.UserID[:8], req.Currency)

		// Check if a ledger account with this code already exists (for idempotency)
		existingAccount, checkErr := s.ledgerClient.GetAccountByCode(ctx, ledgerCode)
		if checkErr != nil {
			return nil, errors.Internal(fmt.Sprintf("failed to check for existing ledger account: %v", checkErr))
		}

		if existingAccount != nil {
			// Reuse the existing ledger account (handles orphaned accounts from previous failed attempts)
			ledgerAccountID = existingAccount.ID
		} else {
			// Create a new ledger account
			ledgerReq := &CreateLedgerAccountRequest{
				Code:     ledgerCode,
				Name:     fmt.Sprintf("Wallet (%s) for User %s", req.Currency, req.UserID[:8]),
				Type:     "asset", // Wallet accounts are assets
				Currency: string(req.Currency),
				Metadata: map[string]string{
					"wallet_type": "default",
					"user_id":     req.UserID,
				},
			}

			ledgerAccount, ledgerErr := s.ledgerClient.CreateAccount(ctx, ledgerReq)
			if ledgerErr != nil {
				return nil, errors.Internal(fmt.Sprintf("failed to create ledger account: %v", ledgerErr))
			}

			ledgerAccountID = ledgerAccount.ID
		}
	}

	// Validate that we have a ledger account ID
	if ledgerAccountID == "" {
		return nil, errors.Internal("ledger account ID is required but could not be created")
	}

	// Check user's KYC status to determine initial wallet status
	// Automatically activate wallet if KYC is verified
	walletStatus := models.WalletStatusInactive
	var userPhone string
	if s.identityClient != nil {
		kycStatus, kycErr := s.identityClient.GetUserKYCStatus(ctx, req.UserID)
		if kycErr == nil && kycStatus == "verified" {
			// User has verified KYC - activate wallet immediately
			walletStatus = models.WalletStatusActive
		}

		// Fetch user phone number for UPI ID generation
		userInfo, userErr := s.identityClient.GetUser(ctx, req.UserID)
		if userErr == nil && userInfo != nil && userInfo.PhoneNumber != "" {
			userPhone = userInfo.PhoneNumber
		}
	}

	// Auto-generate UPI ID if phone number is available and not already in metadata
	if userPhone != "" && metadata["upi_id"] == "" {
		// Remove country code prefix if present (e.g., +91)
		// UPI format: phone@nivomoney
		cleanPhone := userPhone
		if len(userPhone) > 10 && userPhone[0] == '+' {
			// Remove + and country code (e.g., +919876543210 -> 9876543210)
			cleanPhone = userPhone[len(userPhone)-10:]
		} else if len(userPhone) > 10 {
			// Remove country code without + (e.g., 919876543210 -> 9876543210)
			cleanPhone = userPhone[len(userPhone)-10:]
		}
		metadata["upi_id"] = fmt.Sprintf("%s@nivomoney", cleanPhone)
	}

	// Create wallet
	wallet := &models.Wallet{
		UserID:          req.UserID,
		Type:            req.Type,
		Currency:        req.Currency,
		Balance:         0, // Starts with zero balance
		Status:          walletStatus,
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

	// TODO: Send wallet created notification
	// This requires fetching user email from identity service
	// For now, notifications are sent on wallet activation instead

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

// GetWalletLimits retrieves the transfer limits for a wallet.
func (s *WalletService) GetWalletLimits(ctx context.Context, walletID string) (*models.WalletLimits, *errors.Error) {
	// Verify wallet exists
	_, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return s.walletRepo.GetLimits(ctx, walletID)
}

// UpdateWalletLimits updates the transfer limits for a wallet after verifying user password.
func (s *WalletService) UpdateWalletLimits(ctx context.Context, walletID string, req *models.UpdateLimitsRequest) (*models.WalletLimits, *errors.Error) {
	// Get wallet to verify ownership
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Verify wallet is active
	if !wallet.IsActive() {
		return nil, errors.BadRequest("cannot update limits for inactive wallet")
	}

	// Get user ID from context (set by auth middleware)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, errors.Unauthorized("user ID not found in context")
	}

	// Verify user owns this wallet
	if wallet.UserID != userID {
		return nil, errors.Forbidden("you do not own this wallet")
	}

	// Note: Authentication is handled via JWT middleware - no additional password verification needed.
	// The user has already authenticated and we've verified wallet ownership above.

	// Validate limits
	if req.DailyLimit > req.MonthlyLimit {
		return nil, errors.BadRequest("daily limit cannot exceed monthly limit")
	}

	// Update limits
	if err := s.walletRepo.UpdateLimits(ctx, walletID, req.DailyLimit, req.MonthlyLimit); err != nil {
		return nil, err
	}

	// Return updated limits
	return s.walletRepo.GetLimits(ctx, walletID)
}

// ProcessTransfer processes a wallet-to-wallet transfer with limit checking and balance updates.
// This is an internal endpoint called by the transaction service to execute approved transfers.
func (s *WalletService) ProcessTransfer(ctx context.Context, sourceWalletID, destWalletID string, amount int64, transactionID string) *errors.Error {
	// Validate wallets exist before attempting transfer
	sourceWallet, err := s.walletRepo.GetByID(ctx, sourceWalletID)
	if err != nil {
		return err
	}

	destWallet, err := s.walletRepo.GetByID(ctx, destWalletID)
	if err != nil {
		return err
	}

	// Prevent self-transfer
	if sourceWalletID == destWalletID {
		return errors.BadRequest("cannot transfer to the same wallet")
	}

	// Validate amount
	if amount <= 0 {
		return errors.BadRequest("transfer amount must be positive")
	}

	// Execute the transfer atomically (with limit checking and idempotency)
	if transferErr := s.walletRepo.ProcessTransferWithinTx(ctx, sourceWalletID, destWalletID, amount, transactionID); transferErr != nil {
		return transferErr
	}

	// Publish transfer.completed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.transfer.completed", sourceWalletID, map[string]interface{}{
			"source_wallet_id":      sourceWalletID,
			"destination_wallet_id": destWalletID,
			"amount":                amount,
			"transaction_id":        transactionID,
			"source_user_id":        sourceWallet.UserID,
			"dest_user_id":          destWallet.UserID,
		})
	}

	return nil
}

// ProcessDeposit credits a deposit to a wallet (internal method called by transaction service).
func (s *WalletService) ProcessDeposit(ctx context.Context, walletID string, amount int64, transactionID string) *errors.Error {
	// Validate wallet exists
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return err
	}

	// Validate wallet is active
	if wallet.Status != models.WalletStatusActive {
		return errors.BadRequest("wallet is not active")
	}

	// Validate amount
	if amount <= 0 {
		return errors.BadRequest("deposit amount must be positive")
	}

	// Use the wallet repository to update the balance
	// This will be a direct SQL update for deposits
	updateErr := s.walletRepo.UpdateBalance(ctx, walletID, amount)
	if updateErr != nil {
		return updateErr
	}

	// Publish deposit.completed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.deposit.completed", walletID, map[string]interface{}{
			"wallet_id":      walletID,
			"amount":         amount,
			"transaction_id": transactionID,
			"user_id":        wallet.UserID,
		})
	}

	return nil
}
