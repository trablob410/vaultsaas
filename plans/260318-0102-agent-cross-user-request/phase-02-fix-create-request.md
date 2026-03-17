---
phase: 2
title: "Fix CreateRequest to allow non-owner requests"
status: pending
priority: P0
created: 2026-03-18
---

# Phase 2: Fix CreateRequest Handler

## Context

- [workflow/handler.go](../../server/internal/workflow/handler.go) -- `CreateRequest` at line 50
- [workflow/service.go](../../server/internal/workflow/service.go) -- `CreateRequest` service, daily limit query
- [vault/service.go](../../server/internal/vault/service.go) -- `GetSecretByID` at line 204, `Secret` struct at line 14

## Key Insights

- `GetSecretByID` SQL doesn't select `project_id`; `Secret` struct has no `ProjectID` field
- `requester_user_id` is `NOT NULL REFERENCES users(id)` in DB -- needs migration for agent-only requests
- `CreateRequestInput.RequesterUserID` used in daily limit query -- agent requests need equivalent rate-limiting by `ai_agent_id`
- Approve/Reject already use `GetSecretByID` -- no changes needed there

## Requirements

### Functional
- Agent can create access request for secret in same project
- User (project member, non-owner) can create access request
- `requester_user_id` nullable; `ai_agent_id` set for agent requests
- Legacy secrets (no `project_id`): owner-only check preserved

### Non-functional
- No breaking changes to existing human-user flow
- Rate limiting still works for both agents and users

## Architecture

```
CreateRequest handler:
  1. Identify caller: userID = auth.UserIDFromContext, agentID = agent.AgentIDFromContext
  2. GetSecretByID(secretID) -- no owner filter
  3. Authorization guard:
     - If secret.ProjectID != nil:
       - Agent caller: agent_identities.project_id == secret.project_id
       - User caller: project_memberships WHERE project_id AND user_id
     - If secret.ProjectID == nil:
       - User caller: secret.UserID == userID (legacy owner check)
       - Agent caller: 403 (no project = no agent access)
  4. Build CreateRequestInput with correct RequesterUserID / AIAgentID
```

## Related Code Files

**Modify:**
- `server/internal/vault/service.go` -- add `ProjectID` to `Secret` struct; update `GetSecretByID` SQL
- `server/internal/workflow/handler.go` -- rewrite `CreateRequest` authorization logic
- `server/internal/workflow/service.go` -- handle nullable `RequesterUserID`, rate-limit by agent ID

**Create:**
- `server/internal/database/migrations/000025_nullable_requester_user_id.up.sql`
- `server/internal/database/migrations/000025_nullable_requester_user_id.down.sql`

## Implementation Steps

### Step 1: DB migration -- make `requester_user_id` nullable

```sql
-- 000025_nullable_requester_user_id.up.sql
ALTER TABLE access_requests
    ALTER COLUMN requester_user_id DROP NOT NULL,
    DROP CONSTRAINT IF EXISTS access_requests_requester_id_fkey;

ALTER TABLE access_requests
    ADD CONSTRAINT access_requests_requester_user_id_fkey
    FOREIGN KEY (requester_user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add check: at least one of requester_user_id or ai_agent_id must be set
ALTER TABLE access_requests
    ADD CONSTRAINT chk_requester_identity
    CHECK (requester_user_id IS NOT NULL OR ai_agent_id IS NOT NULL);
```

### Step 2: Add `ProjectID` to `Secret` struct

In `vault/service.go`:
- Add `ProjectID *string \`json:"project_id,omitempty"\`` to `Secret` struct
- Update `GetSecretByID` SQL to include `project_id` in SELECT and Scan

### Step 3: Rewrite `CreateRequest` handler

```go
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())
    agentID := agent.AgentIDFromContext(r.Context())

    // Must have at least one identity
    if userID == "" && agentID == "" {
        apierror.Unauthorized(w, "no valid identity")
        return
    }

    // Parse body, validate secretID...

    secret, err := h.vaultSvc.GetSecretByID(r.Context(), secretID)
    // 404 if nil

    // Project-scoped authorization
    if secret.ProjectID != nil {
        if agentID != "" {
            // Agent: check agent_identities.project_id == secret.project_id
            agentIdentity, _ := h.agentSvc.Get(ctx, agentID)
            if agentIdentity == nil || agentIdentity.ProjectID != *secret.ProjectID {
                apierror.Forbidden(w, "agent not in secret's project")
                return
            }
        } else {
            // User: check project_memberships
            // Inline query: SELECT 1 FROM project_memberships WHERE ...
        }
    } else {
        // Legacy: owner-only
        if userID == "" || secret.UserID != userID {
            apierror.Forbidden(w, "only the owner can request access to this secret")
            return
        }
    }

    // Set requester fields
    input := CreateRequestInput{
        RequesterUserID: userID, // "" for agents -- now nullable in DB
        AIAgentID:       agentID,
        RequesterType:   requesterType, // derive from caller identity
        ...
    }
}
```

### Step 4: Update `workflow.Handler` to hold `agentSvc`

- Add `agentSvc` field to `Handler` struct (use interface to avoid import cycle)
- Define interface: `AgentGetter interface { Get(ctx, id) (*agent.Agent, error) }`
- Pass `agentSvc` from `main.go` into `NewHandler`

### Step 5: Update service daily-limit query

In `service.go:CreateRequest`, the daily limit query uses `requester_user_id`:
- If `RequesterUserID` is empty (agent request), query by `ai_agent_id` instead
- Two query paths, or: `WHERE (requester_user_id = $1 OR ai_agent_id = $2) AND secret_id = $3`

### Step 6: Project membership check helper

Add inline query in handler (matches existing pattern in scanner/dynsecret handlers):
```go
var memberCount int
err := h.pool.QueryRow(ctx,
    `SELECT COUNT(*) FROM project_memberships WHERE project_id=$1 AND user_id=$2`,
    *secret.ProjectID, userID,
).Scan(&memberCount)
```

Need `pool` access in handler -- pass via `NewHandler` or use `vaultSvc` pool.

## Todo

- [ ] Create migration 000025
- [ ] Add `ProjectID` to `Secret` struct + `GetSecretByID` SQL
- [ ] Define `AgentGetter` interface in workflow package
- [ ] Rewrite `CreateRequest` handler with dual-caller logic
- [ ] Update `service.CreateRequest` daily-limit for agent callers
- [ ] Wire `agentSvc` and `pool` into `workflow.NewHandler`
- [ ] Update `main.go` `NewHandler` call
- [ ] Compile check

## Success Criteria

- Agent (same project) can create access request -> 201
- Agent (different project) gets 403
- User (project member, non-owner) can create access request -> 201
- User (non-member) gets 403
- Legacy secret (no project_id) owner-only -> works as before
- Daily rate limit applies to both user and agent callers

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Existing FK constraint on `requester_user_id` | Migration explicitly drops and re-adds without NOT NULL |
| Agent can enumerate secrets by UUID | `GetSecretByID` returns 404 only -- no metadata leak beyond existence |
| Audit log uses `userID` -- empty for agents | Update audit call to use `agentID` when `userID` empty |

## Security Considerations

- Project-scope guard prevents cross-project access
- Legacy secrets locked to owner-only (strictest policy)
- `CHECK (requester_user_id IS NOT NULL OR ai_agent_id IS NOT NULL)` -- DB-level invariant
- Audit log must attribute agent requests to `agentID`
