package service

import (
	"context"
	"log"

	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/repository"
	"github.com/vnykmshr/nivo/shared/errors"
)

// VirtualCardService handles business logic for virtual card operations.
type VirtualCardService struct {
	cardRepo   *repository.VirtualCardRepository
	walletRepo *repository.WalletRepository
}

// NewVirtualCardService creates a new virtual card service.
func NewVirtualCardService(cardRepo *repository.VirtualCardRepository, walletRepo *repository.WalletRepository) *VirtualCardService {
	return &VirtualCardService{
		cardRepo:   cardRepo,
		walletRepo: walletRepo,
	}
}

// CreateCard creates a new virtual card for a wallet.
func (s *VirtualCardService) CreateCard(ctx context.Context, walletID, userID string, req *models.CreateVirtualCardRequest) (*models.VirtualCard, *errors.Error) {
	// Verify wallet exists and belongs to user
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.UserID != userID {
		return nil, errors.Forbidden("wallet does not belong to user")
	}

	// Check wallet is active
	if wallet.Status != models.WalletStatusActive {
		return nil, errors.BadRequest("wallet is not active")
	}

	// Create the card
	card := &models.VirtualCard{
		WalletID:       walletID,
		UserID:         userID,
		CardHolderName: req.CardHolderName,
	}

	if createErr := s.cardRepo.Create(ctx, card); createErr != nil {
		return nil, createErr
	}

	log.Printf("[wallet] Virtual card created: card_id=%s, wallet_id=%s", card.ID, walletID)

	return card, nil
}

// GetCard retrieves a virtual card by ID.
func (s *VirtualCardService) GetCard(ctx context.Context, cardID, userID string) (*models.VirtualCard, *errors.Error) {
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	return card, nil
}

// ListCards retrieves all virtual cards for a wallet.
func (s *VirtualCardService) ListCards(ctx context.Context, walletID, userID string) ([]*models.VirtualCardResponse, *errors.Error) {
	// Verify wallet exists and belongs to user
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if wallet.UserID != userID {
		return nil, errors.Forbidden("wallet does not belong to user")
	}

	cards, listErr := s.cardRepo.ListByWallet(ctx, walletID)
	if listErr != nil {
		return nil, listErr
	}

	// Convert to responses (masked data)
	responses := make([]*models.VirtualCardResponse, len(cards))
	for i, card := range cards {
		responses[i] = card.ToResponse()
	}

	return responses, nil
}

// FreezeCard freezes a virtual card.
func (s *VirtualCardService) FreezeCard(ctx context.Context, cardID, userID, reason string) (*models.VirtualCard, *errors.Error) {
	// Verify ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	if card.Status != models.CardStatusActive {
		return nil, errors.BadRequest("can only freeze active cards")
	}

	if freezeErr := s.cardRepo.Freeze(ctx, cardID, reason); freezeErr != nil {
		return nil, freezeErr
	}

	log.Printf("[wallet] Virtual card frozen: card_id=%s, reason=%s", cardID, reason)

	// Return updated card
	return s.cardRepo.GetByID(ctx, cardID)
}

// UnfreezeCard unfreezes a virtual card.
func (s *VirtualCardService) UnfreezeCard(ctx context.Context, cardID, userID string) (*models.VirtualCard, *errors.Error) {
	// Verify ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	if card.Status != models.CardStatusFrozen {
		return nil, errors.BadRequest("can only unfreeze frozen cards")
	}

	if unfreezeErr := s.cardRepo.Unfreeze(ctx, cardID); unfreezeErr != nil {
		return nil, unfreezeErr
	}

	log.Printf("[wallet] Virtual card unfrozen: card_id=%s", cardID)

	// Return updated card
	return s.cardRepo.GetByID(ctx, cardID)
}

// CancelCard cancels a virtual card permanently.
func (s *VirtualCardService) CancelCard(ctx context.Context, cardID, userID, reason string) (*models.VirtualCard, *errors.Error) {
	// Verify ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	if card.Status == models.CardStatusCancelled {
		return nil, errors.BadRequest("card is already cancelled")
	}

	if cancelErr := s.cardRepo.Cancel(ctx, cardID, reason); cancelErr != nil {
		return nil, cancelErr
	}

	log.Printf("[wallet] Virtual card cancelled: card_id=%s, reason=%s", cardID, reason)

	// Return updated card
	return s.cardRepo.GetByID(ctx, cardID)
}

// UpdateCardLimits updates the spending limits for a virtual card.
func (s *VirtualCardService) UpdateCardLimits(ctx context.Context, cardID, userID string, req *models.UpdateCardLimitsRequest) (*models.VirtualCard, *errors.Error) {
	// Verify ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	// Only active and frozen cards can have limits updated
	if card.Status == models.CardStatusCancelled || card.Status == models.CardStatusExpired {
		return nil, errors.BadRequest("cannot update limits for cancelled or expired cards")
	}

	if updateErr := s.cardRepo.UpdateLimits(ctx, cardID, req.DailyLimit, req.MonthlyLimit, req.PerTransactionLimit); updateErr != nil {
		return nil, updateErr
	}

	log.Printf("[wallet] Virtual card limits updated: card_id=%s", cardID)

	// Return updated card
	return s.cardRepo.GetByID(ctx, cardID)
}

// RevealCardDetails reveals the full card details (requires additional security in production).
func (s *VirtualCardService) RevealCardDetails(ctx context.Context, cardID, userID string) (*models.RevealCardDetailsResponse, *errors.Error) {
	// Verify ownership
	card, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, errors.Forbidden("card does not belong to user")
	}

	if card.Status != models.CardStatusActive {
		return nil, errors.BadRequest("can only reveal details for active cards")
	}

	// In production, this would require:
	// 1. Additional authentication (OTP, biometric, etc.)
	// 2. Rate limiting
	// 3. Audit logging

	// For simulation, generate a mock CVV (since we store hashed)
	mockCVV := "***"

	log.Printf("[wallet] Card details revealed: card_id=%s, user_id=%s", cardID, userID)

	return &models.RevealCardDetailsResponse{
		CardNumber:  card.CardNumber,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		CVV:         mockCVV,
	}, nil
}
