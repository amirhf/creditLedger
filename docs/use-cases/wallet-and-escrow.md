# Use Case: Wallet & Escrow Balances

## Business Problem
A Gig Economy platform (e.g., ride-sharing, food delivery) collects money from a customer, holds it in **Escrow** until the job is done, and then releases it to the freelancer/driver, taking a cut.

**Challenges:**
*   Money must be strictly accounted for (Regulated).
*   Funds in "Escrow" belong to neither the customer nor the platform yet.
*   Complex payouts (Driver + Platform Fee + Tax).

## Modeling in Credit Ledger

### Accounts
*   `Customer Wallet (Alice)`: Alice's pre-loaded balance.
*   `Escrow Account (Job #123)`: Temporary holding account for a specific job.
*   `Driver Wallet (Bob)`: Bob's earnings.
*   `Platform Revenue`: The platform's fee.

### Scenarios

#### 1. Booking: Alice requests a ride ($20)
Money moves to Escrow. Alice can't spend it elsewhere, but Bob doesn't have it yet.
*   **Debit:** `Customer Wallet (Alice)`
*   **Credit:** `Escrow Account (Job #123)`
*   **Amount:** 2000
*   **Currency:** `USD`

#### 2. Completion: Ride finished
Distribute funds from Escrow.
*   **Transaction A (Driver Pay - $16):**
    *   **Debit:** `Escrow Account (Job #123)`
    *   **Credit:** `Driver Wallet (Bob)`
    *   **Amount:** 1600
*   **Transaction B (Platform Fee - $4):**
    *   **Debit:** `Escrow Account (Job #123)`
    *   **Credit:** `Platform Revenue`
    *   **Amount:** 400

#### 3. Cancellation: Alice cancels ride
Refund from Escrow.
*   **Debit:** `Escrow Account (Job #123)`
*   **Credit:** `Customer Wallet (Alice)`
*   **Amount:** 2000

## Implementation Guide

The **Posting Orchestrator** service is designed for this. It can coordinate multi-step sagas (Hold -> Capture/Release).

**API Usage:**
You would typically use the **Orchestrator** endpoints (if exposed) or chain standard ledger transfers.

```bash
# Hold Funds
POST /transfers {
  "from": "ALICE", "to": "ESCROW_JOB_123", "amount": 2000
}

# Release to Driver
POST /transfers {
  "from": "ESCROW_JOB_123", "to": "BOB", "amount": 1600
}

# Capture Fee
POST /transfers {
  "from": "ESCROW_JOB_123", "to": "PLATFORM", "amount": 400
}
```
