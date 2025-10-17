# Detect OS
ifeq ($(OS),Windows_NT)
    SHELL := powershell.exe
    .SHELLFLAGS := -NoProfile -Command
    SET_ENV := $$env:
    ENV_SEP := ;
else
    SHELL := /bin/bash
    SET_ENV := 
    ENV_SEP := &&
endif

COMPOSE := docker compose -f deploy/docker-compose.yml

# Infrastructure services (dependencies)
INFRA_SERVICES := redpanda console redis postgres-accounts postgres-ledger postgres-readmodel postgres-orchestrator jaeger prometheus grafana

# Application services
APP_SERVICES := accounts ledger-svc orchestrator read-model gateway


.PHONY: help up up-nobuild down down-clean clean infra-up infra-down services-up services-down \
restart rebuild status ps logs logs-infra logs-services build proto sqlc test \
docker-build run-accounts run-ledger run-orchestrator run-readmodel run-gateway

# Default target - show help
help:
	@echo "=== Credit Ledger Makefile ==="
	@echo ""
	@echo "Quick Start:"
	@echo "  make up              - Build and start everything (infra + services)"
	@echo "  make up-nobuild      - Start everything without rebuilding"
	@echo "  make down            - Stop everything (preserves database data)"
	@echo "  make down-clean      - Stop everything and remove volumes (deletes data)"
	@echo ""
	@echo "Infrastructure:"
	@echo "  make infra-up        - Start only infrastructure (Postgres, Kafka, Redis, etc.)"
	@echo "  make infra-down      - Stop only infrastructure"
	@echo "  make logs-infra      - Tail infrastructure logs"
	@echo ""
	@echo "Services:"
	@echo "  make services-up     - Start only application services (requires infra)"
	@echo "  make services-down   - Stop only application services"
	@echo "  make restart         - Restart services without rebuilding"
	@echo "  make rebuild         - Force rebuild and restart all services"
	@echo "  make logs-services   - Tail service logs"
	@echo ""
	@echo "Development:"
	@echo "  make build           - Build all services locally (Go + npm)"
	@echo "  make docker-build    - Build all Docker images"
	@echo "  make proto           - Generate protobuf code"
	@echo "  make sqlc            - Generate sqlc code"
	@echo "  make test            - Run tests"
	@echo ""
	@echo "Monitoring:"
	@echo "  make status          - Show running containers"
	@echo "  make ps              - Docker ps for this project"
	@echo "  make logs            - Tail all logs"
	@echo ""
	@echo "Local Development (run services outside Docker):"
	@echo "  make run-accounts    - Run accounts service locally"
	@echo "  make run-ledger      - Run ledger service locally"
	@echo "  make run-orchestrator - Run orchestrator service locally"
	@echo "  make run-readmodel   - Run read-model service locally"
	@echo "  make run-gateway     - Run gateway service locally"
	@echo ""
	@echo "Cleanup:"
	@echo "  make down-clean      - Stop everything and remove volumes (deletes data)"
	@echo "  make clean           - Complete cleanup including orphaned containers"

# Start everything with build
up:
	$(COMPOSE) up -d --build
	@echo ""
	@echo "=== All services started ==="
	$(COMPOSE) ps

# Start everything without rebuilding (use when builds are already done)
up-nobuild:
	$(COMPOSE) up -d
	@echo ""
	@echo "=== All services started (no rebuild) ==="
	$(COMPOSE) ps

# Start only infrastructure
infra-up:
	$(COMPOSE) up -d $(INFRA_SERVICES)
	@echo ""
	@echo "=== Infrastructure started ==="
	$(COMPOSE) ps

# Stop only infrastructure
infra-down:
	$(COMPOSE) stop $(INFRA_SERVICES)

# Start only application services (assumes infra is running)
services-up:
	$(COMPOSE) up -d --build $(APP_SERVICES)
	@echo ""
	@echo "=== Application services started ==="
	$(COMPOSE) ps $(APP_SERVICES)

# Stop only application services
services-down:
	$(COMPOSE) stop $(APP_SERVICES)

# Restart services without rebuilding
restart:
	$(COMPOSE) restart $(APP_SERVICES)
	@echo ""
	@echo "=== Services restarted ==="
	$(COMPOSE) ps $(APP_SERVICES)

# Force rebuild and restart all services
rebuild:
	$(COMPOSE) up -d --build --force-recreate $(APP_SERVICES)
	@echo ""
	@echo "=== Services rebuilt and restarted ==="
	$(COMPOSE) ps $(APP_SERVICES)

# Show status of all containers
status:
	@echo "=== Container Status ==="
	$(COMPOSE) ps

# Docker ps alias
ps:
	$(COMPOSE) ps

# Tail all logs
logs:
	$(COMPOSE) logs -f --tail=200

# Tail infrastructure logs only
logs-infra:
	$(COMPOSE) logs -f --tail=200 $(INFRA_SERVICES)

# Tail service logs only
logs-services:
	$(COMPOSE) logs -f --tail=200 $(APP_SERVICES)

# Stop everything (preserves volumes/data)
down:
	$(COMPOSE) down
	@echo "=== Services stopped (data preserved) ==="
	@echo "To remove volumes and delete data, use: make down-clean"

# Stop everything and remove volumes (deletes all data)
down-clean:
	$(COMPOSE) down -v
	@echo "=== Services stopped and volumes removed ==="

# Complete cleanup including orphaned containers
clean:
	$(COMPOSE) down -v --remove-orphans
	@echo "=== Complete cleanup done ==="


build:
	cd services/accounts && go build ./...
	cd services/ledger && go build ./...
	cd services/posting-orchestrator && go build ./...
	cd services/read-model && go build ./...
	cd services/gateway && npm install && npm run build


proto:
	buf generate


sqlc:
	cd services/accounts/internal/store && sqlc generate || exit 0
	cd services/ledger/internal/store && sqlc generate || exit 0
	cd services/read-model/internal/store && sqlc generate || exit 0


test:
	go test ./...


# Build all images
docker-build:
	$(COMPOSE) build accounts ledger-svc orchestrator read-model gateway


# Run single services locally with Go/Node (useful for rapid dev without Docker)
# These assume Docker infra is running via `make up`

ifeq ($(OS),Windows_NT)
run-accounts:
	cd services/accounts; $$env:PORT="7101"; $$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5433/accounts?sslmode=disable"; $$env:KAFKA_BROKERS="localhost:19092"; $$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"; go run ./cmd/accounts

run-ledger:
	cd services/ledger; $$env:PORT="7102"; $$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable"; $$env:KAFKA_BROKERS="localhost:19092"; $$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"; go run ./cmd/ledger

run-orchestrator:
	cd services/posting-orchestrator; $$env:PORT="7103"; $$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5436/orchestrator?sslmode=disable"; $$env:REDIS_URL="redis://localhost:6379"; $$env:LEDGER_URL="http://localhost:7102"; $$env:KAFKA_BROKERS="localhost:19092"; $$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"; go run ./cmd/orchestrator

run-readmodel:
	cd services/read-model; $$env:PORT="7104"; $$env:DATABASE_URL="postgres://ledger:ledgerpw@localhost:5435/readmodel?sslmode=disable"; $$env:KAFKA_BROKERS="localhost:19092"; $$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"; go run ./cmd/readmodel

run-gateway:
	cd services/gateway; $$env:PORT="4000"; $$env:ACCOUNTS_SERVICE_URL="http://localhost:7101"; $$env:ORCHESTRATOR_SERVICE_URL="http://localhost:7103"; $$env:READMODEL_SERVICE_URL="http://localhost:7104"; $$env:OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4318"; npm start
else
run-accounts:
	cd services/accounts && \
	PORT=7101 \
	DATABASE_URL=postgres://ledger:ledgerpw@localhost:5433/accounts?sslmode=disable \
	KAFKA_BROKERS=localhost:19092 \
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
	go run ./cmd/accounts

run-ledger:
	cd services/ledger && \
	PORT=7102 \
	DATABASE_URL=postgres://ledger:ledgerpw@localhost:5434/ledger?sslmode=disable \
	KAFKA_BROKERS=localhost:19092 \
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
	go run ./cmd/ledger

run-orchestrator:
	cd services/posting-orchestrator && \
	PORT=7103 \
	DATABASE_URL=postgres://ledger:ledgerpw@localhost:5436/orchestrator?sslmode=disable \
	REDIS_URL=redis://localhost:6379 \
	LEDGER_URL=http://localhost:7102 \
	KAFKA_BROKERS=localhost:19092 \
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
	go run ./cmd/orchestrator

run-readmodel:
	cd services/read-model && \
	PORT=7104 \
	DATABASE_URL=postgres://ledger:ledgerpw@localhost:5435/readmodel?sslmode=disable \
	KAFKA_BROKERS=localhost:19092 \
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
	go run ./cmd/readmodel

run-gateway:
	cd services/gateway && \
	PORT=4000 \
	ACCOUNTS_SERVICE_URL=http://localhost:7101 \
	ORCHESTRATOR_SERVICE_URL=http://localhost:7103 \
	READMODEL_SERVICE_URL=http://localhost:7104 \
	OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
	npm start
endif