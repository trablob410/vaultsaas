# Phase 5: Dashboard UI Fixes

## Context Links
- [plan.md](./plan.md)
- `dashboard/src/components/approvals/approval-list.tsx` -- current table
- `dashboard/src/types/api.ts` -- AccessRequest type (lines 13-27)
- `server/internal/workflow/service.go` -- ListPending query (lines 128-174)
- `server/internal/workflow/handler.go` -- ListPending response

## Overview
- **Priority:** P1
- **Status:** complete
- **Description:** Fix two bugs in approvals table: (1) shows truncated UUID instead of secret name, (2) shows raw `60m` instead of human-readable duration.

## Key Insights
- `approval-list.tsx` line 82: `req.secret_id.slice(0, 8)…` -- no secret name available.
- `approval-list.tsx` line 84: `{req.duration_minutes}m` -- raw minutes, no formatting.
- Backend `ListPending` already JOINs secrets table but only selects `ar.*` columns. Adding `s.name` to SELECT is trivial.
- `AccessRequest` Go struct and TS type both need `secret_name` field.

## Requirements

### Functional
- Approvals table "Secret" column shows secret name (fallback to truncated ID)
- Duration column shows human-readable format: `30 min`, `1 hr`, `2 hrs`

### Non-Functional
- No new API calls -- secret name included in existing list response

## Architecture

```
ListPending SQL:
  SELECT ar.*, s.name AS secret_name FROM access_requests ar JOIN secrets s ...

Response:
  { requests: [{ ..., secret_name: "Stripe API Key", ... }] }

Dashboard:
  <TableCell>{req.secret_name ?? req.secret_id.slice(0,8)+'...'}</TableCell>
  <TableCell>{formatDuration(req.duration_minutes)}</TableCell>
```

## Related Code Files

### Modify
- `server/internal/workflow/service.go` -- add `s.name` to ListPending SELECT + scan
- `server/internal/workflow/service.go` -- add `SecretName` to `AccessRequest` struct
- `dashboard/src/types/api.ts` -- add `secret_name` to `AccessRequest`
- `dashboard/src/components/approvals/approval-list.tsx` -- render name + format duration

## Implementation Steps

1. Modify `AccessRequest` Go struct in `service.go`:
   ```go
   type AccessRequest struct {
       // ... existing fields ...
       SecretName string `json:"secret_name,omitempty"`
   }
   ```

2. Modify `ListPending` query to include `s.name`:
   - SELECT: add `s.name` after existing columns
   - Scan: add `&r.SecretName` to scan targets
   - Both the count query and the data query already JOIN secrets -- just need the SELECT column

3. Modify `AccessRequest` in `dashboard/src/types/api.ts`:
   ```typescript
   export interface AccessRequest {
     // ... existing fields ...
     secret_name?: string
   }
   ```

4. Modify `approval-list.tsx`:
   - Secret column (line 82):
     ```tsx
     <TableCell className="text-sm font-medium">
       {req.secret_name ?? `${req.secret_id.slice(0, 8)}...`}
     </TableCell>
     ```
   - Duration column (line 84):
     ```tsx
     <TableCell className="text-sm">
       {req.duration_minutes >= 60
         ? `${Math.floor(req.duration_minutes / 60)} hr${Math.floor(req.duration_minutes / 60) > 1 ? 's' : ''}`
         : `${req.duration_minutes} min`}
     </TableCell>
     ```

5. Note: `AccessRequest` TS type has `duration_minutes` but Go struct has `RequestedDurationMinutes` with json tag `requested_duration_minutes`. Check which field name dashboard actually uses -- may need to align. Looking at Go struct: `json:"requested_duration_minutes"`. TS type has `duration_minutes`. **This is a mismatch** -- either fix TS type or add Go alias. Simplest: update TS type to `requested_duration_minutes`.

6. Verify: also update TS type field names that don't match backend:
   - TS `requester_id` vs Go `requester_user_id` -- fix TS to `requester_user_id`
   - TS `duration_minutes` vs Go `requested_duration_minutes` -- fix TS

## Todo List
- [x] Add `SecretName` to Go `AccessRequest` struct
- [x] Add `s.name` to `ListPending` SELECT + scan
- [x] Update TS `AccessRequest` type: add `secret_name`, fix field name mismatches
- [x] Update `approval-list.tsx` secret column
- [x] Update `approval-list.tsx` duration column
- [x] Verify `go build ./...` compiles
- [ ] Visual check in browser

## Success Criteria
- Approvals table shows "Stripe API Key" instead of "a1b2c3d4..."
- Duration shows "30 min" or "1 hr" instead of "30m" or "60m"
- No console errors in browser

## Risk Assessment
- **TS type mismatches**: `requester_id` vs `requester_user_id`, `duration_minutes` vs `requested_duration_minutes`. Could break other components using `AccessRequest` type. Grep for usage before renaming.
- **Null secret_name**: If secret deleted (soft-deleted), JOIN may return null. Use LEFT JOIN or handle null in frontend.

## Security Considerations
- Secret name is metadata, not sensitive. Safe to display.
- No credential values exposed in approvals list.

## Next Steps
- Consider adding secret name to the approval actions dialog as well
- Consider adding pagination to approvals list

## Unresolved Questions
- Should `ListPending` use LEFT JOIN to handle soft-deleted secrets? Currently uses INNER JOIN with `s.deleted_at IS NULL` which would exclude requests for deleted secrets entirely.
