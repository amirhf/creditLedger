-- Add SAGA state tracking fields to transfers table
-- This enables the compensator to track and recover from partial failures

-- Add state tracking columns
ALTER TABLE transfers 
  ADD COLUMN state VARCHAR(50) DEFAULT 'INITIATED',
  ADD COLUMN ledger_call_at TIMESTAMPTZ,
  ADD COLUMN ledger_entry_id UUID,
  ADD COLUMN ledger_response TEXT,
  ADD COLUMN compensation_attempts INT DEFAULT 0,
  ADD COLUMN compensated_at TIMESTAMPTZ,
  ADD COLUMN recovery_attempts INT DEFAULT 0,
  ADD COLUMN last_recovery_at TIMESTAMPTZ;

-- Migrate existing status to state
UPDATE transfers SET state = status WHERE state = 'INITIATED';

-- Index for compensator queries (find stale transfers)
CREATE INDEX idx_transfers_stale 
  ON transfers(state, ledger_call_at) 
  WHERE state IN ('LEDGER_CALLED', 'RECOVERING');

-- Index for monitoring compensations
CREATE INDEX idx_transfers_compensation 
  ON transfers(state, compensation_attempts)
  WHERE state IN ('COMPENSATING', 'COMPENSATED');

-- Index for recovery attempts
CREATE INDEX idx_transfers_recovery 
  ON transfers(state, recovery_attempts)
  WHERE recovery_attempts > 0;

-- Add comments for documentation
COMMENT ON COLUMN transfers.state IS 'Transfer lifecycle state: INITIATED, LEDGER_CALLED, COMPLETED, RECOVERING, COMPENSATING, COMPENSATED, FAILED';
COMMENT ON COLUMN transfers.ledger_call_at IS 'Timestamp when ledger HTTP call was made';
COMMENT ON COLUMN transfers.ledger_entry_id IS 'Journal entry ID returned by ledger service';
COMMENT ON COLUMN transfers.ledger_response IS 'Full HTTP response from ledger (for debugging)';
COMMENT ON COLUMN transfers.compensation_attempts IS 'Number of times compensation was attempted';
COMMENT ON COLUMN transfers.compensated_at IS 'Timestamp when compensation completed successfully';
COMMENT ON COLUMN transfers.recovery_attempts IS 'Number of times recovery was attempted';
COMMENT ON COLUMN transfers.last_recovery_at IS 'Timestamp of last recovery attempt';
