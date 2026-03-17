# Researcher 1: CreateRequest Flow Analysis

## Owner check — exact code path

`handler.go:74` — `CreateRequest` calls:
```go
secret, err := h.vaultSvc.GetSecret(r.Context(), userID, secretID)
```

`vault/service.go:155-173` — `GetSecret` SQL:
```sql
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
```

`userID` comes from `auth.UserIDFromContext` (JWT sub claim). If caller is not secret owner, `GetSecret` returns nil → handler returns 404. Owner check is **implicit** via `AND user_id = $2` predicate.

## Auth context: agent vs user JWT

**User JWT path** (`auth.AuthMiddleware`):
- Validates RS256 JWT, extracts `sub` → stores as `userID` in context key `"userID"`
- `auth.UserIDFromContext(ctx)` returns real `users.id` UUID

**Agent token path** (`agent.AuthMiddleware`):
- Validates SHA-256 hashed Bearer token against `agent_tokens` table
- Stores `token.AgentID` in context key `"agent_id"` — **no `userID` in context**
- `auth.UserIDFromContext` returns `""` for agent requests

**All workflow routes** sit inside `auth.AuthMiddleware` group. Agent bearer tokens will **fail JWT validation** — not wired to any workflow route.

## GetSecretByID vs GetSecret

| Method | SQL filter | Use case |
|---|---|---|
| `GetSecret(ctx, userID, secretID)` | `WHERE id=$1 AND user_id=$2` | Owner-scoped |
| `GetSecretByID(ctx, secretID)` | `WHERE id=$1` | No owner constraint |

`GetSecretByID` exists at `vault/service.go:204`. Returns full `Secret` including `UserID`, `EncryptedDEK`, `StorageKey`.

## Security risks if owner check naively removed

- Any authenticated user could request access to any secret by guessing UUID
- Existence oracle (learns secret exists, name, credential_type)
- Approval queue spam
- Cross-project flooding (daily limit is per requester+secret, not per project)

## Correct fix

1. Replace `GetSecret(ctx, userID, secretID)` with `GetSecretByID(ctx, secretID)`
2. Add project-membership guard: verify requester is member of the project that owns the secret
3. Fallback for `secret.project_id == nil`: owner-only check (legacy secrets)
4. Approval side (`Approve`, `Reject`) already uses `GetSecretByID` + explicit owner check — correct, don't change

## Unresolved Questions

1. Do agents obtain a user JWT via delegation, or is a new auth path needed?
2. Is `access_requests.requester_user_id` nullable? If agent sends request, userID="" → corruption risk
3. `secrets.project_id` nullable — what's the fallback for project-less secrets?
