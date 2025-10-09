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
go build ./...
npm --prefix services/gateway ci && npm --prefix services/gateway run build


proto:
buf generate


sqlc:
find services -name sqlc.yaml -execdir sqlc generate \; || true


test:
go test ./...