package http

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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

// ListAccounts handles GET /v1/accounts
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	currency := r.URL.Query().Get("currency")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Default values
	limit := int32(20)
	offset := int32(0)

	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 && l <= 100 {
			limit = int32(l)
		}
	}

	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	// Get total count
	total, err := h.queries.CountAccounts(ctx, store.CountAccountsParams{
		Column1: currency,
		Column2: status,
	})
	if err != nil {
		h.logger.Printf("Failed to count accounts: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to list accounts")
		return
	}

	// Get accounts
	accounts, err := h.queries.ListAccounts(ctx, store.ListAccountsParams{
		Column1: currency,
		Column2: status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		h.logger.Printf("Failed to list accounts: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to list accounts")
		return
	}

	// Build response
	accountsList := make([]map[string]interface{}, len(accounts))
	for i, acc := range accounts {
		accountsList[i] = map[string]interface{}{
			"id":         acc.ID.String(),
			"currency":   acc.Currency,
			"status":     acc.Status,
			"created_at": acc.CreatedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accountsList,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetAccount handles GET /v1/accounts/:id
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract account ID from URL parameter (chi router)
	accountIDStr := r.PathValue("id")
	if accountIDStr == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_account_id", "Account ID is required")
		return
	}
	
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_account_id", "Account ID must be a valid UUID")
		return
	}

	// Fetch account from database
	account, err := h.queries.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.respondError(w, http.StatusNotFound, "account_not_found", "Account not found")
			return
		}
		h.logger.Printf("Failed to get account: %v", err)
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to get account")
		return
	}

	// Respond with account data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         account.ID.String(),
		"currency":   account.Currency,
		"status":     account.Status,
		"created_at": account.CreatedAt.Format(time.RFC3339),
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
