SHELL := /bin/bash


.PHONY: up down build gen proto sqlc test


up:
	docker compose -f deploy/docker-compose.yml up -d --build
	docker ps


e2e-logs:
	docker compose -f deploy/docker-compose.yml logs -f --tail=200


down:
	docker compose -f deploy/docker-compose.yml down -v


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