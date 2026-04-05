-- +goose up
create table jobs(id uuid primary key,
payload json not null,
 status TEXT CHECK (status IN ('pending', 'processing','success','failed')),
retry_count int not null default 0,
max_retries int not null default 3,
idempotency_key varchar not null unique,
scheduled_at timestamp,
created_at timestamp not null default now(),
updated_at timestamp not null default now(),
next_run_at timestamp not null default now());
-- +goose down 
Drop table jobs;
