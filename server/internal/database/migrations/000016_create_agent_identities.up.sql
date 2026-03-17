CREATE TABLE agent_identities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    agent_type      VARCHAR(100) NOT NULL DEFAULT 'custom',
    auth_method     VARCHAR(50) NOT NULL DEFAULT 'token',
    allowed_scopes  TEXT[] NOT NULL DEFAULT '{}',
    max_session_ttl INT NOT NULL DEFAULT 3600,
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_agent_identities_project ON agent_identities(project_id);
CREATE INDEX idx_agent_identities_status ON agent_identities(status);
