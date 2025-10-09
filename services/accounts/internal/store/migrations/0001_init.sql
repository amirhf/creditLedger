CREATE TABLE IF NOT EXISTS accounts (
        id UUID PRIMARY KEY,
        currency TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'ACTIVE',
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);