CREATE TABLE IF NOT EXISTS jobs (
    id uuid PRIMARY KEY,
    payload json NOT NULL,
    status TEXT CHECK (status IN ('pending', 'processing', 'success', 'failed')),
    type TEXT CHECK (type IN ('email', 'sms', 'push_notification')) NOT NULL,
    retry_count int NOT NULL DEFAULT 0,
    max_retries int NOT NULL DEFAULT 3,
    idempotency_key varchar NOT NULL UNIQUE,
    scheduled_at timestamp,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    next_run_at timestamp DEFAULT now()
);
