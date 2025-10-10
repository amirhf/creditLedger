package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestEntry_Validate_ValidEntry(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err != nil {
		t.Errorf("expected valid entry, got error: %v", err)
	}
}

func TestEntry_Validate_MultipleLines(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()
	accountC := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 500, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 500, Side: SideDebit},
			{AccountID: accountC, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err != nil {
		t.Errorf("expected valid entry with multiple lines, got error: %v", err)
	}
}

func TestEntry_Validate_UnbalancedEntry(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 500, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for unbalanced entry")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "lines" {
		t.Errorf("expected field 'lines', got '%s'", validationErr.Field)
	}
}

func TestEntry_Validate_TooFewLines(t *testing.T) {
	accountA := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for single line entry")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "lines" {
		t.Errorf("expected field 'lines', got '%s'", validationErr.Field)
	}
}

func TestEntry_Validate_NegativeAmount(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: -1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for negative amount")
	}

	validationErr, ok := err.(ValidationError)
	if !ok {
		t.Errorf("expected ValidationError, got %T", err)
	}
	if validationErr.Field != "lines[0].amount" {
		t.Errorf("expected field 'lines[0].amount', got '%s'", validationErr.Field)
	}
}

func TestEntry_Validate_ZeroAmount(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 0, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for zero amount")
	}
}

func TestEntry_Validate_InvalidSide(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: "INVALID"},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for invalid side")
	}
}

func TestEntry_Validate_NilAccountID(t *testing.T) {
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: uuid.Nil, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for nil account ID")
	}
}

func TestEntry_Validate_NilEntryID(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.Nil,
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for nil entry ID")
	}
}

func TestEntry_Validate_NilBatchID(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.Nil,
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for nil batch ID")
	}
}

func TestEntry_Validate_NoDebitLines(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideCredit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for entry with no debit lines")
	}
}

func TestEntry_Validate_NoCreditLines(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()

	entry := &Entry{
		EntryID: uuid.New(),
		BatchID: uuid.New(),
		Lines: []Line{
			{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
			{AccountID: accountB, AmountMinor: 1000, Side: SideDebit},
		},
	}

	err := entry.Validate()
	if err == nil {
		t.Error("expected validation error for entry with no credit lines")
	}
}

func TestNewEntry_Success(t *testing.T) {
	accountA := uuid.New()
	accountB := uuid.New()
	batchID := uuid.New()

	lines := []Line{
		{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
		{AccountID: accountB, AmountMinor: 1000, Side: SideCredit},
	}

	entry, err := NewEntry(batchID, lines)
	if err != nil {
		t.Errorf("expected successful entry creation, got error: %v", err)
	}

	if entry.BatchID != batchID {
		t.Errorf("expected batch ID %s, got %s", batchID, entry.BatchID)
	}

	if entry.EntryID == uuid.Nil {
		t.Error("expected entry ID to be generated")
	}

	if entry.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
}

func TestNewEntry_ValidationFailure(t *testing.T) {
	accountA := uuid.New()
	batchID := uuid.New()

	// Invalid: only one line
	lines := []Line{
		{AccountID: accountA, AmountMinor: 1000, Side: SideDebit},
	}

	entry, err := NewEntry(batchID, lines)
	if err == nil {
		t.Error("expected validation error")
	}

	if entry != nil {
		t.Error("expected nil entry on validation failure")
	}
}
