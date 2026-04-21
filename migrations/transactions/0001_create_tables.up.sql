create table if not exists transactions (
	id uuid not null primary key default gen_random_uuid(),
	external_id uuid not null unique,
	account_id uuid not null,
	amount decimal(10, 2) not null,
	balance decimal(10, 2) not null,
	transaction_type text not null,
	created_at timestamp not null default current_timestamp
);

create table if not exists balances (
	account_id uuid not null primary key,
	current_balance decimal(10, 2) not null default 0.00,
	updated_at timestamp not null default current_timestamp,
	created_at timestamp not null default current_timestamp
);

create table if not exists outbox_events (
	id bigserial primary key,
	account_id uuid not null,
	transaction_id uuid not null unique,
	payload jsonb not null,
	processed boolean default false,
	created_at timestamp not null default current_timestamp
);

CREATE INDEX idx_outbox_unprocessed ON outbox_events (created_at) WHERE processed = false;