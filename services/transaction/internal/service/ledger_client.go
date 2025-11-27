package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LedgerClient handles communication with the Ledger service.
type LedgerClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLedgerClient creates a new Ledger service client.
func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
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
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	ReferenceType string                 `json:"reference_type"`
	ReferenceID   string                 `json:"reference_id"`
	Lines         []LedgerLine           `json:"lines"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// JournalEntry represents a ledger journal entry.
type JournalEntry struct {
	ID            string                 `json:"id"`
	EntryNumber   string                 `json:"entry_number"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	Description   string                 `json:"description"`
	ReferenceType string                 `json:"reference_type"`
	ReferenceID   string                 `json:"reference_id"`
	Lines         []LedgerLine           `json:"lines"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateJournalEntry creates a new journal entry in the ledger.
func (c *LedgerClient) CreateJournalEntry(ctx context.Context, req *CreateJournalEntryRequest) (*JournalEntry, error) {
	url := fmt.Sprintf("%s/api/v1/journal-entries", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Ledger service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ledger service returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response envelope
	var envelope struct {
		Success bool          `json:"success"`
		Data    *JournalEntry `json:"data"`
		Error   *string       `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, fmt.Errorf("create journal entry failed: %s", errMsg)
	}

	return envelope.Data, nil
}
