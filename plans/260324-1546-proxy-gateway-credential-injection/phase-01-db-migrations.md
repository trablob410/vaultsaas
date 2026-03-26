---
phase: 1
title: "DB Migrations — proxy_routes + proxy_endpoint_limits"
status: pending
priority: P1
effort: 2h
---

# Phase 1: DB Migrations

## Overview

Create two tables: `proxy_routes` (credential injection rules) and `proxy_endpoint_limits` (per-agent per-endpoint rate limits).

## Context Links
- [Existing migrations](../../server/internal/database/migrations/) — currently at 000035
- [Agent tables](../../server/internal/agent/service.go) — `agent_identities`, `agent_tokens`
- [Secrets table](../../server/internal/vault/service.go) — `secrets`

## Migration 000036: proxy_routes

```sql
-- 000036_proxy_routes.up.sql
CREATE TABLE proxy_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_identities(id) ON DELETE CASCADE,
    host_pattern TEXT NOT NULL,                    -- e.g. "api.openai.com"
    path_pattern TEXT NOT NULL DEFAULT '/*',       -- e.g. "/v1/*"
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    injection_type TEXT NOT NULL DEFAULT 'header', -- header | query | bearer
    injection_key TEXT NOT NULL DEFAULT 'Authorization',
    injection_format TEXT NOT NULL DEFAULT 'Bearer {value}',
    placeholder_key TEXT UNIQUE,                   -- auto-generated fake key
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, host_pattern, path_pattern)
);

CREATE INDEX idx_proxy_routes_agent ON proxy_routes(agent_id) WHERE enabled = true;
CREATE INDEX idx_proxy_routes_placeholder ON proxy_routes(placeholder_key) WHERE placeholder_key IS NOT NULL;
```

```sql
-- 000036_proxy_routes.down.sql
DROP TABLE IF EXISTS proxy_routes;
```

## Migration 000037: proxy_endpoint_limits

```sql
-- 000037_proxy_endpoint_limits.up.sql
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
```

```sql
-- 000037_proxy_endpoint_limits.down.sql
DROP TABLE IF EXISTS proxy_endpoint_limits;
```

## Files to Create
- `server/internal/database/migrations/000036_proxy_routes.up.sql`
- `server/internal/database/migrations/000036_proxy_routes.down.sql`
- `server/internal/database/migrations/000037_proxy_endpoint_limits.up.sql`
- `server/internal/database/migrations/000037_proxy_endpoint_limits.down.sql`

## Success Criteria
- [x] Migrations run without error (`make migrate-up`)
- [x] Rollback works (`make migrate-down` x2)
- [x] Foreign keys to `agent_identities` and `secrets` enforced
- [x] Unique constraints prevent duplicate routes per agent
