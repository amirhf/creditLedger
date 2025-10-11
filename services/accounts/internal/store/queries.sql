-- Account Operations

-- name: CreateAccount :one
INSERT INTO accounts (id, currency, status, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
SELECT * FROM accounts
WHERE ($1::text IS NULL OR currency = $1)
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAccounts :one
SELECT COUNT(*) FROM accounts
WHERE ($1::text IS NULL OR currency = $1)
  AND ($2::text IS NULL OR status = $2);

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
