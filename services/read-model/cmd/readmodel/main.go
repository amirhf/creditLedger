package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amirhf/credit-ledger/services/read-model/internal/consumer"
	readmodelhttp "github.com/amirhf/credit-ledger/services/read-model/internal/http"
	"github.com/amirhf/credit-ledger/services/read-model/internal/projection"
	"github.com/amirhf/credit-ledger/services/read-model/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Initialize OpenTelemetry tracer
	ctx := context.Background()
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4318" // Default Jaeger OTLP endpoint
	}
	
	tp, err := telemetry.InitTracer(ctx, "read-model-service", otlpEndpoint)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := telemetry.Shutdown(ctx, tp); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	// Load environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:19092"
	}
	brokers := strings.Split(kafkaBrokers, ",")

	port := os.Getenv("PORT")
	if port == "" {
		port = "7104"
	}

	// Connect to database
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Health check database
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Create projector
	projector := projection.NewProjector(dbPool)

	// Create Kafka consumer
	kafkaConsumer := consumer.NewConsumer(brokers, projector)

	// Start Kafka consumer in background
	consumerCtx, cancelConsumer := context.WithCancel(ctx)
	defer cancelConsumer()

	go func() {
		if err := kafkaConsumer.Start(consumerCtx); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}()

	// Setup HTTP server
	r := chi.NewRouter()
	
	// Add OpenTelemetry HTTP middleware
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "read-model-service",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	})
	
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Handle("/metrics", promhttp.Handler())

	// Query endpoints
	handler := readmodelhttp.NewHandler(dbPool)
	r.Get("/v1/accounts/{id}/balance", handler.GetBalance)
	r.Get("/v1/accounts/{id}/statements", handler.GetStatements)

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("read-model listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Cancel consumer context
	cancelConsumer()

	// Shutdown HTTP server with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
