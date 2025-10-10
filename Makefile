SHELL := /bin/bash
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
run-accounts:
	cd services/accounts && go run ./cmd/accounts


run-ledger:
	cd services/ledger && go run ./cmd/ledger


run-orchestrator:
	cd services/posting-orchestrator && go run ./cmd/orchestrator


run-readmodel:
	cd services/read-model && go run ./cmd/readmodel


run-gateway:
	cd services/gateway && npm start


# Start everything (infra + services)
run-all: up services-up logs