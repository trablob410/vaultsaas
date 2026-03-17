CREATE TABLE dynamic_leases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES dynamic_providers(id) ON DELETE CASCADE,
    access_request_id UUID REFERENCES access_requests(id),
    agent_id UUID REFERENCES agent_identities(id),
    secret_data_enc BYTEA NOT NULL,
    ttl_seconds INT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_dynamic_leases_expiry ON dynamic_leases(expires_at) WHERE revoked_at IS NULL;
CREATE INDEX idx_dynamic_leases_provider ON dynamic_leases(provider_id);
