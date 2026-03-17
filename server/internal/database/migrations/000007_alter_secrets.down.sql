DROP INDEX IF EXISTS idx_secrets_source;
DROP INDEX IF EXISTS idx_secrets_credential_type;
ALTER TABLE secrets
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS credential_type;
