# Day 5 Testing Guide: Observability Stack

**Purpose**: Step-by-step guide to test and verify all observability features implemented in Day 5.

---

## Prerequisites

### Required Tools
- Docker Desktop (running)
- PowerShell or Terminal
- Web browser
- Go 1.24+ installed

### Check Docker Status
```powershell
docker --version
docker compose version
```

---

## Step 1: Start Infrastructure Services

### 1.1 Start All Infrastructure

```powershell
cd C:\Users\firou\GolandProjects\creditLedger\deploy
docker compose up -d
```

**Expected Output**:
```
✔ Container deploy-redis-1                 Started
✔ Container deploy-redpanda-1              Started
✔ Container deploy-postgres-accounts-1     Started
✔ Container deploy-postgres-ledger-1       Started
✔ Container deploy-postgres-readmodel-1    Started
✔ Container deploy-postgres-orchestrator-1 Started
✔ Container deploy-jaeger-1                Started
✔ Container deploy-prometheus-1            Started
✔ Container deploy-grafana-1               Started
✔ Container deploy-console-1               Started
```

### 1.2 Verify Infrastructure

```powershell
# Check all containers are running
docker compose ps

# Should show all services as "running"
```

### 1.3 Verify Web UIs

Open in browser and verify each loads:
- **Jaeger**: http://localhost:16686 ✅
- **Prometheus**: http://localhost:9090 ✅
- **Grafana**: http://localhost:3000 ✅ (login: admin/admin)
- **Redpanda Console**: http://localhost:8080 ✅

---

## Step 2: Start Application Services

Open **4 separate PowerShell terminals** for each service.

### 2.1 Start Ledger Service

**Terminal 1**:
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\ledger

# Set environment variables
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7102"

# Run service
go run ./cmd/ledger
```

**Expected Output**:
```
OpenTelemetry tracer initialized for service: ledger-service
Connected to database
Outbox relay worker started
ledger listening on :7102
```

### 2.2 Start Accounts Service

**Terminal 2**:
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\accounts

# Set environment variables
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5433/accounts?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7101"

# Run service
go run ./cmd/accounts
```

**Expected Output**:
```
OpenTelemetry tracer initialized for service: accounts-service
Connected to database
Outbox relay worker started
accounts listening on :7101
```

### 2.3 Start Orchestrator Service

**Terminal 3**:
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\posting-orchestrator

# Set environment variables
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5436/orchestrator?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:REDIS_URL="redis://localhost:6379"
$env:LEDGER_URL="http://localhost:7102"
$env:PORT="7103"

# Run service
go run ./cmd/orchestrator
```

**Expected Output**:
```
OpenTelemetry tracer initialized for service: orchestrator-service
Connected to database
Outbox relay worker started
orchestrator listening on :7103
```

### 2.4 Start Read-Model Service

**Terminal 4**:
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\read-model

# Set environment variables
$env:OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5435/readmodel?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7104"

# Run service
go run ./cmd/readmodel
```

**Expected Output**:
```
OpenTelemetry tracer initialized for service: read-model-service
Connected to database
Starting Kafka consumer for ledger.entry.v1
read-model listening on :7104
```

### 2.5 Verify All Services Are Running

**New Terminal**:
```powershell
# Check health endpoints
Invoke-RestMethod -Uri "http://localhost:7101/healthz"  # Accounts
Invoke-RestMethod -Uri "http://localhost:7102/healthz"  # Ledger
Invoke-RestMethod -Uri "http://localhost:7103/healthz"  # Orchestrator
Invoke-RestMethod -Uri "http://localhost:7104/healthz"  # Read-Model

# All should return HTTP 200
```

---

## Step 3: Test Metrics Collection

### 3.1 Check Metrics Endpoints

```powershell
# Ledger metrics
Invoke-RestMethod -Uri "http://localhost:7102/metrics" | Select-String "ledger_"

# Accounts metrics
Invoke-RestMethod -Uri "http://localhost:7101/metrics" | Select-String "accounts_"

# Orchestrator metrics
Invoke-RestMethod -Uri "http://localhost:7103/metrics" | Select-String "orchestrator_"

# Read-Model metrics
Invoke-RestMethod -Uri "http://localhost:7104/metrics" | Select-String "readmodel_"
```

**Expected**: Should see metric definitions like:
```
# HELP ledger_entries_created_total Total number of ledger entries created
# TYPE ledger_entries_created_total counter
ledger_entries_created_total{currency="USD"} 0
```

### 3.2 Verify Prometheus Scraping

1. Open http://localhost:9090
2. Go to **Status** → **Targets**
3. Verify all 4 services are listed (if configured in prometheus.yml)
4. Or manually query metrics:
   - Go to **Graph** tab
   - Enter query: `up`
   - Click **Execute**
   - Should show services

---

## Step 4: Generate Test Traffic

### 4.1 Create Test Accounts

```powershell
# Create Account A
$responseA = Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" `
  -ContentType "application/json" `
  -Body '{"currency":"USD"}'

$accountA = $responseA.account_id
Write-Host "Account A: $accountA"

# Create Account B
$responseB = Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" `
  -ContentType "application/json" `
  -Body '{"currency":"USD"}'

$accountB = $responseB.account_id
Write-Host "Account B: $accountB"
```

**Expected Output**:
```
Account A: 123e4567-e89b-12d3-a456-426614174000
Account B: 987fcdeb-51a2-43f7-b890-123456789abc
```

### 4.2 Execute Test Transfers

```powershell
# Execute 10 transfers
for ($i = 1; $i -le 10; $i++) {
    Write-Host "Transfer $i..."
    
    $body = @{
        from_account_id = $accountA
        to_account_id = $accountB
        amount_minor = 1000
        currency = "USD"
        idempotency_key = "test-transfer-$i"
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Method POST `
      -Uri "http://localhost:7103/v1/transfers" `
      -ContentType "application/json" `
      -Body $body

    Write-Host "  Transfer ID: $($response.transfer_id)"
    Start-Sleep -Milliseconds 200
}
```

**Expected Output**:
```
Transfer 1...
  Transfer ID: abc123...
Transfer 2...
  Transfer ID: def456...
...
```

### 4.3 Verify Data Flow

Wait 2-3 seconds for events to propagate, then:

```powershell
# Check Account A balance (should be -10000)
Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountA/balance"

# Check Account B balance (should be +10000)
Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountB/balance"

# Check Account A statement
Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountA/statements?limit=10"
```

**Expected**: Balances should reflect the transfers

---

## Step 5: Test Distributed Tracing

### 5.1 View Traces in Jaeger

1. Open http://localhost:16686
2. In **Service** dropdown, select **orchestrator-service**
3. Click **Find Traces**

**Expected**: Should see traces for your transfers

### 5.2 Inspect a Trace

1. Click on any trace
2. Verify trace spans:
   - `POST /v1/transfers` (orchestrator)
   - `POST /v1/entries` (ledger - HTTP call)
   - `processOutbox` (ledger - outbox relay)
   - `publishEvent` (ledger - Kafka publish)
   - `consume EntryPosted` (read-model - Kafka consume)

3. Check span attributes:
   - `event_id`
   - `kafka.topic`
   - `kafka.partition`
   - `aggregate_id`

### 5.3 Verify Trace Propagation

1. Click on the orchestrator span
2. Note the **Trace ID** (e.g., `abc123def456...`)
3. Click on the ledger span
4. Verify it has the **same Trace ID**
5. Click on the read-model span
6. Verify it **also has the same Trace ID**

**Result**: ✅ Single trace spanning all services!

### 5.4 Test Error Tracing

```powershell
# Try to create an invalid entry (should fail)
$invalidBody = @{
    from_account_id = "invalid-uuid"
    to_account_id = $accountB
    amount_minor = 1000
    currency = "USD"
    idempotency_key = "test-error-1"
} | ConvertTo-Json

try {
    Invoke-RestMethod -Method POST `
      -Uri "http://localhost:7103/v1/transfers" `
      -ContentType "application/json" `
      -Body $invalidBody
} catch {
    Write-Host "Expected error: $($_.Exception.Message)"
}
```

**In Jaeger**:
1. Refresh traces
2. Find the error trace (should have red error indicator)
3. Click to inspect
4. Verify error details are recorded in span

---

## Step 6: Test Prometheus Metrics

### 6.1 Query Basic Metrics

Open http://localhost:9090 and execute these queries:

**Request Rate**:
```promql
rate(ledger_entries_created_total[5m])
```
**Expected**: Should show ~0.033 ops/sec (10 entries in 5 min)

**Entry Creation Latency (P95)**:
```promql
histogram_quantile(0.95, rate(ledger_entry_creation_duration_seconds_bucket{status="success"}[5m]))
```
**Expected**: Should show latency in seconds (e.g., 0.05 = 50ms)

**Events Processed**:
```promql
rate(readmodel_events_processed_total[5m])
```
**Expected**: Should match entry creation rate

**Outbox Queue Size**:
```promql
ledger_outbox_queue_size
```
**Expected**: Should be 0 (all events processed)

### 6.2 Query Error Metrics

```promql
rate(ledger_outbox_events_publish_errors_total[5m])
```
**Expected**: Should be 0 (no errors)

```promql
rate(readmodel_event_processing_errors_total[5m])
```
**Expected**: Should be 0 (no errors)

### 6.3 Query Success Rate

```promql
sum(rate(ledger_entry_creation_duration_seconds_count{status="success"}[5m])) 
/ 
sum(rate(ledger_entry_creation_duration_seconds_count[5m]))
```
**Expected**: Should be 1.0 (100% success rate)

### 6.4 View Metrics Graph

1. Switch to **Graph** tab
2. Enter query: `rate(ledger_entries_created_total[5m])`
3. Click **Execute**
4. Should see a time series graph showing entry creation rate

---

## Step 7: Test Grafana Dashboards

### 7.1 Login to Grafana

1. Open http://localhost:3000
2. Login with:
   - Username: `admin`
   - Password: `admin`
3. Skip password change (or change if desired)

### 7.2 Verify Dashboards Loaded

1. Click **Dashboards** (left sidebar)
2. Click **Browse**
3. Should see 3 dashboards:
   - **Credit Ledger - System Overview**
   - **Credit Ledger - Ledger Service**
   - **Credit Ledger - Read Model Service**

### 7.3 Test System Overview Dashboard

1. Click **Credit Ledger - System Overview**
2. Verify panels show data:
   - **Request Rate (QPS)**: Should show activity
   - **Entry Creation Latency**: Should show P50/P95/P99
   - **Error Rate**: Should be 0
   - **Outbox Queue Size**: Should be 0 or low
   - **Total Entries (1h)**: Should show 10
   - **Success Rate**: Should be ~100%

3. Test auto-refresh:
   - Watch panels update every 10 seconds
   - Generate more traffic and see metrics change

### 7.4 Test Ledger Service Dashboard

1. Navigate to **Credit Ledger - Ledger Service**
2. Verify panels:
   - **Entries Created Rate**: Should show activity
   - **Entry Creation Latency**: P50/P95/P99 values
   - **Outbox Events Published**: Should match entries
   - **Outbox Publish Errors**: Should be 0
   - **Outbox Relay Latency**: Should be low (< 100ms)
   - **Total Entries (24h)**: Should show 10

### 7.5 Test Read-Model Service Dashboard

1. Navigate to **Credit Ledger - Read Model Service**
2. Verify panels:
   - **Events Processed Rate**: Should match entry rate
   - **Event Processing Latency**: Should be low
   - **Event Processing Errors**: Should be 0
   - **Projection Lag**: Should be low (< 1s)
   - **Query Rate**: Generate queries to see activity
   - **Query Latency**: Should be very low (< 10ms)

### 7.6 Test Time Range Selection

1. Click time picker (top-right)
2. Select **Last 5 minutes**
3. Verify data updates
4. Try **Last 15 minutes**
5. Try **Last 1 hour**

### 7.7 Test Dashboard Refresh

1. Click refresh icon (top-right)
2. Select **5s** refresh interval
3. Generate traffic and watch real-time updates
4. Change back to **10s** when done

---

## Step 8: Load Testing

### 8.1 Generate High Load

```powershell
# Create 100 transfers rapidly
for ($i = 1; $i -le 100; $i++) {
    $body = @{
        from_account_id = $accountA
        to_account_id = $accountB
        amount_minor = 100
        currency = "USD"
        idempotency_key = "load-test-$i"
    } | ConvertTo-Json

    Start-Job -ScriptBlock {
        param($uri, $body)
        Invoke-RestMethod -Method POST -Uri $uri -ContentType "application/json" -Body $body
    } -ArgumentList "http://localhost:7103/v1/transfers", $body
}

# Wait for jobs to complete
Get-Job | Wait-Job
Get-Job | Receive-Job
Get-Job | Remove-Job
```

### 8.2 Monitor During Load

**In Grafana**:
1. Watch **Request Rate** spike
2. Observe **Latency** percentiles
3. Check **Error Rate** stays at 0
4. Monitor **Outbox Queue Size** (may temporarily increase)
5. Verify **Success Rate** stays high

**In Jaeger**:
1. Refresh traces
2. Should see many traces
3. Click on traces to verify structure
4. Check for any errors (red indicators)

**In Prometheus**:
1. Query: `rate(ledger_entries_created_total[1m])`
2. Should show high rate during load test
3. Query: `ledger_outbox_queue_size`
4. May temporarily increase, then return to 0

### 8.3 Verify Final State

```powershell
# Check final balances
Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountA/balance"
Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountB/balance"

# Account A should be -11000 (10 + 100 transfers)
# Account B should be +11000
```

---

## Step 9: Test Idempotency

### 9.1 Test Duplicate Request Detection

```powershell
# Send same request twice
$body = @{
    from_account_id = $accountA
    to_account_id = $accountB
    amount_minor = 5000
    currency = "USD"
    idempotency_key = "idempotency-test-1"
} | ConvertTo-Json

# First request
$response1 = Invoke-RestMethod -Method POST `
  -Uri "http://localhost:7103/v1/transfers" `
  -ContentType "application/json" `
  -Body $body

Write-Host "First request - Transfer ID: $($response1.transfer_id)"

# Second request (same idempotency key)
$response2 = Invoke-RestMethod -Method POST `
  -Uri "http://localhost:7103/v1/transfers" `
  -ContentType "application/json" `
  -Body $body

Write-Host "Second request - Transfer ID: $($response2.transfer_id)"

# Should return same transfer_id
if ($response1.transfer_id -eq $response2.transfer_id) {
    Write-Host "✅ Idempotency working correctly!"
} else {
    Write-Host "❌ Idempotency failed - different transfer IDs"
}
```

### 9.2 Check Idempotency Metrics

**In Prometheus**:
```promql
orchestrator_idempotency_hits_total
```
**Expected**: Should show 1 hit

**In Grafana**:
- Check if you added an idempotency panel
- Should show the duplicate detection

---

## Step 10: Test Error Scenarios

### 10.1 Test Invalid Input

```powershell
# Invalid UUID
try {
    $body = @{
        from_account_id = "not-a-uuid"
        to_account_id = $accountB
        amount_minor = 1000
        currency = "USD"
        idempotency_key = "error-test-1"
    } | ConvertTo-Json

    Invoke-RestMethod -Method POST `
      -Uri "http://localhost:7103/v1/transfers" `
      -ContentType "application/json" `
      -Body $body
} catch {
    Write-Host "✅ Expected error: $($_.Exception.Response.StatusCode)"
}
```

### 10.2 Test Service Unavailability

```powershell
# Stop ledger service (Ctrl+C in Terminal 1)
# Wait a moment, then try a transfer

try {
    $body = @{
        from_account_id = $accountA
        to_account_id = $accountB
        amount_minor = 1000
        currency = "USD"
        idempotency_key = "error-test-2"
    } | ConvertTo-Json

    Invoke-RestMethod -Method POST `
      -Uri "http://localhost:7103/v1/transfers" `
      -ContentType "application/json" `
      -Body $body
} catch {
    Write-Host "✅ Expected error: Service unavailable"
}

# Restart ledger service (re-run command from Step 2.1)
```

### 10.3 Verify Error Metrics

**In Prometheus**:
```promql
rate(orchestrator_ledger_call_duration_seconds_count{status="error"}[5m])
```
**Expected**: Should show the failed calls

**In Jaeger**:
- Find traces with errors (red indicators)
- Inspect error details in spans

---

## Step 11: Verification Checklist

### ✅ Distributed Tracing
- [ ] Traces visible in Jaeger
- [ ] Single trace spans multiple services
- [ ] Trace ID propagates through HTTP
- [ ] Trace ID propagates through Kafka
- [ ] Span attributes present (event_id, topic, etc.)
- [ ] Errors recorded in spans

### ✅ Metrics Collection
- [ ] All services expose `/metrics` endpoints
- [ ] Metrics visible in Prometheus
- [ ] Business metrics incrementing (entries, transfers, events)
- [ ] Latency histograms recording data
- [ ] Error counters working
- [ ] Success rates calculated correctly

### ✅ Grafana Dashboards
- [ ] All 3 dashboards load successfully
- [ ] Panels show real data
- [ ] Auto-refresh working (10s)
- [ ] Time range selection working
- [ ] Threshold colors correct (green/yellow/red)
- [ ] Gauges show current values
- [ ] Time series graphs show trends

### ✅ System Behavior
- [ ] Entries created successfully
- [ ] Events published to Kafka
- [ ] Events consumed by read-model
- [ ] Balances updated correctly
- [ ] Statements recorded
- [ ] Idempotency working
- [ ] Errors handled gracefully

---

## Step 12: Cleanup

### 12.1 Stop Application Services

In each terminal running a service:
- Press **Ctrl+C** to stop

### 12.2 Stop Infrastructure

```powershell
cd C:\Users\firou\GolandProjects\creditLedger\deploy
docker compose down
```

### 12.3 Clean Up Data (Optional)

```powershell
# Remove volumes (deletes all data)
docker compose down -v
```

---

## Troubleshooting

### Issue: Metrics Not Showing in Prometheus

**Solution**:
1. Check service `/metrics` endpoints are accessible
2. Verify Prometheus configuration in `deploy/prometheus/prometheus.yml`
3. Check Prometheus targets: http://localhost:9090/targets
4. Ensure services are running and generating metrics

### Issue: Dashboards Show "No Data"

**Solution**:
1. Verify Prometheus is running: http://localhost:9090
2. Check datasource in Grafana: Configuration → Data Sources
3. Ensure services are generating traffic
4. Check time range (may need to adjust to "Last 5 minutes")

### Issue: Traces Not Appearing in Jaeger

**Solution**:
1. Verify Jaeger is running: http://localhost:16686
2. Check OTEL_EXPORTER_OTLP_ENDPOINT is set correctly
3. Ensure services are running with OTEL enabled
4. Generate traffic to create traces
5. Check service logs for OTEL initialization messages

### Issue: Services Won't Start

**Solution**:
1. Check database connections (verify PostgreSQL containers running)
2. Check Kafka connection (verify Redpanda container running)
3. Check port conflicts (ensure ports 7101-7104 are free)
4. Review service logs for specific errors

---

## Success Criteria

### ✅ You've successfully tested everything if:

1. **All services start without errors**
2. **Transfers execute successfully**
3. **Balances update correctly**
4. **Traces appear in Jaeger with correct structure**
5. **Metrics are collected and visible in Prometheus**
6. **Grafana dashboards show real-time data**
7. **Error scenarios are handled gracefully**
8. **Idempotency prevents duplicate processing**
9. **Load testing shows system handles concurrent requests**
10. **All panels in all dashboards display data**

---

## Next Steps

After successful testing:

1. **Document any issues found**
2. **Tune metric buckets if needed**
3. **Add more dashboard panels as needed**
4. **Configure alerting rules**
5. **Proceed to Day 6: Gateway Service**

---

## Quick Test Script

Save this as `test-observability.ps1`:

```powershell
# Quick observability test script
Write-Host "=== Credit Ledger Observability Test ===" -ForegroundColor Cyan

# Create accounts
Write-Host "`n1. Creating test accounts..." -ForegroundColor Yellow
$accountA = (Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" -ContentType "application/json" -Body '{"currency":"USD"}').account_id
$accountB = (Invoke-RestMethod -Method POST -Uri "http://localhost:7101/v1/accounts" -ContentType "application/json" -Body '{"currency":"USD"}').account_id
Write-Host "   Account A: $accountA" -ForegroundColor Green
Write-Host "   Account B: $accountB" -ForegroundColor Green

# Execute transfers
Write-Host "`n2. Executing 5 test transfers..." -ForegroundColor Yellow
for ($i = 1; $i -le 5; $i++) {
    $body = @{
        from_account_id = $accountA
        to_account_id = $accountB
        amount_minor = 1000
        currency = "USD"
        idempotency_key = "quick-test-$i"
    } | ConvertTo-Json
    
    Invoke-RestMethod -Method POST -Uri "http://localhost:7103/v1/transfers" -ContentType "application/json" -Body $body | Out-Null
    Write-Host "   Transfer $i completed" -ForegroundColor Green
    Start-Sleep -Milliseconds 200
}

# Wait for processing
Write-Host "`n3. Waiting for event processing..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

# Check balances
Write-Host "`n4. Verifying balances..." -ForegroundColor Yellow
$balanceA = Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountA/balance"
$balanceB = Invoke-RestMethod -Uri "http://localhost:7104/v1/accounts/$accountB/balance"
Write-Host "   Account A balance: $($balanceA.balance_minor)" -ForegroundColor Green
Write-Host "   Account B balance: $($balanceB.balance_minor)" -ForegroundColor Green

# Check observability
Write-Host "`n5. Observability endpoints:" -ForegroundColor Yellow
Write-Host "   Jaeger:     http://localhost:16686" -ForegroundColor Cyan
Write-Host "   Prometheus: http://localhost:9090" -ForegroundColor Cyan
Write-Host "   Grafana:    http://localhost:3000 (admin/admin)" -ForegroundColor Cyan

Write-Host "`n✅ Test completed successfully!" -ForegroundColor Green
Write-Host "   Check the observability UIs to see traces, metrics, and dashboards." -ForegroundColor White
```

Run with:
```powershell
.\test-observability.ps1
```

---

**Happy Testing! 🎉**
