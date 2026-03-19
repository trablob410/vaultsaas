CREATE TABLE request_action_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id  UUID NOT NULL REFERENCES access_requests(id) ON DELETE CASCADE,
    action      TEXT NOT NULL CHECK (action IN ('approve', 'reject')),
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_action_tokens_request_id ON request_action_tokens(request_id);
