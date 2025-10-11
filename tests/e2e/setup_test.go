package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnv holds all infrastructure containers and connection details
type TestEnv struct {
	// Database containers
	PgAccounts     *postgres.PostgresContainer
	PgLedger       *postgres.PostgresContainer
	PgReadModel    *postgres.PostgresContainer
	PgOrchestrator *postgres.PostgresContainer

	// Event broker
	Redpanda *redpanda.Container

	// Cache
	Redis *redis.RedisContainer

	// Connection strings
	AccountsDB     string
	LedgerDB       string
	ReadModelDB    string
	OrchestratorDB string
	RedpandaBroker string
	RedisAddr      string
}

// setupTestEnvironment creates all infrastructure containers
func setupTestEnvironment(ctx context.Context, t *testing.T) (*TestEnv, error) {
	env := &TestEnv{}

	// Start Postgres for Accounts service
	t.Log("Starting Postgres container for Accounts service...")
	pgAccounts, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("accounts"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start accounts postgres: %w", err)
	}
	env.PgAccounts = pgAccounts

	connStr, err := pgAccounts.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts connection string: %w", err)
	}
	env.AccountsDB = connStr
	t.Logf("Accounts DB: %s", connStr)

	// Start Postgres for Ledger service
	t.Log("Starting Postgres container for Ledger service...")
	pgLedger, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("ledger"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start ledger postgres: %w", err)
	}
	env.PgLedger = pgLedger

	connStr, err = pgLedger.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger connection string: %w", err)
	}
	env.LedgerDB = connStr
	t.Logf("Ledger DB: %s", connStr)

	// Start Postgres for Read-Model service
	t.Log("Starting Postgres container for Read-Model service...")
	pgReadModel, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("readmodel"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start readmodel postgres: %w", err)
	}
	env.PgReadModel = pgReadModel

	connStr, err = pgReadModel.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get readmodel connection string: %w", err)
	}
	env.ReadModelDB = connStr
	t.Logf("Read-Model DB: %s", connStr)

	// Start Postgres for Orchestrator service
	t.Log("Starting Postgres container for Orchestrator service...")
	pgOrchestrator, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("orchestrator"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start orchestrator postgres: %w", err)
	}
	env.PgOrchestrator = pgOrchestrator

	connStr, err = pgOrchestrator.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get orchestrator connection string: %w", err)
	}
	env.OrchestratorDB = connStr
	t.Logf("Orchestrator DB: %s", connStr)

	// Start Redpanda (Kafka-compatible)
	t.Log("Starting Redpanda container...")
	redpandaContainer, err := redpanda.RunContainer(ctx,
		testcontainers.WithImage("docker.redpanda.com/redpandadata/redpanda:v23.3.3"),
		redpanda.WithAutoCreateTopics(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start redpanda: %w", err)
	}
	env.Redpanda = redpandaContainer

	broker, err := redpandaContainer.KafkaSeedBroker(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redpanda broker: %w", err)
	}
	env.RedpandaBroker = broker
	t.Logf("Redpanda Broker: %s", broker)

	// Start Redis
	t.Log("Starting Redis container...")
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start redis: %w", err)
	}
	env.Redis = redisContainer

	redisAddr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis connection string: %w", err)
	}
	env.RedisAddr = redisAddr
	t.Logf("Redis: %s", redisAddr)

	t.Log("All infrastructure containers started successfully")
	return env, nil
}

// Teardown stops all containers
func (env *TestEnv) Teardown(ctx context.Context) error {
	var errors []error

	if env.PgAccounts != nil {
		if err := env.PgAccounts.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate accounts postgres: %w", err))
		}
	}

	if env.PgLedger != nil {
		if err := env.PgLedger.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate ledger postgres: %w", err))
		}
	}

	if env.PgReadModel != nil {
		if err := env.PgReadModel.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate readmodel postgres: %w", err))
		}
	}

	if env.PgOrchestrator != nil {
		if err := env.PgOrchestrator.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate orchestrator postgres: %w", err))
		}
	}

	if env.Redpanda != nil {
		if err := env.Redpanda.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate redpanda: %w", err))
		}
	}

	if env.Redis != nil {
		if err := env.Redis.Terminate(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to terminate redis: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during teardown: %v", errors)
	}

	return nil
}

// TestSetup validates the test environment can be created
func TestSetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	env, err := setupTestEnvironment(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup test environment: %v", err)
	}
	defer env.Teardown(ctx)

	t.Log("✓ Test environment setup successful")
	t.Logf("✓ Postgres containers: 4")
	t.Logf("✓ Redpanda broker: %s", env.RedpandaBroker)
	t.Logf("✓ Redis: %s", env.RedisAddr)
}
