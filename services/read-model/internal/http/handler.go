package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/amirhf/credit-ledger/services/read-model/internal/metrics"
	"github.com/amirhf/credit-ledger/services/read-model/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler provides HTTP endpoints for querying balances and statements
type Handler struct {
	db      *pgxpool.Pool
	queries *store.Queries
}

// NewHandler creates a new HTTP handler
func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:      db,
		queries: store.New(db),
	}
}

// BalanceResponse represents the account balance response
type BalanceResponse struct {
	AccountID    string `json:"account_id"`
	Currency     string `json:"currency"`
	BalanceMinor int64  `json:"balance_minor"`
	UpdatedAt    string `json:"updated_at"`
}

// StatementEntry represents a single statement line
type StatementEntry struct {
	ID          int64  `json:"id"`
	AccountID   string `json:"account_id"`
	EntryID     string `json:"entry_id"`
	AmountMinor int64  `json:"amount_minor"`
	Side        string `json:"side"`
	Timestamp   string `json:"timestamp"`
}

// StatementsResponse represents the list of statement entries
type StatementsResponse struct {
	Statements []StatementEntry `json:"statements"`
}

// GetBalance handles GET /v1/accounts/:id/balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		metrics.BalanceQueriesTotal.WithLabelValues("invalid_input").Inc()
		http.Error(w, `{"error":"invalid account_id"}`, http.StatusBadRequest)
		return
	}

	var pgAccountID pgtype.UUID
	if err := pgAccountID.Scan(accountID.String()); err != nil {
		http.Error(w, `{"error":"invalid account_id"}`, http.StatusBadRequest)
		return
	}

	balance, err := h.queries.GetBalance(r.Context(), pgAccountID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No balance record yet - return zero balance with currency from query param or default
			currency := r.URL.Query().Get("currency")
			if currency == "" {
				currency = "USD" // Default currency
			}
			
			metrics.BalanceQueriesTotal.WithLabelValues("zero_balance").Inc()
			metrics.QueryDuration.WithLabelValues("balance", "zero_balance").Observe(time.Since(start).Seconds())
			
			resp := BalanceResponse{
				AccountID:    accountID.String(),
				Currency:     currency,
				BalanceMinor: 0,
				UpdatedAt:    time.Now().Format(time.RFC3339),
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		log.Printf("Error getting balance: %v", err)
		metrics.BalanceQueriesTotal.WithLabelValues("error").Inc()
		metrics.QueryDuration.WithLabelValues("balance", "error").Observe(time.Since(start).Seconds())
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	var accountUUID uuid.UUID
	copy(accountUUID[:], balance.AccountID.Bytes[:])

	resp := BalanceResponse{
		AccountID:    accountUUID.String(),
		Currency:     balance.Currency,
		BalanceMinor: balance.BalanceMinor,
		UpdatedAt:    balance.UpdatedAt.Time.Format(time.RFC3339),
	}

	metrics.BalanceQueriesTotal.WithLabelValues("success").Inc()
	metrics.QueryDuration.WithLabelValues("balance", "success").Observe(time.Since(start).Seconds())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetStatements handles GET /v1/accounts/:id/statements
func (h *Handler) GetStatements(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		metrics.StatementQueriesTotal.WithLabelValues("invalid_input").Inc()
		http.Error(w, `{"error":"invalid account_id"}`, http.StatusBadRequest)
		return
	}

	// Parse query parameters for time range
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var statements []store.Statement
	var queryErr error

	var pgAccountID pgtype.UUID
	if err := pgAccountID.Scan(accountID.String()); err != nil {
		http.Error(w, `{"error":"invalid account_id"}`, http.StatusBadRequest)
		return
	}

	if fromStr != "" && toStr != "" {
		// Time-bounded query
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			http.Error(w, `{"error":"invalid from timestamp (use RFC3339)"}`, http.StatusBadRequest)
			return
		}
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			http.Error(w, `{"error":"invalid to timestamp (use RFC3339)"}`, http.StatusBadRequest)
			return
		}

		var pgFrom, pgTo pgtype.Timestamptz
		if err := pgFrom.Scan(from); err != nil {
			http.Error(w, `{"error":"invalid from timestamp"}`, http.StatusBadRequest)
			return
		}
		if err := pgTo.Scan(to); err != nil {
			http.Error(w, `{"error":"invalid to timestamp"}`, http.StatusBadRequest)
			return
		}

		statements, queryErr = h.queries.GetStatements(r.Context(), store.GetStatementsParams{
			AccountID: pgAccountID,
			Ts:        pgFrom,
			Ts_2:      pgTo,
		})
	} else {
		// Default: get last 100 statements
		limit := int32(100)
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			var parsedLimit int
			if _, err := fmt.Sscanf(limitStr, "%d", &parsedLimit); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
				limit = int32(parsedLimit)
			}
		}

		statements, queryErr = h.queries.GetStatementsByAccount(r.Context(), store.GetStatementsByAccountParams{
			AccountID: pgAccountID,
			Limit:     limit,
		})
	}

	if queryErr != nil {
		log.Printf("Error getting statements: %v", queryErr)
		metrics.StatementQueriesTotal.WithLabelValues("error").Inc()
		metrics.QueryDuration.WithLabelValues("statements", "error").Observe(time.Since(start).Seconds())
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	// Convert to response format
	entries := make([]StatementEntry, len(statements))
	for i, stmt := range statements {
		var stmtAccountID, stmtEntryID uuid.UUID
		copy(stmtAccountID[:], stmt.AccountID.Bytes[:])
		copy(stmtEntryID[:], stmt.EntryID.Bytes[:])

		entries[i] = StatementEntry{
			ID:          stmt.ID,
			AccountID:   stmtAccountID.String(),
			EntryID:     stmtEntryID.String(),
			AmountMinor: stmt.AmountMinor,
			Side:        stmt.Side,
			Timestamp:   stmt.Ts.Time.Format(time.RFC3339),
		}
	}

	resp := StatementsResponse{
		Statements: entries,
	}

	metrics.StatementQueriesTotal.WithLabelValues("success").Inc()
	metrics.QueryDuration.WithLabelValues("statements", "success").Observe(time.Since(start).Seconds())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
