-- Transfers table for tracking transfer requests
CREATE TABLE IF NOT EXISTS transfers (
    id UUID PRIMARY KEY,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('INITIATED', 'COMPLETED', 'FAILED')),
    entry_id UUID,
    failure_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- SAGA state management columns
    state TEXT NOT NULL DEFAULT 'INITIATED' CHECK (state IN ('INITIATED', 'LEDGER_CALLED', 'RECOVERING', 'COMPENSATING', 'COMPENSATED', 'COMPLETED', 'FAILED')),
    ledger_call_at TIMESTAMPTZ,
    ledger_entry_id UUID,
    ledger_response JSONB,
    compensation_attempts INT NOT NULL DEFAULT 0,
    compensated_at TIMESTAMPTZ,
    recovery_attempts INT NOT NULL DEFAULT 0,
    last_recovery_at TIMESTAMPTZ
);

-- Outbox table for transactional event publishing
CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_transfers_idempotency_key ON transfers(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_transfers_status ON transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_state ON transfers(state);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at);
CREATE INDEX IF NOT EXISTS idx_transfers_stale ON transfers(state, ledger_call_at, recovery_attempts) WHERE state IN ('LEDGER_CALLED', 'RECOVERING');
CREATE INDEX IF NOT EXISTS idx_outbox_unsent ON outbox(created_at) WHERE sent_at IS NULL;
