package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/services/ledger/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// AccountRepositoryInterface defines the interface for account repository operations.
type AccountRepositoryInterface interface {
	Create(ctx context.Context, account *models.Account) *errors.Error
	GetByID(ctx context.Context, id string) (*models.Account, *errors.Error)
	GetByCode(ctx context.Context, code string) (*models.Account, *errors.Error)
	List(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error)
	Update(ctx context.Context, account *models.Account) *errors.Error
	GetBalance(ctx context.Context, accountID string) (int64, *errors.Error)
}

// JournalEntryRepositoryInterface defines the interface for journal entry repository operations.
type JournalEntryRepositoryInterface interface {
	Create(ctx context.Context, entry *models.JournalEntry, lines []models.LedgerLine) *errors.Error
	GetByID(ctx context.Context, id string) (*models.JournalEntry, *errors.Error)
	List(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error)
	Post(ctx context.Context, entryID, postedBy string) *errors.Error
	Void(ctx context.Context, entryID, voidedBy, voidReason string) *errors.Error
}

// LedgerService handles business logic for ledger operations.
type LedgerService struct {
	accountRepo AccountRepositoryInterface
	journalRepo JournalEntryRepositoryInterface
}

// NewLedgerService creates a new ledger service.
func NewLedgerService(
	accountRepo AccountRepositoryInterface,
	journalRepo JournalEntryRepositoryInterface,
) *LedgerService {
	return &LedgerService{
		accountRepo: accountRepo,
		journalRepo: journalRepo,
	}
}

// CreateAccount creates a new ledger account.
func (s *LedgerService) CreateAccount(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, *errors.Error) {
	// Validate parent account exists if specified
	if req.ParentID != nil {
		parent, err := s.accountRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		// Ensure parent is same type
		// (In a full implementation, might allow more flexible hierarchies)
		_ = parent
	}

	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Create account
	account := &models.Account{
		Code:     req.Code,
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
		ParentID: req.ParentID,
		Status:   models.AccountStatusActive,
		Balance:  0,
		Metadata: metadata,
	}

	if createErr := s.accountRepo.Create(ctx, account); createErr != nil {
		return nil, createErr
	}

	return account, nil
}

// GetAccount retrieves an account by ID.
func (s *LedgerService) GetAccount(ctx context.Context, accountID string) (*models.Account, *errors.Error) {
	return s.accountRepo.GetByID(ctx, accountID)
}

// GetAccountByCode retrieves an account by code.
func (s *LedgerService) GetAccountByCode(ctx context.Context, code string) (*models.Account, *errors.Error) {
	return s.accountRepo.GetByCode(ctx, code)
}

// ListAccounts retrieves accounts with filters.
func (s *LedgerService) ListAccounts(ctx context.Context, accountType *models.AccountType, status *models.AccountStatus, limit, offset int) ([]*models.Account, *errors.Error) {
	return s.accountRepo.List(ctx, accountType, status, limit, offset)
}

// UpdateAccount updates an account.
func (s *LedgerService) UpdateAccount(ctx context.Context, accountID string, req *models.UpdateAccountRequest) (*models.Account, *errors.Error) {
	// Get existing account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Parse metadata
	metadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid metadata format")
	}

	// Update fields
	account.Name = req.Name
	account.Status = req.Status
	account.Metadata = metadata

	// Save
	if updateErr := s.accountRepo.Update(ctx, account); updateErr != nil {
		return nil, updateErr
	}

	return account, nil
}

// CreateJournalEntry creates a new journal entry.
// This validates the entry follows double-entry bookkeeping rules.
func (s *LedgerService) CreateJournalEntry(ctx context.Context, req *models.CreateJournalEntryRequest) (*models.JournalEntry, *errors.Error) {
	// Validate lines
	if len(req.Lines) < 2 {
		return nil, errors.Validation("journal entry must have at least 2 lines")
	}

	// Validate each line
	for i, line := range req.Lines {
		if err := line.Validate(); err != nil {
			return nil, errors.Validation(fmt.Sprintf("line %d: %v", i, err))
		}

		// Verify account exists
		account, accErr := s.accountRepo.GetByID(ctx, line.AccountID)
		if accErr != nil {
			return nil, errors.Validation(fmt.Sprintf("line %d: invalid account", i))
		}

		// Verify account is active
		if account.Status != models.AccountStatusActive {
			return nil, errors.Validation(fmt.Sprintf("line %d: account %s is not active", i, account.Code))
		}
	}

	// Validate double-entry: total debits must equal total credits
	var totalDebits, totalCredits int64
	for _, line := range req.Lines {
		totalDebits += line.DebitAmount
		totalCredits += line.CreditAmount
	}

	if totalDebits != totalCredits {
		return nil, errors.Validation(fmt.Sprintf("entry not balanced: debits=%d, credits=%d", totalDebits, totalCredits))
	}

	// Parse entry metadata
	entryMetadata, metaErr := req.GetMetadata()
	if metaErr != nil {
		return nil, errors.Validation("invalid entry metadata format")
	}

	// Create entry
	entry := &models.JournalEntry{
		Type:          req.Type,
		Status:        models.EntryStatusDraft,
		Description:   req.Description,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		Metadata:      entryMetadata,
	}

	// Convert input lines to domain lines
	lines := make([]models.LedgerLine, len(req.Lines))
	for i, lineInput := range req.Lines {
		// Parse line metadata
		lineMetadata, lineMetaErr := lineInput.GetMetadata()
		if lineMetaErr != nil {
			return nil, errors.Validation(fmt.Sprintf("line %d: invalid metadata format", i))
		}

		lines[i] = models.LedgerLine{
			AccountID:    lineInput.AccountID,
			DebitAmount:  lineInput.DebitAmount,
			CreditAmount: lineInput.CreditAmount,
			Description:  lineInput.Description,
			Metadata:     lineMetadata,
		}
	}

	// Create in repository (within transaction)
	if createErr := s.journalRepo.Create(ctx, entry, lines); createErr != nil {
		return nil, createErr
	}

	return entry, nil
}

// GetJournalEntry retrieves a journal entry with its lines.
func (s *LedgerService) GetJournalEntry(ctx context.Context, entryID string) (*models.JournalEntry, *errors.Error) {
	return s.journalRepo.GetByID(ctx, entryID)
}

// ListJournalEntries retrieves journal entries with filters.
func (s *LedgerService) ListJournalEntries(ctx context.Context, status *models.EntryStatus, limit, offset int) ([]*models.JournalEntry, *errors.Error) {
	return s.journalRepo.List(ctx, status, limit, offset)
}

// PostJournalEntry posts a draft journal entry to the ledger.
// This makes the entry permanent and updates account balances.
func (s *LedgerService) PostJournalEntry(ctx context.Context, entryID, postedBy string) (*models.JournalEntry, *errors.Error) {
	// Get entry
	entry, err := s.journalRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if entry.Status != models.EntryStatusDraft {
		return nil, errors.BadRequest("only draft entries can be posted")
	}

	// Validate entry is balanced (should already be, but double-check)
	if !entry.IsBalanced() {
		return nil, errors.Validation("entry is not balanced")
	}

	// Post entry (repository handles balance updates via trigger)
	if postErr := s.journalRepo.Post(ctx, entryID, postedBy); postErr != nil {
		return nil, postErr
	}

	// Return updated entry
	return s.journalRepo.GetByID(ctx, entryID)
}

// VoidJournalEntry voids a posted journal entry.
// Note: This doesn't reverse the balance changes. For true reversal, use ReverseJournalEntry.
func (s *LedgerService) VoidJournalEntry(ctx context.Context, entryID, voidedBy, voidReason string) (*models.JournalEntry, *errors.Error) {
	// Get entry
	entry, err := s.journalRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if entry.Status != models.EntryStatusPosted {
		return nil, errors.BadRequest("only posted entries can be voided")
	}

	// Void entry
	if voidErr := s.journalRepo.Void(ctx, entryID, voidedBy, voidReason); voidErr != nil {
		return nil, voidErr
	}

	// Return updated entry
	return s.journalRepo.GetByID(ctx, entryID)
}

// ReverseJournalEntry creates a reversing entry for a posted journal entry.
// This creates a new entry with opposite debit/credit amounts.
func (s *LedgerService) ReverseJournalEntry(ctx context.Context, entryID, reversedBy, reason string) (*models.JournalEntry, *errors.Error) {
	// Get original entry
	originalEntry, err := s.journalRepo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if originalEntry.Status != models.EntryStatusPosted {
		return nil, errors.BadRequest("only posted entries can be reversed")
	}

	// Create reversing entry with opposite amounts
	reversalLines := make([]models.LedgerLineInput, len(originalEntry.Lines))
	for i, line := range originalEntry.Lines {
		reversalLines[i] = models.LedgerLineInput{
			AccountID:    line.AccountID,
			DebitAmount:  line.CreditAmount, // Swap debit/credit
			CreditAmount: line.DebitAmount,
			Description:  fmt.Sprintf("Reversal of %s: %s", originalEntry.EntryNumber, line.Description),
		}
	}

	// Create reversal entry
	reversalReq := &models.CreateJournalEntryRequest{
		Type:          models.EntryTypeReversing,
		Description:   fmt.Sprintf("Reversal of %s: %s", originalEntry.EntryNumber, reason),
		ReferenceType: "journal_entry",
		ReferenceID:   originalEntry.ID,
		Lines:         reversalLines,
	}

	reversalEntry, createErr := s.CreateJournalEntry(ctx, reversalReq)
	if createErr != nil {
		return nil, createErr
	}

	// Auto-post the reversal entry
	reversalEntry, postErr := s.PostJournalEntry(ctx, reversalEntry.ID, reversedBy)
	if postErr != nil {
		return nil, postErr
	}

	// TODO: Mark original entry as reversed and link to reversal
	// This would require an UPDATE on the original entry in the repository

	return reversalEntry, nil
}

// GetAccountBalance retrieves the current balance of an account.
func (s *LedgerService) GetAccountBalance(ctx context.Context, accountID string) (int64, *errors.Error) {
	return s.accountRepo.GetBalance(ctx, accountID)
}
