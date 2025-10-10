-- Transfer Operations

-- name: CreateTransfer :one
INSERT INTO transfers (id, from_account_id, to_account_id, amount_minor, currency, idempotency_key, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers WHERE id = $1 LIMIT 1;

-- name: GetTransferByIdempotencyKey :one
SELECT * FROM transfers WHERE idempotency_key = $1 LIMIT 1;

-- name: UpdateTransferCompleted :exec
UPDATE transfers
SET status = 'COMPLETED', entry_id = $2, updated_at = now()
WHERE id = $1;

-- name: UpdateTransferFailed :exec
UPDATE transfers
SET status = 'FAILED', failure_reason = $2, updated_at = now()
WHERE id = $1;

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
