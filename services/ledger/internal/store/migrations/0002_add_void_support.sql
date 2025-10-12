-- Add void support for journal entries (SAGA compensation)
-- This enables compensating transactions to reverse journal entries

-- Add void tracking columns to journal_entries
ALTER TABLE journal_entries 
  ADD COLUMN voided_by UUID REFERENCES journal_entries(entry_id),
  ADD COLUMN voided_at TIMESTAMPTZ,
  ADD COLUMN void_reason TEXT;

-- Index for finding voided entries (sparse index for performance)
CREATE INDEX idx_journal_entries_voided 
  ON journal_entries(voided_by) 
  WHERE voided_by IS NOT NULL;

-- Index for finding non-voided entries (used in compensation logic)
CREATE INDEX idx_journal_entries_not_voided 
  ON journal_entries(entry_id) 
  WHERE voided_by IS NULL;

-- Add comments for documentation
COMMENT ON COLUMN journal_entries.voided_by IS 'References the entry_id of the void entry that reversed this entry';
COMMENT ON COLUMN journal_entries.voided_at IS 'Timestamp when this entry was voided';
COMMENT ON COLUMN journal_entries.void_reason IS 'Reason for voiding (e.g., transfer_failed, transfer_rollback)';
