# FlowD

Postgres-backed **job queue** in Go: HTTP enqueue, background workers that claim jobs with `FOR UPDATE SKIP LOCKED`, retries with backoff, and a rescuer for stuck `processing` rows.

## Features

- **POST /jobs** — enqueue JSON payload with `type` (`email`, `sms`, `push_notification`). Optional `scheduled_at`, optional `idempotency_key`.
- **Idempotent enqueue** — same `idempotency_key` returns **200** with `idempotent_replay: true` and the existing job; first create returns **201**. Implemented with a **database transaction** (read key → insert; handles races via unique constraint).
- **GET /jobs/{id}** — inspect a job.
- **GET /health** — liveness; pings Postgres.

Workers are stubs (log + success) but the pipeline is real: claim → process → `success` / retry / `failed`.

## Requirements

- Go **1.23+**
- **PostgreSQL** 16+ (or use Docker Compose)
- [sqlc](https://sqlc.dev/) if you change queries under `migrations/queries/`

## Quick start (Docker)

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

curl -sS http://localhost:8080/health | jq .
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

## Design notes

| Topic | Choice |
|--------|--------|
| Safe concurrency | `FOR UPDATE SKIP LOCKED` so many workers can dequeue without double-claim |
| Delivery | At-least-once; combine with **idempotent handlers** in production |
| Retries | `retry_count` incremented on failure; `next_run_at` backoff before `failed` |
| Stuck jobs | Rescuer resets `processing` rows idle longer than 1 minute |

SQL is generated with **sqlc** from `migrations/queries/jobs.sql`. Goose-style files under `migrations/schema/` document migrations; `scripts/schema.sql` is a single-file bootstrap for Docker and CI.

## Tests

```bash
go test ./... -count=1
```

Integration (needs `DB_URL` and schema applied):

```bash
export DB_URL='postgres://postgres:postgres@localhost:5432/flowD_db?sslmode=disable'
go test -tags=integration ./... -count=1
```

## License

MIT (or your choice — add a `LICENSE` file if you need one).
