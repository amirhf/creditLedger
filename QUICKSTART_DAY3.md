# Quick Start - Day 3 Services

## 🚀 Start All Services (3 Terminal Windows)

### Terminal 1: Ledger Service
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\ledger
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
go run ./cmd/ledger
```

### Terminal 2: Accounts Service
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\accounts
$env:DATABASE_URL="postgres://accounts:accountspw@localhost:5433/accounts?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
go run ./cmd/accounts
```

### Terminal 3: Orchestrator Service
```powershell
cd C:\Users\firou\GolandProjects\creditLedger\services\posting-orchestrator
$env:DATABASE_URL="postgres://orchestrator:orchestratorpw@localhost:5435/orchestrator?sslmode=disable"
$env:REDIS_URL="redis://localhost:6379"
$env:KAFKA_BROKERS="localhost:19092"
$env:LEDGER_URL="http://localhost:7102"
go run ./cmd/orchestrator
```

---

## 🧪 Quick Test

```powershell
# 1. Create Account A
$accountA = Invoke-RestMethod -Uri "http://localhost:7101/v1/accounts" -Method POST -ContentType "application/json" -Body '{"currency":"USD"}'
$accountAId = $accountA.account_id

# 2. Create Account B
$accountB = Invoke-RestMethod -Uri "http://localhost:7101/v1/accounts" -Method POST -ContentType "application/json" -Body '{"currency":"USD"}'
$accountBId = $accountB.account_id

# 3. Transfer 1000 cents from A to B
$transfer = @{
    from_account_id = $accountAId
    to_account_id = $accountBId
    amount_minor = 1000
    currency = "USD"
    idempotency_key = "test-$(Get-Date -Format 'yyyyMMddHHmmss')"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:7103/v1/transfers" -Method POST -ContentType "application/json" -Body $transfer
```

---

## 📊 Verify Events

```bash
# Check AccountCreated events
docker exec -it redpanda rpk topic consume ledger.account.v1 --num 5

# Check Transfer events
docker exec -it redpanda rpk topic consume ledger.transfer.v1 --num 5

# Check Entry events
docker exec -it redpanda rpk topic consume ledger.entry.v1 --num 5
```

---

## 🔍 Health Checks

```powershell
curl http://localhost:7101/healthz  # Accounts
curl http://localhost:7102/healthz  # Ledger
curl http://localhost:7103/healthz  # Orchestrator
```

---

## 📝 Service Ports

| Service | Port | Endpoint |
|---------|------|----------|
| Accounts | 7101 | POST /v1/accounts |
| Ledger | 7102 | POST /v1/entries |
| Orchestrator | 7103 | POST /v1/transfers |

---

## 🗄️ Database Ports

| Database | Port | Credentials |
|----------|------|-------------|
| accounts | 5433 | accounts/accountspw |
| ledger | 5434 | ledger/ledgerpw |
| orchestrator | 5435 | orchestrator/orchestratorpw |

---

## ⚡ Full Test Script

```powershell
.\test_day3.ps1
```

This runs comprehensive tests including validation and idempotency checks.

---

## 🛠️ Troubleshooting

**Services won't start?**
```powershell
# Start infrastructure
docker compose -f deploy/docker-compose.yml up -d

# Verify all containers running
docker ps
```

**Database connection errors?**
```powershell
# Check PostgreSQL
docker exec -it postgres-accounts psql -U accounts -d accounts -c "SELECT 1"
docker exec -it postgres-ledger psql -U ledger -d ledger -c "SELECT 1"
docker exec -it postgres-orchestrator psql -U orchestrator -d orchestrator -c "SELECT 1"
```

**Redis connection errors?**
```powershell
docker exec -it redis redis-cli ping
```

**Kafka connection errors?**
```powershell
docker exec -it redpanda rpk cluster info
```

---

## 📚 Documentation

- **DAY3_SETUP.md** - Detailed setup and testing guide
- **DAY3_SUMMARY.md** - Complete implementation summary
- **design.md** - Full system design and progress log
