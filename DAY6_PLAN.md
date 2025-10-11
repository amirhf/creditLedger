# Day 6 Implementation Plan: E2E Tests, Failure Drills & Gateway Completion

**Date**: 2025-10-11  
**Status**: 📋 PLANNING  
**Priority**: HIGH - Validates system reliability and completes public API

---

## Executive Summary

Day 6 focuses on **system validation** and **gateway completion**. This includes:
1. **Gateway Service**: Complete implementation with all endpoints and observability
2. **E2E Testing**: Testcontainers-based integration tests for happy paths
3. **Failure Drills**: Chaos testing to validate resilience patterns
4. **Performance Testing**: Latency measurements and load testing
5. **Documentation**: Comprehensive testing guide and runbook updates

---

## Current System State (Post Day 5)

### ✅ Completed Services
- **Ledger Service**: Journal entries, outbox pattern, OTEL tracing, metrics
- **Accounts Service**: Account creation, outbox pattern, OTEL tracing, metrics
- **Posting-Orchestrator**: Transfer coordination, idempotency, OTEL tracing, metrics
- **Read-Model Service**: Event projections, balance/statement queries, OTEL tracing, metrics

### ✅ Infrastructure
- Redpanda (Kafka), Postgres x3, Redis, Jaeger, Prometheus, Grafana
- Docker Compose with full observability stack
- 3 Grafana dashboards with 30+ panels

### 🚧 Incomplete Components
- **Gateway Service**: Partial implementation (only transfers controller stub)
- **E2E Tests**: None implemented yet
- **Failure Scenarios**: Not tested
- **Performance Benchmarks**: Not measured

---

## Day 6 Objectives

### Primary Goals
1. ✅ Complete Gateway service with all REST endpoints
2. ✅ Add OpenTelemetry to Gateway (trace propagation to Go services)
3. ✅ Implement E2E test suite with Testcontainers-go
4. ✅ Test failure scenarios (service down, Kafka down, etc.)
5. ✅ Measure and document p95 end-to-end latency
6. ✅ Create comprehensive testing documentation

### Success Criteria
- [ ] Gateway exposes all 4 public endpoints with validation
- [ ] E2E tests cover happy path (create accounts → transfer → query balance)
- [ ] Idempotency test proves single effect for duplicate requests
- [ ] Failure tests validate graceful degradation
- [ ] p95 latency < 500ms for end-to-end transfer flow
- [ ] Consumer lag < 100ms under normal load
- [ ] All tests automated and runnable via single command

---

## Phase 1: Gateway Service Completion

**Duration**: ~1.5 hours  
**Priority**: HIGH - Required for E2E testing

### 1.1 Implement All REST Endpoints

**Accounts Endpoints**:
```typescript
POST /accounts
  Body: { currency: string }
  Response: 201 { account_id: string }
  → Calls accounts service: POST http://localhost:7101/v1/accounts

GET /accounts/:id
  Response: 200 { id: string, currency: string, status: string, created_at: string }
  → Calls accounts service: GET http://localhost:7101/v1/accounts/:id
```

**Transfers Endpoints**:
```typescript
POST /transfers
  Body: { from_account_id: string, to_account_id: string, amount_minor: number, currency: string, idempotency_key: string }
  Response: 202 { transfer_id: string, status: string }
  → Calls orchestrator: POST http://localhost:7103/v1/transfers

GET /transfers/:id
  Response: 200 { id: string, from_account_id: string, to_account_id: string, amount_minor: number, currency: string, status: string }
  → Calls orchestrator: GET http://localhost:7103/v1/transfers/:id
```

**Balance/Statement Endpoints**:
```typescript
GET /accounts/:id/balance
  Response: 200 { account_id: string, balance_minor: number, currency: string, updated_at: string }
  → Calls read-model: GET http://localhost:7104/v1/accounts/:id/balance

GET /accounts/:id/statements
  Query: ?from=ISO8601&to=ISO8601&limit=100
  Response: 200 { statements: [{ entry_id, amount_minor, side, ts }] }
  → Calls read-model: GET http://localhost:7104/v1/accounts/:id/statements
```

### 1.2 Request Validation (Zod)

**Schemas**:
```typescript
// accounts.schemas.ts
export const CreateAccountSchema = z.object({
  currency: z.string().length(3).regex(/^[A-Z]{3}$/)
});

// transfers.schemas.ts
export const CreateTransferSchema = z.object({
  from_account_id: z.string().uuid(),
  to_account_id: z.string().uuid(),
  amount_minor: z.number().int().positive(),
  currency: z.string().length(3).regex(/^[A-Z]{3}$/),
  idempotency_key: z.string().min(8).max(128)
});

// statements.schemas.ts
export const GetStatementsQuerySchema = z.object({
  from: z.string().datetime().optional(),
  to: z.string().datetime().optional(),
  limit: z.number().int().min(1).max(1000).default(100)
});
```

### 1.3 HTTP Client Service

**Service**: `src/services/http-client.service.ts`
- Axios-based HTTP client
- Configurable base URLs via environment variables
- Error handling and retry logic
- Request/response logging

**Environment Variables**:
```bash
ACCOUNTS_SERVICE_URL=http://localhost:7101
ORCHESTRATOR_SERVICE_URL=http://localhost:7103
READMODEL_SERVICE_URL=http://localhost:7104
```

### 1.4 Error Handling

**Error Mapping**:
- 400 Bad Request → Validation errors
- 404 Not Found → Resource not found
- 409 Conflict → Idempotency key conflict
- 500 Internal Server Error → Service errors
- 503 Service Unavailable → Downstream service down

**Error Response Format**:
```typescript
{
  error: {
    code: string,
    message: string,
    details?: any
  }
}
```

### 1.5 Files to Create/Modify

**New Files**:
- `src/controllers/accounts.controller.ts`
- `src/controllers/balances.controller.ts`
- `src/schemas/accounts.schemas.ts`
- `src/schemas/transfers.schemas.ts`
- `src/schemas/statements.schemas.ts`
- `src/services/http-client.service.ts`
- `src/services/accounts.service.ts`
- `src/services/orchestrator.service.ts`
- `src/services/readmodel.service.ts`
- `src/filters/http-exception.filter.ts`

**Modified Files**:
- `src/app.module.ts` - Add new controllers and services
- `src/main.ts` - Add global exception filter
- `src/transfers.controller.ts` - Complete implementation
- `package.json` - Add axios dependency

---

## Phase 2: OpenTelemetry Integration (Gateway)

**Duration**: ~45 minutes  
**Priority**: HIGH - Required for distributed tracing

### 2.1 Add OpenTelemetry Dependencies

```bash
npm install @opentelemetry/api \
            @opentelemetry/sdk-node \
            @opentelemetry/auto-instrumentations-node \
            @opentelemetry/exporter-trace-otlp-http \
            @opentelemetry/instrumentation-http \
            @opentelemetry/instrumentation-express \
            @opentelemetry/instrumentation-nestjs-core
```

### 2.2 Tracer Setup

**File**: `src/tracing.ts`
```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';

export function initTracing() {
  const sdk = new NodeSDK({
    serviceName: 'gateway',
    traceExporter: new OTLPTraceExporter({
      url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces'
    }),
    instrumentations: [getNodeAutoInstrumentations()]
  });
  
  sdk.start();
  
  process.on('SIGTERM', () => {
    sdk.shutdown().then(() => console.log('Tracing terminated'));
  });
}
```

### 2.3 HTTP Client Instrumentation

**Propagate traceparent to Go services**:
```typescript
import { trace, context } from '@opentelemetry/api';

// In HTTP client
const span = trace.getActiveSpan();
if (span) {
  const traceId = span.spanContext().traceId;
  const spanId = span.spanContext().spanId;
  headers['traceparent'] = `00-${traceId}-${spanId}-01`;
}
```

### 2.4 Verification

- Start Gateway with OTEL enabled
- Make request: POST /transfers
- Check Jaeger UI for trace spanning Gateway → Orchestrator → Ledger → Kafka → ReadModel

---

## Phase 3: E2E Test Suite (Testcontainers-go)

**Duration**: ~2 hours  
**Priority**: HIGH - Core validation

### 3.1 Test Infrastructure Setup

**Directory**: `tests/e2e/`

**Files**:
- `tests/e2e/setup_test.go` - Testcontainers setup
- `tests/e2e/happy_path_test.go` - Happy path scenarios
- `tests/e2e/idempotency_test.go` - Idempotency tests
- `tests/e2e/failure_test.go` - Failure scenarios
- `tests/e2e/performance_test.go` - Latency benchmarks
- `tests/e2e/helpers.go` - Test utilities

**Dependencies**:
```go
github.com/testcontainers/testcontainers-go v0.26.0
github.com/testcontainers/testcontainers-go/modules/postgres v0.26.0
github.com/testcontainers/testcontainers-go/modules/redpanda v0.26.0
github.com/testcontainers/testcontainers-go/modules/redis v0.26.0
```

### 3.2 Container Setup

**Containers Required**:
1. Postgres x3 (accounts, ledger, read-model)
2. Redpanda (Kafka)
3. Redis (idempotency)

**Setup Code**:
```go
func setupTestEnvironment(ctx context.Context) (*TestEnv, error) {
    // Start Postgres containers
    pgAccounts := testcontainers.PostgresContainer{...}
    pgLedger := testcontainers.PostgresContainer{...}
    pgReadModel := testcontainers.PostgresContainer{...}
    
    // Start Redpanda
    redpanda := testcontainers.RedpandaContainer{...}
    
    // Start Redis
    redis := testcontainers.RedisContainer{...}
    
    // Run migrations
    runMigrations(pgAccounts, "services/accounts/internal/store/migrations")
    runMigrations(pgLedger, "services/ledger/internal/store/migrations")
    runMigrations(pgReadModel, "services/read-model/internal/store/migrations")
    
    // Start services
    startService("accounts", pgAccounts.ConnectionString(), redpanda.Brokers())
    startService("ledger", pgLedger.ConnectionString(), redpanda.Brokers())
    startService("orchestrator", pgOrch.ConnectionString(), redis.ConnectionString())
    startService("read-model", pgReadModel.ConnectionString(), redpanda.Brokers())
    
    return &TestEnv{...}, nil
}
```

### 3.3 Happy Path Test

**Test**: `TestHappyPath_CreateAccountsTransferCheckBalance`

**Flow**:
1. Create account A (USD)
2. Create account B (USD)
3. Transfer $50.00 (5000 minor) from A to B
4. Wait for projection (poll balance endpoint)
5. Assert: A balance = -5000, B balance = +5000
6. Query statements for both accounts
7. Assert: 2 entries each (one DEBIT, one CREDIT)

**Expected Duration**: < 2 seconds

### 3.4 Idempotency Test

**Test**: `TestIdempotency_DuplicateTransferSingleEffect`

**Flow**:
1. Create accounts A and B
2. Execute transfer with idempotency_key="test-idem-1"
3. Execute same transfer again (duplicate)
4. Execute same transfer a third time
5. Wait for projections
6. Assert: Only one transfer executed (balance changed once)
7. Query orchestrator for transfer by idempotency key
8. Assert: Same transfer_id returned for all requests

**Expected Duration**: < 3 seconds

### 3.5 Consumer Crash Recovery Test

**Test**: `TestConsumerCrash_CatchUpAfterRestart`

**Flow**:
1. Create accounts and execute 10 transfers
2. Stop read-model service mid-flight (after 5 transfers)
3. Execute 5 more transfers (events accumulate in Kafka)
4. Restart read-model service
5. Wait for consumer to catch up
6. Assert: All 10 transfers reflected in balances
7. Measure consumer lag (should be < 100ms)

**Expected Duration**: < 5 seconds

---

## Phase 4: Failure Scenario Tests

**Duration**: ~1 hour  
**Priority**: MEDIUM - Validates resilience

### 4.1 Ledger Service Down

**Test**: `TestFailure_LedgerServiceDown`

**Flow**:
1. Stop ledger service
2. Attempt transfer via orchestrator
3. Assert: 500 Internal Server Error
4. Assert: Transfer marked as FAILED in orchestrator DB
5. Assert: TransferFailed event emitted
6. Restart ledger service
7. Execute new transfer
8. Assert: Success

### 4.2 Kafka Down

**Test**: `TestFailure_KafkaDown_OutboxAccumulates`

**Flow**:
1. Create account (success)
2. Stop Redpanda container
3. Create another account (DB write succeeds)
4. Assert: Outbox row created but not sent
5. Restart Redpanda
6. Wait for outbox relay to publish
7. Assert: Event published to Kafka
8. Assert: Consumer processes event

### 4.3 Database Deadlock

**Test**: `TestFailure_DatabaseDeadlock_Retry`

**Flow**:
1. Simulate concurrent transfers to same accounts
2. Trigger deadlock condition
3. Assert: Transaction retried automatically
4. Assert: All transfers eventually succeed
5. Assert: Balances correct

### 4.4 Invalid Account ID

**Test**: `TestValidation_InvalidAccountId`

**Flow**:
1. Attempt transfer with non-existent from_account_id
2. Assert: 400 Bad Request (or 404 Not Found)
3. Assert: No transfer record created
4. Assert: No events emitted

---

## Phase 5: Performance Testing

**Duration**: ~45 minutes  
**Priority**: MEDIUM - Validates SLOs

### 5.1 Latency Benchmark

**Test**: `BenchmarkE2E_TransferLatency`

**Measurement**:
- Create 100 accounts
- Execute 1000 transfers
- Measure end-to-end latency (POST /transfers → balance updated)
- Calculate p50, p95, p99 latencies
- Assert: p95 < 500ms

**Metrics**:
```go
type LatencyMetrics struct {
    P50 time.Duration
    P95 time.Duration
    P99 time.Duration
    Max time.Duration
    Min time.Duration
    Avg time.Duration
}
```

### 5.2 Consumer Lag Measurement

**Test**: `TestPerformance_ConsumerLag`

**Flow**:
1. Execute 100 transfers rapidly
2. Measure time from entry creation to projection update
3. Calculate average lag
4. Assert: Average lag < 100ms

### 5.3 Throughput Test

**Test**: `TestPerformance_Throughput`

**Flow**:
1. Execute transfers in parallel (10 goroutines)
2. Measure transfers per second
3. Assert: > 100 transfers/sec

---

## Phase 6: Documentation

**Duration**: ~30 minutes  
**Priority**: MEDIUM - Knowledge transfer

### 6.1 Testing Guide

**File**: `DAY6_TESTING_GUIDE.md`

**Contents**:
- How to run E2E tests
- How to run specific test suites
- How to interpret test results
- Troubleshooting common issues
- Performance benchmarking guide

### 6.2 Runbook Updates

**File**: `docs/runbook.md` (create if doesn't exist)

**Additions**:
- Failure scenario playbooks
- Recovery procedures
- Performance tuning guide
- Monitoring and alerting setup

### 6.3 Day 6 Summary

**File**: `DAY6_SUMMARY.md`

**Contents**:
- What was implemented
- Test results and metrics
- Known issues and limitations
- Next steps (Day 7)

---

## Implementation Checklist

### Phase 1: Gateway Service ✅
- [ ] Create accounts controller
- [ ] Create balances controller
- [ ] Update transfers controller
- [ ] Implement Zod validation schemas
- [ ] Create HTTP client service
- [ ] Create service layer (accounts, orchestrator, readmodel)
- [ ] Add error handling and filters
- [ ] Add axios dependency
- [ ] Update app.module.ts
- [ ] Test all endpoints manually

### Phase 2: Gateway OTEL ✅
- [ ] Add OpenTelemetry dependencies
- [ ] Create tracing.ts setup
- [ ] Add auto-instrumentation
- [ ] Propagate traceparent to Go services
- [ ] Test trace in Jaeger UI
- [ ] Update main.ts to init tracing

### Phase 3: E2E Tests ✅
- [ ] Create tests/e2e directory
- [ ] Add testcontainers dependencies
- [ ] Implement setup_test.go
- [ ] Implement happy_path_test.go
- [ ] Implement idempotency_test.go
- [ ] Implement consumer_crash_test.go
- [ ] Create test helpers
- [ ] Add Makefile target for E2E tests

### Phase 4: Failure Tests ✅
- [ ] Implement ledger_down_test.go
- [ ] Implement kafka_down_test.go
- [ ] Implement deadlock_test.go
- [ ] Implement validation_test.go
- [ ] Document failure scenarios

### Phase 5: Performance Tests ✅
- [ ] Implement latency benchmark
- [ ] Implement consumer lag test
- [ ] Implement throughput test
- [ ] Document performance results

### Phase 6: Documentation ✅
- [ ] Create DAY6_TESTING_GUIDE.md
- [ ] Update runbook.md
- [ ] Create DAY6_SUMMARY.md
- [ ] Update README.md with test instructions

---

## Expected Outcomes

### Deliverables
1. ✅ Fully functional Gateway service with 6 endpoints
2. ✅ Gateway integrated with OpenTelemetry
3. ✅ E2E test suite with 10+ test cases
4. ✅ Failure scenario tests (4 scenarios)
5. ✅ Performance benchmarks and metrics
6. ✅ Comprehensive testing documentation

### Metrics
- **Test Coverage**: > 80% for critical paths
- **E2E Test Duration**: < 30 seconds for full suite
- **p95 Latency**: < 500ms (target)
- **Consumer Lag**: < 100ms (target)
- **Throughput**: > 100 transfers/sec (target)

### Quality Gates
- [ ] All E2E tests pass
- [ ] All failure tests pass
- [ ] Performance targets met
- [ ] No critical bugs found
- [ ] Documentation complete

---

## Risk Assessment

### High Risk
- **Testcontainers complexity**: May require significant setup time
  - Mitigation: Start with simple containers, add complexity incrementally
  
- **Flaky tests**: Network timing issues in E2E tests
  - Mitigation: Use polling with timeouts, generous wait times

### Medium Risk
- **Performance targets**: May not meet 500ms p95 initially
  - Mitigation: Profile and optimize hot paths, acceptable to document actual performance

- **Gateway TypeScript/Go integration**: Type mismatches
  - Mitigation: Strict validation, comprehensive error handling

### Low Risk
- **Documentation time**: May take longer than estimated
  - Mitigation: Can be completed after Day 6 if needed

---

## Timeline

**Total Estimated Time**: 6-7 hours

| Phase | Duration | Priority |
|-------|----------|----------|
| Gateway Service | 1.5 hours | HIGH |
| Gateway OTEL | 0.75 hours | HIGH |
| E2E Tests | 2 hours | HIGH |
| Failure Tests | 1 hour | MEDIUM |
| Performance Tests | 0.75 hours | MEDIUM |
| Documentation | 0.5 hours | MEDIUM |

---

## Next Steps (Day 7)

After Day 6 completion:
1. **Reconciliation Job**: Nightly balance recomputation
2. **Replay Procedure**: Rebuild read-model from events
3. **Architecture Diagrams**: C4 diagrams for documentation
4. **Demo Script**: Portfolio video preparation
5. **README Polish**: Quickstart guide and screenshots

---

## Success Criteria Summary

✅ **Day 6 Complete When**:
- [ ] Gateway exposes all 6 REST endpoints
- [ ] Gateway has OpenTelemetry tracing
- [ ] E2E test suite passes (happy path + idempotency + crash recovery)
- [ ] Failure tests validate resilience patterns
- [ ] Performance benchmarks documented
- [ ] Testing guide published

**Status**: 📋 READY TO START
