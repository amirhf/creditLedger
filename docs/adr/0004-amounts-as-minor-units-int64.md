# 4. Amounts as Minor Units (Int64)

Date: 2025-12-01

## Status

Accepted

## Context

Financial systems must be precise. Using floating-point numbers (e.g., `float64` or `double`) for currency introduces rounding errors (e.g., `0.1 + 0.2 != 0.3`) that are unacceptable in accounting.

## Decision

All monetary amounts will be stored and transmitted as **integers representing minor units**.

1.  **Integer Math:** We use `int64` for all internal calculations.
2.  **Minor Units:**
    *   USD: cents (100 cents = $1.00).
    *   Credits: atomic units (1 credit = 1 unit, or 1000 milli-credits if fractional credits are needed).
3.  **No Floats:** We explicitly ban floating-point types for amount fields in Protobuf, Database, and Go structs.

## Consequences

### Positive
*   **Precision:** Zero rounding errors. Addition and subtraction are exact.
*   **Simplicity:** Integer arithmetic is fast and standard across all languages and databases.

### Negative
*   **Display Logic:** The frontend or API Gateway must handle formatting (dividing by 100 for USD) for display purposes.
*   **Overflow:** While `int64` is huge, it theoretically has a limit. (For reference, `int64` max is ~9 quintillion, sufficient for any realistic global money supply in cents).
