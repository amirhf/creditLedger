-- Balance Queries

-- name: GetBalance :one
SELECT account_id, currency, balance_minor, updated_at
FROM balances
WHERE account_id = $1;

-- name: UpsertBalance :exec
INSERT INTO balances (account_id, currency, balance_minor, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (account_id) DO UPDATE
SET balance_minor = balances.balance_minor + EXCLUDED.balance_minor,
    updated_at = now();

-- name: SetBalance :exec
INSERT INTO balances (account_id, currency, balance_minor, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (account_id) DO UPDATE
SET balance_minor = EXCLUDED.balance_minor,
    updated_at = now();

-- Statement Queries

-- name: CreateStatement :exec
INSERT INTO statements (account_id, entry_id, amount_minor, side, ts)
VALUES ($1, $2, $3, $4, $5);

-- name: GetStatements :many
SELECT id, account_id, entry_id, amount_minor, side, ts
FROM statements
WHERE account_id = $1
  AND ts >= $2
  AND ts <= $3
ORDER BY ts DESC;

-- name: GetStatementsByAccount :many
SELECT id, account_id, entry_id, amount_minor, side, ts
FROM statements
WHERE account_id = $1
ORDER BY ts DESC
LIMIT $2;

-- Event Deduplication Queries

-- name: IsEventProcessed :one
SELECT EXISTS(SELECT 1 FROM event_dedup WHERE event_id = $1);

-- name: MarkEventProcessed :exec
INSERT INTO event_dedup (event_id, processed_at)
VALUES ($1, now())
ON CONFLICT (event_id) DO NOTHING;

-- name: CleanupOldEvents :exec
DELETE FROM event_dedup
WHERE processed_at < $1;

-- Transfer Queries

-- name: CreateTransfer :exec
INSERT INTO transfers (id, from_account_id, to_account_id, amount_minor, currency, status, idempotency_key, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
ON CONFLICT (id) DO NOTHING;

-- name: UpdateTransferStatus :exec
UPDATE transfers
SET status = $2, updated_at = now()
WHERE id = $1;

-- name: GetTransfer :one
SELECT id, from_account_id, to_account_id, amount_minor, currency, status, idempotency_key, created_at, updated_at
FROM transfers
WHERE id = $1;

-- name: ListTransfers :many
SELECT id, from_account_id, to_account_id, amount_minor, currency, status, idempotency_key, created_at, updated_at
FROM transfers
WHERE 
  (sqlc.narg('from_account_id')::uuid IS NULL OR from_account_id = sqlc.narg('from_account_id')) AND
  (sqlc.narg('to_account_id')::uuid IS NULL OR to_account_id = sqlc.narg('to_account_id')) AND
  (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')) AND
  (sqlc.narg('currency')::text IS NULL OR currency = sqlc.narg('currency'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountTransfers :one
SELECT COUNT(*) FROM transfers
WHERE 
  (sqlc.narg('from_account_id')::uuid IS NULL OR from_account_id = sqlc.narg('from_account_id')) AND
  (sqlc.narg('to_account_id')::uuid IS NULL OR to_account_id = sqlc.narg('to_account_id')) AND
  (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')) AND
  (sqlc.narg('currency')::text IS NULL OR currency = sqlc.narg('currency'));
