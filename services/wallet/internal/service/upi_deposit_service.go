package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
)

// UPIDepositService handles business logic for UPI deposits.
type UPIDepositService struct {
	upiRepo        *repository.UPIDepositRepository
	walletRepo     WalletRepositoryInterface
	eventPublisher *events.Publisher
}

// NewUPIDepositService creates a new UPI deposit service.
func NewUPIDepositService(
	upiRepo *repository.UPIDepositRepository,
	walletRepo WalletRepositoryInterface,
	eventPublisher *events.Publisher,
) *UPIDepositService {
	return &UPIDepositService{
		upiRepo:        upiRepo,
		walletRepo:     walletRepo,
		eventPublisher: eventPublisher,
	}
}

// InitiateDeposit creates a new UPI deposit request.
func (s *UPIDepositService) InitiateDeposit(ctx context.Context, walletID, userID string, amount int64) (*models.UPIDepositResponse, *errors.Error) {
	// Validate wallet exists and belongs to user
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.UserID != userID {
		return nil, errors.Forbidden("wallet does not belong to user")
	}

	// Validate wallet is active
	if wallet.Status != models.WalletStatusActive {
		return nil, errors.BadRequest("wallet is not active")
	}

	// Validate amount (min ₹1, max ₹1,00,000)
	if amount < 100 { // 100 paise = ₹1
		return nil, errors.BadRequest("minimum deposit amount is ₹1")
	}
	if amount > 10000000 { // 1 crore paise = ₹1,00,000
		return nil, errors.BadRequest("maximum deposit amount is ₹1,00,000")
	}

	// Get or generate UPI VPA for wallet
	upiVPA, err := s.upiRepo.GetWalletUPIVPA(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Generate unique UPI reference
	upiReference := repository.GenerateUPIReference()

	// Create deposit record
	deposit := &models.UPIDeposit{
		WalletID:     walletID,
		UserID:       userID,
		Amount:       amount,
		UPIReference: upiReference,
		Status:       models.UPIDepositStatusPending,
	}

	if createErr := s.upiRepo.Create(ctx, deposit); createErr != nil {
		return nil, createErr
	}

	// Generate UPI payment string
	upiString := s.generateUPIString(upiVPA, amount, upiReference)

	// For simulation: auto-complete deposit after delay
	go s.simulateDepositCompletion(deposit.ID, walletID, amount)

	// Publish deposit.initiated event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.upi_deposit.initiated", walletID, map[string]interface{}{
			"deposit_id":    deposit.ID,
			"wallet_id":     walletID,
			"user_id":       userID,
			"amount":        amount,
			"upi_reference": upiReference,
		})
	}

	return &models.UPIDepositResponse{
		Deposit:   deposit,
		UPIString: upiString,
		QRCodeURL: fmt.Sprintf("/api/v1/qr?data=%s", upiString),
		ExpiresIn: "5 minutes",
		Message:   "UPI deposit initiated. Complete payment in your UPI app or scan the QR code.",
	}, nil
}

// GetDeposit retrieves a UPI deposit by ID.
func (s *UPIDepositService) GetDeposit(ctx context.Context, depositID, userID string) (*models.UPIDeposit, *errors.Error) {
	deposit, err := s.upiRepo.GetByID(ctx, depositID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if deposit.UserID != userID {
		return nil, errors.Forbidden("deposit does not belong to user")
	}

	return deposit, nil
}

// ListDeposits retrieves UPI deposits for a user.
func (s *UPIDepositService) ListDeposits(ctx context.Context, userID string, limit int) ([]*models.UPIDeposit, *errors.Error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.upiRepo.ListByUserID(ctx, userID, limit)
}

// GetWalletUPIDetails retrieves UPI details for a wallet.
func (s *UPIDepositService) GetWalletUPIDetails(ctx context.Context, walletID, userID string) (*models.WalletUPIDetails, *errors.Error) {
	// Validate wallet exists and belongs to user
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.UserID != userID {
		return nil, errors.Forbidden("wallet does not belong to user")
	}

	// Get or generate UPI VPA
	upiVPA, err := s.upiRepo.GetWalletUPIVPA(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return &models.WalletUPIDetails{
		WalletID:  walletID,
		UPIVPA:    upiVPA,
		QRCodeURL: fmt.Sprintf("/api/v1/qr?vpa=%s", upiVPA),
	}, nil
}

// CompleteDeposit marks a deposit as completed and credits the wallet.
// This would be called by a webhook handler in production.
func (s *UPIDepositService) CompleteDeposit(ctx context.Context, depositID string) *errors.Error {
	// Get deposit
	deposit, err := s.upiRepo.GetByID(ctx, depositID)
	if err != nil {
		return err
	}

	// Verify deposit is pending
	if deposit.Status != models.UPIDepositStatusPending {
		return errors.BadRequest("deposit is not pending")
	}

	// Mark deposit as completed
	if completeErr := s.upiRepo.Complete(ctx, depositID); completeErr != nil {
		return completeErr
	}

	// Credit wallet balance
	if updateErr := s.walletRepo.UpdateBalance(ctx, deposit.WalletID, deposit.Amount); updateErr != nil {
		// Attempt to revert deposit status on balance update failure
		_ = s.upiRepo.Fail(ctx, depositID, "failed to credit wallet")
		return updateErr
	}

	// Publish deposit.completed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishWalletEvent("wallet.upi_deposit.completed", deposit.WalletID, map[string]interface{}{
			"deposit_id":    deposit.ID,
			"wallet_id":     deposit.WalletID,
			"user_id":       deposit.UserID,
			"amount":        deposit.Amount,
			"upi_reference": deposit.UPIReference,
		})
	}

	return nil
}

// generateUPIString generates a UPI payment string.
func (s *UPIDepositService) generateUPIString(vpa string, amount int64, reference string) string {
	// Amount in rupees for UPI string
	amountRupees := float64(amount) / 100.0
	return fmt.Sprintf("upi://pay?pa=%s&pn=NivoMoney&am=%.2f&tr=%s&cu=INR&tn=Wallet%%20Deposit",
		vpa, amountRupees, reference)
}

// simulateDepositCompletion simulates UPI payment completion after a delay.
// In production, this would be handled by a UPI webhook callback.
func (s *UPIDepositService) simulateDepositCompletion(depositID, walletID string, amount int64) {
	// Wait for simulated payment (3 seconds)
	time.Sleep(3 * time.Second)

	ctx := context.Background()

	// Complete the deposit
	if err := s.CompleteDeposit(ctx, depositID); err != nil {
		log.Printf("[UPI] Failed to complete simulated deposit %s: %v", depositID, err)
		return
	}

	log.Printf("[UPI] Simulated deposit %s completed: ₹%.2f credited to wallet %s",
		depositID, float64(amount)/100.0, walletID)
}

// ExpirePendingDeposits expires any pending deposits that have passed their expiry time.
// This should be run periodically (e.g., every minute) by a background job.
func (s *UPIDepositService) ExpirePendingDeposits(ctx context.Context) (int64, *errors.Error) {
	return s.upiRepo.ExpirePending(ctx)
}
