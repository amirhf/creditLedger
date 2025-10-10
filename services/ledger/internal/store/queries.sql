-- Journal Entry Operations

-- name: CreateJournalEntry :one
INSERT INTO journal_entries (entry_id, batch_id, ts)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetJournalEntry :one
SELECT * FROM journal_entries
WHERE entry_id = $1;

-- name: GetJournalEntriesByBatch :many
SELECT * FROM journal_entries
WHERE batch_id = $1
ORDER BY ts;

-- Journal Line Operations

-- name: CreateJournalLine :one
INSERT INTO journal_lines (entry_id, account_id, amount_minor, side)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetJournalLinesByEntry :many
SELECT * FROM journal_lines
WHERE entry_id = $1
ORDER BY id;

-- name: GetJournalLinesByAccount :many
SELECT jl.*, je.ts
FROM journal_lines jl
JOIN journal_entries je ON jl.entry_id = je.entry_id
WHERE jl.account_id = $1
ORDER BY je.ts DESC
LIMIT $2;

-- Outbox Operations

-- name: CreateOutboxEvent :one
INSERT INTO outbox (id, aggregate_type, aggregate_id, event_type, payload, headers, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUnsentOutboxEvents :many
SELECT * FROM outbox
WHERE sent_at IS NULL
ORDER BY created_at
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEventSent :exec
UPDATE outbox
SET sent_at = now()
WHERE id = $1;

-- name: GetOutboxEvent :one
SELECT * FROM outbox
WHERE id = $1;
