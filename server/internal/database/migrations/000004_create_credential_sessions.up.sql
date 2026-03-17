CREATE TABLE credential_sessions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_request_id UUID         NOT NULL REFERENCES access_requests(id) ON DELETE CASCADE,
    status            VARCHAR(20)  NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'expired', 'revoked')),
    expires_at        TIMESTAMPTZ  NOT NULL,
    usage_count       INTEGER      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_credential_sessions_request ON credential_sessions (access_request_id);
CREATE INDEX idx_credential_sessions_active ON credential_sessions (id) WHERE status = 'active';
