-- Journal Entries: Core double-entry ledger table
CREATE TABLE IF NOT EXISTS journal_entries (
    entry_id UUID PRIMARY KEY,
    batch_id UUID NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Journal Lines: Individual debit/credit lines for each entry
CREATE TABLE IF NOT EXISTS journal_lines (
    id BIGSERIAL PRIMARY KEY,
    entry_id UUID NOT NULL REFERENCES journal_entries(entry_id) ON DELETE CASCADE,
    account_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    side TEXT NOT NULL CHECK (side IN ('DEBIT', 'CREDIT'))
);

-- Outbox: Transactional outbox pattern for event publishing
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

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_journal_lines_account ON journal_lines(account_id, entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_lines_entry ON journal_lines(entry_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_batch ON journal_entries(batch_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_ts ON journal_entries(ts);
CREATE INDEX IF NOT EXISTS idx_outbox_unsent ON outbox(created_at) WHERE sent_at IS NULL;
