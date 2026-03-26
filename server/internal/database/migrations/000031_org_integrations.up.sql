CREATE TABLE org_integrations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL,
    access_token_enc BYTEA NOT NULL,
    workspace_id    VARCHAR(100),
    team_name       VARCHAR(255),
    bot_user_id     VARCHAR(100),
    metadata_json   JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, provider)
);

CREATE INDEX idx_org_integrations_org ON org_integrations(org_id);
