package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/amirhf/credit-ledger/services/accounts/internal/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Relay polls the outbox table and publishes events to Kafka
type Relay struct {
	db           *sql.DB
	queries      *store.Queries
	publisher    *Publisher
	logger       *log.Logger
	pollInterval time.Duration
	batchSize    int32
}

// NewRelay creates a new outbox relay
func NewRelay(db *sql.DB, publisher *Publisher, logger *log.Logger) *Relay {
	return &Relay{
		db:           db,
		queries:      store.New(db),
		publisher:    publisher,
		logger:       logger,
		pollInterval: 100 * time.Millisecond, // Poll every 100ms
		batchSize:    10,                      // Process up to 10 events per batch
	}
}

// Start begins the relay worker loop
func (r *Relay) Start(ctx context.Context) error {
	r.logger.Println("Outbox relay worker started")
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Println("Outbox relay worker stopping...")
			return ctx.Err()
		case <-ticker.C:
			if err := r.processOutbox(ctx); err != nil {
				r.logger.Printf("Error processing outbox: %v", err)
				// Continue processing on error (with exponential backoff in production)
			}
		}
	}
}

// processOutbox fetches unsent events and publishes them
func (r *Relay) processOutbox(ctx context.Context) error {
	tracer := otel.Tracer("outbox-relay")
	ctx, span := tracer.Start(ctx, "processOutbox")
	defer span.End()

	// Start a transaction for SELECT FOR UPDATE SKIP LOCKED
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Fetch unsent events with row-level locking
	events, err := qtx.GetUnsentOutboxEvents(ctx, r.batchSize)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to fetch unsent events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to process
	}

	span.SetAttributes(attribute.Int("event_count", len(events)))
	r.logger.Printf("Processing %d outbox events", len(events))

	// Process each event
	for _, event := range events {
		if err := r.publishEvent(ctx, qtx, event); err != nil {
			r.logger.Printf("Failed to publish event %s: %v", event.ID, err)
			// Continue with other events even if one fails
			continue
		}
	}

	// Commit the transaction to mark events as sent
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// publishEvent publishes a single event to Kafka and marks it as sent
func (r *Relay) publishEvent(ctx context.Context, qtx *store.Queries, event store.Outbox) error {
	tracer := otel.Tracer("outbox-relay")
	ctx, span := tracer.Start(ctx, "publishEvent",
		trace.WithAttributes(
			attribute.String("event_id", event.ID.String()),
			attribute.String("event_type", event.EventType),
			attribute.String("aggregate_id", event.AggregateID.String()),
			attribute.String("aggregate_type", event.AggregateType),
		),
	)
	defer span.End()

	// Determine the topic based on event type
	topic := r.getTopicForEvent(event.EventType)
	span.SetAttributes(attribute.String("kafka.topic", topic))

	// Parse headers
	var headers map[string]interface{}
	if err := json.Unmarshal(event.Headers, &headers); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to unmarshal headers: %w", err)
	}

	// Add event metadata to headers
	headersMap := make(map[string]string)
	for k, v := range headers {
		headersMap[k] = fmt.Sprintf("%v", v)
	}
	headersMap["event_id"] = event.ID.String()
	headersMap["aggregate_id"] = event.AggregateID.String()
	headersMap["aggregate_type"] = event.AggregateType

	// Publish to Kafka (trace context will be injected by publisher)
	key := []byte(event.AggregateID.String())
	if err := r.publisher.Publish(ctx, topic, key, event.Payload, headersMap); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to publish to kafka: %w", err)
	}

	// Mark event as sent
	if err := qtx.MarkOutboxEventSent(ctx, event.ID); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to mark event as sent: %w", err)
	}

	r.logger.Printf("Published event %s to topic %s", event.ID, topic)
	return nil
}

// getTopicForEvent maps event types to Kafka topics
func (r *Relay) getTopicForEvent(eventType string) string {
	switch eventType {
	case "AccountCreated":
		return "ledger.account.v1"
	case "EntryPosted":
		return "ledger.entry.v1"
	case "TransferInitiated", "TransferCompleted", "TransferFailed":
		return "ledger.transfer.v1"
	default:
		return "ledger.events.v1" // Default topic
	}
}
