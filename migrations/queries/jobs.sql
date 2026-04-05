-- name: InsertJob :one
INSERT INTO jobs(id, payload, status, retry_count, max_retries, idempotency_key, scheduled_at,  created_at, updated_at,next_run_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;

-- name: GetJobByScheduledAt :one
update jobs
set status = 'processing', updated_at = now()
WHERE id = (
    SELECT id FROM jobs
    WHERE status = 'pending' AND (next_run_at IS NULL OR next_run_at >= now())
    ORDER BY next_run_at ASC
    For update SKIP LOCKED
    LIMIT 1
)
returning id;

-- name: UpdateJobStatusSuccess :exec
update jobs
set status = 'success', updated_at = now(), next_run_at = NULL
WHERE id = $1;

-- name: UpdateJobStatusNotSuccess :exec
UPDATE jobs
SET 
    status = CASE 
        WHEN retry_count < max_retries THEN 'pending' 
        ELSE 'failed' 
    END,
    next_run_at = CASE 
        WHEN retry_count < max_retries THEN NOW() + INTERVAL '5 seconds' 
        ELSE NULL 
    END,
    updated_at = NOW()
WHERE id = $1;
-- name: IncrementRetryCount :one
UPDATE jobs
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE id = $1
returning retry_count;


-- name: GetStuckJobs :one
SELECT id FROM jobs
WHERE status = 'processing' AND updated_at < NOW() - INTERVAL '1 minute'
FOR UPDATE SKIP LOCKED
limit 1;

-- name: ResetStuckJob :exec
UPDATE jobs
SET status = 'pending', updated_at = NOW(), next_run_at = NOW() + INTERVAL '5 seconds'
WHERE id = $1;  