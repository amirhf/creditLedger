package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	ledgerv1 "github.com/amirhf/credit-ledger/proto/gen/go/ledger/v1"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/domain"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/idem"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/resilience"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/store"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/protobuf/proto"
)

// CreateTransferRequest represents the HTTP request body
type CreateTransferRequest struct {
	FromAccountID  string `json:"from_account_id"`
	ToAccountID    string `json:"to_account_id"`
	AmountMinor    int64  `json:"amount_minor"`
	Currency       string `json:"currency"`
	IdempotencyKey string `json:"idempotency_key"`
}

// CreateTransferResponse represents the HTTP response
type CreateTransferResponse struct {
	TransferID string `json:"transfer_id"`
	Status     string `json:"status"`
	EntryID    string `json:"entry_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// LedgerEntryRequest represents the request to ledger service
type LedgerEntryRequest struct {
	BatchID string             `json:"batch_id"`
	Lines   []LedgerLineRequest `json:"lines"`
}

type LedgerLineRequest struct {
	AccountID   string `json:"account_id"`
	AmountMinor int64  `json:"amount_minor"`
	Side        string `json:"side"`
}

// LedgerEntryResponse represents the response from ledger service
type LedgerEntryResponse struct {
	EntryID string `json:"entry_id"`
	BatchID string `json:"batch_id"`
}

// Handler handles HTTP requests for the orchestrator service
type Handler struct {
	db             *sql.DB
	queries        *store.Queries
	idemGuard      *idem.Guard
	ledgerURL      string
	httpClient     *http.Client
	logger         *log.Logger
	circuitBreaker *resilience.CircuitBreaker
	retryConfig    resilience.RetryConfig
}

// NewHandler creates a new HTTP handler
func NewHandler(db *sql.DB, idemGuard *idem.Guard, ledgerURL string, logger *log.Logger) *Handler {
	return &Handler{
		db:        db,
		queries:   store.New(db),
		idemGuard: idemGuard,
		ledgerURL: ledgerURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
		logger:         logger,
		circuitBreaker: resilience.NewCircuitBreaker(resilience.DefaultCircuitBreakerConfig(), logger),
		retryConfig:    resilience.DefaultRetryConfig(),
	}
}

// CreateTransfer handles POST /v1/transfers
func (h *Handler) CreateTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	// Parse UUIDs
	fromAccountID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_from_account_id", "from_account_id must be a valid UUID")
		return
	}

	toAccountID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_to_account_id", "to_account_id must be a valid UUID")
		return
	}

	// Check idempotency - first check database for existing transfer
	existingTransfer, err := h.queries.GetTransferByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil {
		// Transfer already exists, return the existing result
		h.logger.Printf("Idempotent request detected: %s", req.IdempotencyKey)
		h.respondTransfer(w, existingTransfer)
		return
	} else if err != sql.ErrNoRows {
		h.logger.Printf("Failed to check idempotency: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to process transfer")
		return
	}

	// Try to claim idempotency key in Redis (optional, falls back to database)
	claimed, err := h.idemGuard.Claim(ctx, fmt.Sprintf("transfer:%s", req.IdempotencyKey), 5*time.Minute)
	if err != nil {
		// Redis is unavailable, log warning and continue (database idempotency check already passed)
		h.logger.Printf("Redis unavailable for idempotency check (continuing with database fallback): %v", err)
	} else if !claimed {
		// Another request is processing this idempotency key
		h.respondError(w, http.StatusConflict, "duplicate_request", "Transfer with this idempotency key is already being processed")
		return
	}

	// Create transfer with validation
	transfer, err := domain.NewTransfer(fromAccountID, toAccountID, req.AmountMinor, req.Currency, req.IdempotencyKey)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Execute transfer coordination
	if err := h.executeTransfer(ctx, transfer); err != nil {
		h.logger.Printf("Failed to execute transfer: %v", err)
		h.respondError(w, http.StatusInternalServerError, "transfer_failed", err.Error())
		return
	}

	// Fetch the completed transfer
	dbTransfer, err := h.queries.GetTransfer(ctx, transfer.ID)
	if err != nil {
		h.logger.Printf("Failed to fetch transfer: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Transfer completed but failed to fetch result")
		return
	}

	h.respondTransfer(w, dbTransfer)
}

// executeTransfer coordinates the transfer by calling ledger and emitting events
func (h *Handler) executeTransfer(ctx context.Context, transfer *domain.Transfer) error {
	// Start transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Insert transfer record
	now := time.Now()
	_, err = qtx.CreateTransfer(ctx, store.CreateTransferParams{
		ID:             transfer.ID,
		FromAccountID:  transfer.FromAccountID,
		ToAccountID:    transfer.ToAccountID,
		AmountMinor:    transfer.AmountMinor,
		Currency:       transfer.Currency,
		IdempotencyKey: transfer.IdempotencyKey,
		Status:         string(domain.StatusInitiated),
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return fmt.Errorf("failed to create transfer record: %w", err)
	}

	// Create TransferInitiated event
	initiatedEvent := &ledgerv1.TransferInitiated{
		TransferId: transfer.ID.String(),
		From:       transfer.FromAccountID.String(),
		To:         transfer.ToAccountID.String(),
		Amount: &ledgerv1.Money{
			Units:    transfer.AmountMinor,
			Currency: transfer.Currency,
		},
		IdemKey:  transfer.IdempotencyKey,
		TsUnixMs: now.UnixMilli(),
	}

	initiatedPayload, _ := proto.Marshal(initiatedEvent)
	initiatedHeaders := map[string]interface{}{
		"event_name": "TransferInitiated",
		"schema":     "ledger.v1.TransferInitiated",
	}
	initiatedHeadersJSON, _ := json.Marshal(initiatedHeaders)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            uuid.New(),
		AggregateType: "Transfer",
		AggregateID:   transfer.ID,
		EventType:     "TransferInitiated",
		Payload:       initiatedPayload,
		Headers:       initiatedHeadersJSON,
		CreatedAt:     now,
	})
	if err != nil {
		return fmt.Errorf("failed to create initiated event: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Record that we're about to call the ledger (enables compensator recovery)
	ledgerCallTime := time.Now()
	if err := h.queries.RecordLedgerCall(ctx, store.RecordLedgerCallParams{
		ID:           transfer.ID,
		LedgerCallAt: sql.NullTime{Time: ledgerCallTime, Valid: true},
	}); err != nil {
		h.logger.Printf("Failed to record ledger call: %v", err)
		// Continue anyway - this is just for recovery tracking
	}

	// Call ledger service to create journal entry
	entryID, ledgerResponse, err := h.callLedgerService(ctx, transfer)
	if err != nil {
		// Mark transfer as failed
		if err := h.markTransferFailed(ctx, transfer.ID, err.Error()); err != nil {
			h.logger.Printf("Failed to mark transfer as failed: %v", err)
		}
		return fmt.Errorf("ledger service call failed: %w", err)
	}

	// Mark transfer as completed with ledger entry details
	if err := h.markTransferCompletedWithResponse(ctx, transfer.ID, entryID, ledgerResponse); err != nil {
		return fmt.Errorf("failed to mark transfer as completed: %w", err)
	}

	h.logger.Printf("Transfer %s completed with entry %s", transfer.ID, entryID)
	return nil
}

// callLedgerService calls the ledger service to create a journal entry with retry and circuit breaker
func (h *Handler) callLedgerService(ctx context.Context, transfer *domain.Transfer) (uuid.UUID, string, error) {
	var entryID uuid.UUID
	var respBodyStr string
	
	// Wrap the call with circuit breaker
	err := h.circuitBreaker.Execute(func() error {
		// Retry the call with exponential backoff
		return resilience.Retry(ctx, h.retryConfig, func(ctx context.Context) error {
			var err error
			entryID, respBodyStr, err = h.doLedgerCall(ctx, transfer)
			return err
		}, h.logger)
	})
	
	if err != nil {
		// Log circuit breaker state on failure
		state, failures, _ := h.circuitBreaker.GetMetrics()
		h.logger.Printf("Ledger call failed (circuit: %s, failures: %d): %v", state, failures, err)
		return uuid.Nil, respBodyStr, err
	}
	
	return entryID, respBodyStr, nil
}

// doLedgerCall performs the actual HTTP call to ledger service
func (h *Handler) doLedgerCall(ctx context.Context, transfer *domain.Transfer) (uuid.UUID, string, error) {
	// Create ledger entry request with double-entry lines
	batchID := transfer.ID // Use transfer ID as batch ID
	ledgerReq := LedgerEntryRequest{
		BatchID: batchID.String(),
		Lines: []LedgerLineRequest{
			{
				AccountID:   transfer.FromAccountID.String(),
				AmountMinor: transfer.AmountMinor,
				Side:        "DEBIT",
			},
			{
				AccountID:   transfer.ToAccountID.String(),
				AmountMinor: transfer.AmountMinor,
				Side:        "CREDIT",
			},
		},
	}

	reqBody, err := json.Marshal(ledgerReq)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to ledger service
	req, err := http.NewRequestWithContext(ctx, "POST", h.ledgerURL+"/v1/entries", bytes.NewBuffer(reqBody))
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to call ledger service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for logging/debugging
	respBody, _ := io.ReadAll(resp.Body)
	respBodyStr := string(respBody)

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		json.Unmarshal(respBody, &errResp)
		return uuid.Nil, respBodyStr, fmt.Errorf("ledger service returned %d: %s", resp.StatusCode, errResp.Message)
	}

	var ledgerResp LedgerEntryResponse
	if err := json.Unmarshal(respBody, &ledgerResp); err != nil {
		return uuid.Nil, respBodyStr, fmt.Errorf("failed to decode response: %w", err)
	}

	entryID, err := uuid.Parse(ledgerResp.EntryID)
	if err != nil {
		return uuid.Nil, respBodyStr, fmt.Errorf("invalid entry_id in response: %w", err)
	}

	return entryID, respBodyStr, nil
}

// markTransferCompletedWithResponse marks a transfer as completed with ledger response
func (h *Handler) markTransferCompletedWithResponse(ctx context.Context, transferID, entryID uuid.UUID, ledgerResponse string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Update transfer with ledger entry details using RecordLedgerSuccess
	if err := qtx.RecordLedgerSuccess(ctx, store.RecordLedgerSuccessParams{
		ID:             transferID,
		LedgerEntryID:  uuid.NullUUID{UUID: entryID, Valid: true},
		LedgerResponse: sql.NullString{String: ledgerResponse, Valid: true},
	}); err != nil {
		return err
	}

	// Create TransferCompleted event
	completedEvent := &ledgerv1.TransferCompleted{
		TransferId: transferID.String(),
		TsUnixMs:   time.Now().UnixMilli(),
	}

	completedPayload, _ := proto.Marshal(completedEvent)
	completedHeaders := map[string]interface{}{
		"event_name": "TransferCompleted",
		"schema":     "ledger.v1.TransferCompleted",
	}
	completedHeadersJSON, _ := json.Marshal(completedHeaders)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            uuid.New(),
		AggregateType: "Transfer",
		AggregateID:   transferID,
		EventType:     "TransferCompleted",
		Payload:       completedPayload,
		Headers:       completedHeadersJSON,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// markTransferCompleted marks a transfer as completed and emits event (legacy)
func (h *Handler) markTransferCompleted(ctx context.Context, transferID, entryID uuid.UUID) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Update transfer status
	if err := qtx.UpdateTransferCompleted(ctx, store.UpdateTransferCompletedParams{
		ID:      transferID,
		EntryID: uuid.NullUUID{UUID: entryID, Valid: true},
	}); err != nil {
		return err
	}

	// Create TransferCompleted event
	now := time.Now()
	completedEvent := &ledgerv1.TransferCompleted{
		TransferId: transferID.String(),
		TsUnixMs:   now.UnixMilli(),
	}

	completedPayload, _ := proto.Marshal(completedEvent)
	completedHeaders := map[string]interface{}{
		"event_name": "TransferCompleted",
		"schema":     "ledger.v1.TransferCompleted",
	}
	completedHeadersJSON, _ := json.Marshal(completedHeaders)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            uuid.New(),
		AggregateType: "Transfer",
		AggregateID:   transferID,
		EventType:     "TransferCompleted",
		Payload:       completedPayload,
		Headers:       completedHeadersJSON,
		CreatedAt:     now,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// markTransferFailed marks a transfer as failed and emits event
func (h *Handler) markTransferFailed(ctx context.Context, transferID uuid.UUID, reason string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Update transfer status
	if err := qtx.UpdateTransferFailed(ctx, store.UpdateTransferFailedParams{
		ID:            transferID,
		FailureReason: sql.NullString{String: reason, Valid: true},
	}); err != nil {
		return err
	}

	// Create TransferFailed event
	now := time.Now()
	failedEvent := &ledgerv1.TransferFailed{
		TransferId: transferID.String(),
		Reason:     reason,
		TsUnixMs:   now.UnixMilli(),
	}

	failedPayload, _ := proto.Marshal(failedEvent)
	failedHeaders := map[string]interface{}{
		"event_name": "TransferFailed",
		"schema":     "ledger.v1.TransferFailed",
	}
	failedHeadersJSON, _ := json.Marshal(failedHeaders)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            uuid.New(),
		AggregateType: "Transfer",
		AggregateID:   transferID,
		EventType:     "TransferFailed",
		Payload:       failedPayload,
		Headers:       failedHeadersJSON,
		CreatedAt:     now,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

// respondTransfer sends a transfer response
func (h *Handler) respondTransfer(w http.ResponseWriter, transfer store.Transfer) {
	resp := CreateTransferResponse{
		TransferID: transfer.ID.String(),
		Status:     transfer.Status,
	}
	if transfer.EntryID.Valid {
		resp.EntryID = transfer.EntryID.UUID.String()
	}

	w.Header().Set("Content-Type", "application/json")
	if transfer.Status == string(domain.StatusCompleted) {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
	json.NewEncoder(w).Encode(resp)
}

// GetTransfer handles GET /v1/transfers/:id
func (h *Handler) GetTransfer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract transfer ID from URL parameter (chi router)
	transferIDStr := r.PathValue("id")
	if transferIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_transfer_id", "Transfer ID is required")
		return
	}
	
	transferID, err := uuid.Parse(transferIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_transfer_id", "Transfer ID must be a valid UUID")
		return
	}

	// Fetch transfer from database
	transfer, err := h.queries.GetTransfer(ctx, transferID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.respondError(w, http.StatusNotFound, "transfer_not_found", "Transfer not found")
			return
		}
		h.logger.Printf("Failed to get transfer: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to get transfer")
		return
	}

	// Build response
	resp := map[string]interface{}{
		"id":              transfer.ID.String(),
		"from_account_id": transfer.FromAccountID.String(),
		"to_account_id":   transfer.ToAccountID.String(),
		"amount_minor":    transfer.AmountMinor,
		"currency":        transfer.Currency,
		"status":          transfer.Status,
		"idempotency_key": transfer.IdempotencyKey,
		"created_at":      transfer.CreatedAt.Format(time.RFC3339),
	}
	if transfer.EntryID.Valid {
		resp["entry_id"] = transfer.EntryID.UUID.String()
	}
	if transfer.FailureReason.Valid {
		resp["failure_reason"] = transfer.FailureReason.String
	}

	// Respond with transfer data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// respondError sends an error response
func (h *Handler) respondError(w http.ResponseWriter, status int, error string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   error,
		Message: message,
	})
}
