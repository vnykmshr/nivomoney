package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
	"github.com/vnykmshr/nivo/shared/response"
)

// UPIDepositHandler handles HTTP requests for UPI deposit operations.
type UPIDepositHandler struct {
	upiService *service.UPIDepositService
}

// NewUPIDepositHandler creates a new UPI deposit handler.
func NewUPIDepositHandler(upiService *service.UPIDepositService) *UPIDepositHandler {
	return &UPIDepositHandler{
		upiService: upiService,
	}
}

// InitiateDeposit handles POST /api/v1/wallets/{id}/deposit/upi
func (h *UPIDepositHandler) InitiateDeposit(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	walletID := r.PathValue("id")
	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.InitiateUPIDepositRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Initiate deposit
	depositResponse, serviceErr := h.upiService.InitiateDeposit(r.Context(), walletID, userID, req.Amount)
	if serviceErr != nil {
		response.Error(w, serviceErr)
		return
	}

	response.JSON(w, http.StatusAccepted, depositResponse)
}

// GetWalletUPIDetails handles GET /api/v1/wallets/{id}/upi
func (h *UPIDepositHandler) GetWalletUPIDetails(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	walletID := r.PathValue("id")
	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	details, err := h.upiService.GetWalletUPIDetails(r.Context(), walletID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, details)
}

// GetDeposit handles GET /api/v1/deposits/upi/{id}
func (h *UPIDepositHandler) GetDeposit(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	depositID := r.PathValue("id")
	if depositID == "" {
		response.Error(w, errors.BadRequest("deposit ID is required"))
		return
	}

	deposit, err := h.upiService.GetDeposit(r.Context(), depositID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, deposit)
}

// ListDeposits handles GET /api/v1/deposits/upi
func (h *UPIDepositHandler) ListDeposits(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Parse limit from query params
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	deposits, err := h.upiService.ListDeposits(r.Context(), userID, limit)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, map[string]interface{}{
		"deposits": deposits,
		"count":    len(deposits),
	})
}
