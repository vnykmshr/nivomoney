package handler

import (
	"io"
	"net/http"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/wallet/internal/models"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// WalletHandler handles HTTP requests for wallet operations.
type WalletHandler struct {
	walletService *service.WalletService
}

// NewWalletHandler creates a new wallet handler.
func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// CreateWallet handles POST /api/v1/wallets
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request (gopantic v1.2.0+ supports json.RawMessage)
	req, parseErr := model.ParseInto[models.CreateWalletRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	wallet, createErr := h.walletService.CreateWallet(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, wallet)
}

// GetWallet handles GET /api/v1/wallets/:id
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("id")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	wallet, err := h.walletService.GetWallet(r.Context(), walletID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, wallet)
}

// ListUserWallets handles GET /api/v1/users/:userId/wallets
func (h *WalletHandler) ListUserWallets(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")

	if userID == "" {
		response.Error(w, errors.BadRequest("user ID is required"))
		return
	}

	// Optional status filter from query params
	var status *models.WalletStatus
	statusParam := r.URL.Query().Get("status")
	if statusParam != "" {
		s := models.WalletStatus(statusParam)
		status = &s
	}

	wallets, err := h.walletService.ListUserWallets(r.Context(), userID, status)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, wallets)
}

// ActivateWallet handles POST /api/v1/wallets/:id/activate
func (h *WalletHandler) ActivateWallet(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("id")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	wallet, err := h.walletService.ActivateWallet(r.Context(), walletID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, wallet)
}

// FreezeWallet handles POST /api/v1/wallets/:id/freeze
func (h *WalletHandler) FreezeWallet(w http.ResponseWriter, r *http.Request) {
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
	req, parseErr := model.ParseInto[models.FreezeWalletRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	wallet, freezeErr := h.walletService.FreezeWallet(r.Context(), walletID, req.Reason)
	if freezeErr != nil {
		response.Error(w, freezeErr)
		return
	}

	response.OK(w, wallet)
}

// UnfreezeWallet handles POST /api/v1/wallets/:id/unfreeze
func (h *WalletHandler) UnfreezeWallet(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("id")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	wallet, err := h.walletService.UnfreezeWallet(r.Context(), walletID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, wallet)
}

// CloseWallet handles POST /api/v1/wallets/:id/close
func (h *WalletHandler) CloseWallet(w http.ResponseWriter, r *http.Request) {
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
	req, parseErr := model.ParseInto[models.CloseWalletRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	wallet, closeErr := h.walletService.CloseWallet(r.Context(), walletID, req.Reason)
	if closeErr != nil {
		response.Error(w, closeErr)
		return
	}

	response.OK(w, wallet)
}

// GetWalletBalance handles GET /api/v1/wallets/:id/balance
func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("id")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	balance, err := h.walletService.GetWalletBalance(r.Context(), walletID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, balance)
}

// GetWalletLimits handles GET /api/v1/wallets/:id/limits
func (h *WalletHandler) GetWalletLimits(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("id")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	limits, err := h.walletService.GetWalletLimits(r.Context(), walletID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, limits)
}

// UpdateWalletLimits handles PUT /api/v1/wallets/:id/limits
func (h *WalletHandler) UpdateWalletLimits(w http.ResponseWriter, r *http.Request) {
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
	req, parseErr := model.ParseInto[models.UpdateLimitsRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	limits, updateErr := h.walletService.UpdateWalletLimits(r.Context(), walletID, &req)
	if updateErr != nil {
		response.Error(w, updateErr)
		return
	}

	response.OK(w, limits)
}

// ProcessTransfer handles POST /internal/v1/wallets/transfer (internal endpoint)
// This endpoint is called by the transaction service to execute wallet-to-wallet transfers.
func (h *WalletHandler) ProcessTransfer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.ProcessTransferRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	// Process the transfer
	transferErr := h.walletService.ProcessTransfer(
		r.Context(),
		req.SourceWalletID,
		req.DestinationWalletID,
		req.Amount,
		req.TransactionID,
	)
	if transferErr != nil {
		response.Error(w, transferErr)
		return
	}

	response.OK(w, map[string]interface{}{
		"success":           true,
		"source_wallet_id":  req.SourceWalletID,
		"dest_wallet_id":    req.DestinationWalletID,
		"amount":            req.Amount,
		"transaction_id":    req.TransactionID,
	})
}
