CREATE TABLE IF NOT EXISTS outbox (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type text NOT NULL,
    aggregate_id text NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    attempts int NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_error text NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz NULL
);

CREATE INDEX IF NOT EXISTS outbox_status_next_attempt_idx ON outbox(status, next_attempt_at);

