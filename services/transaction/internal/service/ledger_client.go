package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
)

// LedgerClient handles communication with the Ledger service.
type LedgerClient struct {
	*clients.BaseClient
}

// NewLedgerClient creates a new Ledger service client.
func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.DefaultTimeout),
	}
}

// LedgerLine represents a ledger entry line (debit or credit).
type LedgerLine struct {
	AccountID    string `json:"account_id"`
	DebitAmount  int64  `json:"debit_amount"`
	CreditAmount int64  `json:"credit_amount"`
	Description  string `json:"description"`
}

// CreateJournalEntryRequest represents a journal entry creation request.
type CreateJournalEntryRequest struct {
	Type          string         `json:"type"`
	Description   string         `json:"description"`
	ReferenceType string         `json:"reference_type"`
	ReferenceID   string         `json:"reference_id"`
	Lines         []LedgerLine   `json:"lines"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

// JournalEntry represents a ledger journal entry.
type JournalEntry struct {
	ID            string         `json:"id"`
	EntryNumber   string         `json:"entry_number"`
	Type          string         `json:"type"`
	Status        string         `json:"status"`
	Description   string         `json:"description"`
	ReferenceType string         `json:"reference_type"`
	ReferenceID   string         `json:"reference_id"`
	Lines         []LedgerLine   `json:"lines"`
	Metadata      map[string]any `json:"metadata"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// CreateJournalEntry creates a new journal entry in the ledger.
func (c *LedgerClient) CreateJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*JournalEntry, *errors.Error) {
	var result JournalEntry
	if err := c.Post(ctx, "/api/v1/journal-entries", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PostJournalEntry posts a draft journal entry to the ledger (finalizes it).
func (c *LedgerClient) PostJournalEntry(ctx context.Context, entryID string) (*JournalEntry, *errors.Error) {
	var result JournalEntry
	path := fmt.Sprintf("/api/v1/journal-entries/%s/post", entryID)
	if err := c.Post(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateAndPostJournalEntry creates a journal entry and posts it in one operation.
func (c *LedgerClient) CreateAndPostJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*JournalEntry, *errors.Error) {
	// Create the draft entry
	entry, createErr := c.CreateJournalEntry(ctx, req)
	if createErr != nil {
		return nil, createErr
	}

	// Post the entry to finalize it
	postedEntry, postErr := c.PostJournalEntry(ctx, entry.ID)
	if postErr != nil {
		return entry, postErr // Return draft entry even if posting fails
	}

	return postedEntry, nil
}
