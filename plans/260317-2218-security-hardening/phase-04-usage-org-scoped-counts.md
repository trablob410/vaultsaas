# Phase 4 — Usage Org-Scoped Counts (C4)

## Context Links
- `server/internal/usage/tracker.go:60-71` — global COUNT(*) queries
- DB schema: `secrets.project_id`, `projects.workspace_id`, `workspaces.org_id`
- `agent_identities` — need to verify FK chain to org

## Overview
- **Priority:** CRITICAL
- **Status:** pending
- **Description:** `GetCurrent` for `secrets_count` and `agents_count` runs unscoped `SELECT COUNT(*) FROM secrets` / `agent_identities`. Returns global counts instead of org-specific. Free tier limits effectively unenforced (or over-enforced if another org has many secrets).

## Key Insights
- Schema chain: `secrets -> projects -> workspaces -> organizations`
- Agent chain: `agent_identities` has `project_id` -> same chain to org
- The `orgID` param is already passed to `GetCurrent` but unused for these metrics
- `requests_today` is already correctly scoped via `org_id`

## Requirements
**Functional:** secrets_count and agents_count must be scoped to the org.
**Non-functional:** Query must be efficient; consider adding indexes if needed.

## Related Code Files
| Action | File |
|--------|------|
| Modify | `server/internal/usage/tracker.go` |

## Implementation Steps

1. **Update `secrets_count` query:**
   ```go
   case "secrets_count":
       var count int
       err := t.db.QueryRow(ctx, `
           SELECT COUNT(*)
           FROM secrets s
           JOIN projects p ON p.id = s.project_id
           JOIN workspaces w ON w.id = p.workspace_id
           WHERE w.org_id = $1
       `, orgID).Scan(&count)
       return count, err
   ```

2. **Update `agents_count` query:**
   ```go
   case "agents_count":
       var count int
       err := t.db.QueryRow(ctx, `
           SELECT COUNT(*)
           FROM agent_identities a
           JOIN projects p ON p.id = a.project_id
           JOIN workspaces w ON w.id = p.workspace_id
           WHERE w.org_id = $1
       `, orgID).Scan(&count)
       return count, err
   ```

3. **Remove TODO comments** on lines 62 and 68

4. **Verify column names** — check that `agent_identities` has `project_id`:
   - Migration `000016_create_agent_identities.up.sql` should confirm

## Todo List
- [ ] Verify agent_identities.project_id exists in migration 000016
- [ ] Update secrets_count query with org scope
- [ ] Update agents_count query with org scope
- [ ] Remove TODO comments
- [ ] Verify compilation
- [ ] Test with multiple orgs to confirm isolation

## Success Criteria
- `CheckLimit` returns counts scoped to the requesting org
- Org A's secrets don't count against Org B's limits
- Existing `requests_today` behavior unchanged

## Risk Assessment
- **JOIN performance**: 3-table join on COUNT. For free tier (~50 secrets max) this is negligible.
- **Missing FK**: If agent_identities lacks project_id, query will fail. Verify in migration.

## Security Considerations
- Org isolation is a security boundary — this fix prevents cross-org data leakage in usage counts
- No new auth checks needed; orgID already derived from authenticated context

## Next Steps
- Consider adding composite indexes if performance becomes an issue at scale
