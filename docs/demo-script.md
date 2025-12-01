# Demo Script: 10-Minute Architecture Walkthrough

**Target Audience:** CTOs, Lead Engineers, or Recruiters.
**Goal:** Demonstrate that this is a production-grade architecture, not just a "todo app" for money.

---

## Setup (Before the call)
1.  Run `make up` (or `make up-nobuild`) to start the stack.
2.  Open 4 tabs in your browser:
    *   **Swagger:** `http://localhost:4000/api`
    *   **Redpanda Console:** `http://localhost:8080`
    *   **Jaeger:** `http://localhost:16686`
    *   **Grafana:** `http://localhost:3000` (Login: admin/admin)

---

## The Script

### 1. Intro & Context (2 mins)
*   **"Hi, today I'm showing you 'Credit Ledger'. It's a reference architecture for handling financial credits, wallets, or usage-based billing at scale."**
*   "Most people start with a simple `credits` column in a database. That breaks when you have concurrency issues or need audit logs."
*   "This system is different. It's **Event-Driven** and uses **Double-Entry Accounting**. It guarantees that money is never created or lost, even if parts of the system fail."

### 2. The "Happy Path" (3 mins)
*   **Action:** Go to **Swagger UI**.
*   **Step 1: Create Accounts.**
    *   Expand `POST /accounts`.
    *   "First, I'll create two accounts: 'Alice' and 'Bob'."
    *   Execute twice. Copy the `accountId`s.
*   **Step 2: Transfer Funds.**
    *   Expand `POST /transfers`.
    *   "Now I'll move $10.00 (1000 cents) from Alice to Bob."
    *   Fill in IDs, amount `1000`, currency `USD`.
    *   **Important:** Highlight the `idempotencyKey` field. "This key ensures safety. If I click this button 5 times, the money only moves ONCE."
    *   Execute the request.
*   **Step 3: Verify Balance.**
    *   Expand `GET /accounts/{id}/balance`.
    *   Check Bob's balance. It should be `1000`.

### 3. The "Under the Hood" (3 mins)
*   "Okay, the API worked. But the magic is what happened in the background."
*   **Action:** Go to **Redpanda Console**.
    *   Click `Topics` -> `ledger.transfer.v1`.
    *   "Here is the event that was published. The Ledger service didn't just update the DB; it emitted this immutable event."
*   **Action:** Go to **Jaeger**.
    *   Select Service: `gateway`. Click "Find Traces".
    *   Click the top trace.
    *   "Look at this distributed trace. You can see the request hit the **Gateway**, then the **Orchestrator**, then the **Ledger**."
    *   "You can see the **Outbox** pattern here—the database write happened, and then the event was relayed asynchronously."

### 4. Idempotency Demo (2 mins)
*   **Action:** Go back to **Swagger**.
*   "Let's try to double-charge Alice."
*   **Step:** Click "Execute" on the `POST /transfers` endpoint *again* (using the SAME `idempotencyKey`).
*   **Result:** "See? It returns `200 OK` (or `201`), but if we check the balance..."
*   **Action:** Check Bob's balance again.
*   **Result:** "It's still `1000`. The system recognized the key and prevented the duplicate transaction."

### 5. Conclusion
*   "This architecture solves the hard problems of fintech: **Concurrency**, **Auditability**, and **Observability**. It's designed to be a template that teams can pick up and use for their own wallet or credit systems."
