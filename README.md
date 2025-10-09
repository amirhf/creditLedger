# Credit Ledger (Go + TS)


## Quickstart
1. Copy `.env.example` to `.env` and adjust if needed.
2. Start infra: `make up`
3. Generate code: `make proto sqlc`
4. Build services: `make build`
5. Tail logs: `make e2e-logs`


Infra: Redpanda (Kafka API) + Console, Postgres x3, Redis, Jaeger, Prometheus, Grafana.


Services (Go): accounts, ledger, posting-orchestrator, read-model
Gateway (TS/Nest): REST API and OpenAPI docs.


## Useful URLs
- Redpanda Console: http://localhost:8080
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686
- Gateway API: http://localhost:4000