ALTER TABLE credential_sessions
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE credential_sessions
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS credential_type;
