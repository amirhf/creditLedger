# Day 3 Setup & Testing Guide

## What Was Implemented

Day 3 adds two critical services to the credit ledger system:

1. **Accounts Service** - Manages account creation with event publishing
2. **Posting-Orchestrator Service** - Coordinates transfers with idempotency

## Prerequisites

Ensure the infrastructure is running:
```powershell
docker compose -f deploy/docker-compose.yml up -d
```

This starts:
- PostgreSQL (3 instances on ports 5433, 5434, 5435)
- Redpanda (Kafka) on port 19092
- Redis on port 6379
- Jaeger on port 16686
- Prometheus on port 9090
- Grafana on port 3000

## Running the Services

### 1. Run Ledger Service (from Day 2)

```powershell
cd services/ledger
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7102"
go run ./cmd/ledger
```

### 2. Run Accounts Service (NEW)

```powershell
cd services/accounts
$env:DATABASE_URL="postgres://accounts:accountspw@localhost:5433/accounts?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:PORT="7101"
go run ./cmd/accounts
```

### 3. Run Posting-Orchestrator Service (NEW)

```powershell
cd services/posting-orchestrator
$env:DATABASE_URL="postgres://orchestrator:orchestratorpw@localhost:5435/orchestrator?sslmode=disable"
$env:REDIS_URL="redis://localhost:6379"
$env:KAFKA_BROKERS="localhost:19092"
$env:LEDGER_URL="http://localhost:7102"
$env:PORT="7103"
go run ./cmd/orchestrator
```

## Database Migrations

The services will need their database schemas created. You can run migrations manually:

### Accounts Database
```powershell
docker exec -it postgres-accounts psql -U accounts -d accounts -f /path/to/services/accounts/internal/store/migrations/0001_init.sql
```

Or connect and run:
```sql
-- Connect to accounts database
psql -h localhost -p 5433 -U accounts -d accounts

-- Run the migration SQL from services/accounts/internal/store/migrations/0001_init.sql
```

### Orchestrator Database
```sql
-- Connect to orchestrator database
psql -h localhost -p 5435 -U orchestrator -d orchestrator

-- Run the migration SQL from services/posting-orchestrator/internal/store/migrations/0001_init.sql
```

## Testing

### Automated Test Script

Run the provided PowerShell test script:
```powershell
.\test_day3.ps1
```

This will:
1. Create two accounts (A and B)
2. Execute a transfer from A to B
3. Test idempotency (retry same transfer)
4. Test validation (invalid inputs)

### Manual Testing

#### 1. Create an Account
```powershell
$body = @{
    currency = "USD"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7101/v1/accounts" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

Expected response:
```json
{
  "account_id": "550e8400-e29b-41d4-a716-446655440000",
  "currency": "USD",
  "status": "ACTIVE"
}
```

#### 2. Create a Transfer
```powershell
$transfer = @{
    from_account_id = "550e8400-e29b-41d4-a716-446655440000"
    to_account_id = "660e8400-e29b-41d4-a716-446655440001"
    amount_minor = 1000
    currency = "USD"
    idempotency_key = "unique-key-123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7103/v1/transfers" `
    -Method POST `
    -ContentType "application/json" `
    -Body $transfer
```

Expected response:
```json
{
  "transfer_id": "770e8400-e29b-41d4-a716-446655440002",
  "status": "COMPLETED",
  "entry_id": "880e8400-e29b-41d4-a716-446655440003"
}
```

#### 3. Verify Events in Kafka

Check AccountCreated events:
```bash
docker exec -it redpanda rpk topic consume ledger.account.v1 --num 10
```

Check Transfer events:
```bash
docker exec -it redpanda rpk topic consume ledger.transfer.v1 --num 10
```

Check Entry events:
```bash
docker exec -it redpanda rpk topic consume ledger.entry.v1 --num 10
```

## Architecture Flow

```
Client
  │
  ├─> POST /v1/accounts ──> Accounts Service
  │                           │
  │                           ├─> accounts DB (transactional)
  │                           ├─> outbox table (same tx)
  │                           └─> Kafka: AccountCreated event
  │
  └─> POST /v1/transfers ──> Orchestrator Service
                               │
                               ├─> Check Redis (idempotency)
                               ├─> transfers DB (INITIATED)
                               ├─> outbox: TransferInitiated event
                               │
                               ├─> HTTP POST to Ledger Service
                               │     └─> Creates journal entry (DR/CR)
                               │
                               ├─> Update transfers DB (COMPLETED)
                               └─> outbox: TransferCompleted event
```

## Key Features Demonstrated

### 1. Transactional Outbox Pattern
- Domain write + event in single transaction
- At-least-once delivery guarantee
- No distributed transactions needed

### 2. Idempotency
- **Redis Layer**: Fast duplicate detection with SETNX
- **Database Layer**: Long-term deduplication via unique constraint
- Idempotent responses return original result

### 3. Event-Driven Architecture
- Services emit events for state changes
- Events include metadata (event_id, aggregate_id, schema)
- Consumers can rebuild state from events

### 4. Saga-like Orchestration
- Orchestrator coordinates multi-service transaction
- Emits lifecycle events (Initiated, Completed, Failed)
- Compensating actions on failure

### 5. Graceful Shutdown
- Signal handling (SIGINT, SIGTERM)
- Context cancellation for background workers
- HTTP server shutdown with timeout

## Troubleshooting

### Service won't start - Database connection failed
- Verify PostgreSQL containers are running: `docker ps`
- Check connection string matches docker-compose.yml
- Ensure migrations have been run

### Transfer fails - Ledger service unreachable
- Verify ledger service is running on port 7102
- Check LEDGER_URL environment variable
- Test ledger health: `curl http://localhost:7102/healthz`

### Events not appearing in Kafka
- Check outbox relay worker logs
- Verify Kafka is running: `docker exec -it redpanda rpk cluster info`
- Check outbox table for unsent events:
  ```sql
  SELECT * FROM outbox WHERE sent_at IS NULL;
  ```

### Idempotency not working
- Verify Redis is running: `docker exec -it redis redis-cli ping`
- Check Redis connection in orchestrator logs
- Verify idempotency_key is being sent in request

## Health Checks

All services expose health endpoints:
```powershell
curl http://localhost:7101/healthz  # Accounts
curl http://localhost:7102/healthz  # Ledger
curl http://localhost:7103/healthz  # Orchestrator
```

## Metrics

Prometheus metrics available at:
```
http://localhost:7101/metrics  # Accounts
http://localhost:7102/metrics  # Ledger
http://localhost:7103/metrics  # Orchestrator
```

## Next Steps (Day 4)

- Implement Read-Model service for balance queries
- Add Gateway service for unified API
- Implement statement queries
- Add end-to-end integration tests
