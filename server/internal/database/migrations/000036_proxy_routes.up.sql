CREATE TABLE proxy_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_identities(id) ON DELETE CASCADE,
    host_pattern TEXT NOT NULL,
    path_pattern TEXT NOT NULL DEFAULT '/*',
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    injection_type TEXT NOT NULL DEFAULT 'header',
    injection_key TEXT NOT NULL DEFAULT 'Authorization',
    injection_format TEXT NOT NULL DEFAULT 'Bearer {value}',
    placeholder_key TEXT UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, host_pattern, path_pattern)
);

CREATE INDEX idx_proxy_routes_agent ON proxy_routes(agent_id) WHERE enabled = true;
CREATE INDEX idx_proxy_routes_placeholder ON proxy_routes(placeholder_key) WHERE placeholder_key IS NOT NULL;
