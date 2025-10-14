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


.PHONY: up down build gen proto sqlc test logs services-up services-down \
docker-build docker-push run-accounts run-ledger run-orchestrator run-readmodel run-gateway run-all

up:
	$(COMPOSE) up -d --build
	docker ps

services-up:
	$(COMPOSE) up -d --build accounts ledger-svc orchestrator read-model gateway

logs:
	$(COMPOSE) logs -f --tail=200


down:
	$(COMPOSE) down -v


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


# Start everything (infra + services)
run-all: up services-up logs