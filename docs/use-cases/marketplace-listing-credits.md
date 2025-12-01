# Use Case: Marketplace Listing Credits

## Business Problem
A classifieds or freelance marketplace (e.g., Craigslist, Upwork) charges users to "bump" their listings or post premium jobs. Users buy "packs" of credits (e.g., 10 bumps for $20) and spend them over time.

**Challenges:**
*   Need to separate "purchased cash balance" from "promotional credits".
*   High volume of small transactions.
*   Double-spend prevention (bumping the same listing twice simultaneously).

## Modeling in Credit Ledger

We use two different "currencies" or separate accounts to distinguish between Real Money balances and Credit balances.

### Accounts
*   `User Wallet (Bob)`: Holds `USD` (real money).
*   `User Credit Balance (Bob)`: Holds `BUMPS` (virtual currency).
*   `Platform Fees`: Captures USD from credit purchases.
*   `Service Inventory`: Source of BUMPS.

### Scenarios

#### 1. Bob deposits $50
*   **Debit:** `External Payment Gateway (Stripe)`
*   **Credit:** `User Wallet (Bob)`
*   **Currency:** `USD`
*   **Amount:** 5000 (cents)

#### 2. Bob buys "10 Bumps Pack" for $20
This is a **Currency Exchange** (Atomic Swap).
*   **Leg 1:**
    *   **Debit:** `User Wallet (Bob)`
    *   **Credit:** `Platform Fees`
    *   **Currency:** `USD`
    *   **Amount:** 2000
*   **Leg 2:**
    *   **Debit:** `Service Inventory`
    *   **Credit:** `User Credit Balance (Bob)`
    *   **Currency:** `BUMPS`
    *   **Amount:** 10

#### 3. Bob bumps a listing (Costs 1 Bump)
*   **Debit:** `User Credit Balance (Bob)`
*   **Credit:** `Service Inventory` (Burn)
*   **Currency:** `BUMPS`
*   **Amount:** 1

## Implementation Guide

**Idempotency is Key:**
When Bob clicks "Bump", the frontend generates a UUID (`idempotencyKey`). If Bob clicks twice, the Ledger rejects the second request, ensuring he is only charged once.

```bash
# Bob spends 1 BUMP
POST /transfers {
  "from": "BOB_CREDIT_ACCT_ID",
  "to": "INVENTORY_ACCT_ID",
  "amount": 1,
  "currency": "BUMPS",
  "idempotencyKey": "bump-listing-123-action"
}
```
