package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"

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

	// Status filter - validate against known values
	if statusParam := r.URL.Query().Get("status"); statusParam != "" {
		status := models.TransactionStatus(statusParam)
		// Validate status is a known value
		validStatuses := []models.TransactionStatus{
			models.TransactionStatusPending,
			models.TransactionStatusProcessing,
			models.TransactionStatusCompleted,
			models.TransactionStatusFailed,
			models.TransactionStatusReversed,
			models.TransactionStatusCancelled,
		}
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			response.Error(w, errors.BadRequest("invalid status value"))
			return
		}
		filter.Status = &status
	}

	// Type filter - validate against known values
	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		txType := models.TransactionType(typeParam)
		// Validate type is a known value
		validTypes := []models.TransactionType{
			models.TransactionTypeTransfer,
			models.TransactionTypeDeposit,
			models.TransactionTypeWithdrawal,
			models.TransactionTypeReversal,
			models.TransactionTypeFee,
			models.TransactionTypeRefund,
		}
		isValid := false
		for _, validType := range validTypes {
			if txType == validType {
				isValid = true
				break
			}
		}
		if !isValid {
			response.Error(w, errors.BadRequest("invalid type value"))
			return
		}
		filter.Type = &txType
	}

	// Search filter (searches description and reference)
	// Limit search query length to prevent performance issues
	if searchParam := r.URL.Query().Get("search"); searchParam != "" {
		if len(searchParam) > 200 {
			response.Error(w, errors.BadRequest("search query too long (max 200 characters)"))
			return
		}
		filter.Search = &searchParam
	}

	// Amount range filters (in smallest unit - paise)
	if minAmountParam := r.URL.Query().Get("min_amount"); minAmountParam != "" {
		minAmount, err := strconv.ParseInt(minAmountParam, 10, 64)
		if err != nil {
			response.Error(w, errors.BadRequest("invalid min_amount value"))
			return
		}
		if minAmount < 0 {
			response.Error(w, errors.BadRequest("min_amount cannot be negative"))
			return
		}
		filter.MinAmount = &minAmount
	}

	if maxAmountParam := r.URL.Query().Get("max_amount"); maxAmountParam != "" {
		maxAmount, err := strconv.ParseInt(maxAmountParam, 10, 64)
		if err != nil {
			response.Error(w, errors.BadRequest("invalid max_amount value"))
			return
		}
		if maxAmount < 0 {
			response.Error(w, errors.BadRequest("max_amount cannot be negative"))
			return
		}
		filter.MaxAmount = &maxAmount
	}

	// Validate amount range
	if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
		response.Error(w, errors.BadRequest("min_amount cannot be greater than max_amount"))
		return
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

// SearchAllTransactions handles GET /api/v1/admin/transactions/search (admin operation)
func (h *TransactionHandler) SearchAllTransactions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	filter := &models.TransactionFilter{}

	// Transaction ID (exact match)
	if txID := r.URL.Query().Get("transaction_id"); txID != "" {
		filter.TransactionID = &txID
	}

	// User ID (via wallet ownership)
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filter.UserID = &userID
	}

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

	// Search filter (description/reference)
	if searchParam := r.URL.Query().Get("search"); searchParam != "" {
		if len(searchParam) < 2 {
			response.Error(w, errors.BadRequest("search query must be at least 2 characters"))
			return
		}
		if len(searchParam) > 200 {
			response.Error(w, errors.BadRequest("search query too long (max 200 characters)"))
			return
		}
		filter.Search = &searchParam
	}

	// Amount range filters
	if minAmountParam := r.URL.Query().Get("min_amount"); minAmountParam != "" {
		minAmount, err := strconv.ParseInt(minAmountParam, 10, 64)
		if err != nil || minAmount < 0 {
			response.Error(w, errors.BadRequest("invalid min_amount value"))
			return
		}
		filter.MinAmount = &minAmount
	}

	if maxAmountParam := r.URL.Query().Get("max_amount"); maxAmountParam != "" {
		maxAmount, err := strconv.ParseInt(maxAmountParam, 10, 64)
		if err != nil || maxAmount < 0 {
			response.Error(w, errors.BadRequest("invalid max_amount value"))
			return
		}
		filter.MaxAmount = &maxAmount
	}

	// Validate amount range
	if filter.MinAmount != nil && filter.MaxAmount != nil && *filter.MinAmount > *filter.MaxAmount {
		response.Error(w, errors.BadRequest("min_amount cannot be greater than max_amount"))
		return
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

	transactions, err := h.transactionService.SearchAllTransactions(r.Context(), filter)
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

// ProcessTransfer handles POST /internal/v1/transactions/:id/process (internal endpoint)
// This endpoint processes a pending transfer transaction by executing the wallet-to-wallet transfer.
func (h *TransactionHandler) ProcessTransfer(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("id")

	if transactionID == "" {
		response.Error(w, errors.BadRequest("transaction ID is required"))
		return
	}

	processErr := h.transactionService.ProcessTransfer(r.Context(), transactionID)
	if processErr != nil {
		response.Error(w, processErr)
		return
	}

	response.OK(w, map[string]interface{}{
		"success":        true,
		"transaction_id": transactionID,
		"message":        "Transfer processed successfully",
	})
}

// ========================================================================
// Spending Category Endpoints
// ========================================================================

// UpdateTransactionCategory handles PATCH /api/v1/transactions/:id/category
func (h *TransactionHandler) UpdateTransactionCategory(w http.ResponseWriter, r *http.Request) {
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
	req, parseErr := model.ParseInto[models.UpdateCategoryRequest](body)
	if parseErr != nil {
		response.Error(w, errors.Validation(parseErr.Error()))
		return
	}

	category := models.SpendingCategory(req.Category)
	transaction, updateErr := h.transactionService.UpdateTransactionCategory(r.Context(), transactionID, category)
	if updateErr != nil {
		response.Error(w, updateErr)
		return
	}

	response.OK(w, transaction)
}

// GetSpendingSummary handles GET /api/v1/wallets/:walletId/spending-summary
func (h *TransactionHandler) GetSpendingSummary(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("walletId")

	if walletID == "" {
		response.Error(w, errors.BadRequest("wallet ID is required"))
		return
	}

	// Parse date range from query params (default to current month)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Default to current month if not provided
	if startDate == "" {
		now := r.Context().Value("now")
		if now == nil {
			startDate = firstDayOfMonth()
		}
	}
	if endDate == "" {
		endDate = lastDayOfMonth()
	}

	summary, err := h.transactionService.GetSpendingSummary(r.Context(), walletID, startDate, endDate)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, summary)
}

// AutoCategorizeTransaction handles POST /api/v1/transactions/:id/auto-categorize
func (h *TransactionHandler) AutoCategorizeTransaction(w http.ResponseWriter, r *http.Request) {
	transactionID := r.PathValue("id")

	if transactionID == "" {
		response.Error(w, errors.BadRequest("transaction ID is required"))
		return
	}

	transaction, err := h.transactionService.AutoCategorizeTransaction(r.Context(), transactionID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.OK(w, transaction)
}

// firstDayOfMonth returns the first day of the current month in ISO format.
func firstDayOfMonth() string {
	now := currentTime()
	return now.Format("2006-01") + "-01"
}

// lastDayOfMonth returns the last day of the current month in ISO format.
func lastDayOfMonth() string {
	now := currentTime()
	firstOfNextMonth := now.AddDate(0, 1, -now.Day()+1)
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Format("2006-01-02")
}

// currentTime returns the current time (can be mocked for testing).
var currentTime = func() time.Time {
	return time.Now()
}
