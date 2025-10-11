package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EntriesCreated tracks the total number of ledger entries created
	EntriesCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_entries_created_total",
			Help: "Total number of ledger entries created",
		},
		[]string{"currency"},
	)

	// EntryCreationDuration tracks the duration of entry creation operations
	EntryCreationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_entry_creation_duration_seconds",
			Help:    "Duration of ledger entry creation operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// OutboxEventsPublished tracks the total number of outbox events published
	OutboxEventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_outbox_events_published_total",
			Help: "Total number of outbox events published to Kafka",
		},
		[]string{"event_type", "topic"},
	)

	// OutboxEventsPublishErrors tracks the total number of outbox publish errors
	OutboxEventsPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_outbox_events_publish_errors_total",
			Help: "Total number of outbox event publish errors",
		},
		[]string{"event_type", "error_type"},
	)

	// OutboxRelayLatency tracks the time from event creation to Kafka publish
	OutboxRelayLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ledger_outbox_relay_latency_seconds",
			Help:    "Time from outbox event creation to Kafka publish",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)

	// OutboxQueueSize tracks the current number of unsent events in the outbox
	OutboxQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ledger_outbox_queue_size",
			Help: "Current number of unsent events in the outbox table",
		},
	)

	// LedgerBalance tracks the total balance across all accounts (for monitoring)
	LedgerBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ledger_total_balance_minor",
			Help: "Total balance across all accounts in minor units",
		},
		[]string{"currency"},
	)
)
