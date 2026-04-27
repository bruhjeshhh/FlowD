CREATE TABLE IF NOT EXISTS jobs (
    id uuid PRIMARY KEY,
    payload json NOT NULL,
    status TEXT CHECK (status IN ('pending', 'processing', 'success', 'failed', 'cancelled')),
    type TEXT CHECK (type IN ('email', 'sms', 'push_notification')) NOT NULL,
    retry_count int NOT NULL DEFAULT 0,
    max_retries int NOT NULL DEFAULT 3,
    idempotency_key varchar NOT NULL UNIQUE,
    scheduled_at timestamp,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    next_run_at timestamp DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webhooks (
    id uuid PRIMARY KEY,
    url text NOT NULL,
    job_type text NOT NULL,
    event text CHECK (event IN ('job_success', 'job_failed')) NOT NULL,
    secret text,
    created_at timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_job_type_event ON webhooks(job_type, event);
