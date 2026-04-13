# FlowD

Postgres-backed **job queue** in Go: HTTP enqueue, background workers that claim jobs with `FOR UPDATE SKIP LOCKED`, retries with backoff, and a rescuer for stuck `processing` rows. FlowD is a lightweight, Postgres-backed job queue designed for reliability without external dependencies like Kafka or Redis. It demonstrates safe concurrent job processing, retries with backoff, and observability using production-grade patterns.

## Features

- **POST /jobs** — enqueue JSON payload with `type` (`email`, `sms`, `push_notification`). Optional `scheduled_at` (ISO 8601 datetime for delayed execution), optional `idempotency_key`.
- **Idempotent enqueue** — same `idempotency_key` returns **200** with `idempotent_replay: true` and the existing job; first create returns **201**. Implemented with a **database transaction** (read key → insert; handles races via unique constraint).
- **GET /jobs/{id}** — inspect a job.
- **GET /jobs?status=failed** — **dead-letter queue**: list terminal failed jobs (`limit` default 50, max 200; `offset` for paging). Use for ops dashboards or manual replay workflows.
- **GET /health** — liveness; pings Postgres.
- **GET /metrics** — Prometheus metrics endpoint.

Workers are stubs (log + success) but the pipeline is real: claim → process → `success` / retry / `failed`.

**Structured logs** — the process uses `log/slog` with JSON to stdout. Workers and the rescuer attach `worker_id`, `job_id`, and `job_type` where applicable so you can trace a job through claim → success or retry.

## Usage

### Enqueue a job

```bash
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "payload": {
      "type": "email",
      "data": {
        "to": "user@example.com",
        "subject": "Hello",
        "body": "Your message here"
      }
    }
  }'
```

Response (201 Created):

```json
{
  "job": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "email",
    "retry_count": 0,
    "max_retries": 3,
    "status": "pending"
  },
  "idempotent_replay": false
}
```

### Enqueue with idempotency key

Use `idempotency_key` to prevent duplicate jobs:

```bash
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "idempotency_key": "welcome-email-user-123",
    "payload": {
      "type": "email",
      "data": { "to": "user@example.com", "subject": "Welcome!", "body": "..." }
    }
  }'
```

Repeating the same request returns 200 with `idempotent_replay: true` instead of creating a duplicate.

### Schedule a delayed job

```bash
curl -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "scheduled_at": "2026-04-13T15:00:00Z",
    "payload": {
      "type": "email",
      "data": { "to": "user@example.com", "subject": "Reminder", "body": "..." }
    }
  }'
```

### Batch enqueue

Submit up to 100 jobs at once:

```bash
curl -X POST http://localhost:8080/jobs/batch \
  -H 'Content-Type: application/json' \
  -d '{
    "jobs": [
      {
        "idempotency_key": "batch-1",
        "payload": { "type": "email", "data": { "to": "a@b.com", "subject": "Hi", "body": "A" } }
      },
      {
        "idempotency_key": "batch-2",
        "payload": { "type": "email", "data": { "to": "c@d.com", "subject": "Hi", "body": "B" } }
      }
    ]
  }'
```

### Get job status

```bash
curl http://localhost:8080/jobs/550e8400-e29b-41d4-a716-446655440000
```

### List dead-letter queue (failed jobs)

```bash
curl 'http://localhost:8080/jobs?status=failed&limit=50&offset=0'
```

### Replay a failed job

```bash
curl -X POST http://localhost:8080/jobs/550e8400-e29b-41d4-a716-446655440000/replay
```

### Cancel a job

```bash
curl -X DELETE http://localhost:8080/jobs/550e8400-e29b-41d4-a716-446655440000
```

**Prometheus metrics** at `GET /metrics`:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `flowd_jobs_enqueued_total` | Counter | `type` | Total jobs enqueued |
| `flowd_jobs_processed_total` | Counter | `type`, `status` | Total jobs processed |
| `flowd_job_duration_seconds` | Histogram | `type` | Job processing duration |
| `flowd_jobs_in_queue` | Gauge | `status` | Jobs in queue by status |
| `flowd_workers_active` | Gauge | - | Active workers |
| `flowd_http_requests_total` | Counter | `method`, `path`, `status` | HTTP requests |
| `flowd_http_request_duration_seconds` | Histogram | `method`, `path` | HTTP request duration |

## Requirements

- Go **1.23+**
- **PostgreSQL** 16+ (or use Docker Compose)
- [sqlc](https://sqlc.dev/) if you change queries under `migrations/queries/`

## Quick Start (Docker)

```bash
docker compose up --build
```

Then:

```bash
curl -sS -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "idempotency_key": "demo-1",
    "payload": {
      "type": "email",
      "data": { "to": "you@example.com", "subject": "hi", "body": "hello" }
    }
  }' | jq .

# Delayed job (runs in 1 hour)
curl -sS -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "idempotency_key": "demo-delayed",
    "scheduled_at": "2026-04-08T15:00:00Z",
    "payload": {
      "type": "email",
      "data": { "to": "you@example.com", "subject": "reminder", "body": "hello" }
    }
  }' | jq .

curl -sS http://localhost:8080/health | jq .

curl -sS 'http://localhost:8080/jobs?status=failed&limit=10' | jq .
```

Repeat the same `idempotency_key` to see `idempotent_replay: true`.

## Local run (Postgres already up)

Apply schema once (matches Compose init script):

```bash
psql "$DB_URL" -f scripts/schema.sql
export DB_URL='postgres://user:pass@localhost:5432/dbname?sslmode=disable'
go run .
```

Optional: `WORKER_COUNT` (default `4`).

## Environment

| Variable | Required | Description |
|----------|----------|-------------|
| `DB_URL` | yes | PostgreSQL URL (e.g. `postgres://user:pass@host:5432/dbname?sslmode=disable`) |
| `WORKER_COUNT` | no | Number of worker goroutines (default `4`) |

## Architecture

Clients enqueue and query jobs over HTTP. Workers and the rescuer loop independently and coordinate only through Postgres: workers claim the next eligible row with `FOR UPDATE SKIP LOCKED`; the rescuer moves stuck `processing` rows back to `pending`.

```mermaid
flowchart LR
  client[Client] --> httpAPI[HTTP_API]
  httpAPI --> postgres[(Postgres)]
  workers[Workers] --> postgres
  rescuer[Rescuer] --> postgres
```

## Design notes

| Topic | Choice |
|--------|--------|
| Safe concurrency | `FOR UPDATE SKIP LOCKED` so many workers can dequeue without double-claim |
| Delivery | At-least-once; combine with **idempotent handlers** in production |
| Retries | `retry_count` incremented on failure; after `max_retries` the job is `failed` (dead-letter) |
| Exponential backoff | Before terminal failure, `next_run_at = now + (5s × 2^n)` where `n` is the **current** `retry_count` before the failing attempt (so first retry waits 5s, then 10s, 20s, …). The exponent is capped at 30 to bound the delay. |
| Dead-letter queue | Jobs in `failed` stay in the same table; list them with **GET /jobs?status=failed** for inspection or to drive a separate replay process. |
| Stuck jobs | Rescuer resets `processing` rows idle longer than 1 minute |
| Delayed jobs | Jobs with `scheduled_at` in the future are held until that time (worker filters with `scheduled_at <= now()` and `next_run_at <= now()`) |

SQL is generated with **sqlc** from `migrations/queries/jobs.sql`. Goose-style files under `migrations/schema/` document migrations; `scripts/schema.sql` is a single-file bootstrap for Docker and CI.

## Tests

```bash
go vet ./...
go test ./... -count=1
```

Integration (needs `DB_URL` and schema applied):

```bash
export DB_URL='postgres://postgres:postgres@localhost:5432/flowD_db?sslmode=disable'
go test -tags=integration ./... -count=1
```

## Motivation

FlowD was built to provide a lightweight, Postgres-based job queue that leverages database features for reliability without the operational overhead of external systems like Redis or RabbitMQ. It embraces simplicity: if you already run Postgres, you get durable job queuing with strong consistency guarantees.

Key motivations:
- **Zero external dependencies** — uses Postgres as the single source of truth
- **Safe concurrency** — `FOR UPDATE SKIP LOCKED` prevents race conditions without advisory locks
- **Observability** — structured JSON logs and Prometheus metrics out of the box
- **Operational simplicity** — dead-letter queue, rescuer for stuck jobs, and idempotent enqueue built in

## Contributing

Contributions are welcome. To get started:

1. **Fork** the repository and create a branch from `main`
2. **Write tests** for any new functionality (`go test ./... -count=1`)
3. **Run linting** with `go vet ./...`
4. **Keep commits atomic** — one feature or fix per commit
5. **Follow existing code style** — this project uses standard Go conventions

For larger changes, please open an issue first to discuss the proposed changes.

### Adding new job types

To add a new job type (e.g., `push_notification`):

1. Add a handler function in `internal/worker/` (e.g., `handlepushnotifications.go`)
2. Register it in `internal/worker/jobhandler.go`
3. Add tests in the same package

### Database changes

If you modify SQL queries:

1. Update `migrations/queries/jobs.sql`
2. Run `sqlc generate` to regenerate `internal/database/jobs.sql.go`
3. Review the generated code before committing

