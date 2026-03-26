CREATE TABLE notification_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_type  VARCHAR(20) NOT NULL,
    recipient     TEXT NOT NULL,
    subject       TEXT DEFAULT '',
    payload_json  JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count   INTEGER NOT NULL DEFAULT 0,
    max_retries   INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error    TEXT DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_jobs_pending ON notification_jobs(status, next_retry_at)
  WHERE status IN ('pending', 'retrying');
