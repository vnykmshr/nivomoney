package handler

import (
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
	"github.com/vnykmshr/nivo/shared/response"
)

// VirtualCardHandler handles HTTP requests for virtual card operations.
type VirtualCardHandler struct {
	cardService *service.VirtualCardService
}

// NewVirtualCardHandler creates a new virtual card handler.
func NewVirtualCardHandler(cardService *service.VirtualCardService) *VirtualCardHandler {
	return &VirtualCardHandler{
		cardService: cardService,
	}
}

// CreateCard handles POST /api/v1/wallets/:walletId/cards
func (h *VirtualCardHandler) CreateCard(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("walletId")
	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateVirtualCardRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	card, createErr := h.cardService.CreateCard(r.Context(), walletID, userID, &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	// Return full card details on creation (only time CVV is shown)
	response.Created(w, map[string]interface{}{
		"card": card.ToResponse(),
		"details": &models.RevealCardDetailsResponse{
			CardNumber:  card.CardNumber,
			ExpiryMonth: card.ExpiryMonth,
			ExpiryYear:  card.ExpiryYear,
			CVV:         card.CVV,
		},
		"message": "Save your card details securely. CVV will not be shown again.",
	})
}

// ListCards handles GET /api/v1/wallets/:walletId/cards
func (h *VirtualCardHandler) ListCards(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("walletId")
	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	cards, err := h.cardService.ListCards(r.Context(), walletID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, cards)
}

// GetCard handles GET /api/v1/cards/:id
func (h *VirtualCardHandler) GetCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	card, err := h.cardService.GetCard(r.Context(), cardID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, card.ToResponse())
}

// FreezeCard handles POST /api/v1/cards/:id/freeze
func (h *VirtualCardHandler) FreezeCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.FreezeCardRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	card, freezeErr := h.cardService.FreezeCard(r.Context(), cardID, userID, req.Reason)
	if freezeErr != nil {
		response.Error(w, freezeErr)
		return
	}

	response.OK(w, card.ToResponse())
}

// UnfreezeCard handles POST /api/v1/cards/:id/unfreeze
func (h *VirtualCardHandler) UnfreezeCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	card, unfreezeErr := h.cardService.UnfreezeCard(r.Context(), cardID, userID)
	if unfreezeErr != nil {
		response.Error(w, unfreezeErr)
		return
	}

	response.OK(w, card.ToResponse())
}

// CancelCard handles DELETE /api/v1/cards/:id
func (h *VirtualCardHandler) CancelCard(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CancelCardRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	card, cancelErr := h.cardService.CancelCard(r.Context(), cardID, userID, req.Reason)
	if cancelErr != nil {
		response.Error(w, cancelErr)
		return
	}

	response.OK(w, card.ToResponse())
}

// UpdateCardLimits handles PATCH /api/v1/cards/:id/limits
func (h *VirtualCardHandler) UpdateCardLimits(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.UpdateCardLimitsRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	card, updateErr := h.cardService.UpdateCardLimits(r.Context(), cardID, userID, &req)
	if updateErr != nil {
		response.Error(w, updateErr)
		return
	}

	response.OK(w, card.ToResponse())
}

// RevealCardDetails handles GET /api/v1/cards/:id/reveal
func (h *VirtualCardHandler) RevealCardDetails(w http.ResponseWriter, r *http.Request) {
	cardID := r.PathValue("id")
	if cardID == "" {
		response.Error(w, errors.BadRequest("card ID is required"))
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	details, err := h.cardService.RevealCardDetails(r.Context(), cardID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, details)
}
