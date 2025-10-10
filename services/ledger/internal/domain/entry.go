package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Side represents the debit or credit side of a journal line
type Side string

const (
	SideDebit  Side = "DEBIT"
	SideCredit Side = "CREDIT"
)

// Line represents a single debit or credit line in a journal entry
type Line struct {
	AccountID   uuid.UUID
	AmountMinor int64
	Side        Side
}

// Entry represents a complete journal entry with multiple lines
type Entry struct {
	EntryID   uuid.UUID
	BatchID   uuid.UUID
	Lines     []Line
	Timestamp time.Time
}

// ValidationError represents a domain validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks the double-entry accounting invariants
func (e *Entry) Validate() error {
	// Must have at least 2 lines (one debit, one credit minimum)
	if len(e.Lines) < 2 {
		return ValidationError{
			Field:   "lines",
			Message: fmt.Sprintf("entry must have at least 2 lines, got %d", len(e.Lines)),
		}
	}

	var debitSum int64
	var creditSum int64
	var hasDebit, hasCredit bool

	for i, line := range e.Lines {
		// All amounts must be positive
		if line.AmountMinor <= 0 {
			return ValidationError{
				Field:   fmt.Sprintf("lines[%d].amount", i),
				Message: fmt.Sprintf("amount must be positive, got %d", line.AmountMinor),
			}
		}

		// Side must be valid
		if line.Side != SideDebit && line.Side != SideCredit {
			return ValidationError{
				Field:   fmt.Sprintf("lines[%d].side", i),
				Message: fmt.Sprintf("invalid side: %s", line.Side),
			}
		}

		// Account ID must be valid
		if line.AccountID == uuid.Nil {
			return ValidationError{
				Field:   fmt.Sprintf("lines[%d].account_id", i),
				Message: "account_id cannot be nil",
			}
		}

		// Sum debits and credits
		switch line.Side {
		case SideDebit:
			debitSum += line.AmountMinor
			hasDebit = true
		case SideCredit:
			creditSum += line.AmountMinor
			hasCredit = true
		}
	}

	// Must have at least one debit and one credit
	if !hasDebit {
		return ValidationError{
			Field:   "lines",
			Message: "entry must have at least one debit line",
		}
	}
	if !hasCredit {
		return ValidationError{
			Field:   "lines",
			Message: "entry must have at least one credit line",
		}
	}

	// Double-entry invariant: debits must equal credits
	if debitSum != creditSum {
		return ValidationError{
			Field:   "lines",
			Message: fmt.Sprintf("debits (%d) must equal credits (%d)", debitSum, creditSum),
		}
	}

	// Entry ID must be valid
	if e.EntryID == uuid.Nil {
		return ValidationError{
			Field:   "entry_id",
			Message: "entry_id cannot be nil",
		}
	}

	// Batch ID must be valid
	if e.BatchID == uuid.Nil {
		return ValidationError{
			Field:   "batch_id",
			Message: "batch_id cannot be nil",
		}
	}

	return nil
}

// NewEntry creates a new journal entry with validation
func NewEntry(batchID uuid.UUID, lines []Line) (*Entry, error) {
	entry := &Entry{
		EntryID:   uuid.New(),
		BatchID:   batchID,
		Lines:     lines,
		Timestamp: time.Now(),
	}

	if err := entry.Validate(); err != nil {
		return nil, err
	}

	return entry, nil
}
