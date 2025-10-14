-- Remove void support from journal_entries

-- Drop indexes
DROP INDEX IF EXISTS idx_journal_entries_not_voided;
DROP INDEX IF EXISTS idx_journal_entries_voided;

-- Drop columns
ALTER TABLE journal_entries 
  DROP COLUMN IF EXISTS void_reason,
  DROP COLUMN IF EXISTS voided_at,
  DROP COLUMN IF EXISTS voided_by;
