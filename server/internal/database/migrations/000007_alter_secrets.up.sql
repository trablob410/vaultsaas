-- Add credential_type, source, version to secrets
ALTER TABLE secrets
    ADD COLUMN credential_type VARCHAR(50) NOT NULL DEFAULT 'api_key'
        CHECK (credential_type IN ('api_key', 'db_credential', 'ssh_key', 'oauth_token', 'cloud_credential', 'personal_session')),
    ADD COLUMN source          VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN version         INTEGER      NOT NULL DEFAULT 1;

CREATE INDEX idx_secrets_credential_type ON secrets (credential_type);
CREATE INDEX idx_secrets_source ON secrets (source) WHERE source != '';
