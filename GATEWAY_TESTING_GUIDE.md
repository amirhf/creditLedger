# Gateway Visual Testing Guide

This guide shows you how to visually test the Gateway API using different tools.

---

## Prerequisites

### 1. Start Infrastructure
```powershell
cd deploy
docker compose up -d
```

### 2. Start All Go Services

**Terminal 1 - Accounts Service**:
```powershell
cd services/accounts
$env:DATABASE_URL="postgres://accounts:accountspw@localhost:5433/accounts?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
C:\Users\firou\sdk\go1.24.8\bin\go.exe run ./cmd/accounts
```

**Terminal 2 - Ledger Service**:
```powershell
cd services/ledger
$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
C:\Users\firou\sdk\go1.24.8\bin\go.exe run ./cmd/ledger
```

**Terminal 3 - Orchestrator Service**:
```powershell
cd services/posting-orchestrator
$env:DATABASE_URL="postgres://orchestrator:orchestratorpw@localhost:5436/orchestrator?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
$env:REDIS_URL="localhost:6379"
$env:LEDGER_URL="http://localhost:7102"
C:\Users\firou\sdk\go1.24.8\bin\go.exe run ./cmd/orchestrator
```

**Terminal 4 - Read-Model Service**:
```powershell
cd services/read-model
$env:DATABASE_URL="postgres://readmodel:readmodelpw@localhost:5435/readmodel?sslmode=disable"
$env:KAFKA_BROKERS="localhost:19092"
C:\Users\firou\sdk\go1.24.8\bin\go.exe run ./cmd/readmodel
```

### 3. Start Gateway Service

**Terminal 5 - Gateway**:
```powershell
cd services/gateway
npm install  # First time only
npm run start:dev
```

Wait for: `Gateway service listening on port 4000`

---

## Method 1: Swagger UI (Recommended for Beginners) 🌟

### Access Swagger UI
1. Open browser: **http://localhost:4000/api**
2. You'll see an interactive API documentation page

### How to Use Swagger UI

#### Step 1: Create Account A
1. Find **POST /accounts** endpoint
2. Click **"Try it out"**
3. Edit the request body:
   ```json
   {
     "currency": "USD"
   }
   ```
4. Click **"Execute"**
5. **Copy the `account_id`** from the response (you'll need it later)

#### Step 2: Create Account B
1. Repeat the same process
2. **Copy this `account_id`** too

#### Step 3: Create Transfer
1. Find **POST /transfers** endpoint
2. Click **"Try it out"**
3. Edit the request body (paste your account IDs):
   ```json
   {
     "from_account_id": "PASTE_ACCOUNT_A_ID_HERE",
     "to_account_id": "PASTE_ACCOUNT_B_ID_HERE",
     "amount_minor": 5000,
     "currency": "USD",
     "idempotency_key": "my-first-transfer"
   }
   ```
4. Click **"Execute"**
5. **Copy the `transfer_id`**

#### Step 4: Check Balance
1. Wait 2 seconds (for projection to update)
2. Find **GET /accounts/{id}/balance** endpoint
3. Click **"Try it out"**
4. Paste **Account B ID** in the `id` field
5. Click **"Execute"**
6. You should see `balance_minor: 5000`!

#### Step 5: View Statements
1. Find **GET /accounts/{id}/statements** endpoint
2. Click **"Try it out"**
3. Paste **Account A ID**
4. Set `limit` to `10`
5. Click **"Execute"**
6. You'll see the transaction history!

---

## Method 2: VS Code REST Client Extension

### Install Extension
1. Open VS Code
2. Go to Extensions (Ctrl+Shift+X)
3. Search for **"REST Client"** by Huachao Mao
4. Click Install

### Use the Test File
1. Open `gateway-api-tests.http` in VS Code
2. You'll see **"Send Request"** links above each request
3. Click **"Send Request"** to execute
4. Response appears in a new tab

### Quick Test Flow
1. Click "Send Request" on **Test 1: Create Account A**
2. Copy the `account_id` from response
3. Replace `@accountA = PASTE_ACCOUNT_ID_HERE` with actual ID
4. Repeat for Account B
5. Continue clicking "Send Request" for each test

**Pro Tip**: Results appear in a split view with syntax highlighting!

---

## Method 3: Postman

### Import Collection
1. Open Postman
2. Click **Import**
3. Select `Gateway-API.postman_collection.json`
4. Collection appears in left sidebar

### Run Tests
1. Click on **"Credit Ledger Gateway API"** collection
2. Click **"Run"** button (top right)
3. Select all requests
4. Click **"Run Credit Ledger Gateway API"**
5. Watch tests execute automatically!

**Features**:
- ✅ Auto-saves account IDs and transfer IDs to variables
- ✅ Automated test assertions
- ✅ Visual pass/fail indicators
- ✅ Response time tracking

### Manual Testing
1. Click individual requests in the collection
2. Click **"Send"**
3. View response in the bottom panel

---

## Method 4: PowerShell Script (Automated)

### Run All Tests
```powershell
.\test_gateway.ps1
```

### What It Does
- ✅ Checks if all services are running
- ✅ Creates 2 accounts
- ✅ Executes transfers
- ✅ Queries balances and statements
- ✅ Tests validation (invalid inputs)
- ✅ Tests idempotency
- ✅ Shows colored output (✓ green, ✗ red)

**Best for**: Quick smoke testing after changes

---

## Method 5: cURL (Command Line)

### Create Account
```powershell
curl -X POST http://localhost:4000/accounts `
  -H "Content-Type: application/json" `
  -d '{"currency":"USD"}'
```

### Get Account
```powershell
curl http://localhost:4000/accounts/YOUR_ACCOUNT_ID_HERE
```

### Create Transfer
```powershell
curl -X POST http://localhost:4000/transfers `
  -H "Content-Type: application/json" `
  -d '{
    "from_account_id": "ACCOUNT_A_ID",
    "to_account_id": "ACCOUNT_B_ID",
    "amount_minor": 5000,
    "currency": "USD",
    "idempotency_key": "test-123"
  }'
```

### Get Balance
```powershell
curl http://localhost:4000/accounts/YOUR_ACCOUNT_ID/balance
```

---

## Method 6: Browser DevTools (Manual HTTP)

### Using Fetch API
1. Open browser: http://localhost:4000/api (Swagger page)
2. Press **F12** to open DevTools
3. Go to **Console** tab
4. Paste and run:

```javascript
// Create Account
fetch('http://localhost:4000/accounts', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ currency: 'USD' })
})
.then(r => r.json())
.then(data => {
  console.log('Account created:', data);
  window.accountId = data.account_id; // Save for later
});

// Get Balance (after creating account)
fetch(`http://localhost:4000/accounts/${window.accountId}/balance`)
.then(r => r.json())
.then(data => console.log('Balance:', data));
```

---

## Observability While Testing

### View Traces in Jaeger
1. Open: **http://localhost:16686**
2. Select service: **gateway** (after Phase 2)
3. Click **"Find Traces"**
4. See end-to-end request flow!

### View Metrics in Grafana
1. Open: **http://localhost:3000**
2. Login: admin/admin
3. Go to **Dashboards** → **System Overview**
4. Watch metrics update in real-time as you test!

### View Prometheus Metrics
1. Open: **http://localhost:9090**
2. Query: `http_requests_total`
3. See request counts per endpoint

---

## Common Test Scenarios

### Scenario 1: Happy Path
1. Create 2 accounts
2. Transfer money between them
3. Check both balances
4. View statements
5. **Expected**: All succeed, balances correct

### Scenario 2: Validation Errors
1. Try invalid currency: `"currency": "US"` → **400 Bad Request**
2. Try invalid UUID: `/accounts/invalid` → **400 Bad Request**
3. Try same account transfer → **400 Bad Request**
4. Try negative amount → **400 Bad Request**

### Scenario 3: Idempotency
1. Create transfer with key `"test-123"`
2. Repeat exact same request
3. **Expected**: Same `transfer_id` returned both times

### Scenario 4: Not Found
1. Try to get non-existent account
2. **Expected**: **404 Not Found**

---

## Troubleshooting

### Gateway won't start
```
Error: Cannot find module 'axios'
```
**Solution**: Run `npm install` in `services/gateway`

### Services not responding
```
Error: connect ECONNREFUSED
```
**Solution**: Check all 4 Go services are running (see Prerequisites)

### Balance is 0 after transfer
**Solution**: Wait 2-3 seconds for projection to update, then retry

### Swagger UI not loading
**Solution**: 
1. Run `npm install` to get @nestjs/swagger
2. Rebuild: `npm run build`
3. Restart Gateway

---

## Quick Reference

| Tool | Best For | Difficulty |
|------|----------|-----------|
| **Swagger UI** | Beginners, visual exploration | ⭐ Easy |
| **REST Client** | VS Code users, quick tests | ⭐⭐ Easy |
| **Postman** | Comprehensive testing, teams | ⭐⭐ Medium |
| **PowerShell Script** | Automated smoke tests | ⭐ Easy |
| **cURL** | Command-line users | ⭐⭐⭐ Advanced |
| **Browser DevTools** | JavaScript developers | ⭐⭐⭐ Advanced |

---

## Next Steps

After testing the Gateway:
1. ✅ Verify all endpoints work
2. ✅ Check Jaeger for traces (Phase 2)
3. ✅ Run E2E tests (Phase 3)
4. ✅ Test failure scenarios (Phase 4)

**Recommended**: Start with **Swagger UI** for the best visual experience!
