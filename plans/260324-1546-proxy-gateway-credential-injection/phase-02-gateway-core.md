---
phase: 2
title: "Gateway Core — HTTP Forward Proxy + Auth + Route Matching + Credential Injection"
status: pending
priority: P1
effort: 10h
---

# Phase 2: Gateway Core

## Overview

Build the HTTP forward proxy that intercepts agent requests, matches routes, decrypts secrets, and injects real credentials. This is the heart of the feature.

## Context Links
- [Agent auth middleware](../../server/internal/agent/middleware.go) — reuse `ValidateToken`
- [Vault decrypt](../../server/pkg/crypto/) — `DecryptAES256GCM`
- [Audit logger](../../server/internal/audit/) — `Logger.Log()`
- [Rate limiter](../../server/internal/ratelimit/) — existing Redis sliding window
- [Main entry](../../server/cmd/server/main.go) — add gateway listener

## Architecture

```
┌──────────────────────────────────────────────┐
│  gateway package                             │
│                                              │
│  Server                                      │
│  ├── listenAndServe(:10256)                  │
│  ├── handleHTTP()      ← GET/POST/PUT/etc    │
│  └── handleConnect()   ← HTTPS CONNECT       │
│                                              │
│  RouteStore (DB queries)                     │
│  ├── FindRoute(agentID, host, path)          │
│  ├── ListRoutes(agentID)                     │
│  ├── CreateRoute(...)                        │
│  └── DeleteRoute(...)                        │
│                                              │
│  Injector                                    │
│  ├── InjectCredential(req, route, secret)    │
│  └── matchPattern(pattern, value) bool       │
│                                              │
│  Handler (REST API for CRUD)                 │
│  ├── ListRoutes     GET /proxy-routes        │
│  ├── CreateRoute    POST /proxy-routes       │
│  ├── UpdateRoute    PUT /proxy-routes/{id}   │
│  └── DeleteRoute    DELETE /proxy-routes/{id}│
└──────────────────────────────────────────────┘
```

## Request Flow (Detailed)

### HTTP requests (non-CONNECT)
1. Agent sends `GET http://api.openai.com/v1/models` with `Proxy-Authorization: Bearer <agent_token>`
2. Gateway extracts + validates agent token via `agent.Service.ValidateToken()`
3. Extract target host + path from request URL
4. Query `proxy_routes` for matching route: `WHERE agent_id = $1 AND enabled = true`
5. For each route: `matchPattern(host_pattern, host)` AND `matchPattern(path_pattern, path)`
6. If no match → forward request as-is (transparent proxy, no injection)
7. If match → decrypt secret value using `crypto.DecryptAES256GCM(masterKey, encryptedDEK)` then decrypt blob
8. Apply injection: replace placeholder in header/query with real value based on `injection_type`/`injection_format`
9. Forward request via `httputil.ReverseProxy` or direct `http.Client`
10. Audit log: `{agent_id, host, path, route_id, timestamp}`
11. Return response to agent

### HTTPS CONNECT tunnel
1. Agent sends `CONNECT api.openai.com:443`
2. Gateway validates agent token
3. Establish TCP tunnel (no MITM — we don't break TLS)
4. **Limitation:** Cannot inject credentials in CONNECT tunnel (encrypted)
5. For HTTPS credential injection: agent must use `http://` URL through proxy (gateway upgrades to HTTPS on outbound side)

**Key insight:** For credential injection to work with HTTPS APIs, the proxy receives the plaintext HTTP request (agent sends `http://api.openai.com/...` to proxy), then the proxy makes the HTTPS request to the real API. This is standard forward proxy behavior — no MITM needed.

## Pattern Matching

Simple glob-style matching:
- `*` matches any sequence of characters within a path segment
- `/*` matches any path
- Exact match for host (no wildcards on host for security)

```go
// matchPath("/*", "/v1/chat/completions") → true
// matchPath("/v1/*", "/v1/chat/completions") → true
// matchPath("/v1/chat/*", "/v1/models") → false
func matchPath(pattern, path string) bool
```

## Injection Types

| Type | `injection_key` | `injection_format` | Behavior |
|------|-----------------|-------------------|----------|
| `bearer` | `Authorization` | `Bearer {value}` | Set/replace Authorization header |
| `header` | any header name | `{value}` or custom | Set/replace named header |
| `query` | query param name | `{value}` | Append/replace query parameter |

`{value}` is replaced with decrypted secret plaintext.

## Files to Create

| File | LOC | Purpose |
|------|-----|---------|
| `server/internal/gateway/server.go` | ~120 | HTTP proxy server, listener, request routing |
| `server/internal/gateway/injector.go` | ~80 | Credential injection + pattern matching |
| `server/internal/gateway/store.go` | ~100 | DB queries for proxy_routes |
| `server/internal/gateway/handler.go` | ~120 | REST API handlers for route CRUD |

## Files to Modify

| File | Change |
|------|--------|
| `server/cmd/server/main.go` | Start gateway listener on GATEWAY_PORT |
| `server/internal/config/config.go` | Add `GatewayPort`, `GatewayEnabled` fields |

## Implementation Steps

1. **Config** — Add `GATEWAY_PORT` and `GATEWAY_ENABLED` to config struct
2. **Store** — `RouteStore` with `FindMatchingRoute(ctx, agentID, host, path)`, `ListRoutes`, `CreateRoute`, `UpdateRoute`, `DeleteRoute`
3. **Injector** — `InjectCredential(req *http.Request, route *ProxyRoute, secretValue string)` + `matchPath(pattern, path) bool`
4. **Server** — `gateway.Server` struct holding store, agent service, vault service, audit logger, master key. Methods: `ListenAndServe()`, `handleRequest()`, `handleConnect()`
5. **Handler** — REST handlers for proxy route CRUD (JWT-only, admin/owner scope)
6. **Main** — If `GATEWAY_ENABLED`, spawn `gateway.Server` in goroutine

## Key Code Snippets

### Proxy handler core logic
```go
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Auth
    token := extractProxyAuth(r)
    agentToken, err := s.agentSvc.ValidateToken(r.Context(), token)
    if err != nil {
        http.Error(w, "unauthorized", 407) // Proxy Authentication Required
        return
    }

    // 2. Route match
    host := r.URL.Host
    path := r.URL.Path
    route, err := s.store.FindMatchingRoute(r.Context(), agentToken.AgentID, host, path)

    // 3. Inject if matched
    if route != nil {
        secretValue, err := s.decryptSecret(r.Context(), route.SecretID)
        InjectCredential(r, route, secretValue)
    }

    // 4. Forward
    s.proxy.ServeHTTP(w, r)

    // 5. Audit
    s.auditLog.Log(r.Context(), audit.Entry{...})
}
```

## Security Considerations
- Agent token validated on every request (no caching to avoid stale tokens)
- Secret decrypted in memory, never logged, zeroed after injection
- Proxy-Authorization header stripped before forwarding
- Placeholder key stripped/replaced before forwarding
- No MITM: HTTPS connection integrity preserved for CONNECT tunnels
- Rate limiting checked before credential injection (fail fast)

## Success Criteria
- [ ] `GATEWAY_ENABLED=true GATEWAY_PORT=10256` starts proxy listener
- [ ] Agent authenticates via `Proxy-Authorization: Bearer <token>`
- [ ] Route matching works for exact host + glob path
- [ ] Credential injected into header/query per route config
- [ ] Requests without matching route forwarded transparently
- [ ] Audit log entry created for every proxied request
- [ ] 407 returned for invalid/missing agent token
- [ ] REST CRUD endpoints for proxy routes work
