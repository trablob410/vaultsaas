CREATE TABLE cli_auth_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token      TEXT,              -- set when login completes
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '5 minutes',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
