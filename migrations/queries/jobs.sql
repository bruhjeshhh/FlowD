-- name: InsertJob :one
INSERT INTO jobs(id, payload, status, retry_count, max_retries, idempotency_key, scheduled_at,  created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
RETURNING *;