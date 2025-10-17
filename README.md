
# Credit Ledger (Go + TS + Kafka + Postgres)

> Event-driven, **double-entry** credit ledger with **transactional outbox**, **idempotent consumers**, replayable **CQRS** projections, and full **observability** (Jaeger + Grafana + Redpanda Console).

[![Go](https://img.shields.io/badge/Go-1.22+-blue)](#)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-blue)](#)
[![CI](https://img.shields.io/badge/CI-GitHub%20Actions-green)](#)
[![Docker](https://img.shields.io/badge/Docker-Compose-success)](#)
[![License](https://img.shields.io/badge/license-MIT-lightgrey)](#)

**Live demo:** [https://creditledger-gateway-staging.fly.dev/api](https://creditledger-gateway-staging.fly.dev/api)

**Repo:** [https://github.com/amirhf/creditLedger](https://github.com/amirhf/creditLedger)

---

Infra: Redpanda (Kafka API) + Console, Postgres x3, Redis, Jaeger, Prometheus, Grafana.


Services (Go): accounts, ledger, posting-orchestrator, read-model
Gateway (TS/Nest): REST API and OpenAPI docs.


## Useful URLs
- Redpanda Console: http://localhost:8080
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686
- Gateway API: http://localhost:4000

## Quickstart (local, one command)

```bash
cp .env.example .env
make up                # Build and start everything (infra + services)
make logs              # Tail logs (optional)
```

**Pro tip:** If you've already built the images, use `make up-nobuild` to start faster without rebuilding.

**Useful URLs**

| Tool                     | URL                                                                                     | Notes                     |
| ------------------------ |-----------------------------------------------------------------------------------------| ------------------------- |
| Gateway                  | [http://localhost:4000](http://localhost:4000)                                          |                           |
| **Gateway – Swagger UI** | **[http://localhost:4000/api](http://localhost:4000/api)**                              | **Primary “Try it” path** |
| Redpanda Console         | [http://localhost:8080](http://localhost:8080)                                          |                           |
| Jaeger                   | [http://localhost:16686](http://localhost:16686)                                        |                           |
| Grafana                  | [http://localhost:3000](http://localhost:3000) (admin/admin)                            |                           |

> **Development workflows:**
> - Full stack: `make up` (or `make up-nobuild` if already built)
> - Infra only: `make infra-up` then run services locally with `make run-*`
> - Quick restart: `make restart` (no rebuild)
> - Force rebuild: `make rebuild`
> - See all options: `make help`

---

## Try it now — **Swagger UI** (no tooling needed)

1. Open **Swagger UI** → **[http://localhost:4000/api](http://localhost:4000/api)**
2. **POST `/accounts`** twice → copy the two returned `accountId` values as **A** and **B**.
3. **POST `/transfers`** with:

    * `from`: **A**
    * `to`: **B**
    * `amount`: `1234` (minor units)
    * `currency`: `USD`
    * `idempotencyKey`: `demo-1`
4. **POST `/transfers`** again with the **same** `idempotencyKey` → observe **exactly-once effect** (no double credit).
5. **GET `/accounts/{id}/balance`** for **B** → expect `1234`.
6. (Optional) **GET `/accounts/{id}/statements`** to view lines.

Then jump into the ops views:

* **Jaeger** → find the latest trace (Gateway → Orchestrator → Ledger → Outbox → Read-model).
* **Redpanda Console** → topics `ledger.entry.v1`, `ledger.transfer.v1`.
* **Grafana** → panels: **Outbox Age**, **Consumer Lag**, **p95 latency**.

---

## Try it with cURL (backup to Swagger)

```bash
A=$(curl -s -XPOST :4000/accounts -H 'content-type: application/json' -d '{"currency":"USD"}' | jq -r .accountId)
B=$(curl -s -XPOST :4000/accounts -H 'content-type: application/json' -d '{"currency":"USD"}' | jq -r .accountId)

curl -s -XPOST :4000/transfers \
  -H 'content-type: application/json' \
  -d "{\"from\":\"$A\",\"to\":\"$B\",\"amount\":1234,\"currency\":\"USD\",\"idempotencyKey\":\"demo-1\"}" | jq

curl -s :4000/accounts/$B/balance | jq
```

*Postman users:* import `Gateway-API.postman_collection.json` (if present) and run **Create → Transfer → Balance**.

---

## How it works (in 60 seconds)

* **Gateway** (HTTP/Swagger) accepts requests and publishes commands.
* **Ledger** applies domain rules (double-entry) and writes to the **transactional outbox**.
* **Outbox daemon** relays events to Kafka/Redpanda ensuring at-least-once delivery without dual-write hazards.
* **Consumers** are **idempotent** and feed **read models** (CQRS) for fast queries.
* **Observability** via Jaeger traces across services, metrics in Grafana, and event inspection in Redpanda Console.

---

## Troubleshooting

* **Jaeger empty:** wait a few seconds after a transfer; refresh.
* **Redpanda Console shows no topics:** services may still be starting; perform a transfer to produce events.
* **DB connection refused:** ensure all three Postgres containers are healthy (`docker ps`), then retry `make services-up`.
* **Amount units:** All amounts are **minor units** (`int64`). `1234` = `$12.34` for USD.
