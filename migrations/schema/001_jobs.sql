-- +goose up
create table jobs(id uuid primary key,
payload json not null,
 Status ENUM('Pending', 'Processing','Success','Failed') ,
retry_count int not null default 0,
max_retries int not null default 3,
idempotency_key varchar not null unique,
scheduled_at timestamp  default now(),
created_at timestamp not null default now(),
updated_at timestamp not null default now());
-- +goose down 
Drop table jobs
