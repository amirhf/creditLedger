package outbox

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Publisher handles publishing events to Kafka
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(brokers []string) *Publisher {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		Async:                  false, // Synchronous for reliability
	}

	return &Publisher{
		writer: writer,
	}
}

// Publish sends a message to Kafka with headers and trace context
func (p *Publisher) Publish(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	// Convert headers to Kafka format
	kafkaHeaders := make([]kafka.Header, 0, len(headers)+2)
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	// Inject trace context into Kafka headers
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	msg := kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
	}

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka writer
func (p *Publisher) Close() error {
	return p.writer.Close()
}
