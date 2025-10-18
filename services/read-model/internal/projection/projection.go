package projection

import (
	"context"
	"fmt"
	"log"
	"time"

	ledgerv1 "github.com/amirhf/credit-ledger/proto/gen/go/ledger/v1"
	"github.com/amirhf/credit-ledger/services/read-model/internal/store"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

// Projector applies EntryPosted events to balances and statements tables
type Projector struct {
	db      *pgxpool.Pool
	queries *store.Queries
}

// NewProjector creates a new projector instance
func NewProjector(db *pgxpool.Pool) *Projector {
	return &Projector{
		db:      db,
		queries: store.New(db),
	}
}

// ProcessEntryPosted applies an EntryPosted event to the read model
// Returns error if processing fails; idempotent via event_id deduplication
func (p *Projector) ProcessEntryPosted(ctx context.Context, eventID uuid.UUID, payload []byte) error {
	// Check if event already processed (idempotency)
	var pgEventID pgtype.UUID
	if err := pgEventID.Scan(eventID.String()); err != nil {
		return fmt.Errorf("convert event_id to pgtype: %w", err)
	}

	processed, err := p.queries.IsEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("check event processed: %w", err)
	}
	if processed {
		log.Printf("Event %s already processed, skipping", eventID)
		return nil
	}

	// Deserialize event
	var event ledgerv1.EntryPosted
	if err := proto.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal EntryPosted: %w", err)
	}

	// Begin transaction for atomic projection update
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := p.queries.WithTx(tx)

	// Process each line in the entry
	entryID, err := uuid.Parse(event.EntryId)
	if err != nil {
		return fmt.Errorf("parse entry_id: %w", err)
	}

	ts := time.Unix(0, event.TsUnixMs*int64(time.Millisecond))

	for _, line := range event.Lines {
		accountID, err := uuid.Parse(line.AccountId)
		if err != nil {
			return fmt.Errorf("parse account_id: %w", err)
		}

		var pgAccountID pgtype.UUID
		if err := pgAccountID.Scan(accountID.String()); err != nil {
			return fmt.Errorf("convert account_id to pgtype: %w", err)
		}

		var pgEntryID pgtype.UUID
		if err := pgEntryID.Scan(entryID.String()); err != nil {
			return fmt.Errorf("convert entry_id to pgtype: %w", err)
		}

		currency := line.Amount.Currency
		amountMinor := line.Amount.Units

		// Determine balance delta based on side
		// In traditional accounting, for asset accounts:
		// - DEBIT increases the asset (but in transfers, FROM account is debited = money out)
		// - CREDIT decreases the asset (but in transfers, TO account is credited = money in)
		// However, the orchestrator uses: FROM=DEBIT (out), TO=CREDIT (in)
		// So we need to invert: DEBIT = decrease balance, CREDIT = increase balance
		var balanceDelta int64
		if line.Side == ledgerv1.Side_DEBIT {
			balanceDelta = -amountMinor // Money leaving (FROM account)
		} else {
			balanceDelta = amountMinor // Money arriving (TO account)
		}

		// Update balance (UPSERT with delta)
		err = qtx.UpsertBalance(ctx, store.UpsertBalanceParams{
			AccountID:    pgAccountID,
			Currency:     currency,
			BalanceMinor: balanceDelta,
		})
		if err != nil {
			return fmt.Errorf("upsert balance for account %s: %w", accountID, err)
		}

		// Append to statements
		sideStr := "DEBIT"
		if line.Side == ledgerv1.Side_CREDIT {
			sideStr = "CREDIT"
		}

		var pgTs pgtype.Timestamptz
		if err := pgTs.Scan(ts); err != nil {
			return fmt.Errorf("convert timestamp to pgtype: %w", err)
		}

		err = qtx.CreateStatement(ctx, store.CreateStatementParams{
			AccountID:   pgAccountID,
			EntryID:     pgEntryID,
			AmountMinor: amountMinor,
			Side:        sideStr,
			Ts:          pgTs,
		})
		if err != nil {
			return fmt.Errorf("create statement for account %s: %w", accountID, err)
		}
	}

	// Mark event as processed
	err = qtx.MarkEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("Processed EntryPosted event %s with %d lines", eventID, len(event.Lines))
	return nil
}

// ProcessTransferInitiated creates a transfer record in the read model
func (p *Projector) ProcessTransferInitiated(ctx context.Context, eventID uuid.UUID, event *ledgerv1.TransferInitiated) error {
	// Check if event already processed (idempotency)
	var pgEventID pgtype.UUID
	if err := pgEventID.Scan(eventID.String()); err != nil {
		return fmt.Errorf("convert event_id to pgtype: %w", err)
	}

	processed, err := p.queries.IsEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("check event processed: %w", err)
	}
	if processed {
		log.Printf("Transfer event %s already processed, skipping", eventID)
		return nil
	}

	// Parse UUIDs
	transferID, err := uuid.Parse(event.TransferId)
	if err != nil {
		return fmt.Errorf("parse transfer_id: %w", err)
	}

	fromAccountID, err := uuid.Parse(event.From)
	if err != nil {
		return fmt.Errorf("parse from_account_id: %w", err)
	}

	toAccountID, err := uuid.Parse(event.To)
	if err != nil {
		return fmt.Errorf("parse to_account_id: %w", err)
	}

	// Begin transaction
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := p.queries.WithTx(tx)

	// Convert to pgtype
	var pgTransferID, pgFromAccountID, pgToAccountID pgtype.UUID
	if err := pgTransferID.Scan(transferID.String()); err != nil {
		return fmt.Errorf("convert transfer_id: %w", err)
	}
	if err := pgFromAccountID.Scan(fromAccountID.String()); err != nil {
		return fmt.Errorf("convert from_account_id: %w", err)
	}
	if err := pgToAccountID.Scan(toAccountID.String()); err != nil {
		return fmt.Errorf("convert to_account_id: %w", err)
	}

	createdAt := time.Unix(0, event.TsUnixMs*int64(time.Millisecond))
	var pgCreatedAt pgtype.Timestamptz
	if err := pgCreatedAt.Scan(createdAt); err != nil {
		return fmt.Errorf("convert created_at: %w", err)
	}

	// Create transfer record
	err = qtx.CreateTransfer(ctx, store.CreateTransferParams{
		ID:            pgTransferID,
		FromAccountID: pgFromAccountID,
		ToAccountID:   pgToAccountID,
		AmountMinor:   event.Amount.Units,
		Currency:      event.Amount.Currency,
		Status:        "INITIATED",
		IdempotencyKey: event.IdemKey,
		CreatedAt:     pgCreatedAt,
	})
	if err != nil {
		return fmt.Errorf("create transfer: %w", err)
	}

	// Mark event as processed
	err = qtx.MarkEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("Processed TransferInitiated event %s for transfer %s", eventID, transferID)
	return nil
}

// ProcessTransferCompleted updates a transfer status to COMPLETED
func (p *Projector) ProcessTransferCompleted(ctx context.Context, eventID uuid.UUID, event *ledgerv1.TransferCompleted) error {
	// Check if event already processed (idempotency)
	var pgEventID pgtype.UUID
	if err := pgEventID.Scan(eventID.String()); err != nil {
		return fmt.Errorf("convert event_id to pgtype: %w", err)
	}

	processed, err := p.queries.IsEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("check event processed: %w", err)
	}
	if processed {
		log.Printf("Transfer event %s already processed, skipping", eventID)
		return nil
	}

	// Parse transfer ID
	transferID, err := uuid.Parse(event.TransferId)
	if err != nil {
		return fmt.Errorf("parse transfer_id: %w", err)
	}

	// Begin transaction
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := p.queries.WithTx(tx)

	// Convert to pgtype
	var pgTransferID pgtype.UUID
	if err := pgTransferID.Scan(transferID.String()); err != nil {
		return fmt.Errorf("convert transfer_id: %w", err)
	}

	// Update transfer status
	err = qtx.UpdateTransferStatus(ctx, store.UpdateTransferStatusParams{
		ID:     pgTransferID,
		Status: "COMPLETED",
	})
	if err != nil {
		return fmt.Errorf("update transfer status: %w", err)
	}

	// Mark event as processed
	err = qtx.MarkEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("Processed TransferCompleted event %s for transfer %s", eventID, transferID)
	return nil
}

// ProcessTransferFailed updates a transfer status to FAILED
func (p *Projector) ProcessTransferFailed(ctx context.Context, eventID uuid.UUID, event *ledgerv1.TransferFailed) error {
	// Check if event already processed (idempotency)
	var pgEventID pgtype.UUID
	if err := pgEventID.Scan(eventID.String()); err != nil {
		return fmt.Errorf("convert event_id to pgtype: %w", err)
	}

	processed, err := p.queries.IsEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("check event processed: %w", err)
	}
	if processed {
		log.Printf("Transfer event %s already processed, skipping", eventID)
		return nil
	}

	// Parse transfer ID
	transferID, err := uuid.Parse(event.TransferId)
	if err != nil {
		return fmt.Errorf("parse transfer_id: %w", err)
	}

	// Begin transaction
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := p.queries.WithTx(tx)

	// Convert to pgtype
	var pgTransferID pgtype.UUID
	if err := pgTransferID.Scan(transferID.String()); err != nil {
		return fmt.Errorf("convert transfer_id: %w", err)
	}

	// Update transfer status
	err = qtx.UpdateTransferStatus(ctx, store.UpdateTransferStatusParams{
		ID:     pgTransferID,
		Status: "FAILED",
	})
	if err != nil {
		return fmt.Errorf("update transfer status: %w", err)
	}

	// Mark event as processed
	err = qtx.MarkEventProcessed(ctx, pgEventID)
	if err != nil {
		return fmt.Errorf("mark event processed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("Processed TransferFailed event %s for transfer %s", eventID, transferID)
	return nil
}
