package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/amirhf/credit-ledger/services/ledger/internal/domain"
	"github.com/amirhf/credit-ledger/services/ledger/internal/metrics"
	"github.com/amirhf/credit-ledger/services/ledger/internal/store"
	ledgerv1 "github.com/amirhf/credit-ledger/proto/gen/go/ledger/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// CreateEntryRequest represents the HTTP request body for creating a journal entry
type CreateEntryRequest struct {
	BatchID string         `json:"batch_id"`
	Lines   []LineRequest  `json:"lines"`
}

type LineRequest struct {
	AccountID   string `json:"account_id"`
	AmountMinor int64  `json:"amount_minor"`
	Side        string `json:"side"` // "DEBIT" or "CREDIT"
}

// CreateEntryResponse represents the HTTP response
type CreateEntryResponse struct {
	EntryID string `json:"entry_id"`
	BatchID string `json:"batch_id"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handler handles HTTP requests for the ledger service
type Handler struct {
	db      *sql.DB
	queries *store.Queries
	logger  *log.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(db *sql.DB, logger *log.Logger) *Handler {
	return &Handler{
		db:      db,
		queries: store.New(db),
		logger:  logger,
	}
}

// CreateEntry handles POST /v1/entries
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	start := time.Now()

	// Parse request body
	var req CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	// Validate and parse batch ID
	batchID, err := uuid.Parse(req.BatchID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_batch_id", "Batch ID must be a valid UUID")
		return
	}

	// Convert request lines to domain lines
	lines := make([]domain.Line, len(req.Lines))
	for i, lineReq := range req.Lines {
		accountID, err := uuid.Parse(lineReq.AccountID)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "invalid_account_id", 
				fmt.Sprintf("Line %d: account_id must be a valid UUID", i))
			return
		}

		var side domain.Side
		switch lineReq.Side {
		case "DEBIT":
			side = domain.SideDebit
		case "CREDIT":
			side = domain.SideCredit
		default:
			h.respondError(w, http.StatusBadRequest, "invalid_side", 
				fmt.Sprintf("Line %d: side must be 'DEBIT' or 'CREDIT'", i))
			return
		}

		lines[i] = domain.Line{
			AccountID:   accountID,
			AmountMinor: lineReq.AmountMinor,
			Side:        side,
		}
	}

	// Create and validate entry
	entry, err := domain.NewEntry(batchID, lines)
	if err != nil {
		if validationErr, ok := err.(domain.ValidationError); ok {
			h.respondError(w, http.StatusBadRequest, "validation_error", validationErr.Error())
			return
		}
		h.respondError(w, http.StatusBadRequest, "invalid_entry", err.Error())
		return
	}

	// Start database transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		h.logger.Printf("Failed to begin transaction: %v", err)
		h.respondError(w, http.StatusInternalServerError, "database_error", "Failed to process entry")
		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Insert journal entry
	_, err = qtx.CreateJournalEntry(ctx, store.CreateJournalEntryParams{
		EntryID: entry.EntryID,
		BatchID: entry.BatchID,
		Ts:      entry.Timestamp,
	})
	if err != nil {
		h.logger.Printf("Failed to create journal entry: %v", err)
		h.respondError(w, http.StatusInternalServerError, "database_error", "Failed to create entry")
		return
	}

	// Insert journal lines
	for _, line := range entry.Lines {
		_, err = qtx.CreateJournalLine(ctx, store.CreateJournalLineParams{
			EntryID:     entry.EntryID,
			AccountID:   line.AccountID,
			AmountMinor: line.AmountMinor,
			Side:        string(line.Side),
		})
		if err != nil {
			h.logger.Printf("Failed to create journal line: %v", err)
			h.respondError(w, http.StatusInternalServerError, "database_error", "Failed to create entry lines")
			return
		}
	}

	// Create outbox event (EntryPosted)
	event, err := h.createEntryPostedEvent(entry)
	if err != nil {
		h.logger.Printf("Failed to create event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "event_error", "Failed to create event")
		return
	}

	eventPayload, err := proto.Marshal(event)
	if err != nil {
		h.logger.Printf("Failed to marshal event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "event_error", "Failed to serialize event")
		return
	}

	headers := map[string]interface{}{
		"event_name": "EntryPosted",
		"schema":     "ledger.v1.EntryPosted",
	}
	headersJSON, _ := json.Marshal(headers)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            uuid.New(),
		AggregateType: "journal_entry",
		AggregateID:   entry.EntryID,
		EventType:     "EntryPosted",
		Payload:       eventPayload,
		Headers:       headersJSON,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		h.logger.Printf("Failed to create outbox event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "database_error", "Failed to create outbox event")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		h.logger.Printf("Failed to commit transaction: %v", err)
		metrics.EntryCreationDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
		h.respondError(w, http.StatusInternalServerError, "database_error", "Failed to commit entry")
		return
	}

	// Record metrics
	metrics.EntriesCreated.WithLabelValues("USD").Inc() // TODO: extract currency from entry
	metrics.EntryCreationDuration.WithLabelValues("success").Observe(time.Since(start).Seconds())

	// Return success response
	h.respondJSON(w, http.StatusCreated, CreateEntryResponse{
		EntryID: entry.EntryID.String(),
		BatchID: entry.BatchID.String(),
	})
}

// createEntryPostedEvent converts a domain entry to a protobuf event
func (h *Handler) createEntryPostedEvent(entry *domain.Entry) (*ledgerv1.EntryPosted, error) {
	lines := make([]*ledgerv1.EntryLine, len(entry.Lines))
	for i, line := range entry.Lines {
		var side ledgerv1.Side
		switch line.Side {
		case domain.SideDebit:
			side = ledgerv1.Side_DEBIT
		case domain.SideCredit:
			side = ledgerv1.Side_CREDIT
		default:
			side = ledgerv1.Side_SIDE_UNSPECIFIED
		}

		lines[i] = &ledgerv1.EntryLine{
			AccountId: line.AccountID.String(),
			Amount: &ledgerv1.Money{
				Units:    line.AmountMinor,
				Currency: "USD", // TODO: get from account or entry
			},
			Side: side,
		}
	}

	return &ledgerv1.EntryPosted{
		EntryId:  entry.EntryID.String(),
		BatchId:  entry.BatchID.String(),
		Lines:    lines,
		TsUnixMs: entry.Timestamp.UnixMilli(),
	}, nil
}

// respondJSON sends a JSON response
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, error string, message string) {
	h.respondJSON(w, status, ErrorResponse{
		Error:   error,
		Message: message,
	})
}
