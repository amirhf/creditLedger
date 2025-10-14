-- Add SAGA state management columns to transfers table
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'INITIATED' CHECK (state IN ('INITIATED', 'LEDGER_CALLED', 'RECOVERING', 'COMPENSATING', 'COMPENSATED', 'COMPLETED', 'FAILED'));
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS ledger_call_at TIMESTAMPTZ;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS ledger_entry_id UUID;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS ledger_response JSONB;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS compensation_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS compensated_at TIMESTAMPTZ;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS recovery_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE transfers ADD COLUMN IF NOT EXISTS last_recovery_at TIMESTAMPTZ;

-- Create indexes for SAGA state management
CREATE INDEX IF NOT EXISTS idx_transfers_state ON transfers(state);
CREATE INDEX IF NOT EXISTS idx_transfers_stale ON transfers(state, ledger_call_at, recovery_attempts) WHERE state IN ('LEDGER_CALLED', 'RECOVERING');
