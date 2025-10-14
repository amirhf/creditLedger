package outbox

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
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

	// Add SASL authentication if credentials provided (for managed Kafka providers)
	if username := os.Getenv("KAFKA_SASL_USERNAME"); username != "" {
		password := os.Getenv("KAFKA_SASL_PASSWORD")
		mechanismType := os.Getenv("KAFKA_SASL_MECHANISM") // "PLAIN" or "SCRAM-SHA-256"
		if mechanismType == "" {
			mechanismType = "PLAIN" // Default to PLAIN for Confluent Cloud
		}

		var mechanism sasl.Mechanism
		var err error

		switch mechanismType {
		case "PLAIN":
			log.Println("Configuring Kafka SASL authentication with PLAIN")
			mechanism = plain.Mechanism{
				Username: username,
				Password: password,
			}
		case "SCRAM-SHA-256":
			log.Println("Configuring Kafka SASL authentication with SCRAM-SHA-256")
			mechanism, err = scram.Mechanism(scram.SHA256, username, password)
			if err != nil {
				log.Fatalf("Failed to create SCRAM mechanism: %v", err)
			}
		default:
			log.Fatalf("Unsupported SASL mechanism: %s. Use PLAIN or SCRAM-SHA-256", mechanismType)
		}

		writer.Transport = &kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{},
		}
		log.Printf("Kafka SASL authentication configured successfully with %s", mechanismType)
	}

	return &Publisher{
		writer: writer,
	}
}

// Publish sends a message to Kafka with headers
func (p *Publisher) Publish(ctx context.Context, topic string, key []byte, value []byte, headers map[string]string) error {
	// Convert headers to Kafka format
	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
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
