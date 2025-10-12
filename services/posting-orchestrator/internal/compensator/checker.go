package compensator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// LedgerChecker checks the ledger service for entry existence
type LedgerChecker struct {
	ledgerURL  string
	httpClient *http.Client
	logger     *log.Logger
}

// LedgerEntryResponse represents the response from ledger's GetEntryByBatch endpoint
type LedgerEntryResponse struct {
	EntryID string `json:"entry_id"`
	BatchID string `json:"batch_id"`
	Voided  bool   `json:"voided"`
}

// NewLedgerChecker creates a new ledger checker
func NewLedgerChecker(ledgerURL string, logger *log.Logger) *LedgerChecker {
	return &LedgerChecker{
		ledgerURL: ledgerURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// CheckEntry checks if a journal entry exists for the given transfer ID
// Returns (exists bool, entryID string, error)
func (c *LedgerChecker) CheckEntry(ctx context.Context, transferID string) (bool, string, error) {
	// The transfer ID is used as the batch ID in the ledger
	url := fmt.Sprintf("%s/v1/entries/by-batch/%s", c.ledgerURL, transferID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("ledger request failed: %w", err)
	}
	defer resp.Body.Close()

	// If entry not found, that's OK - it means ledger never processed it
	if resp.StatusCode == http.StatusNotFound {
		return false, "", nil
	}

	// Any other non-2xx status is an error
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("ledger returned status %d", resp.StatusCode)
	}

	// Parse response
	var entryResp LedgerEntryResponse
	if err := json.NewDecoder(resp.Body).Decode(&entryResp); err != nil {
		return false, "", fmt.Errorf("failed to parse ledger response: %w", err)
	}

	// Check if entry is voided (shouldn't happen in recovery, but be defensive)
	if entryResp.Voided {
		c.logger.Printf("WARNING: Entry %s for transfer %s is voided", entryResp.EntryID, transferID)
	}

	return true, entryResp.EntryID, nil
}
