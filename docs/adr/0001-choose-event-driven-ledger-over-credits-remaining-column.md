# 1. Choose Event-Driven Ledger over "Credits Remaining" Column

Date: 2025-12-01

## Status

Accepted

## Context

For many SaaS applications and marketplaces, the initial requirement is simple: "track how many credits a user has." The obvious solution is a `credits_remaining` integer column in the `users` table.

However, as the system grows, we face several challenges:
1.  **Concurrency:** Two processes (e.g., a subscription renewal and a usage charge) updating the balance simultaneously can lead to race conditions and lost updates.
2.  **Auditability:** When a user asks "why is my balance 37?", we cannot answer without a complete history of transactions.
3.  **Scalability:** Locking the user row for every balance update creates a bottleneck.
4.  **Fragmented Logic:** Business logic for "giving credits" and "spending credits" gets scattered across different services.

## Decision

We will implement an **Event-Driven, Double-Entry Ledger** instead of a simple balance column.

1.  **Double-Entry:** Every movement of value is recorded as a transaction with at least two postings (debit and credit) that must sum to zero. This ensures value is never created or destroyed inadvertently.
2.  **Event-Driven:** The ledger emits events (e.g., `TransferRecorded`) which downstream services consume to update read models (balances).
3.  **Immutable Log:** The source of truth is the append-only log of transactions, not the current balance.

## Consequences

### Positive
*   **Audit Trail:** Complete history of every transaction is preserved by design.
*   **Debuggability:** We can replay the event log to reconstruct the state at any point in time.
*   **Decoupling:** The service calculating the balance (Read Model) is decoupled from the service recording the transaction (Ledger).
*   **Correctness:** Double-entry accounting rules prevents "money from nowhere" bugs.

### Negative
*   **Complexity:** significantly more code and infrastructure (Kafka, separate services) than a single database column.
*   **Eventual Consistency:** The "current balance" seen by the user might lag slightly behind the accepted writes (milliseconds to seconds).
*   **Storage:** Requires more storage space for the transaction log compared to a single mutable value.
