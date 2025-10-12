package compensator

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/store"
	"github.com/google/uuid"
)

// Compensator is the background worker that recovers stuck transfers
type Compensator struct {
	db           *sql.DB
	queries      *store.Queries
	ledgerURL    string
	checker      *LedgerChecker
	pollInterval time.Duration // how often to check for stale transfers
	staleTimeout time.Duration // how long before a transfer is considered stale
	logger       *log.Logger
}

// NewCompensator creates a new compensator instance
func NewCompensator(db *sql.DB, ledgerURL string, logger *log.Logger) *Compensator {
	return &Compensator{
		db:           db,
		queries:      store.New(db),
		ledgerURL:    ledgerURL,
		checker:      NewLedgerChecker(ledgerURL, logger),
		pollInterval: 30 * time.Second, // check every 30 seconds
		staleTimeout: 5 * time.Minute,  // transfers older than 5 minutes
		logger:       logger,
	}
}

// Start begins the compensator background worker
func (c *Compensator) Start(ctx context.Context) error {
	c.logger.Println("Compensator worker started")
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	// Run immediately on start
	c.processStaleTransfers(ctx)

	for {
		select {
		case <-ticker.C:
			c.processStaleTransfers(ctx)
		case <-ctx.Done():
			c.logger.Println("Compensator worker stopped")
			return ctx.Err()
		}
	}
}

// processStaleTransfers finds and recovers stale transfers
func (c *Compensator) processStaleTransfers(ctx context.Context) {
	// Find transfers that have been in LEDGER_CALLED or RECOVERING state for too long
	staleThreshold := sql.NullTime{
		Time:  time.Now().Add(-c.staleTimeout),
		Valid: true,
	}
	
	transfers, err := c.queries.GetStaleTransfers(ctx, staleThreshold)
	if err != nil {
		c.logger.Printf("Failed to get stale transfers: %v", err)
		return
	}

	if len(transfers) > 0 {
		c.logger.Printf("Found %d stale transfer(s) to recover", len(transfers))
	}

	for _, transfer := range transfers {
		if err := c.recoverTransfer(ctx, transfer); err != nil {
			c.logger.Printf("Failed to recover transfer %s: %v", transfer.ID, err)
		}
	}
}

// recoverTransfer attempts to recover a single stuck transfer
func (c *Compensator) recoverTransfer(ctx context.Context, transfer store.Transfer) error {
	age := "unknown"
	if transfer.LedgerCallAt.Valid {
		age = time.Since(transfer.LedgerCallAt.Time).String()
	}
	
	c.logger.Printf("Recovering transfer %s (state: %s, age: %v)", 
		transfer.ID, 
		transfer.State.String, 
		age)

	// Mark as recovering (idempotent - only works if in LEDGER_CALLED state)
	if err := c.queries.MarkTransferRecovering(ctx, transfer.ID); err != nil {
		// Already in RECOVERING or different state, continue anyway
	}

	// Increment recovery attempts
	if err := c.queries.IncrementRecoveryAttempt(ctx, transfer.ID); err != nil {
		return err
	}

	// Check ledger state
	entryExists, entryID, err := c.checker.CheckEntry(ctx, transfer.ID.String())
	if err != nil {
		c.logger.Printf("Failed to check ledger for transfer %s: %v", transfer.ID, err)
		// Ledger is unreachable, will retry on next poll
		return err
	}

	if entryExists {
		// Happy path: ledger succeeded but orchestrator crashed before marking completed
		c.logger.Printf("Transfer %s: ledger entry %s exists, marking as completed", transfer.ID, entryID)
		return c.markCompleted(ctx, transfer.ID.String(), entryID)
	}

	// Sad path: ledger never processed the transfer or it was lost
	c.logger.Printf("Transfer %s: no ledger entry found, marking as failed", transfer.ID)
	return c.markFailed(ctx, transfer.ID.String(), "ledger_entry_not_found")
}

// markCompleted marks a transfer as completed
func (c *Compensator) markCompleted(ctx context.Context, transferID, entryID string) error {
	// Parse UUID for the transferID
	transferUUID, err := uuid.Parse(transferID)
	if err != nil {
		return err
	}

	entryUUID, err := uuid.Parse(entryID)
	if err != nil {
		return err
	}

	if err := c.queries.RecordLedgerSuccess(ctx, store.RecordLedgerSuccessParams{
		ID:            transferUUID,
		LedgerEntryID: uuid.NullUUID{UUID: entryUUID, Valid: true},
		LedgerResponse: sql.NullString{
			String: `{"entry_id":"` + entryID + `","recovered":true}`,
			Valid:  true,
		},
	}); err != nil {
		return err
	}

	c.logger.Printf("Transfer %s recovered and marked as COMPLETED", transferID)
	return nil
}

// markFailed marks a transfer as failed
func (c *Compensator) markFailed(ctx context.Context, transferID, reason string) error {
	transferUUID, err := uuid.Parse(transferID)
	if err != nil {
		return err
	}

	if err := c.queries.UpdateTransferFailed(ctx, store.UpdateTransferFailedParams{
		ID:            transferUUID,
		FailureReason: sql.NullString{String: reason, Valid: true},
	}); err != nil {
		return err
	}

	c.logger.Printf("Transfer %s marked as FAILED (reason: %s)", transferID, reason)
	return nil
}
