package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AccountStatus represents the status of an account
type AccountStatus string

const (
	StatusActive    AccountStatus = "ACTIVE"
	StatusSuspended AccountStatus = "SUSPENDED"
)

// Account represents a financial account
type Account struct {
	ID       uuid.UUID
	Currency string
	Status   AccountStatus
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewAccount creates a new account with validation
func NewAccount(currency string) (*Account, error) {
	// Validate currency
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return nil, &ValidationError{Field: "currency", Message: "currency is required"}
	}
	if len(currency) != 3 {
		return nil, &ValidationError{Field: "currency", Message: "currency must be a 3-letter ISO code (e.g., USD, EUR)"}
	}

	return &Account{
		ID:       uuid.New(),
		Currency: currency,
		Status:   StatusActive,
	}, nil
}

// Validate checks if the account is valid
func (a *Account) Validate() error {
	if a.ID == uuid.Nil {
		return &ValidationError{Field: "id", Message: "id cannot be nil"}
	}
	if a.Currency == "" {
		return &ValidationError{Field: "currency", Message: "currency is required"}
	}
	if len(a.Currency) != 3 {
		return &ValidationError{Field: "currency", Message: "currency must be 3 characters"}
	}
	if a.Status != StatusActive && a.Status != StatusSuspended {
		return &ValidationError{Field: "status", Message: "status must be ACTIVE or SUSPENDED"}
	}
	return nil
}
