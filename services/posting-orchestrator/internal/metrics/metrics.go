package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TransfersInitiated tracks the total number of transfers initiated
	TransfersInitiated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orchestrator_transfers_initiated_total",
			Help: "Total number of transfers initiated",
		},
		[]string{"currency"},
	)

	// TransferDuration tracks the end-to-end duration of transfer operations
	TransferDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orchestrator_transfer_duration_seconds",
			Help:    "Duration of transfer operations from request to ledger response",
			Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"status"},
	)

	// IdempotencyHits tracks the number of idempotency key hits
	IdempotencyHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orchestrator_idempotency_hits_total",
			Help: "Total number of duplicate requests caught by idempotency",
		},
	)

	// LedgerCallDuration tracks the duration of calls to the ledger service
	LedgerCallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orchestrator_ledger_call_duration_seconds",
			Help:    "Duration of HTTP calls to the ledger service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// OutboxEventsPublished tracks the total number of outbox events published
	OutboxEventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orchestrator_outbox_events_published_total",
			Help: "Total number of outbox events published to Kafka",
		},
		[]string{"event_type", "topic"},
	)

	// OutboxEventsPublishErrors tracks the total number of outbox publish errors
	OutboxEventsPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orchestrator_outbox_events_publish_errors_total",
			Help: "Total number of outbox event publish errors",
		},
		[]string{"event_type", "error_type"},
	)

	// OutboxRelayLatency tracks the time from event creation to Kafka publish
	OutboxRelayLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "orchestrator_outbox_relay_latency_seconds",
			Help:    "Time from outbox event creation to Kafka publish",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)
)
