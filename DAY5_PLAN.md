# Day 5 Implementation Plan: Observability

**Date**: 2025-10-11  
**Status**: PLANNING  
**Priority**: MEDIUM - Enhances debugging, monitoring, and production readiness

---

## Overview

Day 5 focuses on implementing comprehensive observability across all services to enable:
- **Distributed Tracing**: End-to-end request tracking across services and Kafka
- **Metrics**: Performance monitoring, consumer lag, error rates
- **Structured Logging**: Contextual logs with trace correlation

---

## Current State (Days 1-4 Complete)

### ✅ Implemented Services
1. **Ledger Service** (Port 7102) - Journal entries with outbox pattern
2. **Accounts Service** (Port 7101) - Account creation with events
3. **Posting-Orchestrator** (Port 7103) - Transfer coordination with idempotency
4. **Read-Model** (Port 7104) - CQRS projections for balances/statements
5. **Gateway** (Port 4000) - NestJS REST API (scaffold exists)

### ✅ Infrastructure
- Redpanda (Kafka API) - Port 19092
- PostgreSQL x3 (accounts, ledger, readmodel) - Ports 5433-5435
- Redis - Port 6379
- Jaeger - Port 16686 (UI), 4318 (OTLP)
- Prometheus - Port 9090
- Grafana - Port 3000

### 🎯 Gap Analysis
- ❌ No OpenTelemetry instrumentation
- ❌ No custom Prometheus metrics
- ❌ No Grafana dashboards
- ❌ Logging lacks trace context
- ❌ No trace propagation through Kafka

---

## Day 5 Objectives

### 1. OpenTelemetry Tracing
**Goal**: End-to-end visibility from Gateway → Orchestrator → Ledger → Kafka → Read-Model

**Components**:
- HTTP server/client instrumentation
- Database (pgx) tracing
- Kafka producer/consumer spans
- Context propagation via `traceparent` header

**Services to Instrument**:
- ✅ Ledger service
- ✅ Accounts service
- ✅ Posting-Orchestrator service
- ✅ Read-Model service
- ✅ Gateway service (TypeScript/NestJS)

### 2. Prometheus Metrics
**Goal**: Real-time monitoring of system health and performance

**Custom Metrics**:
- Request counters and latency histograms
- Outbox relay metrics (age, publish rate)
- Consumer lag per topic/partition
- Projection latency
- Error rates by type

### 3. Grafana Dashboards
**Goal**: Unified visualization of system metrics

**Panels**:
- Request rate and latency (p50, p95, p99)
- Outbox age and publish rate
- Consumer lag
- Error rates
- Database connection pool stats

### 4. Structured Logging
**Goal**: Contextual logs with trace correlation

**Enhancements**:
- Add `trace_id` to all log statements
- Add domain IDs (`transfer_id`, `entry_id`, `account_id`)
- Consistent log levels (INFO, WARN, ERROR)
- JSON format for log aggregation

---

## Implementation Tasks

### Phase 1: OpenTelemetry Setup (Go Services)

#### Task 1.1: Add OTEL Dependencies
**Files**: `services/*/go.mod`

```go
require (
    go.opentelemetry.io/otel v1.21.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
    go.opentelemetry.io/otel/sdk v1.21.0
    go.opentelemetry.io/otel/trace v1.21.0
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
    go.opentelemetry.io/contrib/instrumentation/github.com/jackc/pgx/v5/otelgithub.com/jackc/pgx/v5 v0.1.0
    go.opentelemetry.io/contrib/instrumentation/github.com/segmentio/kafka-go/otelsegmentio v0.46.1
)
```

#### Task 1.2: Create OTEL Initialization Package
**File**: `services/ledger/internal/telemetry/tracer.go` (replicate for all services)

```go
package telemetry

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
    exporter, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(jaegerEndpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
        ),
    )
    if err != nil {
        return nil, err
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

#### Task 1.3: Instrument HTTP Handlers
**Files**: `services/*/internal/http/*.go`

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Wrap Chi router
r := chi.NewRouter()
r.Use(func(next http.Handler) http.Handler {
    return otelhttp.NewHandler(next, "http-server")
})
```

#### Task 1.4: Instrument HTTP Clients
**Files**: `services/posting-orchestrator/internal/http/handler.go`

```go
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}
```

#### Task 1.5: Instrument Database (pgx)
**Files**: `services/*/cmd/*/main.go`

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
    "go.opentelemetry.io/contrib/instrumentation/github.com/jackc/pgx/v5/otelgithub.com/jackc/pgx/v5"
)

config, err := pgxpool.ParseConfig(databaseURL)
config.ConnConfig.Tracer = otelpgx.NewTracer()
pool, err := pgxpool.NewWithConfig(ctx, config)
```

#### Task 1.6: Instrument Kafka Producer
**Files**: `services/*/internal/outbox/publisher.go`

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/segmentio/kafka-go/otelsegmentio"
)

writer := &kafka.Writer{
    // ... existing config
}
writer = otelsegmentio.WrapWriter(writer)
```

#### Task 1.7: Instrument Kafka Consumer
**Files**: `services/read-model/internal/consumer/consumer.go`

```go
reader := kafka.NewReader(kafka.ReaderConfig{
    // ... existing config
})
reader = otelsegmentio.WrapReader(reader)
```

#### Task 1.8: Propagate Trace Context via Kafka Headers
**Files**: `services/*/internal/outbox/relay.go`

```go
import (
    "go.opentelemetry.io/otel/propagation"
)

// When publishing to Kafka
carrier := propagation.MapCarrier{}
otel.GetTextMapPropagator().Inject(ctx, carrier)
for k, v := range carrier {
    headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
}

// When consuming from Kafka
carrier := propagation.MapCarrier{}
for _, h := range msg.Headers {
    carrier[h.Key] = string(h.Value)
}
ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
```

---

### Phase 2: Prometheus Metrics (Go Services)

#### Task 2.1: Add Prometheus Dependencies
**Files**: `services/*/go.mod`

```go
require (
    github.com/prometheus/client_golang v1.17.0
)
```

#### Task 2.2: Create Metrics Package
**File**: `services/ledger/internal/metrics/metrics.go`

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    EntriesCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "ledger_entries_created_total",
        Help: "Total number of journal entries created",
    })

    ValidationErrors = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "ledger_entry_validation_errors_total",
        Help: "Total validation errors by type",
    }, []string{"error_type"})

    OutboxRelayLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "outbox_relay_latency_seconds",
        Help:    "Outbox relay processing latency",
        Buckets: prometheus.DefBuckets,
    })

    OutboxEventsPublished = promauto.NewCounter(prometheus.CounterOpts{
        Name: "outbox_events_published_total",
        Help: "Total events published from outbox",
    })
)
```

#### Task 2.3: Instrument Ledger Service
**Files**: `services/ledger/internal/http/handler.go`, `services/ledger/internal/outbox/relay.go`

```go
// In handler
metrics.EntriesCreated.Inc()

// In relay
start := time.Now()
// ... publish logic
metrics.OutboxRelayLatency.Observe(time.Since(start).Seconds())
metrics.OutboxEventsPublished.Inc()
```

#### Task 2.4: Add Consumer Lag Metrics
**File**: `services/read-model/internal/metrics/metrics.go`

```go
var (
    ConsumerLag = promauto.NewGaugeVec(prometheus.GaugeOpts{
        Name: "consumer_lag",
        Help: "Consumer lag per topic/partition",
    }, []string{"topic", "partition"})

    ProjectionLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "projection_latency_seconds",
        Help:    "Projection processing latency",
        Buckets: prometheus.DefBuckets,
    })

    EventsProcessed = promauto.NewCounter(prometheus.CounterOpts{
        Name: "events_processed_total",
        Help: "Total events processed by projector",
    })
)
```

#### Task 2.5: Expose Metrics Endpoint
**Files**: `services/*/cmd/*/main.go`

```go
import (
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Add metrics endpoint
r.Handle("/metrics", promhttp.Handler())
```

#### Task 2.6: Create Metrics for Each Service

**Accounts Service**:
- `accounts_created_total`
- `account_creation_errors_total`

**Orchestrator Service**:
- `transfers_initiated_total`
- `transfers_completed_total`
- `transfers_failed_total`
- `transfer_latency_seconds`
- `idempotency_hits_total`

**Read-Model Service**:
- `events_processed_total`
- `projection_latency_seconds`
- `consumer_lag`
- `balance_queries_total`
- `statement_queries_total`

---

### Phase 3: Gateway Service (TypeScript/NestJS)

#### Task 3.1: Add OpenTelemetry to Gateway
**File**: `services/gateway/src/telemetry/tracer.ts`

```typescript
import { NodeSDK } from '@opentelemetry/sdk-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http';
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express';

export function initTracer(serviceName: string, jaegerEndpoint: string) {
  const sdk = new NodeSDK({
    serviceName,
    traceExporter: new OTLPTraceExporter({
      url: `${jaegerEndpoint}/v1/traces`,
    }),
    instrumentations: [
      new HttpInstrumentation(),
      new ExpressInstrumentation(),
    ],
  });

  sdk.start();
  return sdk;
}
```

#### Task 3.2: Add Prometheus Metrics to Gateway
**File**: `services/gateway/src/metrics/metrics.module.ts`

```typescript
import { PrometheusModule } from '@willsoto/nestjs-prometheus';

@Module({
  imports: [
    PrometheusModule.register({
      path: '/metrics',
      defaultMetrics: { enabled: true },
    }),
  ],
})
export class MetricsModule {}
```

---

### Phase 4: Grafana Dashboard

#### Task 4.1: Create Dashboard JSON
**File**: `deploy/grafana/dashboards/credit-ledger.json`

**Panels**:
1. **Request Rate** (rate of HTTP requests per service)
2. **Request Latency** (p50, p95, p99 histograms)
3. **Error Rate** (4xx, 5xx responses)
4. **Outbox Age** (time since created_at for unsent events)
5. **Outbox Publish Rate** (events/sec)
6. **Consumer Lag** (by topic/partition)
7. **Projection Latency** (histogram)
8. **Database Connections** (active connections per pool)
9. **Kafka Producer/Consumer Metrics**
10. **Transfer Success Rate** (completed vs failed)

#### Task 4.2: Configure Grafana Datasource
**File**: `deploy/grafana/provisioning/datasources/prometheus.yml`

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

#### Task 4.3: Configure Dashboard Provisioning
**File**: `deploy/grafana/provisioning/dashboards/dashboards.yml`

```yaml
apiVersion: 1
providers:
  - name: 'Credit Ledger'
    folder: ''
    type: file
    options:
      path: /etc/grafana/dashboards
```

---

### Phase 5: Structured Logging Enhancements

#### Task 5.1: Add Trace Context to Logs
**Files**: All `services/*/cmd/*/main.go` and handlers

```go
import (
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
)

// Create logger with trace context
func logWithTrace(ctx context.Context, logger *zap.Logger) *zap.Logger {
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        return logger.With(
            zap.String("trace_id", span.SpanContext().TraceID().String()),
            zap.String("span_id", span.SpanContext().SpanID().String()),
        )
    }
    return logger
}
```

#### Task 5.2: Add Domain Context to Logs
**Files**: All handlers and domain logic

```go
logger.Info("transfer initiated",
    zap.String("transfer_id", transferID.String()),
    zap.String("from_account", fromAccount.String()),
    zap.String("to_account", toAccount.String()),
    zap.Int64("amount_minor", amountMinor),
    zap.String("currency", currency),
)
```

#### Task 5.3: Standardize Log Levels
- **INFO**: Normal operations (request received, event published)
- **WARN**: Recoverable errors (retry, validation failure)
- **ERROR**: Unrecoverable errors (DB connection lost, Kafka down)

---

## Testing Plan

### Test 1: Trace Propagation
**Objective**: Verify end-to-end trace across all services

**Steps**:
1. POST /transfers via Gateway
2. Open Jaeger UI (http://localhost:16686)
3. Search for trace by transfer_id
4. Verify spans:
   - Gateway → Orchestrator (HTTP)
   - Orchestrator → Ledger (HTTP)
   - Ledger → Kafka (Producer)
   - Kafka → Read-Model (Consumer)
   - Database operations (pgx)

**Expected**: Single trace with 10+ spans showing full request flow

### Test 2: Metrics Collection
**Objective**: Verify Prometheus scrapes all services

**Steps**:
1. Execute multiple transfers
2. Open Prometheus UI (http://localhost:9090)
3. Query metrics:
   - `rate(ledger_entries_created_total[1m])`
   - `histogram_quantile(0.95, rate(outbox_relay_latency_seconds_bucket[1m]))`
   - `consumer_lag{topic="ledger.entry.v1"}`

**Expected**: All metrics present with non-zero values

### Test 3: Grafana Dashboard
**Objective**: Verify dashboard displays real-time data

**Steps**:
1. Open Grafana UI (http://localhost:3000)
2. Navigate to Credit Ledger dashboard
3. Execute load test (100 transfers)
4. Verify panels update with data

**Expected**: All panels show graphs with data points

### Test 4: Log Correlation
**Objective**: Verify logs contain trace_id

**Steps**:
1. Execute transfer with known idempotency key
2. Extract trace_id from Jaeger
3. Search logs for trace_id
4. Verify logs from all services appear

**Expected**: Logs from Gateway, Orchestrator, Ledger, Read-Model with same trace_id

---

## Environment Variables

Add to all services:

```bash
# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=<service-name>

# Existing
DATABASE_URL=...
KAFKA_BROKERS=...
PORT=...
```

---

## Success Criteria

### Must Have
- ✅ All services export traces to Jaeger
- ✅ Trace propagation works across HTTP and Kafka
- ✅ Custom Prometheus metrics exposed on /metrics
- ✅ Grafana dashboard displays real-time metrics
- ✅ Logs contain trace_id for correlation

### Nice to Have
- ✅ Consumer lag < 100ms under load
- ✅ p95 latency < 500ms for transfers
- ✅ Outbox age < 1 second
- ✅ Zero dropped traces

---

## File Checklist

### New Files
- [ ] `services/ledger/internal/telemetry/tracer.go`
- [ ] `services/ledger/internal/metrics/metrics.go`
- [ ] `services/accounts/internal/telemetry/tracer.go`
- [ ] `services/accounts/internal/metrics/metrics.go`
- [ ] `services/posting-orchestrator/internal/telemetry/tracer.go`
- [ ] `services/posting-orchestrator/internal/metrics/metrics.go`
- [ ] `services/read-model/internal/telemetry/tracer.go`
- [ ] `services/read-model/internal/metrics/metrics.go`
- [ ] `services/gateway/src/telemetry/tracer.ts`
- [ ] `services/gateway/src/metrics/metrics.module.ts`
- [ ] `deploy/grafana/dashboards/credit-ledger.json`
- [ ] `deploy/grafana/provisioning/datasources/prometheus.yml`
- [ ] `deploy/grafana/provisioning/dashboards/dashboards.yml`
- [ ] `DAY5_TESTING.md`
- [ ] `DAY5_SUMMARY.md`

### Modified Files
- [ ] `services/ledger/go.mod` (add OTEL deps)
- [ ] `services/ledger/cmd/ledger/main.go` (init tracer, metrics endpoint)
- [ ] `services/ledger/internal/http/handler.go` (instrument, add metrics)
- [ ] `services/ledger/internal/outbox/relay.go` (propagate trace, add metrics)
- [ ] `services/ledger/internal/outbox/publisher.go` (instrument Kafka)
- [ ] Similar changes for accounts, orchestrator, read-model services
- [ ] `deploy/docker-compose.yml` (add Grafana volumes for provisioning)

---

## Estimated Effort

- **Phase 1 (OTEL)**: 4-6 hours
- **Phase 2 (Metrics)**: 3-4 hours
- **Phase 3 (Gateway)**: 2-3 hours
- **Phase 4 (Grafana)**: 2-3 hours
- **Phase 5 (Logging)**: 1-2 hours
- **Testing**: 2-3 hours

**Total**: 14-21 hours (2-3 days)

---

## Dependencies

### Go Packages
```
go.opentelemetry.io/otel v1.21.0
go.opentelemetry.io/otel/sdk v1.21.0
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.21.0
go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
go.opentelemetry.io/contrib/instrumentation/github.com/jackc/pgx/v5/otelgithub.com/jackc/pgx/v5 v0.1.0
github.com/prometheus/client_golang v1.17.0
```

### TypeScript Packages
```json
{
  "@opentelemetry/sdk-node": "^0.45.0",
  "@opentelemetry/exporter-trace-otlp-http": "^0.45.0",
  "@opentelemetry/instrumentation-http": "^0.45.0",
  "@opentelemetry/instrumentation-express": "^0.34.0",
  "@willsoto/nestjs-prometheus": "^6.0.0"
}
```

---

## Next Steps After Day 5

### Day 6: E2E Tests & Failure Drills
- Testcontainers-go integration tests
- Consumer crash and recovery
- Kafka downtime simulation
- Performance benchmarks

### Day 7: Reconciliation & Polish
- Nightly reconciliation job
- Gateway service completion
- Architecture diagrams
- Demo script

---

## References

- [OpenTelemetry Go Docs](https://opentelemetry.io/docs/instrumentation/go/)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/)
- [Jaeger Tracing](https://www.jaegertracing.io/docs/)

---

**Status**: Ready to begin implementation
**Next Action**: Start with Phase 1 - OpenTelemetry setup for Ledger service
