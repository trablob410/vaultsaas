-- Add credential_type and revoked_at, drop updated_at
ALTER TABLE credential_sessions
    ADD COLUMN credential_type VARCHAR(50),
    ADD COLUMN revoked_at      TIMESTAMPTZ;

ALTER TABLE credential_sessions
    DROP COLUMN IF EXISTS updated_at;
