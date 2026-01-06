package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vnykmshr/nivo/services/transaction/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/events"
)

// TransactionRepositoryInterface defines the interface for transaction repository operations.
type TransactionRepositoryInterface interface {
	Create(ctx context.Context, transaction *models.Transaction) *errors.Error
	GetByID(ctx context.Context, id string) (*models.Transaction, *errors.Error)
	ListByWallet(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
	SearchAll(ctx context.Context, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error)
	UpdateMetadata(ctx context.Context, id string, metadata map[string]string) *errors.Error
	CompleteWithMetadata(ctx context.Context, id string, metadata map[string]string) *errors.Error
	UpdateStatus(ctx context.Context, id string, status models.TransactionStatus, failureReason *string) *errors.Error
	UpdateCategory(ctx context.Context, id string, category models.SpendingCategory) *errors.Error
	GetCategoryPatterns(ctx context.Context) ([]*models.CategoryPattern, *errors.Error)
	GetCategorySummary(ctx context.Context, walletID string, startDate, endDate string) ([]models.CategorySummary, *errors.Error)
}

// TransactionService handles business logic for transaction operations.
type TransactionService struct {
	transactionRepo TransactionRepositoryInterface
	riskClient      *RiskClient
	walletClient    *WalletClient
	ledgerClient    *LedgerClient
	eventPublisher  *events.Publisher
}

// NewTransactionService creates a new transaction service.
func NewTransactionService(transactionRepo TransactionRepositoryInterface, riskClient *RiskClient, walletClient *WalletClient, ledgerClient *LedgerClient, eventPublisher *events.Publisher) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		riskClient:      riskClient,
		walletClient:    walletClient,
		ledgerClient:    ledgerClient,
		eventPublisher:  eventPublisher,
	}
}

// CreateTransfer creates a transfer transaction between wallets.
func (s *TransactionService) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Validate source and destination are different
	if req.SourceWalletID == req.DestinationWalletID {
		return nil, errors.BadRequest("source and destination wallets must be different")
	}

	// Create transaction
	sourceWalletID := req.SourceWalletID
	destWalletID := req.DestinationWalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:                models.TransactionTypeTransfer,
		Status:              models.TransactionStatusPending,
		SourceWalletID:      &sourceWalletID,
		DestinationWalletID: &destWalletID,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Description:         req.Description,
		Reference:           reference,
		Metadata:            metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(transaction.Status),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"source_wallet_id":      transaction.SourceWalletID,
			"destination_wallet_id": transaction.DestinationWalletID,
			"description":           transaction.Description,
		})
	}

	// Evaluate risk for the transaction
	if evalErr := s.evaluateTransactionRisk(ctx, transaction); evalErr != nil {
		log.Printf("[transaction] Risk evaluation failed for transaction %s: %v", transaction.ID, evalErr)
		// Continue processing even if risk evaluation fails (fail open for now)
	}

	// TODO: In production, trigger async processing:
	// 1. Verify source wallet has sufficient balance
	// 2. Create hold on source wallet
	// 3. Create ledger entry
	// 4. Update wallet balances
	// 5. Mark transaction as completed
	// For now, transaction remains in pending state

	return transaction, nil
}

// CreateDeposit creates a deposit transaction to a wallet.
func (s *TransactionService) CreateDeposit(ctx context.Context, req *models.CreateDepositRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	destWalletID := req.WalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusPending,
		DestinationWalletID: &destWalletID,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Description:         req.Description,
		Reference:           reference,
		Metadata:            metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(transaction.Status),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"destination_wallet_id": transaction.DestinationWalletID,
			"description":           transaction.Description,
		})
	}

	// TODO: Trigger async processing for deposit
	// 1. Verify external payment received
	// 2. Create ledger entry
	// 3. Update wallet balance
	// 4. Mark transaction as completed

	return transaction, nil
}

// InitiateUPIDeposit initiates a UPI deposit and returns virtual UPI ID for payment.
func (s *TransactionService) InitiateUPIDeposit(ctx context.Context, req *models.CreateUPIDepositRequest) (*models.UPIDepositResponse, *errors.Error) {
	// Generate virtual UPI ID (mock format: nivomoney.{wallet_suffix}@yesbank)
	walletSuffix := req.WalletID[len(req.WalletID)-8:]
	virtualUPIID := fmt.Sprintf("nivomoney.%s@yesbank", walletSuffix)

	// Set description
	description := req.Description
	if description == "" {
		description = "UPI Deposit"
	}

	// Create deposit transaction with UPI metadata
	destWalletID := req.WalletID
	upiTransactionID := fmt.Sprintf("UPI%d", time.Now().UnixNano())

	transaction := &models.Transaction{
		Type:                models.TransactionTypeDeposit,
		Status:              models.TransactionStatusPending,
		DestinationWalletID: &destWalletID,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Description:         description,
		Metadata: map[string]string{
			"payment_method":     "upi",
			"virtual_upi_id":     virtualUPIID,
			"upi_transaction_id": upiTransactionID,
		},
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.upi_deposit.initiated", transaction.ID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(transaction.Status),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"destination_wallet_id": transaction.DestinationWalletID,
			"virtual_upi_id":        virtualUPIID,
		})
	}

	// Calculate expiry (30 minutes from now)
	expiresAt := time.Now().Add(30 * time.Minute).Format(time.RFC3339)

	// Prepare response with mock QR code
	response := &models.UPIDepositResponse{
		Transaction:  transaction,
		VirtualUPIID: virtualUPIID,
		QRCode:       generateMockQRCode(virtualUPIID, req.Amount),
		ExpiresAt:    expiresAt,
		Instructions: []string{
			"Open any UPI app (Google Pay, PhonePe, Paytm, etc.)",
			fmt.Sprintf("Pay to UPI ID: %s", virtualUPIID),
			fmt.Sprintf("Amount: ₹%.2f", float64(req.Amount)/100),
			"Your wallet will be credited instantly upon successful payment",
			"This is a simulation - use the 'Simulate Payment' button to complete",
		},
	}

	return response, nil
}

// CompleteUPIDeposit completes a UPI deposit (simulates webhook from payment gateway).
func (s *TransactionService) CompleteUPIDeposit(ctx context.Context, req *models.CompleteUPIDepositRequest) (*models.Transaction, *errors.Error) {
	// Get transaction
	transaction, err := s.transactionRepo.GetByID(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}

	// Validate transaction
	if transaction.Type != models.TransactionTypeDeposit {
		return nil, errors.BadRequest("transaction is not a deposit")
	}

	// Check if already completed (idempotency)
	if transaction.Status == models.TransactionStatusCompleted {
		log.Printf("[transaction] UPI deposit already completed: transaction_id=%s", transaction.ID)
		return transaction, nil
	}

	if transaction.Status != models.TransactionStatusPending {
		return nil, errors.BadRequest("transaction is not in pending status")
	}

	// Check if it's a UPI deposit
	if transaction.Metadata == nil || transaction.Metadata["payment_method"] != "upi" {
		return nil, errors.BadRequest("transaction is not a UPI deposit")
	}

	// Update transaction based on status
	if req.Status == "success" {
		// Update metadata with external UPI transaction ID
		updatedMetadata := make(map[string]string)
		for k, v := range transaction.Metadata {
			updatedMetadata[k] = v
		}
		updatedMetadata["external_upi_transaction_id"] = req.UPITransactionID

		// Complete transaction atomically with metadata update (provides idempotency)
		if updateErr := s.transactionRepo.CompleteWithMetadata(ctx, transaction.ID, updatedMetadata); updateErr != nil {
			return nil, updateErr
		}

		// Refetch to get updated transaction
		transaction, err = s.transactionRepo.GetByID(ctx, req.TransactionID)
		if err != nil {
			return nil, err
		}

		// Publish completion event
		if s.eventPublisher != nil {
			s.eventPublisher.PublishTransactionEvent("transaction.upi_deposit.completed", transaction.ID, map[string]interface{}{
				"type":                  string(transaction.Type),
				"status":                string(transaction.Status),
				"amount":                transaction.Amount,
				"currency":              transaction.Currency,
				"destination_wallet_id": transaction.DestinationWalletID,
				"upi_transaction_id":    req.UPITransactionID,
			})
		}

		log.Printf("[transaction] UPI deposit completed: transaction_id=%s, amount=%d", transaction.ID, transaction.Amount)

		// Credit the deposit to the wallet
		// NOTE: In a fully event-driven architecture, this would be handled by the Wallet service
		// listening to the "transaction.upi_deposit.completed" event. For now, we call it directly.
		if s.walletClient != nil && transaction.DestinationWalletID != nil {
			depositReq := &DepositRequest{
				WalletID:      *transaction.DestinationWalletID,
				Amount:        transaction.Amount,
				TransactionID: transaction.ID,
				Description:   transaction.Description,
			}
			if creditErr := s.walletClient.CreditDeposit(ctx, depositReq); creditErr != nil {
				log.Printf("[transaction] Warning: Failed to credit deposit to wallet: %v", creditErr)
				// Don't fail the transaction - it's already marked as completed
				// Manual intervention may be needed to reconcile the wallet balance
			} else {
				log.Printf("[transaction] Wallet credited: wallet_id=%s, amount=%d", *transaction.DestinationWalletID, transaction.Amount)
			}
		}
	} else {
		// Mark as failed
		failureReason := "UPI payment failed"
		if updateErr := s.transactionRepo.UpdateStatus(ctx, transaction.ID, models.TransactionStatusFailed, &failureReason); updateErr != nil {
			return nil, updateErr
		}

		// Refetch to get updated transaction
		transaction, err = s.transactionRepo.GetByID(ctx, req.TransactionID)
		if err != nil {
			return nil, err
		}

		// Publish failure event
		if s.eventPublisher != nil {
			s.eventPublisher.PublishTransactionEvent("transaction.upi_deposit.failed", transaction.ID, map[string]interface{}{
				"type":           string(transaction.Type),
				"status":         string(transaction.Status),
				"failure_reason": failureReason,
			})
		}

		log.Printf("[transaction] UPI deposit failed: transaction_id=%s", transaction.ID)
	}

	return transaction, nil
}

// generateMockQRCode generates a mock base64 QR code string.
func generateMockQRCode(upiID string, amount int64) string {
	// In a real implementation, this would generate an actual QR code
	// For now, return a mock base64 string (1x1 transparent PNG)
	return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
}

// CreateWithdrawal creates a withdrawal transaction from a wallet.
func (s *TransactionService) CreateWithdrawal(ctx context.Context, req *models.CreateWithdrawalRequest) (*models.Transaction, *errors.Error) {
	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	sourceWalletID := req.WalletID
	var reference *string
	if req.Reference != "" {
		reference = &req.Reference
	}

	transaction := &models.Transaction{
		Type:           models.TransactionTypeWithdrawal,
		Status:         models.TransactionStatusPending,
		SourceWalletID: &sourceWalletID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Description:    req.Description,
		Reference:      reference,
		Metadata:       metadata,
	}

	if createErr := s.transactionRepo.Create(ctx, transaction); createErr != nil {
		return nil, createErr
	}

	// Publish transaction.created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.created", transaction.ID, map[string]interface{}{
			"type":             string(transaction.Type),
			"status":           string(transaction.Status),
			"amount":           transaction.Amount,
			"currency":         transaction.Currency,
			"source_wallet_id": transaction.SourceWalletID,
			"description":      transaction.Description,
		})
	}

	// TODO: Trigger async processing for withdrawal
	// 1. Verify wallet has sufficient balance
	// 2. Create hold on wallet
	// 3. Initiate external payment
	// 4. Create ledger entry
	// 5. Update wallet balance
	// 6. Mark transaction as completed

	return transaction, nil
}

// GetTransaction retrieves a transaction by ID.
func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*models.Transaction, *errors.Error) {
	return s.transactionRepo.GetByID(ctx, id)
}

// ListWalletTransactions retrieves transactions for a wallet.
func (s *TransactionService) ListWalletTransactions(ctx context.Context, walletID string, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	return s.transactionRepo.ListByWallet(ctx, walletID, filter)
}

// SearchAllTransactions searches transactions across all wallets (admin operation).
func (s *TransactionService) SearchAllTransactions(ctx context.Context, filter *models.TransactionFilter) ([]*models.Transaction, *errors.Error) {
	// Validate filter parameters
	if filter != nil {
		// Validate limit (max 100)
		if filter.Limit <= 0 || filter.Limit > 100 {
			filter.Limit = 50 // Default limit
		}
		if filter.Offset < 0 {
			filter.Offset = 0
		}
	}

	return s.transactionRepo.SearchAll(ctx, filter)
}

// ReverseTransaction reverses a completed transaction.
func (s *TransactionService) ReverseTransaction(ctx context.Context, transactionID, reason string) (*models.Transaction, *errors.Error) {
	// Get original transaction
	originalTx, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Validate transaction can be reversed
	if !originalTx.IsCompleted() {
		return nil, errors.BadRequest("only completed transactions can be reversed")
	}

	if originalTx.Type == models.TransactionTypeReversal {
		return nil, errors.BadRequest("cannot reverse a reversal transaction")
	}

	// Create reversal transaction
	parentID := transactionID
	reversalTx := &models.Transaction{
		Type:                models.TransactionTypeReversal,
		Status:              models.TransactionStatusPending,
		SourceWalletID:      originalTx.DestinationWalletID, // Reverse direction
		DestinationWalletID: originalTx.SourceWalletID,
		Amount:              originalTx.Amount,
		Currency:            originalTx.Currency,
		Description:         "Reversal: " + reason,
		ParentTransactionID: &parentID,
		Metadata:            map[string]string{"reversal_reason": reason},
	}

	if createErr := s.transactionRepo.Create(ctx, reversalTx); createErr != nil {
		return nil, createErr
	}

	// TODO: Trigger async processing for reversal
	// 1. Create reversal ledger entry
	// 2. Update wallet balances
	// 3. Mark reversal as completed
	// 4. Mark original transaction as reversed

	return reversalTx, nil
}

// ProcessTransfer processes a pending transfer transaction by executing the wallet transfer
// with limit checking and balance updates. This is typically called after risk evaluation.
func (s *TransactionService) ProcessTransfer(ctx context.Context, transactionID string) *errors.Error {
	// Get the transaction
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}

	// Validate transaction is pending
	if transaction.Status != models.TransactionStatusPending {
		return errors.BadRequest(fmt.Sprintf("transaction is not pending (status: %s)", transaction.Status))
	}

	// Validate transaction is a transfer
	if transaction.Type != models.TransactionTypeTransfer {
		return errors.BadRequest(fmt.Sprintf("transaction is not a transfer (type: %s)", transaction.Type))
	}

	// Validate required fields
	if transaction.SourceWalletID == nil || transaction.DestinationWalletID == nil {
		return errors.BadRequest("transfer must have both source and destination wallets")
	}

	// Call wallet service to execute the transfer (includes limit checking and balance updates)
	if s.walletClient == nil {
		log.Printf("[transaction] Wallet client not configured, cannot process transfer")
		return errors.Internal("wallet client not configured")
	}

	transferReq := &TransferRequest{
		SourceWalletID:      *transaction.SourceWalletID,
		DestinationWalletID: *transaction.DestinationWalletID,
		Amount:              transaction.Amount,
		TransactionID:       transaction.ID,
		Description:         transaction.Description,
	}

	transferErr := s.walletClient.ExecuteTransfer(ctx, transferReq)
	if transferErr != nil {
		// Transfer failed - update transaction status
		failureReason := transferErr.Error()
		updateErr := s.transactionRepo.UpdateStatus(ctx, transactionID, models.TransactionStatusFailed, &failureReason)
		if updateErr != nil {
			log.Printf("[transaction] Failed to update failed transaction status: %v", updateErr)
		}

		log.Printf("[transaction] Transfer failed for transaction %s: %v", transactionID, transferErr)
		return errors.Internal(fmt.Sprintf("transfer failed: %s", failureReason))
	}

	// TODO: Create ledger entry (if needed)
	// For now, the wallet service handles the balance updates directly

	// Mark transaction as completed
	completeErr := s.transactionRepo.UpdateStatus(ctx, transactionID, models.TransactionStatusCompleted, nil)
	if completeErr != nil {
		log.Printf("[transaction] Failed to mark transaction as completed: %v", completeErr)
		return completeErr
	}

	// Publish transaction.completed event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishTransactionEvent("transaction.completed", transactionID, map[string]interface{}{
			"type":                  string(transaction.Type),
			"status":                string(models.TransactionStatusCompleted),
			"amount":                transaction.Amount,
			"currency":              transaction.Currency,
			"source_wallet_id":      transaction.SourceWalletID,
			"destination_wallet_id": transaction.DestinationWalletID,
		})
	}

	log.Printf("[transaction] Transfer completed successfully for transaction %s", transactionID)
	return nil
}

// evaluateTransactionRisk evaluates risk for a transaction using the Risk Service.
func (s *TransactionService) evaluateTransactionRisk(ctx context.Context, transaction *models.Transaction) error {
	if s.riskClient == nil {
		log.Printf("[transaction] Risk client not configured, skipping risk evaluation")
		return nil
	}

	// Extract user ID from wallet ownership (for now, use a placeholder)
	// In production, you would fetch the wallet owner from the wallet service
	userID := "unknown"

	// Prepare risk evaluation request
	riskReq := &RiskEvaluationRequest{
		TransactionID:   transaction.ID,
		UserID:          userID,
		Amount:          transaction.Amount,
		Currency:        string(transaction.Currency),
		TransactionType: string(transaction.Type),
	}

	if transaction.SourceWalletID != nil {
		riskReq.FromWalletID = *transaction.SourceWalletID
	}
	if transaction.DestinationWalletID != nil {
		riskReq.ToWalletID = *transaction.DestinationWalletID
	}

	// Call risk service
	result, err := s.riskClient.EvaluateTransaction(ctx, riskReq)
	if err != nil {
		return err
	}

	// Log risk evaluation result
	log.Printf("[transaction] Risk evaluation for transaction %s: action=%s, score=%d, allowed=%v",
		transaction.ID, result.Action, result.RiskScore, result.Allowed)

	// Store risk information in transaction metadata
	if transaction.Metadata == nil {
		transaction.Metadata = make(map[string]string)
	}
	transaction.Metadata["risk_score"] = fmt.Sprintf("%d", result.RiskScore)
	transaction.Metadata["risk_action"] = result.Action
	transaction.Metadata["risk_event_id"] = result.EventID

	if len(result.TriggeredRules) > 0 {
		transaction.Metadata["risk_triggered_rules"] = fmt.Sprintf("%d", len(result.TriggeredRules))
	}

	// Handle risk actions
	if !result.Allowed {
		log.Printf("[transaction] Transaction %s BLOCKED by risk evaluation: %s", transaction.ID, result.Reason)
		// In production, you would update the transaction status to failed
		// For now, just log the blocking decision
	} else if result.Action == "flag" {
		log.Printf("[transaction] Transaction %s FLAGGED by risk evaluation: %s", transaction.ID, result.Reason)
		// In production, you might notify compliance team or require manual review
	}

	return nil
}

// ========================================================================
// Spending Category Operations
// ========================================================================

// UpdateTransactionCategory updates the category of a transaction.
func (s *TransactionService) UpdateTransactionCategory(ctx context.Context, transactionID string, category models.SpendingCategory) (*models.Transaction, *errors.Error) {
	// Validate category
	if !models.ValidCategories[category] {
		return nil, errors.Validation("invalid spending category")
	}

	// Verify transaction exists
	if _, err := s.transactionRepo.GetByID(ctx, transactionID); err != nil {
		return nil, err
	}

	// Update category
	if updateErr := s.transactionRepo.UpdateCategory(ctx, transactionID, category); updateErr != nil {
		return nil, updateErr
	}

	// Refetch to get updated transaction
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	log.Printf("[transaction] Category updated for transaction %s: %s", transactionID, category)
	return transaction, nil
}

// GetSpendingSummary retrieves spending summary grouped by category for a wallet.
func (s *TransactionService) GetSpendingSummary(ctx context.Context, walletID, startDate, endDate string) (*models.CategorySummaryResponse, *errors.Error) {
	summaries, err := s.transactionRepo.GetCategorySummary(ctx, walletID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate total spent
	var totalSpent int64
	for _, s := range summaries {
		totalSpent += s.TotalAmount
	}

	response := &models.CategorySummaryResponse{
		Categories: summaries,
		TotalSpent: totalSpent,
	}
	response.Period.StartDate = startDate
	response.Period.EndDate = endDate

	return response, nil
}

// AutoCategorizeTransaction automatically categorizes a transaction based on its description.
func (s *TransactionService) AutoCategorizeTransaction(ctx context.Context, transactionID string) (*models.Transaction, *errors.Error) {
	// Get transaction
	transaction, err := s.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Get category patterns
	patterns, patternErr := s.transactionRepo.GetCategoryPatterns(ctx)
	if patternErr != nil {
		return nil, patternErr
	}

	// Match description against patterns (case-insensitive)
	descLower := strings.ToLower(transaction.Description)
	var matchedCategory models.SpendingCategory

	for _, pattern := range patterns {
		if strings.Contains(descLower, strings.ToLower(pattern.Pattern)) {
			matchedCategory = pattern.Category
			break // Patterns are ordered by priority, take first match
		}
	}

	// If no match found, default to 'other'
	if matchedCategory == "" {
		matchedCategory = models.CategoryOther
	}

	// Update if different from current
	if transaction.Category != matchedCategory {
		if updateErr := s.transactionRepo.UpdateCategory(ctx, transactionID, matchedCategory); updateErr != nil {
			return nil, updateErr
		}
		log.Printf("[transaction] Auto-categorized transaction %s as %s", transactionID, matchedCategory)
	}

	// Refetch
	return s.transactionRepo.GetByID(ctx, transactionID)
}

// ========================================================================
// Statement Export Operations
// ========================================================================

// StatementRequest represents a request for a statement export.
type StatementRequest struct {
	WalletID  string
	StartDate string
	EndDate   string
	Format    string // "csv" or "pdf"
}

// StatementData represents the data for a statement export.
type StatementData struct {
	WalletID     string
	StartDate    string
	EndDate      string
	Transactions []*models.Transaction
	TotalCredits int64
	TotalDebits  int64
	NetBalance   int64
	GeneratedAt  string
}

// GetStatementData retrieves statement data for a wallet within a date range.
func (s *TransactionService) GetStatementData(ctx context.Context, walletID, startDate, endDate string) (*StatementData, *errors.Error) {
	// Create filter for date range
	filter := &models.TransactionFilter{
		Limit: 1000, // Reasonable limit for statement
	}

	// Fetch transactions for the wallet
	transactions, err := s.transactionRepo.ListByWallet(ctx, walletID, filter)
	if err != nil {
		return nil, err
	}

	// Filter by date and calculate totals
	var filteredTx []*models.Transaction
	var totalCredits, totalDebits int64

	for _, tx := range transactions {
		// Only include completed transactions
		if tx.Status != models.TransactionStatusCompleted {
			continue
		}

		// Filter by date (simple string comparison for ISO dates)
		txDate := tx.CreatedAt.Format("2006-01-02")
		if txDate < startDate || txDate > endDate {
			continue
		}

		filteredTx = append(filteredTx, tx)

		// Calculate credits/debits from wallet perspective
		if tx.DestinationWalletID != nil && *tx.DestinationWalletID == walletID {
			totalCredits += tx.Amount
		}
		if tx.SourceWalletID != nil && *tx.SourceWalletID == walletID {
			totalDebits += tx.Amount
		}
	}

	return &StatementData{
		WalletID:     walletID,
		StartDate:    startDate,
		EndDate:      endDate,
		Transactions: filteredTx,
		TotalCredits: totalCredits,
		TotalDebits:  totalDebits,
		NetBalance:   totalCredits - totalDebits,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	}, nil
}

// GenerateCSV generates a CSV statement from statement data.
func (s *TransactionService) GenerateCSV(data *StatementData) []byte {
	var buf strings.Builder

	// Write header
	buf.WriteString("Date,Transaction ID,Type,Description,Category,Debit,Credit,Status\n")

	// Write transactions
	for _, tx := range data.Transactions {
		date := tx.CreatedAt.Format("2006-01-02 15:04:05")
		txType := string(tx.Type)
		desc := escapeCSV(tx.Description)
		category := string(tx.Category)
		status := string(tx.Status)

		var debit, credit string
		if tx.SourceWalletID != nil && *tx.SourceWalletID == data.WalletID {
			debit = formatAmount(tx.Amount)
		}
		if tx.DestinationWalletID != nil && *tx.DestinationWalletID == data.WalletID {
			credit = formatAmount(tx.Amount)
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n",
			date, tx.ID, txType, desc, category, debit, credit, status)
		buf.WriteString(line)
	}

	// Write summary
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("Statement Period:,%s to %s\n", data.StartDate, data.EndDate))
	buf.WriteString(fmt.Sprintf("Total Credits:,,%s\n", formatAmount(data.TotalCredits)))
	buf.WriteString(fmt.Sprintf("Total Debits:,%s\n", formatAmount(data.TotalDebits)))
	buf.WriteString(fmt.Sprintf("Net Balance:,%s\n", formatAmount(data.NetBalance)))
	buf.WriteString(fmt.Sprintf("Generated:,%s\n", data.GeneratedAt))

	return []byte(buf.String())
}

// GeneratePDF generates a simple text-based PDF statement.
func (s *TransactionService) GeneratePDF(data *StatementData) []byte {
	// Create a simple text content for PDF
	var content strings.Builder

	content.WriteString("NIVO NEOBANK - ACCOUNT STATEMENT\n")
	content.WriteString("================================\n\n")
	content.WriteString(fmt.Sprintf("Wallet ID: %s\n", data.WalletID))
	content.WriteString(fmt.Sprintf("Period: %s to %s\n", data.StartDate, data.EndDate))
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", data.GeneratedAt))
	content.WriteString("TRANSACTION DETAILS\n")
	content.WriteString("-------------------\n\n")

	// Column headers
	content.WriteString(fmt.Sprintf("%-20s %-12s %-30s %-12s %15s %15s\n",
		"Date", "Type", "Description", "Category", "Debit", "Credit"))
	content.WriteString(strings.Repeat("-", 110) + "\n")

	// Transactions
	for _, tx := range data.Transactions {
		date := tx.CreatedAt.Format("2006-01-02 15:04")
		txType := string(tx.Type)
		desc := truncateString(tx.Description, 28)
		category := string(tx.Category)

		var debit, credit string
		if tx.SourceWalletID != nil && *tx.SourceWalletID == data.WalletID {
			debit = formatAmount(tx.Amount)
		}
		if tx.DestinationWalletID != nil && *tx.DestinationWalletID == data.WalletID {
			credit = formatAmount(tx.Amount)
		}

		content.WriteString(fmt.Sprintf("%-20s %-12s %-30s %-12s %15s %15s\n",
			date, txType, desc, category, debit, credit))
	}

	content.WriteString(strings.Repeat("-", 110) + "\n\n")

	// Summary
	content.WriteString("SUMMARY\n")
	content.WriteString("-------\n")
	content.WriteString(fmt.Sprintf("Total Credits:  %s\n", formatAmount(data.TotalCredits)))
	content.WriteString(fmt.Sprintf("Total Debits:   %s\n", formatAmount(data.TotalDebits)))
	content.WriteString(fmt.Sprintf("Net Balance:    %s\n", formatAmount(data.NetBalance)))
	content.WriteString("\n")
	content.WriteString("This is a computer-generated statement and does not require a signature.\n")

	// For a production system, use a proper PDF library like gofpdf
	// For now, return the text content as-is (can be rendered as PDF by frontend)
	return []byte(content.String())
}

// escapeCSV escapes a string for CSV output.
func escapeCSV(s string) string {
	// Prevent CSV injection by prefixing cells that start with formula characters
	// with a single quote. These characters can trigger formula execution in
	// spreadsheet applications like Excel.
	if len(s) > 0 {
		firstChar := s[0]
		if firstChar == '=' || firstChar == '+' || firstChar == '-' || firstChar == '@' ||
			firstChar == '\t' || firstChar == '\r' {
			s = "'" + s
		}
	}

	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

// formatAmount formats an amount in paise to rupees string.
func formatAmount(paise int64) string {
	if paise == 0 {
		return ""
	}
	rupees := float64(paise) / 100
	return fmt.Sprintf("%.2f", rupees)
}

// truncateString truncates a string to a maximum length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-2] + ".."
}
