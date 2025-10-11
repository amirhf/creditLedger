# E2E Integration Tests

End-to-end integration tests for the Credit Ledger system using Testcontainers.

## Overview

These tests validate the entire system by:
- Spinning up real infrastructure (Postgres, Redpanda/Kafka, Redis)
- Starting all microservices (accounts, ledger, orchestrator, read-model, gateway)
- Executing complete workflows
- Verifying correctness, idempotency, and resilience

## Prerequisites

- **Go 1.22+** installed
- **Docker** running (for Testcontainers)
- All service binaries built or source code available
- Sufficient Docker resources (4GB RAM minimum)

## Test Suite

### 1. Happy Path Test (`happy_path_test.go`)
Tests the basic workflow:
- Create two accounts
- Execute transfer from A to B
- Verify balances updated correctly
- Verify statements have correct entries

**Duration**: ~10 seconds  
**Command**: `go test -v -run TestHappyPath`

### 2. Idempotency Test (`idempotency_test.go`)
Tests duplicate request handling:
- Execute same transfer 3 times with same idempotency key
- Verify only one transfer created
- Verify balance changed only once

**Duration**: ~5 seconds  
**Command**: `go test -v -run TestIdempotency`

### 3. Resilience Test (`resilience_test.go`)
Tests consumer crash recovery:
- Execute transfers while consumer is running
- Stop read-model service
- Execute more transfers (events accumulate in Kafka)
- Restart read-model service
- Verify all events eventually processed

**Duration**: ~15 seconds  
**Command**: `go test -v -run TestResilience`

### 4. Performance Benchmarks (`performance_test.go`)
Measures end-to-end latency:
- Execute N transfers
- Measure time from API call to balance update
- Calculate p50, p95, p99 percentiles

**Duration**: Variable (depends on N)  
**Command**: `go test -v -bench=BenchmarkE2E -benchmem`

## Running Tests

### Run All Tests
```bash
cd tests/e2e
C:\Users\firou\sdk\go1.24.8\bin\go test -v ./... -timeout 5m
```

### Run Specific Test
```bash
C:\Users\firou\sdk\go1.24.8\bin\go test -v -run TestHappyPath
```

### Run with Race Detector
```bash
C:\Users\firou\sdk\go1.24.8\bin\go test -v -race ./...
```

### Run Benchmarks
```bash
C:\Users\firou\sdk\go1.24.8\bin\go test -v -bench=. -benchmem
```

### Run with Coverage
```bash
C:\Users\firou\sdk\go1.24.8\bin\go test -v -cover -coverprofile=coverage.out ./...
C:\Users\firou\sdk\go1.24.8\bin\go tool cover -html=coverage.out
```

## Test Architecture

### Infrastructure (Testcontainers)
- **Postgres x4**: One for each service (accounts, ledger, orchestrator, read-model)
- **Redpanda**: Kafka-compatible event broker
- **Redis**: Idempotency key storage

### Services
- **Accounts**: Account creation and management
- **Ledger**: Double-entry journal
- **Orchestrator**: Transfer coordination
- **Read-Model**: Balance projections
- **Gateway**: Public REST API

### Test Flow
```
1. Start containers (Postgres, Redpanda, Redis)
2. Run migrations on all databases
3. Start all microservices
4. Wait for services to be ready
5. Execute test scenarios
6. Assert expected outcomes
7. Stop services and containers
```

## Troubleshooting

### Tests timeout
- Increase test timeout: `-timeout 10m`
- Check Docker has sufficient resources
- Verify Docker daemon is running

### Port conflicts
- Tests use dynamic ports from Testcontainers
- Stop existing services before running tests

### Containers fail to start
- Pull images manually: `docker pull postgres:15`, `docker pull docker.redpanda.com/redpandadata/redpanda:latest`
- Check Docker disk space
- Increase Docker memory limit

### Flaky tests
- Increase wait timeouts in helpers
- Add debug logging: `t.Logf(...)`
- Check service logs for errors

### Services crash
- Check service logs in test output
- Verify migrations ran successfully
- Check environment variables are correct

## Expected Results

### Happy Path Test
- ✅ Two accounts created
- ✅ Transfer executes successfully
- ✅ Account A balance: -5000
- ✅ Account B balance: +5000
- ✅ Transfer status: COMPLETED
- ✅ Statements have correct entries

### Idempotency Test
- ✅ Same transfer ID returned for all requests
- ✅ Balance changed only once (not tripled)
- ✅ Only one journal entry created

### Resilience Test
- ✅ Consumer stops when service down
- ✅ Events accumulate in Kafka
- ✅ Consumer catches up after restart
- ✅ All transfers eventually processed
- ✅ No duplicate processing

### Performance Benchmarks
- ✅ p50 latency: < 100ms
- ✅ p95 latency: < 500ms (target)
- ✅ p99 latency: < 1000ms
- ✅ Throughput: > 50 transfers/sec

## CI/CD Integration

To run these tests in CI:

```yaml
# Example GitHub Actions workflow
jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: Run E2E Tests
        run: |
          cd tests/e2e
          go test -v ./... -timeout 10m
```

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| p50 Latency | < 100ms | Time from API call to balance update |
| p95 Latency | < 500ms | 95th percentile end-to-end |
| p99 Latency | < 1000ms | 99th percentile end-to-end |
| Throughput | > 50/sec | Transfers processed per second |
| Consumer Lag | < 100ms | Time from produce to consume |

## Files

- `setup_test.go` - Testcontainers infrastructure setup
- `services.go` - Service lifecycle management
- `helpers.go` - HTTP client, wait utilities, assertions
- `happy_path_test.go` - Basic workflow validation
- `idempotency_test.go` - Duplicate request handling
- `resilience_test.go` - Consumer crash recovery
- `performance_test.go` - Latency benchmarks

## Contributing

When adding new tests:
1. Use existing helpers for common operations
2. Add proper cleanup (`defer env.Teardown(ctx)`)
3. Use meaningful test names
4. Add assertions with clear messages
5. Keep tests isolated (no shared state)
6. Document expected behavior

## References

- [Testcontainers-go Documentation](https://golang.testcontainers.org/)
- [Testify Assertions](https://github.com/stretchr/testify)
- Main project README: `../../README.md`
- System design: `../../design.md`
