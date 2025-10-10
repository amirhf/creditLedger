# Day 4 Testing Guide

## End-to-End Test: Read-Model Projections

This guide walks through testing the complete flow from account creation to balance queries.

### Prerequisites

1. All infrastructure running (Postgres, Kafka, Redis)
2. All services running:
   - Accounts service (port 7101)
   - Ledger service (port 7102)
   - Orchestrator service (port 7103)
   - **Read-model service (port 7104)** ← NEW

### Start Read-Model Service

```powershell
cd services/read-model

# Set environment variables
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5435/readmodel?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7104"

# Run the service
C:\Users\firou\sdk\go1.24.8\bin\go.exe run ./cmd/readmodel
```

### Test Flow

#### Step 1: Create Two Accounts

```powershell
# Create Account A
$accountA = Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" `
  -ContentType "application/json" `
  -Body '{"currency":"USD"}'

Write-Host "Account A: $($accountA.account_id)"

# Create Account B
$accountB = Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" `
  -ContentType "application/json" `
  -Body '{"currency":"USD"}'

Write-Host "Account B: $($accountB.account_id)"
```

#### Step 2: Execute Transfer

```powershell
# Transfer 5000 cents ($50) from A to B
$transfer = Invoke-RestMethod -Method POST -Uri "http://localhost:7103/v1/transfers" `
  -ContentType "application/json" `
  -Body (@{
    from_account_id = $accountA.account_id
    to_account_id = $accountB.account_id
    amount_minor = 5000
    currency = "USD"
    idempotency_key = "test-day4-$(Get-Date -Format 'yyyyMMddHHmmss')"
  } | ConvertTo-Json)

Write-Host "Transfer ID: $($transfer.transfer_id)"
Write-Host "Entry ID: $($transfer.entry_id)"
Write-Host "Status: $($transfer.status)"
```

#### Step 3: Wait for Projection (Consumer Lag)

```powershell
# Wait 2 seconds for Kafka consumer to process events
Start-Sleep -Seconds 2
```

#### Step 4: Query Balances

```powershell
# Get Account A balance (should be -5000)
$balanceA = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountA.account_id)/balance"

Write-Host "`nAccount A Balance:"
Write-Host "  Account ID: $($balanceA.account_id)"
Write-Host "  Currency: $($balanceA.currency)"
Write-Host "  Balance: $($balanceA.balance_minor) cents"
Write-Host "  Updated: $($balanceA.updated_at)"

# Get Account B balance (should be +5000)
$balanceB = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountB.account_id)/balance"

Write-Host "`nAccount B Balance:"
Write-Host "  Account ID: $($balanceB.account_id)"
Write-Host "  Currency: $($balanceB.currency)"
Write-Host "  Balance: $($balanceB.balance_minor) cents"
Write-Host "  Updated: $($balanceB.updated_at)"
```

#### Step 5: Query Statements

```powershell
# Get Account A statements
$statementsA = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountA.account_id)/statements"

Write-Host "`nAccount A Statements:"
foreach ($stmt in $statementsA.statements) {
  Write-Host "  - Entry: $($stmt.entry_id)"
  Write-Host "    Amount: $($stmt.amount_minor) $($stmt.side)"
  Write-Host "    Time: $($stmt.timestamp)"
}

# Get Account B statements
$statementsB = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountB.account_id)/statements"

Write-Host "`nAccount B Statements:"
foreach ($stmt in $statementsB.statements) {
  Write-Host "  - Entry: $($stmt.entry_id)"
  Write-Host "    Amount: $($stmt.amount_minor) $($stmt.side)"
  Write-Host "    Time: $($stmt.timestamp)"
}
```

#### Step 6: Test Idempotency

```powershell
# Execute another transfer
$transfer2 = Invoke-RestMethod -Method POST -Uri "http://localhost:7103/v1/transfers" `
  -ContentType "application/json" `
  -Body (@{
    from_account_id = $accountA.account_id
    to_account_id = $accountB.account_id
    amount_minor = 2500
    currency = "USD"
    idempotency_key = "test-day4-second-$(Get-Date -Format 'yyyyMMddHHmmss')"
  } | ConvertTo-Json)

Write-Host "`nSecond Transfer:"
Write-Host "  Transfer ID: $($transfer2.transfer_id)"
Write-Host "  Entry ID: $($transfer2.entry_id)"

# Wait for projection
Start-Sleep -Seconds 2

# Check updated balances
$balanceA2 = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountA.account_id)/balance"
$balanceB2 = Invoke-RestMethod -Method GET -Uri "http://localhost:7104/v1/accounts/$($accountB.account_id)/balance"

Write-Host "`nUpdated Balances:"
Write-Host "  Account A: $($balanceA2.balance_minor) cents (should be -7500)"
Write-Host "  Account B: $($balanceB2.balance_minor) cents (should be +7500)"
```

### Expected Results

1. **Account A Balance**: `-7500` cents (2 debits: -5000, -2500)
2. **Account B Balance**: `+7500` cents (2 credits: +5000, +2500)
3. **Statements**: Each account should have 2 statement entries
4. **Idempotency**: Re-running the same transfer with the same idempotency key should return the original result without creating duplicate entries

### Verification Checklist

- [ ] Read-model service starts without errors
- [ ] Kafka consumer connects and starts consuming
- [ ] Balances are correctly calculated (DEBIT increases, CREDIT decreases)
- [ ] Statements show all transaction history
- [ ] Multiple transfers update balances correctly
- [ ] Event deduplication works (no duplicate processing)
- [ ] Timestamps are in RFC3339 format
- [ ] UUIDs are properly formatted

### Database Verification

```powershell
# Connect to read-model database
docker exec -it postgres-readmodel psql -U ledger -d readmodel

# Check balances
SELECT account_id, currency, balance_minor, updated_at FROM balances;

# Check statements
SELECT account_id, entry_id, amount_minor, side, ts FROM statements ORDER BY ts DESC;

# Check event deduplication
SELECT COUNT(*) FROM event_dedup;
```

### Kafka Verification

```powershell
# Check consumer group lag
docker exec -it redpanda rpk group describe read-model-projections

# Consume recent events
docker exec -it redpanda rpk topic consume ledger.entry.v1 --num 10
```

### Troubleshooting

**Issue**: Balances not updating
- Check if read-model service is running
- Check Kafka consumer logs for errors
- Verify `ledger.entry.v1` topic has messages
- Check database connection

**Issue**: "Account not found" error
- Wait longer for projection (increase sleep time)
- Check if EntryPosted events are being published
- Verify consumer group is consuming from the beginning

**Issue**: Incorrect balance calculations
- Check DEBIT/CREDIT logic in projection.go
- Verify double-entry invariant in ledger service
- Recompute balances from statements table

### Performance Notes

- Consumer processes events in ~100ms batches
- Projection lag should be < 1 second under normal load
- Event deduplication prevents duplicate processing
- Balances are updated atomically with statements
