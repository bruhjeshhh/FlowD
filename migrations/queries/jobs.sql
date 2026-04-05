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
    For update SKIP LOCKED
    LIMIT 1
)
returning id;

-- name: UpdateJobStatusSuccess :exec
update jobs
set status = 'success', updated_at = now()
WHERE id = $1;

-- name: UpdateJobStatusNotSuccess :exec
UPDATE jobs
SET 
    status = CASE 
        WHEN retry_count < max_retries THEN 'pending' 
        ELSE 'failed' 
    END,
    updated_at = NOW()
WHERE id = $1;

-- name: IncrementRetryCount :one
UPDATE jobs
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE id = $1
returning retry_count;
