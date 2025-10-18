-- Add transfers table to read-model for querying transfer history
-- This table is populated by consuming TransferInitiated and TransferCompleted events

CREATE TABLE transfers (
    id UUID PRIMARY KEY,
    from_account_id UUID NOT NULL,
    to_account_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('INITIATED', 'COMPLETED', 'FAILED')),
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for efficient querying
CREATE INDEX idx_transfers_from_account ON transfers(from_account_id, created_at DESC);
CREATE INDEX idx_transfers_to_account ON transfers(to_account_id, created_at DESC);
CREATE INDEX idx_transfers_status ON transfers(status);
CREATE INDEX idx_transfers_currency ON transfers(currency);
CREATE INDEX idx_transfers_created_at ON transfers(created_at DESC);
CREATE INDEX idx_transfers_idempotency_key ON transfers(idempotency_key);
