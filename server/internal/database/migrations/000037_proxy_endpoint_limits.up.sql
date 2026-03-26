CREATE TABLE proxy_endpoint_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_identities(id) ON DELETE CASCADE,
    host_pattern TEXT NOT NULL,
    path_pattern TEXT NOT NULL DEFAULT '/*',
    rpm INTEGER NOT NULL DEFAULT 60,
    blocked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, host_pattern, path_pattern)
);

CREATE INDEX idx_proxy_endpoint_limits_agent ON proxy_endpoint_limits(agent_id);
