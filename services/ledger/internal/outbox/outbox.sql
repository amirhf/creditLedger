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
CREATE INDEX IF NOT EXISTS idx_outbox_unsent ON outbox (sent_at);