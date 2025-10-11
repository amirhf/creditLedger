package consumer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/amirhf/credit-ledger/services/read-model/internal/metrics"
	"github.com/amirhf/credit-ledger/services/read-model/internal/projection"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Consumer reads events from Kafka and applies them to projections
type Consumer struct {
	reader    *kafka.Reader
	projector *projection.Projector
}

// NewConsumer creates a Kafka consumer for the ledger.entry.v1 topic
func NewConsumer(brokers []string, projector *projection.Projector) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "ledger.entry.v1",
		GroupID:        "read-model-projections",
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset, // Start from beginning for new consumers
	})

	return &Consumer{
		reader:    reader,
		projector: projector,
	}
}

// Start begins consuming messages and blocks until context is canceled
func (c *Consumer) Start(ctx context.Context) error {
	log.Println("Starting Kafka consumer for ledger.entry.v1")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Kafka consumer")
			return c.reader.Close()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context canceled, exit gracefully
				return c.reader.Close()
			}
			log.Printf("Error fetching message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Extract trace context from Kafka headers
		carrier := propagation.MapCarrier{}
		for _, h := range msg.Headers {
			carrier[h.Key] = string(h.Value)
		}
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		// Create span for message processing
		tracer := otel.Tracer("kafka-consumer")
		msgCtx, span := tracer.Start(msgCtx, "consume EntryPosted",
			trace.WithAttributes(
				attribute.String("kafka.topic", msg.Topic),
				attribute.Int("kafka.partition", msg.Partition),
				attribute.Int64("kafka.offset", msg.Offset),
			),
		)
		defer span.End()

		// Extract event_id from headers
		eventID, err := extractEventID(msg.Headers)
		if err != nil {
			span.RecordError(err)
			log.Printf("Error extracting event_id: %v, skipping message", err)
			c.reader.CommitMessages(ctx, msg)
			continue
		}
		span.SetAttributes(attribute.String("event_id", eventID.String()))

		// Process the event
		start := time.Now()
		err = c.projector.ProcessEntryPosted(msgCtx, eventID, msg.Value)
		duration := time.Since(start).Seconds()
		
		if err != nil {
			span.RecordError(err)
			metrics.EventProcessingErrors.WithLabelValues("EntryPosted", "processing_error").Inc()
			metrics.EventProcessingDuration.WithLabelValues("EntryPosted", "error").Observe(duration)
			log.Printf("Error processing event %s: %v", eventID, err)
			// Don't commit on error - will retry
			time.Sleep(time.Second)
			continue
		}

		// Record successful processing
		metrics.EventsProcessed.WithLabelValues("EntryPosted").Inc()
		metrics.EventProcessingDuration.WithLabelValues("EntryPosted", "success").Observe(duration)

		// Commit the message
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			span.RecordError(err)
			log.Printf("Error committing message: %v", err)
		}
	}
}

// extractEventID extracts the event_id from Kafka message headers
func extractEventID(headers []kafka.Header) (uuid.UUID, error) {
	for _, h := range headers {
		if h.Key == "event_id" {
			return uuid.Parse(string(h.Value))
		}
	}
	return uuid.Nil, fmt.Errorf("event_id header not found")
}
