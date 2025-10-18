package consumer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	ledgerv1 "github.com/amirhf/credit-ledger/proto/gen/go/ledger/v1"
	"github.com/amirhf/credit-ledger/services/read-model/internal/metrics"
	"github.com/amirhf/credit-ledger/services/read-model/internal/projection"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// TransferConsumer reads transfer events from Kafka and applies them to projections
type TransferConsumer struct {
	reader    *kafka.Reader
	projector *projection.Projector
}

// NewTransferConsumer creates a Kafka consumer for the ledger.transfer.v1 topic
func NewTransferConsumer(brokers []string, projector *projection.Projector) *TransferConsumer {
	config := kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          "ledger.transfer.v1",
		GroupID:        "read-model-transfer-projections",
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset, // Start from beginning for new consumers
	}

	// Add SASL authentication if credentials provided
	if username := os.Getenv("KAFKA_SASL_USERNAME"); username != "" {
		password := os.Getenv("KAFKA_SASL_PASSWORD")
		mechanismType := os.Getenv("KAFKA_SASL_MECHANISM")
		if mechanismType == "" {
			mechanismType = "PLAIN"
		}

		var mechanism sasl.Mechanism
		var err error

		switch mechanismType {
		case "PLAIN":
			log.Println("Configuring Kafka transfer consumer SASL authentication with PLAIN")
			mechanism = plain.Mechanism{
				Username: username,
				Password: password,
			}
		case "SCRAM-SHA-256":
			log.Println("Configuring Kafka transfer consumer SASL authentication with SCRAM-SHA-256")
			mechanism, err = scram.Mechanism(scram.SHA256, username, password)
			if err != nil {
				log.Fatalf("Failed to create SCRAM mechanism: %v", err)
			}
		default:
			log.Fatalf("Unsupported SASL mechanism: %s", mechanismType)
		}

		config.Dialer = &kafka.Dialer{
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		}
	}

	reader := kafka.NewReader(config)

	return &TransferConsumer{
		reader:    reader,
		projector: projector,
	}
}

// Start begins consuming transfer messages and blocks until context is canceled
func (c *TransferConsumer) Start(ctx context.Context) error {
	log.Println("Starting Kafka consumer for ledger.transfer.v1")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Kafka transfer consumer")
			return c.reader.Close()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return c.reader.Close()
			}
			log.Printf("Error fetching transfer message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		// Extract trace context from Kafka headers
		carrier := propagation.MapCarrier{}
		for _, h := range msg.Headers {
			carrier[h.Key] = string(h.Value)
		}
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		// Extract event type from headers
		eventType, err := extractEventType(msg.Headers)
		if err != nil {
			log.Printf("Error extracting event_type: %v, skipping message", err)
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		// Create span for message processing
		tracer := otel.Tracer("kafka-transfer-consumer")
		msgCtx, span := tracer.Start(msgCtx, fmt.Sprintf("consume %s", eventType),
			trace.WithAttributes(
				attribute.String("kafka.topic", msg.Topic),
				attribute.Int("kafka.partition", msg.Partition),
				attribute.Int64("kafka.offset", msg.Offset),
				attribute.String("event_type", eventType),
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

		// Process the event based on type
		start := time.Now()
		var processErr error

		switch eventType {
		case "TransferInitiated":
			processErr = c.processTransferInitiated(msgCtx, eventID, msg.Value)
		case "TransferCompleted":
			processErr = c.processTransferCompleted(msgCtx, eventID, msg.Value)
		case "TransferFailed":
			processErr = c.processTransferFailed(msgCtx, eventID, msg.Value)
		default:
			log.Printf("Unknown event type: %s, skipping", eventType)
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		duration := time.Since(start).Seconds()

		if processErr != nil {
			span.RecordError(processErr)
			metrics.EventProcessingErrors.WithLabelValues(eventType, "processing_error").Inc()
			metrics.EventProcessingDuration.WithLabelValues(eventType, "error").Observe(duration)
			log.Printf("Error processing event %s (%s): %v", eventID, eventType, processErr)
			time.Sleep(time.Second)
			continue
		}

		// Record successful processing
		metrics.EventsProcessed.WithLabelValues(eventType).Inc()
		metrics.EventProcessingDuration.WithLabelValues(eventType, "success").Observe(duration)

		// Commit the message
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			span.RecordError(err)
			log.Printf("Error committing message: %v", err)
		}
	}
}

func (c *TransferConsumer) processTransferInitiated(ctx context.Context, eventID uuid.UUID, payload []byte) error {
	var event ledgerv1.TransferInitiated
	if err := proto.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal TransferInitiated: %w", err)
	}

	return c.projector.ProcessTransferInitiated(ctx, eventID, &event)
}

func (c *TransferConsumer) processTransferCompleted(ctx context.Context, eventID uuid.UUID, payload []byte) error {
	var event ledgerv1.TransferCompleted
	if err := proto.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal TransferCompleted: %w", err)
	}

	return c.projector.ProcessTransferCompleted(ctx, eventID, &event)
}

func (c *TransferConsumer) processTransferFailed(ctx context.Context, eventID uuid.UUID, payload []byte) error {
	var event ledgerv1.TransferFailed
	if err := proto.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal TransferFailed: %w", err)
	}

	return c.projector.ProcessTransferFailed(ctx, eventID, &event)
}

// extractEventType extracts the event_type from Kafka message headers
func extractEventType(headers []kafka.Header) (string, error) {
	for _, h := range headers {
		if h.Key == "event_type" {
			return string(h.Value), nil
		}
		// Also check in headers JSON for backward compatibility
		if h.Key == "headers" {
			var headerMap map[string]interface{}
			if err := json.Unmarshal(h.Value, &headerMap); err == nil {
				if eventName, ok := headerMap["event_name"].(string); ok {
					return eventName, nil
				}
			}
		}
	}
	return "", fmt.Errorf("event_type header not found")
}
