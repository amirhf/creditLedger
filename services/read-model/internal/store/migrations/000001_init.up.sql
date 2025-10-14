-- Read-Model Database Schema
-- Purpose: Maintain queryable projections of balances and statements from EntryPosted events

-- Balances table: Current balance per account (UPSERT on event consumption)
CREATE TABLE balances (
    account_id UUID PRIMARY KEY,
    currency TEXT NOT NULL,
    balance_minor BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_balances_currency ON balances(currency);

-- Statements table: Append-only transaction history per account
CREATE TABLE statements (
    id BIGSERIAL PRIMARY KEY,
    account_id UUID NOT NULL,
    entry_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL,
    side TEXT NOT NULL CHECK (side IN ('DEBIT', 'CREDIT')),
    ts TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_statements_account_ts ON statements(account_id, ts DESC);
CREATE INDEX idx_statements_entry ON statements(entry_id);

-- Event deduplication table: Ensures idempotent event processing
CREATE TABLE event_dedup (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_event_dedup_processed_at ON event_dedup(processed_at);
