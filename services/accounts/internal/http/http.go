package http

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	ledgerv1 "github.com/amirhf/credit-ledger/proto/gen/go/ledger/v1"
	"github.com/amirhf/credit-ledger/services/accounts/internal/domain"
	"github.com/amirhf/credit-ledger/services/accounts/internal/store"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// CreateAccountRequest represents the HTTP request body
type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

// CreateAccountResponse represents the HTTP response
type CreateAccountResponse struct {
	AccountID string `json:"account_id"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handler handles HTTP requests for the accounts service
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

// CreateAccount handles POST /v1/accounts
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Failed to parse request body")
		return
	}

	// Create account with validation
	account, err := domain.NewAccount(req.Currency)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Start transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		h.logger.Printf("Failed to begin transaction: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	// Insert account
	now := time.Now()
	dbAccount, err := qtx.CreateAccount(ctx, store.CreateAccountParams{
		ID:        account.ID,
		Currency:  account.Currency,
		Status:    string(account.Status),
		CreatedAt: now,
	})
	if err != nil {
		h.logger.Printf("Failed to create account: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
		return
	}

	// Create AccountCreated event
	event := &ledgerv1.AccountCreated{
		AccountId: account.ID.String(),
		Currency:  account.Currency,
		TsUnixMs:  now.UnixMilli(),
	}

	// Serialize event to protobuf
	payload, err := proto.Marshal(event)
	if err != nil {
		h.logger.Printf("Failed to marshal event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
		return
	}

	// Create outbox event
	eventID := uuid.New()
	headers := map[string]interface{}{
		"event_name": "AccountCreated",
		"schema":     "ledger.v1.AccountCreated",
	}
	headersJSON, _ := json.Marshal(headers)

	_, err = qtx.CreateOutboxEvent(ctx, store.CreateOutboxEventParams{
		ID:            eventID,
		AggregateType: "Account",
		AggregateID:   account.ID,
		EventType:     "AccountCreated",
		Payload:       payload,
		Headers:       headersJSON,
		CreatedAt:     now,
	})
	if err != nil {
		h.logger.Printf("Failed to create outbox event: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		h.logger.Printf("Failed to commit transaction: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create account")
		return
	}

	h.logger.Printf("Created account %s with currency %s", account.ID, account.Currency)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateAccountResponse{
		AccountID: dbAccount.ID.String(),
		Currency:  dbAccount.Currency,
		Status:    dbAccount.Status,
	})
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
