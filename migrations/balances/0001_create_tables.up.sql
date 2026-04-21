create table if not exists balances (
	account_id uuid not null primary key,
	current_balance decimal(10, 2) not null default 0.00 CHECK (current_balance >= 0),
	updated_at timestamp not null default current_timestamp
);

CREATE TABLE IF NOT EXISTS inbox_events (
	id             bigserial primary key,
	account_id     uuid not null,
	transaction_id uuid not null unique,
	payload        jsonb not null,
	created_at     timestamp not null default current_timestamp
);
