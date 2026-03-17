---
phase: 1
title: "Wire agent auth middleware to workflow routes"
status: pending
priority: P0
created: 2026-03-18
---

# Phase 1: Wire Agent Auth Middleware

## Context

- [agent/middleware.go](../../server/internal/agent/middleware.go) -- exists, validates SHA-256 hashed tokens
- [auth/middleware.go](../../server/internal/auth/middleware.go) -- JWT RS256 validation
- [main.go](../../server/cmd/server/main.go) -- route wiring (line 168-192)

## Key Insights

- Both middlewares use same pattern: `Bearer <token>` header, store ID in context
- JWT tokens are RS256-signed JWTs; agent tokens are raw base64 strings -- distinct formats
- A dual-auth middleware can try JWT first, fall back to agent token validation
- Only `POST /secrets/{id}/access-requests` needs dual-auth initially; other workflow routes stay user-only

## Requirements

### Functional
- Create `DualAuthMiddleware(jwtMgr, agentSvc)` that accepts either auth type
- Set both `userID` and `agentID` context keys appropriately
- Wire to `CreateRequest` route

### Non-functional
- No perf regression -- JWT validation should short-circuit before agent token DB lookup
- Agent token path must not break existing user JWT auth

## Architecture

```
Request → DualAuthMiddleware
  ├─ Try JWT validate → success → set userID in ctx → next
  └─ Try agent token validate → success → set agentID in ctx → next
  └─ Both fail → 401
```

## Related Code Files

**Modify:**
- `server/internal/auth/middleware.go` -- add `DualAuthMiddleware` func
- `server/cmd/server/main.go` -- wire new middleware to CreateRequest route

**Read only:**
- `server/internal/agent/middleware.go` -- reference for agent context key
- `server/internal/agent/service.go` -- `ValidateToken` signature

## Implementation Steps

1. In `auth/middleware.go`, create `DualAuthMiddleware(jwtMgr *JWTManager, agentSvc AgentValidator)`:
   - Define `AgentValidator` interface: `ValidateToken(ctx, token) (*agent.Token, error)`
   - Extract Bearer token from `Authorization` header
   - Try `jwtMgr.ValidateAccessToken(token)` -- if success, set `userIDKey` context
   - If JWT fails, try `agentSvc.ValidateToken(ctx, token)` -- if success, set `agentIDKey` context
   - If both fail, return 401
   - Import agent package for `agentIDKey` context key -- or define a shared key

2. **Context key sharing**: `agent.agentIDKey` is unexported (`contextKey("agent_id")`).
   - Option A: Export it from agent package (`AgentIDKey`)
   - Option B: Use `agent.AgentIDFromContext` in handler; set value using same key type
   - **Chosen: Option A** -- export the context key constant, simpler

3. In `main.go`, replace the single `CreateRequest` route:
   ```go
   // Before (inside auth.AuthMiddleware group):
   r.Post("/secrets/{secret_id}/access-requests", workflowHandler.CreateRequest)

   // After: move to its own group with dual-auth
   r.Group(func(r chi.Router) {
       r.Use(auth.DualAuthMiddleware(jwtMgr, agentSvc))
       r.Post("/secrets/{secret_id}/access-requests", workflowHandler.CreateRequest)
   })
   ```

4. Remove the old `CreateRequest` route from the JWT-only group.

## Todo

- [ ] Define `AgentValidator` interface in `auth` package
- [ ] Implement `DualAuthMiddleware`
- [ ] Export agent context key or use shared mechanism
- [ ] Wire in `main.go`
- [ ] Compile check: `cd server && go build ./...`

## Success Criteria

- Agent bearer token authenticates to `POST /secrets/{id}/access-requests`
- User JWT still works for same route
- Invalid tokens return 401
- All other workflow routes remain JWT-only

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Import cycle: `auth` imports `agent` | Use interface `AgentValidator` -- no direct import of `agent` package |
| Agent token accidentally accepted on other routes | Only wire `DualAuthMiddleware` to CreateRequest route |

## Security Considerations

- Agent tokens are SHA-256 hashed, never stored plain -- same security model
- Dual-auth tries JWT first (fast, no DB) before agent token (DB lookup) -- rate-limited by existing `apiLimiter`
- Agent tokens have expiry and revocation checks in `ValidateToken`
