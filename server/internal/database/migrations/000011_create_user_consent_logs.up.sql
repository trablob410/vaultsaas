CREATE TABLE user_consent_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret_id       UUID         NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    credential_type VARCHAR(50)  NOT NULL,
    consent_type    VARCHAR(50)  NOT NULL CHECK (consent_type IN ('access_grant', 'risk_acknowledgment')),
    granted         BOOLEAN      NOT NULL DEFAULT false,
    ip_address      INET,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_consent_logs_user_id ON user_consent_logs (user_id);
CREATE INDEX idx_user_consent_logs_secret_id ON user_consent_logs (secret_id);
