-- +goose up
create table jobs(id uuid primary key,
payload json not null,
 status TEXT CHECK (status IN ('pending', 'processing','success','failed')),
 type TEXT CHECK (type IN ('email','sms','push_notification')) not null,
retry_count int not null default 0,
max_retries int not null default 3,
idempotency_key varchar not null unique,
scheduled_at timestamp,
created_at timestamp not null default now(),
updated_at timestamp not null default now(),
next_run_at timestamp default now());

create table dead_letter_jobs(
id uuid primary key,
job_id uuid not null,
payload json not null,
status TEXT CHECK (status IN ('dead')) not null,
type TEXT NOT NULL,
retry_count int not null,
max_retries int not null,
idempotency_key varchar,
scheduled_at timestamp,
created_at timestamp not null default now(),
failure_reason TEXT,
original_error TEXT);
-- +goose down 
Drop table jobs;
