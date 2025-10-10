package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// TransferStatus represents the status of a transfer
type TransferStatus string

const (
	StatusInitiated TransferStatus = "INITIATED"
	StatusCompleted TransferStatus = "COMPLETED"
	StatusFailed    TransferStatus = "FAILED"
)

// Transfer represents a money transfer between accounts
type Transfer struct {
	ID              uuid.UUID
	FromAccountID   uuid.UUID
	ToAccountID     uuid.UUID
	AmountMinor     int64
	Currency        string
	IdempotencyKey  string
	Status          TransferStatus
	EntryID         *uuid.UUID
	FailureReason   *string
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewTransfer creates a new transfer with validation
func NewTransfer(fromAccountID, toAccountID uuid.UUID, amountMinor int64, currency, idempotencyKey string) (*Transfer, error) {
	// Validate amount
	if amountMinor <= 0 {
		return nil, &ValidationError{Field: "amount_minor", Message: "amount must be positive"}
	}

	// Validate currency
	if len(currency) != 3 {
		return nil, &ValidationError{Field: "currency", Message: "currency must be a 3-letter ISO code"}
	}

	// Validate accounts are different
	if fromAccountID == toAccountID {
		return nil, &ValidationError{Field: "accounts", Message: "from and to accounts must be different"}
	}

	// Validate idempotency key
	if idempotencyKey == "" {
		return nil, &ValidationError{Field: "idempotency_key", Message: "idempotency key is required"}
	}

	return &Transfer{
		ID:             uuid.New(),
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		AmountMinor:    amountMinor,
		Currency:       currency,
		IdempotencyKey: idempotencyKey,
		Status:         StatusInitiated,
	}, nil
}

// Validate checks if the transfer is valid
func (t *Transfer) Validate() error {
	if t.ID == uuid.Nil {
		return &ValidationError{Field: "id", Message: "id cannot be nil"}
	}
	if t.FromAccountID == uuid.Nil {
		return &ValidationError{Field: "from_account_id", Message: "from_account_id cannot be nil"}
	}
	if t.ToAccountID == uuid.Nil {
		return &ValidationError{Field: "to_account_id", Message: "to_account_id cannot be nil"}
	}
	if t.FromAccountID == t.ToAccountID {
		return &ValidationError{Field: "accounts", Message: "from and to accounts must be different"}
	}
	if t.AmountMinor <= 0 {
		return &ValidationError{Field: "amount_minor", Message: "amount must be positive"}
	}
	if len(t.Currency) != 3 {
		return &ValidationError{Field: "currency", Message: "currency must be 3 characters"}
	}
	if t.IdempotencyKey == "" {
		return &ValidationError{Field: "idempotency_key", Message: "idempotency key is required"}
	}
	return nil
}
