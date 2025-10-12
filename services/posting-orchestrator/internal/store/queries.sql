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
SET status = 'COMPLETED', state = 'COMPLETED', entry_id = $2, updated_at = now()
WHERE id = $1;

-- name: UpdateTransferFailed :exec
UPDATE transfers
SET status = 'FAILED', state = 'FAILED', failure_reason = $2, updated_at = now()
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

-- SAGA State Management Operations

-- name: UpdateTransferState :exec
UPDATE transfers
SET state = $2, updated_at = now()
WHERE id = $1;

-- name: RecordLedgerCall :exec
UPDATE transfers
SET state = 'LEDGER_CALLED', ledger_call_at = $2, updated_at = now()
WHERE id = $1;

-- name: RecordLedgerSuccess :exec
UPDATE transfers
SET state = 'COMPLETED', 
    ledger_entry_id = $2, 
    ledger_response = $3,
    updated_at = now()
WHERE id = $1;

-- name: GetStaleTransfers :many
SELECT * FROM transfers
WHERE state IN ('LEDGER_CALLED', 'RECOVERING')
  AND ledger_call_at < $1
  AND recovery_attempts < 5
ORDER BY ledger_call_at ASC
LIMIT 100;

-- name: IncrementRecoveryAttempt :exec
UPDATE transfers
SET recovery_attempts = recovery_attempts + 1,
    last_recovery_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: MarkTransferRecovering :exec
UPDATE transfers
SET state = 'RECOVERING',
    updated_at = now()
WHERE id = $1 AND state = 'LEDGER_CALLED';

-- name: MarkTransferCompensating :exec
UPDATE transfers
SET state = 'COMPENSATING',
    compensation_attempts = compensation_attempts + 1,
    updated_at = now()
WHERE id = $1;

-- name: MarkTransferCompensated :exec
UPDATE transfers
SET state = 'COMPENSATED',
    compensated_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: GetTransfersByState :many
SELECT * FROM transfers
WHERE state = $1
ORDER BY created_at DESC
LIMIT $2;
