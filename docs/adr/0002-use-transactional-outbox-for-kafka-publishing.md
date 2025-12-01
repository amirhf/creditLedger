# 2. Use Transactional Outbox for Kafka Publishing

Date: 2025-12-01

## Status

Accepted

## Context

We need to update our local database (e.g., record a transaction) and publish an event to Kafka (e.g., `TransferRecorded`) so other services can react.

If we try to do both independently:
1.  **Dual-Write Problem:** If we write to DB but fail to publish to Kafka, the system is inconsistent. If we publish to Kafka but the DB transaction rolls back, we've sent a phantom event.
2.  **Ordering:** Events might arrive at Kafka in a different order than the database transactions committed.

## Decision

We will use the **Transactional Outbox Pattern**.

1.  **Atomic Write:** When a transaction occurs, we insert the business data (Ledger Entry) AND the event payload into an `outbox` table in the **same** database transaction.
2.  **Guaranteed Delivery:** A separate background process (the "Relay" or "Publisher") polls the `outbox` table and publishes messages to Kafka.
3.  **At-Least-Once:** If the relay crashes after publishing but before marking the outbox entry as "sent", it will resend the message upon restart. Consumers must therefore be idempotent.

## Consequences

### Positive
*   **Consistency:** Guarantees that if a transaction is committed, the corresponding event will eventually be published.
*   **Ordering:** The outbox table preserves the order of events as they occurred in the database.
*   **Resilience:** Kafka downtime does not block the main application flow (writes just queue up in the outbox).

### Negative
*   **Latency:** There is a small delay (polling interval) between the transaction commit and the event appearing in Kafka.
*   **Complexity:** Requires implementing and managing the relay process/worker.
*   **Idempotency Required:** Consumers must handle potential duplicate messages.
