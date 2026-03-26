---
phase: 3
title: "Placeholder Key Pattern + Per-Endpoint Rate Limiting"
status: pending
priority: P1
effort: 6h
---

# Phase 3: Placeholder Keys + Endpoint Rate Limiting

## Overview

Two features that make the gateway production-ready:
1. **Placeholder keys** — auto-generated fake API keys agents use; gateway swaps them for real ones
2. **Per-endpoint rate limiting** — per-agent, per-host+path rate limits and blocking

## Part A: Placeholder Key Pattern

### Concept

When a proxy route is created, Valt auto-generates a unique placeholder key (e.g., `valt_pk_a1b2c3d4e5f6`). Agent uses this as its "API key." Gateway recognizes the placeholder in request headers/body and swaps it with the real decrypted secret.

### Why Placeholders?

- Agent config contains only fake keys → prompt injection extracts useless strings
- Multiple agents can have different placeholders for same secret → traceability
- Placeholder format is recognizable → easy to audit for accidental leaks

### Placeholder Format

```
valt_pk_{base64url_random_16_bytes}
```

Example: `valt_pk_7Kx9mPqR2sT4vW6y`

### Implementation

**Generate on route creation:**
```go
func generatePlaceholder() string {
    b := make([]byte, 16)
    crypto/rand.Read(b)
    return "valt_pk_" + base64.RawURLEncoding.EncodeToString(b)
}
```

**Lookup optimization:** Index on `placeholder_key` column. When gateway receives request, scan headers for `valt_pk_*` pattern → lookup route by placeholder → inject.

**Two matching modes (checked in order):**
1. **Placeholder match** — scan request headers for any `valt_pk_*` value → direct route lookup
2. **Host+path match** — fallback to pattern matching if no placeholder found

### Files to Modify

| File | Change |
|------|--------|
| `server/internal/gateway/store.go` | `FindByPlaceholder(ctx, placeholder)` query |
| `server/internal/gateway/server.go` | Add placeholder scan before host+path match |
| `server/internal/gateway/injector.go` | Replace placeholder value in headers with real secret |
| `server/internal/gateway/handler.go` | Auto-generate placeholder on route create; return in response |

### API Response Enhancement

```json
// POST /api/v1/proxy-routes response
{
  "id": "uuid",
  "agent_id": "uuid",
  "host_pattern": "api.openai.com",
  "path_pattern": "/v1/*",
  "placeholder_key": "valt_pk_7Kx9mPqR2sT4vW6y",
  "injection_type": "bearer",
  "enabled": true
}
```

Agent config becomes:
```yaml
# Agent just needs this — fake key, safe to commit
OPENAI_API_KEY: valt_pk_7Kx9mPqR2sT4vW6y
HTTPS_PROXY: http://valt-gateway:10256
```

---

## Part B: Per-Endpoint Rate Limiting

### Concept

Beyond global agent rate limiting (60 rpm in Redis), add per-endpoint limits. Admin can set:
- `api.openai.com /v1/*` → 30 rpm for agent-A
- `api.stripe.com /*` → blocked for agent-B

### Implementation

**Check order in gateway:**
1. Validate agent token
2. Check `proxy_endpoint_limits` for agent+host+path
3. If `blocked = true` → 403 immediately
4. If `rpm` exceeded → 429
5. Proceed with credential injection

**Rate limit storage:** Reuse existing Redis sliding window from `ratelimit` package. Key format:
```
ratelimit:proxy:{agentID}:{host}:{path_pattern}
```

If Redis unavailable: fall back to in-memory counter (like existing fail-open behavior).

### Files to Create

| File | LOC | Purpose |
|------|-----|---------|
| `server/internal/gateway/endpoint_limiter.go` | ~80 | Per-endpoint rate limit logic |

### Files to Modify

| File | Change |
|------|--------|
| `server/internal/gateway/store.go` | Add `FindEndpointLimit(ctx, agentID, host, path)` |
| `server/internal/gateway/server.go` | Check endpoint limit before injection |
| `server/internal/gateway/handler.go` | CRUD endpoints for endpoint limits |

### REST API for Endpoint Limits

```
GET    /api/v1/proxy-endpoint-limits?agent_id=X
POST   /api/v1/proxy-endpoint-limits
PUT    /api/v1/proxy-endpoint-limits/{id}
DELETE /api/v1/proxy-endpoint-limits/{id}
```

## Success Criteria
- [ ] Route creation auto-generates `valt_pk_*` placeholder
- [ ] Gateway matches requests by placeholder key in headers
- [ ] Placeholder swapped with real secret before forwarding
- [ ] Per-endpoint rate limits enforced (429 on exceed)
- [ ] Blocked endpoints return 403
- [ ] Rate limit keys scoped to agent+endpoint (not global)
- [ ] CRUD API for endpoint limits works
