-- Remove SAGA state management columns and indexes
DROP INDEX IF EXISTS idx_transfers_stale;
DROP INDEX IF EXISTS idx_transfers_state;

ALTER TABLE transfers DROP COLUMN IF EXISTS last_recovery_at;
ALTER TABLE transfers DROP COLUMN IF EXISTS recovery_attempts;
ALTER TABLE transfers DROP COLUMN IF EXISTS compensated_at;
ALTER TABLE transfers DROP COLUMN IF EXISTS compensation_attempts;
ALTER TABLE transfers DROP COLUMN IF EXISTS ledger_response;
ALTER TABLE transfers DROP COLUMN IF EXISTS ledger_entry_id;
ALTER TABLE transfers DROP COLUMN IF EXISTS ledger_call_at;
ALTER TABLE transfers DROP COLUMN IF EXISTS state;
