---
title: "Proxy Gateway with Credential Injection"
description: "OneCLI-inspired HTTP proxy gateway — transparent credential injection, placeholder keys, host+path routing, per-endpoint rate limiting"
status: pending
priority: P1
effort: 28h
branch: feat/proxy-gateway
tags: [proxy, gateway, credential-injection, agent, security]
blockedBy: []
blocks: []
created: 2026-03-24
---

# Proxy Gateway with Credential Injection

Adds an HTTP forward proxy to Valt so ANY agent framework can use Valt secrets without MCP integration. Agents set `HTTPS_PROXY` → Valt swaps placeholder keys for real credentials at request time. Agent never sees real key (prompt injection defense).

## Architecture

```
Agent (any framework)
  │  HTTPS_PROXY=http://valt:10256
  │  Proxy-Authorization: Bearer <agent_token>
  │  Authorization: Bearer PLACEHOLDER_abc123  ← fake key
  ▼
┌─────────────────────────────────────┐
│  Go Gateway (port 10256)            │
│  1. Validate agent token            │
│  2. Match host+path → proxy_routes  │
│  3. Check endpoint rate limit       │
│  4. Decrypt secret (AES-256-GCM)    │
│  5. Swap placeholder → real key     │
│  6. Forward request to target API   │
│  7. Audit log                       │
└─────────────────────────────────────┘
  │
  ▼
External API (api.openai.com, api.stripe.com, etc.)
```

**Key decision:** Proxy lives in same Go binary as main server (different port). Reuses existing DB pool, agent auth, vault decrypt, audit log. KISS — no separate service.

## Phases

| # | Phase | Priority | Effort | Status |
|---|-------|----------|--------|--------|
| 1 | [DB Migrations](./phase-01-db-migrations.md) | P1 | 2h | pending |
| 2 | [Gateway Core](./phase-02-gateway-core.md) | P1 | 10h | pending |
| 3 | [Placeholder Keys + Endpoint Rate Limiting](./phase-03-placeholder-and-ratelimit.md) | P1 | 6h | pending |
| 4 | [Dashboard UI](./phase-04-dashboard-ui.md) | P2 | 6h | pending |
| 5 | [Testing + Docs](./phase-05-testing-docs.md) | P2 | 4h | pending |

## Dependency Graph

```
Phase 1 (DB) ──▶ Phase 2 (Gateway Core) ──▶ Phase 3 (Placeholder + Rate Limit)
                                           ──▶ Phase 4 (Dashboard UI)
                                                      ──▶ Phase 5 (Testing + Docs)
```

## DB Migrations

| Migration | Phase | Purpose |
|-----------|-------|---------|
| 000036 | 1 | `proxy_routes` table |
| 000037 | 1 | `proxy_endpoint_limits` table |

## New Go Packages

| Package | Purpose |
|---------|---------|
| `server/internal/gateway/` | HTTP forward proxy, route matching, credential injection |

## New Env Vars

| Variable | Default | Notes |
|----------|---------|-------|
| `GATEWAY_PORT` | `10256` | Proxy gateway listen port |
| `GATEWAY_ENABLED` | `false` | Feature flag |

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Go proxy in same binary | Reuse DB pool, auth, vault, audit. KISS. |
| HTTP CONNECT for HTTPS | Standard proxy protocol, no MITM/CA cert needed |
| Placeholder per-route | Each route gets unique placeholder, not global per-agent |
| AES-256-GCM decrypt at forward time | Consistent with existing vault encryption model |
| Audit via existing audit package | No new audit infra needed |
