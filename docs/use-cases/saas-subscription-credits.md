# Use Case: SaaS Subscription & Usage Credits

## Business Problem
A SaaS platform (e.g., AI generation tool, email sender) sells "credits" as part of monthly subscriptions. Users consume credits when they perform actions (generate an image, send an email).

**Challenges:**
*   Users complain credits disappear.
*   Product team wants to give "bonus credits" that expire before "paid credits".
*   Need to audit exactly when and why credits were deducted.

## Modeling in Credit Ledger

We map this to the ledger using a **Liability Account** for the platform (representing unredeemed credits) and **Asset Accounts** for users (representing their right to use the service). *Note: In some models, user credits are liabilities of the platform. Here we treat the user's balance as "Store Value".*

### Accounts
*   `User Account (Alice)`: Holds Alice's credits.
*   `Revenue Account`: Where credits go when "spent" (recognized revenue).
*   `Issuance Account`: The source of new credits (minting).

### Scenarios

#### 1. Monthly Subscription Grant (Alice gets 1000 credits)
*   **Debit:** `Issuance Account`
*   **Credit:** `User Account (Alice)`
*   **Amount:** 1000
*   **Context:** `{"subscription_id": "sub_123", "month": "Nov 2025"}`

#### 2. Usage (Alice generates an image - costs 5 credits)
*   **Debit:** `User Account (Alice)`
*   **Credit:** `Revenue Account`
*   **Amount:** 5
*   **Context:** `{"action": "image_gen", "resource_id": "img_888"}`

#### 3. Refund (Image generation failed)
*   **Debit:** `Revenue Account`
*   **Credit:** `User Account (Alice)`
*   **Amount:** 5
*   **Context:** `{"reason": "generation_failed", "original_tx": "tx_999"}`

## Implementation Guide

**API Calls:**

```bash
# 1. Create Alice's Account
POST /accounts { "ownerId": "alice_user_id", "currency": "CREDITS" }

# 2. Grant Credits (Minting)
POST /transfers {
  "from": "ISSUANCE_ACCOUNT_ID",
  "to": "ALICE_ACCOUNT_ID",
  "amount": 1000,
  "currency": "CREDITS",
  "metadata": { "reason": "subscription_renew" }
}

# 3. Consume Credits (Usage)
POST /transfers {
  "from": "ALICE_ACCOUNT_ID",
  "to": "REVENUE_ACCOUNT_ID",
  "amount": 5,
  "currency": "CREDITS"
}
```
