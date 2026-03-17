# Researcher 2: RBAC & Agent Identity Analysis

## Agent identity → project membership

- `agent_identities` has `project_id` FK → `projects.id` (NOT NULL)
- `project_memberships` only links `user_id ↔ project_id` — no agent rows
- `rbac.Middleware` queries `project_memberships WHERE user_id = requester` → **useless for agents**

## project_id linkage: secrets ↔ agent_identities

- `secrets.project_id` added in migration 000015 — nullable
- `agent_identities.project_id` — NOT NULL
- **Shared axis**: if `secret.project_id == agent.project_id`, they are in the same project

## Auth context: agent token vs user JWT

| Aspect | User JWT | Agent Token |
|---|---|---|
| Auth middleware | `auth.AuthMiddleware` validates RS256 JWT, puts `userID` in ctx | Same middleware — fails JWT validation |
| Context key | `userID` | Nothing — no agent context |
| Identity downstream | `auth.UserIDFromContext(ctx)` | Returns `""` |

**Critical finding:** `agent.Service.ValidateToken()` is never wired as HTTP middleware in `main.go`. Agent tokens cannot authenticate to ANY workflow route. The MCP server cannot call `CreateRequest` at all.

## Recommended guard to replace owner check

```
secret = GetSecretByID(secretID)          // no owner filter
→ 404 if not found
→ if secret.project_id != nil {
    check project_memberships WHERE project_id = secret.project_id AND user_id = callerID
    → 403 if not a member
  } else {
    // legacy: fall back to owner-only
    if secret.user_id != callerID → 403
  }
```

This is a **2-query inline** pattern already used in scanner/dynsecret route handlers.

## Does fix need new DB query?

No new migration or query needed:
- `vaultSvc.GetSecretByID` already exists
- Project membership lookup pattern in `rbac/middleware.go` can be inlined

## Unresolved Questions

1. **Agent token auth path missing** — if agents should call `CreateRequest` directly (as `requester_type=ai_agent`), a dual-auth middleware (accept JWT OR agent token) is needed, plus agent context set in request context
2. **`userID` empty for agents** — `RequesterUserID: userID` in `CreateRequestInput` would be `""` if no JWT; check `access_requests.requester_user_id` nullability
3. **`secrets.project_id` nullable** — fallback behavior needed for legacy secrets
4. **Agents have no `project_memberships` row** — for agent requests, the membership check must use `agent_identities.project_id == secret.project_id` directly, not `project_memberships`
