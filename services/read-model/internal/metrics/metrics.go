package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// EventsProcessed tracks the total number of events processed
	EventsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "readmodel_events_processed_total",
			Help: "Total number of events processed by the read model",
		},
		[]string{"event_type"},
	)

	// EventProcessingDuration tracks the duration of event processing
	EventProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "readmodel_event_processing_duration_seconds",
			Help:    "Duration of event processing operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"event_type", "status"},
	)

	// EventProcessingErrors tracks the total number of event processing errors
	EventProcessingErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "readmodel_event_processing_errors_total",
			Help: "Total number of event processing errors",
		},
		[]string{"event_type", "error_type"},
	)

	// ProjectionLag tracks the lag between event timestamp and processing time
	ProjectionLag = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "readmodel_projection_lag_seconds",
			Help:    "Lag between event timestamp and processing time",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		},
	)

	// DuplicateEventsSkipped tracks the number of duplicate events skipped
	DuplicateEventsSkipped = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "readmodel_duplicate_events_skipped_total",
			Help: "Total number of duplicate events skipped due to idempotency",
		},
	)

	// BalanceQueriesTotal tracks the total number of balance queries
	BalanceQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "readmodel_balance_queries_total",
			Help: "Total number of balance queries",
		},
		[]string{"status"},
	)

	// StatementQueriesTotal tracks the total number of statement queries
	StatementQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "readmodel_statement_queries_total",
			Help: "Total number of statement queries",
		},
		[]string{"status"},
	)

	// QueryDuration tracks the duration of query operations
	QueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "readmodel_query_duration_seconds",
			Help:    "Duration of query operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type", "status"},
	)
)
