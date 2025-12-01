# Use Case: Loyalty & Points System

## Business Problem
A retail brand wants to reward customers with "Points" for purchases. Points can be redeemed for discounts or products.

**Challenges:**
*   Points have an expiry date (e.g., 1 year).
*   Points are liability on the company books until redeemed.
*   High concurrency (Black Friday sales).

## Modeling in Credit Ledger

### Accounts
*   `User Loyalty Account`: Holds the user's points.
*   `Liability Account`: Represents the total outstanding points liability.
*   `Redemption Account`: Where points go to die (expense recognized).

### Scenarios

#### 1. Earn: User buys $100 item (Gets 100 Points)
Company increases its liability.
*   **Debit:** `Liability Account` (Increasing liability)
*   **Credit:** `User Loyalty Account`
*   **Amount:** 100
*   **Currency:** `PTS`
*   **Context:** `{"expiry": "2026-12-01"}`

#### 2. Burn: User redeems 50 Points for a mug
Company decreases liability.
*   **Debit:** `User Loyalty Account`
*   **Credit:** `Redemption Account`
*   **Amount:** 50
*   **Currency:** `PTS`

#### 3. Expiry: Points expire
We simply reverse the issuance (or move to a specific "Expired" revenue account).
*   **Debit:** `User Loyalty Account`
*   **Credit:** `Liability Account` (or `Breakage Income`)
*   **Amount:** 50
*   **Context:** `{"reason": "expired"}`

## Implementation Guide

**Managing Expiry:**
The core Ledger captures the *movements*. A separate "Expiry Worker" service (not included in this core repo, but easy to add) would:
1.  Query the Read Model for points granted > 1 year ago.
2.  Check if they are still unspent (FIFO logic).
3.  Issue a "Expiry Transfer" command to the Ledger.

**FIFO Logic:**
The ledger entries are time-ordered. To implement "First-In-First-Out" usage (oldest points spent first), the Read Model service aggregates the balance. The Ledger just records "Alice spent 50 points". The logic of *which* 50 points were spent is often derived from the history or handled by the logic initiating the transfer.
