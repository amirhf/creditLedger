# 3. Keep Read Models Separate (CQRS Projections)

Date: 2025-12-01

## Status

Accepted

## Context

In a ledger system, write patterns and read patterns are very different:
*   **Writes:** High-volume, append-only inserts of transaction logs. Complex validation (double-entry constraints).
*   **Reads:** Aggregations (balance calculation), filtering by date/type, and high-concurrency user queries ("how many credits do I have?").

Querying the raw ledger for every balance check ("SELECT SUM(...)") is expensive and locks the write tables, hurting performance.

## Decision

We will apply **Command Query Responsibility Segregation (CQRS)**.

1.  **Write Side (Ledger Service):** Optimized for writing immutable entries. It does *not* maintain the current balance for querying.
2.  **Read Side (Read Model Service):** Consumes events from Kafka and updates a specialized `balances` table (and potentially others) optimized for fast lookups.
3.  **Projections:** The logic that translates a stream of events into a structural database state is called a "Projection."

## Consequences

### Positive
*   **Performance:** Writes and reads can scale independently. Complex queries don't slow down transaction processing.
*   **Flexibility:** We can build multiple read models from the same event stream (e.g., one for user balances, one for daily analytics) without changing the write side.
*   **Replayability:** If we want to change how we calculate balances or add a new view, we can replay the event history to build a new read model.

### Negative
*   **Eventual Consistency:** Users might see a stale balance for a few milliseconds after a transaction.
*   **Complexity:** Requires maintaining two separate data models and the synchronization pipeline (Kafka).
