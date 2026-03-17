CREATE TABLE access_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id       UUID         NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    requester_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requester_type  VARCHAR(20)  NOT NULL CHECK (requester_type IN ('human', 'agent')),
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'revoked')),
    reason          TEXT,
    duration        INTERVAL     NOT NULL DEFAULT '1 hour',
    decided_by      UUID         REFERENCES users(id),
    decided_at      TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_access_requests_secret_id ON access_requests (secret_id);
CREATE INDEX idx_access_requests_requester_id ON access_requests (requester_id);
CREATE INDEX idx_access_requests_status ON access_requests (status);
CREATE INDEX idx_access_requests_pending ON access_requests (id) WHERE status = 'pending';
