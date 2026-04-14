-- name: InsertJob :one
INSERT INTO jobs(
    id, payload, status, type, retry_count, max_retries, idempotency_key,
    scheduled_at, created_at, updated_at, next_run_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetJobByIdempotencyKey :one
SELECT * FROM jobs
WHERE idempotency_key = $1
LIMIT 1;

-- name: GetJobByID :one
SELECT * FROM jobs
WHERE id = $1
LIMIT 1;

-- name: ListJobsByStatus :many
SELECT * FROM jobs
WHERE status = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetJobByScheduledAt :one
UPDATE jobs
SET status = 'processing', updated_at = now()
WHERE id = (
    SELECT j.id FROM jobs j
    WHERE j.status = 'pending'
      AND (j.scheduled_at IS NULL OR j.scheduled_at <= now())
      AND (j.next_run_at IS NULL OR j.next_run_at <= now())
    ORDER BY j.next_run_at ASC NULLS LAST, j.created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;

-- name: UpdateJobStatusSuccess :exec
UPDATE jobs
SET status = 'success', updated_at = now(), next_run_at = NULL
WHERE id = $1;

-- name: UpdateJobStatusNotSuccess :exec
UPDATE jobs
SET
    retry_count = jobs.retry_count + 1,
    status = CASE
        WHEN jobs.retry_count + 1 < jobs.max_retries THEN 'pending'
        ELSE 'failed'
    END,
    next_run_at = CASE
        WHEN jobs.retry_count + 1 < jobs.max_retries THEN NOW() + (
            INTERVAL '1 second' * (
                5 * POWER(2::numeric, LEAST(jobs.retry_count::numeric, 30))
            )
        )
        ELSE NULL
    END,
    updated_at = NOW()
WHERE id = $1;

-- name: IncrementRetryCount :one
UPDATE jobs
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE id = $1
RETURNING retry_count;

-- name: GetStuckJobs :one
SELECT id FROM jobs
WHERE status = 'processing' AND updated_at < NOW() - INTERVAL '1 minute'
FOR UPDATE SKIP LOCKED
LIMIT 1;

-- name: ResetStuckJob :exec
UPDATE jobs
SET status = 'pending', updated_at = NOW(), next_run_at = NOW() + INTERVAL '5 seconds'
WHERE id = $1;

-- name: CountJobsByStatus :many
SELECT status, COUNT(*) as count FROM jobs GROUP BY status;

-- Create DLQ table
-- name: CreateDeadLetterJob :one
INSERT INTO dead_letter_jobs(
    id, job_id, payload, status, type, retry_count, max_retries,
    idempotency_key, scheduled_at, created_at, failure_reason, original_error
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListDeadLetterJobs :many
SELECT * FROM dead_letter_jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetDeadLetterJobByID :one
SELECT * FROM dead_letter_jobs
WHERE id = $1
LIMIT 1;

-- name: CancelJob :one
UPDATE jobs
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1 AND status IN ('pending', 'processing')
RETURNING *;

-- name: CountDeadLetterJobs :one
SELECT COUNT(*) as count FROM dead_letter_jobs;

-- name: DeleteDeadLetterJob :exec
DELETE FROM dead_letter_jobs WHERE id = $1;

-- name: CleanupOldJobs :exec
DELETE FROM jobs
WHERE (status = 'success' OR status = 'cancelled')
  AND updated_at < NOW() - INTERVAL '1 hour' * $1;

-- name: CleanupOldDeadLetterJobs :exec
DELETE FROM dead_letter_jobs
WHERE created_at < NOW() - INTERVAL '1 hour' * $1;
