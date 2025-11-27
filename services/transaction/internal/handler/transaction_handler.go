package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/services/transaction/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// TransactionHandler handles HTTP requests for transaction operations.
type TransactionHandler struct {
	transactionService *service.TransactionService
}

// NewTransactionHandler creates a new transaction handler.
func NewTransactionHandler(transactionService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// CreateTransfer handles POST /api/v1/transactions/transfer
func (h *TransactionHandler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateTransferRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	transaction, createErr := h.transactionService.CreateTransfer(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, transaction)
}

// CreateDeposit handles POST /api/v1/transactions/deposit
func (h *TransactionHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateDepositRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	transaction, createErr := h.transactionService.CreateDeposit(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, transaction)
}

// InitiateUPIDeposit handles POST /api/v1/transactions/deposit/upi
func (h *TransactionHandler) InitiateUPIDeposit(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateUPIDepositRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	depositResponse, createErr := h.transactionService.InitiateUPIDeposit(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, depositResponse)
}

// CompleteUPIDeposit handles POST /api/v1/transactions/deposit/upi/complete
func (h *TransactionHandler) CompleteUPIDeposit(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CompleteUPIDepositRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	transaction, completeErr := h.transactionService.CompleteUPIDeposit(r.Context(), &req)
	if completeErr != nil {
		response.Error(w, completeErr)
		return
	}

	response.OK(w, transaction)
}

// CreateWithdrawal handles POST /api/v1/transactions/withdrawal
func (h *TransactionHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.CreateWithdrawalRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	transaction, createErr := h.transactionService.CreateWithdrawal(r.Context(), &req)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.Created(w, transaction)
}

// GetTransaction handles GET /api/v1/transactions/:id
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("id")

	if transactionID == "" {
		response.Error(w, errors.BadRequest("transaction ID is required"))
		return
	}

	transaction, err := h.transactionService.GetTransaction(r.Context(), transactionID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, transaction)
}

// ListWalletTransactions handles GET /api/v1/wallets/:walletId/transactions
func (h *TransactionHandler) ListWalletTransactions(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("walletId")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	// Parse query parameters for filtering
	filter := &models.TransactionFilter{}

	// Status filter
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status := models.TransactionStatus(statusParam)
		filter.Status = &status
	}

	// Type filter
	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		txType := models.TransactionType(typeParam)
		filter.Type = &txType
	}

	// Pagination
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			filter.Limit = limit
		}
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	transactions, err := h.transactionService.ListWalletTransactions(r.Context(), walletID, filter)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, transactions)
}

// ReverseTransaction handles POST /api/v1/transactions/:id/reverse
func (h *TransactionHandler) ReverseTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("id")

	if transactionID == "" {
		response.Error(w, errors.BadRequest("transaction ID is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Parse and validate request
	req, parseErr := model.ParseInto[models.ReverseTransactionRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	reversalTx, reverseErr := h.transactionService.ReverseTransaction(r.Context(), transactionID, req.Reason)
	if reverseErr != nil {
		response.Error(w, reverseErr)
		return
	}

	response.Created(w, reversalTx)
}
