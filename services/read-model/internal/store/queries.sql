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
