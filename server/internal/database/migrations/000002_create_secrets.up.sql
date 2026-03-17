CREATE TABLE secrets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    storage_key   TEXT         NOT NULL,
    encrypted_dek BYTEA        NOT NULL,
    policy        JSONB        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_secrets_user_id ON secrets (user_id);
CREATE INDEX idx_secrets_not_deleted ON secrets (id) WHERE deleted_at IS NULL;
