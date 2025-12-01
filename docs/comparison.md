# Comparison: Credit Ledger vs. Industry Standards

When building a financial system, you have three main choices:
1.  **Build it yourself** (using a reference architecture like this one).
2.  **Use a specialized database** (TigerBeetle).
3.  **Use a ledger platform** (Formance, Blnk).

Here is how **Credit Ledger** fits into that landscape.

| Feature | **Credit Ledger** (This Repo) | **TigerBeetle** | **Formance / Blnk** |
| :--- | :--- | :--- | :--- |
| **Type** | **Reference Architecture** | Specialized Database | Financial Platform |
| **Primary Language** | Go | Zig | Go |
| **Storage** | Postgres (Ledger) + Kafka | Custom (LSM Tree) | Postgres |
| **Performance** | High (Thousands TPS) | Extreme (Millions TPS) | High |
| **Deployment** | You own the code (Microservices) | Single Binary / SaaS | Single Binary / Cloud |
| **Extensibility** | **Full Control** (Edit the code) | Limited (Scripts/Webhooks) | Configurable (DSL/Scripts) |
| **Learning Curve** | Medium (Requires Go/Kafka knowledge) | Medium (New DB paradigms) | Low/Medium (API driven) |
| **Hiring / Staffing** | **Easy** (Commodity Go/SQL skills) | Niche (Zig / System DBs) | Niche (Platform specific DSLs) |

---

## Strategic Comparison (The "Why" beyond technical specs)

While performance specs (TPS) are interesting, the **Total Cost of Ownership (TCO)** is usually decided by other factors.

### 1. The "Boring Tech" Advantage
**Credit Ledger** relies on the "boring" stack: **Go, Postgres, Kafka**.
*   **Risk:** Low. These technologies are battle-tested for decades.
*   **Hiring:** You can hire almost any backend engineer to maintain this.
*   **Bus Factor:** If your lead engineer leaves, you don't need to find a specialist in a niche financial database to replace them.

### 2. Business Logic Proximity
Real-world ledgers are rarely just "move money." They are "move money *if* X, *minus* tax Y, *triggering* event Z."
*   **Black Box Ledgers:** Force you to implement logic via webhooks (latency) or proprietary scripting languages (complexity).
*   **Credit Ledger:** You own the code. You can inject complex, synchronous business rules (e.g., fraud checks, tax splitting) directly into the Go transaction flow.

---

## When to use **Credit Ledger** (This Architecture)

Choose this if:
*   **You want to own the code.** You need to customize the logic deeply (e.g., complex tax rules, weird recurring billing cycles) and don't want to be limited by a vendor's API or scripting language.
*   **You are already running Go/Kafka/Postgres.** This fits naturally into your existing Kubernetes cluster without introducing a new, exotic database technology.
*   **You are learning.** You want to understand *how* a ledger works under the hood (Outbox pattern, CQRS, Idempotency) rather than treating it as a black box.
*   **"Good Enough" Scale.** You need 100-5,000 TPS, not 1,000,000 TPS. (Most B2B SaaS apps fit here).

## When to use **TigerBeetle**

Choose TigerBeetle if:
*   **Performance is critical.** You are building a high-frequency trading engine, a payment switch, or a core banking system handling millions of transactions.
*   **Safety is paramount.** You want formally verified correctness guarantees that exceed what a standard Postgres setup offers.
*   **You want a "Ledger Database".** You treat the ledger as a specialized infrastructure component, distinct from your application logic.

## When to use **Formance / Blnk**

Choose Formance or Blnk if:
*   **Speed to Market.** You want a "Stripe-like" API for your internal ledger immediately without writing backend code.
*   **Orchestration needed.** You need built-in tools for complex money movements (e.g., "hold here, wait for webhook, then split funds 3 ways").
*   **Managed Service.** You prefer a SaaS solution (Formance Cloud) over hosting it yourself.

---

## Summary

**Credit Ledger** is an **educational and architectural template**. It proves that you can build a robust, correct financial engine using standard open-source tools (Go, Postgres, Kafka) if you apply the right patterns. It is perfect for teams who want full control and ownership of their financial stack.
