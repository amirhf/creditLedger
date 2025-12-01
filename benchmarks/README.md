# Benchmarks

This directory contains scripts to load test the Credit Ledger architecture.

## Methodology

We use [k6](https://k6.io/) to simulate traffic.
The standard test (`script.js`):
1.  Creates two accounts in the `setup()` phase.
2.  Spins up concurrent virtual users (VUs).
3.  Each VU continuously posts transfers between the two accounts with unique `idempotencyKey`s.
4.  We measure the HTTP response time (end-to-end latency) and success rate.

## Prerequisites

*   Docker
*   The stack must be running (`make up`)

## Running the Benchmark (via Docker)

You can run k6 without installing it locally using the official Docker image:

```bash
# Run the test connecting to the host's network
docker run --rm -i \
  --network="host" \
  -v $(pwd)/benchmarks/script.js:/script.js \
  grafana/k6 run /script.js
```

*Note: On Mac/Windows, `network="host"` might have issues connecting to `localhost`. If so, replace `localhost` with `host.docker.internal` in the script or pass it via env var:*

```bash
docker run --rm -i \
  -v $(pwd)/benchmarks/script.js:/script.js \
  -e BASE_URL=http://host.docker.internal:4000 \
  grafana/k6 run /script.js
```

## Interpretation

*   **http_req_duration**: The time taken for the API to accept the request, write to the Ledger DB, and return.
*   **http_reqs**: Total requests processed (Throughput).

## Monitoring

While the test runs, check:
1.  **Grafana:** See if `Consumer Lag` spikes. If lag grows indefinitely, the Kafka consumers (read model / projections) cannot keep up with the producer.
2.  **CPU/Memory:** Check Docker stats (`docker stats`) to see if any service is bottlenecking.
