# Phase 2 — Scanner Authorization (C2)

## Context Links
- `server/internal/scanner/handler.go` — 4 routes with `_ = auth.UserIDFromContext(...)`
- `server/internal/rbac/middleware.go` — reusable RBAC middleware
- `server/internal/rbac/rbac.go` — resource/action constants (missing `ResourceScans`)
- `server/internal/scanner/service.go` — scan_results has `project_id`

## Overview
- **Priority:** CRITICAL
- **Status:** pending
- **Description:** 4 scanner routes (`listScans`, `listFindings`, `importFinding`, `dismissFinding`) extract userID but never verify project membership. Any authenticated user can access any project's scans.

## Key Insights
- `createScan` and `listScans` have `project_id` in URL — can use `rbac.Middleware` directly
- `listFindings`, `importFinding`, `dismissFinding` use `scan_id` in URL — need scan->project lookup
- `rbac.Middleware` expects a chi URL param name for project_id
- RBAC resource constants need `ResourceScans` added

## Requirements
**Functional:** Only project members can CRUD scans/findings for that project.
**Non-functional:** Read ops need `ActionRead`; write ops (import, dismiss) need `ActionWrite`.

## Architecture
Two approaches for scan_id-scoped routes:
- **Option A (chosen):** Split routes into two groups — project-scoped routes use `rbac.Middleware`; scan-scoped routes do inline project lookup + membership check
- **Option B:** Add a service method `GetScanProjectID(scanID)` and write a custom middleware. Overkill for 3 routes.

## Related Code Files
| Action | File |
|--------|------|
| Modify | `server/internal/scanner/handler.go` |
| Modify | `server/internal/rbac/rbac.go` (add ResourceScans) |
| Modify | `server/internal/scanner/service.go` (add GetScanProjectID) |

## Implementation Steps

1. **Add `ResourceScans` to rbac.go:**
   ```go
   ResourceScans Resource = "scans"
   ```
   Add to all role perm maps: owner/admin get read+write, member gets read+write, viewer gets read.

2. **Add `GetScanProjectID` to scanner service.go:**
   ```go
   func (s *Service) GetScanProjectID(ctx context.Context, scanID string) (string, error) {
       var projectID string
       err := s.db.QueryRow(ctx, `SELECT project_id FROM scan_results WHERE id = $1`, scanID).Scan(&projectID)
       return projectID, err
   }
   ```

3. **Update handler to accept `db *pgxpool.Pool`** (needed for inline RBAC check on scan-scoped routes):
   ```go
   type Handler struct {
       service *Service
       db      *pgxpool.Pool
   }
   func NewHandler(service *Service, db *pgxpool.Pool) *Handler { ... }
   ```

4. **Apply `rbac.Middleware` to project-scoped routes:**
   ```go
   func (h *Handler) Routes() chi.Router {
       r := chi.NewRouter()
       r.With(rbac.Middleware(h.db, "project_id", rbac.ResourceScans, rbac.ActionWrite)).
           Post("/projects/{project_id}/scans", h.createScan)
       r.With(rbac.Middleware(h.db, "project_id", rbac.ResourceScans, rbac.ActionRead)).
           Get("/projects/{project_id}/scans", h.listScans)
       // scan-scoped routes — inline check in handler
       r.Get("/scans/{scan_id}/findings", h.listFindings)
       r.Post("/scans/{scan_id}/findings/{finding_id}/import", h.importFinding)
       r.Post("/scans/{scan_id}/findings/{finding_id}/dismiss", h.dismissFinding)
       return r
   }
   ```

5. **Add inline RBAC check helper in handler:**
   ```go
   func (h *Handler) checkScanAccess(ctx context.Context, w http.ResponseWriter, scanID, userID string, action rbac.Action) bool {
       projectID, err := h.service.GetScanProjectID(ctx, scanID)
       if err != nil {
           apierror.NotFound(w, "scan not found")
           return false
       }
       var role string
       err = h.db.QueryRow(ctx, `SELECT role FROM project_memberships WHERE project_id = $1 AND user_id = $2`, projectID, userID).Scan(&role)
       if err != nil {
           apierror.Forbidden(w, "access denied")
           return false
       }
       if !rbac.Can(rbac.RoleFromProjectMembership(role), rbac.ResourceScans, action) {
           apierror.Forbidden(w, "insufficient permissions")
           return false
       }
       return true
   }
   ```

6. **Update `listFindings`** — replace `_ = auth.UserIDFromContext(...)` with actual check:
   ```go
   userID := auth.UserIDFromContext(r.Context())
   if !h.checkScanAccess(r.Context(), w, scanID, userID, rbac.ActionRead) { return }
   ```

7. **Update `importFinding` and `dismissFinding`** — same pattern with `rbac.ActionWrite`

8. **Update main.go** — pass `pool` to `scanner.NewHandler`:
   ```go
   scannerHandler := scanner.NewHandler(scannerSvc, pool)
   ```

## Todo List
- [ ] Add `ResourceScans` to rbac.go + role perms
- [ ] Add `GetScanProjectID` to scanner service
- [ ] Add `db` field to scanner Handler
- [ ] Apply rbac.Middleware to project-scoped routes
- [ ] Add `checkScanAccess` helper
- [ ] Update listFindings with RBAC check
- [ ] Update importFinding with RBAC check
- [ ] Update dismissFinding with RBAC check
- [ ] Update main.go scanner handler init
- [ ] Verify compilation

## Success Criteria
- Unauthenticated requests get 401
- Non-member requests get 403
- Project members with correct role can access scans/findings
- Viewer role can read but not import/dismiss

## Risk Assessment
- **Extra DB query**: scan-scoped routes need 2 queries (scan->project, project->membership). Acceptable for security.
- **No test coverage**: scanner handler has no tests. Manual verification needed.

## Security Considerations
- Default-deny: if project_memberships lookup fails, access denied
- viewer role cannot import/dismiss (write ops)

## Next Steps
- Phase 3 applies same pattern to dynsecret routes
