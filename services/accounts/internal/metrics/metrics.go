package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AccountsCreated tracks the total number of accounts created
	AccountsCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "accounts_created_total",
			Help: "Total number of accounts created",
		},
		[]string{"currency"},
	)

	// AccountCreationDuration tracks the duration of account creation operations
	AccountCreationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "accounts_creation_duration_seconds",
			Help:    "Duration of account creation operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// OutboxEventsPublished tracks the total number of outbox events published
	OutboxEventsPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "accounts_outbox_events_published_total",
			Help: "Total number of outbox events published to Kafka",
		},
		[]string{"event_type", "topic"},
	)

	// OutboxEventsPublishErrors tracks the total number of outbox publish errors
	OutboxEventsPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "accounts_outbox_events_publish_errors_total",
			Help: "Total number of outbox event publish errors",
		},
		[]string{"event_type", "error_type"},
	)

	// OutboxRelayLatency tracks the time from event creation to Kafka publish
	OutboxRelayLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "accounts_outbox_relay_latency_seconds",
			Help:    "Time from outbox event creation to Kafka publish",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)
)
