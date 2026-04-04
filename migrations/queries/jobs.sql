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

-- name: GetJobByScheduledAt :one
update jobs
set status = 'processing', updated_at = now()
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending'
    ORDER BY scheduled_at
    LIMIT 1
)
returning id;

-- name: UpdateJobStatus :exec
update jobs
set status = 'success', updated_at = now()
WHERE id = $1;