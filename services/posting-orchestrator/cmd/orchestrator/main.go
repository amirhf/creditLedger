package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/compensator"
	orchestratorhttp "github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/http"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/idem"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/outbox"
	"github.com/amirhf/credit-ledger/services/posting-orchestrator/internal/telemetry"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
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
	
	tp, err := telemetry.InitTracer(ctx, "orchestrator-service", otlpEndpoint)
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := telemetry.Shutdown(ctx, tp); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping database to verify connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Get Redis URL from environment
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379" // Default for local development
	}

	// Create idempotency guard
	idemGuard := idem.NewGuard(redisURL)

	// Get Kafka brokers from environment
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:19092" // Default for local development
	}
	brokers := strings.Split(kafkaBrokers, ",")
	log.Printf("Kafka brokers: %v", brokers)

	// Create Kafka publisher
	publisher := outbox.NewPublisher(brokers)
	defer publisher.Close()

	// Create outbox relay
	relay := outbox.NewRelay(db, publisher, log.Default())

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start outbox relay worker in background
	go func() {
		if err := relay.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Outbox relay stopped with error: %v", err)
		}
	}()

	// Get ledger service URL
	ledgerURL := os.Getenv("LEDGER_URL")
	if ledgerURL == "" {
		ledgerURL = "http://localhost:7102" // Default for local development
	}

	// Create and start compensator worker in background
	comp := compensator.NewCompensator(db, ledgerURL, log.Default())
	go func() {
		if err := comp.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Compensator worker stopped with error: %v", err)
		}
	}()

	// Create handler
	handler := orchestratorhttp.NewHandler(db, idemGuard, ledgerURL, log.Default())

	// Setup router
	r := chi.NewRouter()
	
	// Add OpenTelemetry HTTP middleware
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "orchestrator-service",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)
	})
	
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Handle("/metrics", promhttp.Handler())
	r.Post("/v1/transfers", handler.CreateTransfer)
	r.Get("/v1/transfers/{id}", handler.GetTransfer)

	// Setup HTTP server
	addr := ":7103"
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start HTTP server in background
	go func() {
		log.Printf("orchestrator listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")

	// Cancel context to stop relay worker
	cancel()

	// Shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
