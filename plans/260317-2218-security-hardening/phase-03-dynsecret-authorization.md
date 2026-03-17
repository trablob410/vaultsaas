# Phase 3 — DynSecret Authorization (C3)

## Context Links
- `server/internal/dynsecret/handler.go` — 5 routes, 3 have zero auth checks
- `server/internal/rbac/middleware.go` — reusable middleware
- `server/internal/dynsecret/service.go:67` — `GetProvider` returns `ProviderConfig` with `ProjectID`

## Overview
- **Priority:** CRITICAL
- **Status:** pending
- **Description:** `createLease`, `listLeases`, `revokeLease` read no userID at all. `createProvider`/`listProviders` get userID but never verify project membership. Any authenticated user can operate on any provider/lease.

## Key Insights
- `createProvider`/`listProviders` have `project_id` in URL — direct rbac.Middleware
- `createLease`/`listLeases` have `provider_id` in URL — need provider->project lookup (already exists via `GetProvider`)
- `revokeLease` has `lease_id` — need lease->provider->project lookup
- Reuse same inline check pattern as Phase 2

## Requirements
**Functional:** Only project members can manage providers/leases. Write ops need `ActionWrite`.
**Non-functional:** No new resource constant needed if we reuse `ResourceSecret` for dynamic secrets (they are secrets). Or add `ResourceDynSecret`. Decision: add `ResourceDynSecret` for clarity.

## Related Code Files
| Action | File |
|--------|------|
| Modify | `server/internal/dynsecret/handler.go` |
| Modify | `server/internal/rbac/rbac.go` (add ResourceDynSecret) |

## Implementation Steps

1. **Add `ResourceDynSecret` to rbac.go:**
   ```go
   ResourceDynSecret Resource = "dynsecret"
   ```
   Same permissions as ResourceSecret in all role maps.

2. **Add `db` field to dynsecret Handler:**
   ```go
   type Handler struct {
       service *Service
       db      *pgxpool.Pool
   }
   func NewHandler(service *Service, db *pgxpool.Pool) *Handler { ... }
   ```

3. **Apply rbac.Middleware to project-scoped routes:**
   ```go
   r.With(rbac.Middleware(h.db, "project_id", rbac.ResourceDynSecret, rbac.ActionWrite)).
       Post("/projects/{project_id}/providers", h.createProvider)
   r.With(rbac.Middleware(h.db, "project_id", rbac.ResourceDynSecret, rbac.ActionRead)).
       Get("/projects/{project_id}/providers", h.listProviders)
   ```

4. **Add inline check helper for provider-scoped routes:**
   ```go
   func (h *Handler) checkProviderAccess(ctx context.Context, w http.ResponseWriter, providerID, userID string, action rbac.Action) bool {
       pc, err := h.service.GetProvider(ctx, providerID)
       if err != nil {
           apierror.NotFound(w, "provider not found")
           return false
       }
       var role string
       err = h.db.QueryRow(ctx, `SELECT role FROM project_memberships WHERE project_id = $1 AND user_id = $2`, pc.ProjectID, userID).Scan(&role)
       if err != nil {
           apierror.Forbidden(w, "access denied")
           return false
       }
       if !rbac.Can(rbac.RoleFromProjectMembership(role), rbac.ResourceDynSecret, action) {
           apierror.Forbidden(w, "insufficient permissions")
           return false
       }
       return true
   }
   ```

5. **Add lease->provider lookup helper for revokeLease:**
   Add `GetLeaseProviderID` to service:
   ```go
   func (s *Service) GetLeaseProviderID(ctx context.Context, leaseID string) (string, error) {
       var providerID string
       err := s.db.QueryRow(ctx, `SELECT provider_id FROM dynamic_leases WHERE id = $1`, leaseID).Scan(&providerID)
       return providerID, err
   }
   ```

6. **Update `createLease`** — add userID check:
   ```go
   userID := auth.UserIDFromContext(r.Context())
   if !h.checkProviderAccess(r.Context(), w, providerID, userID, rbac.ActionWrite) { return }
   ```

7. **Update `listLeases`** — same with ActionRead

8. **Update `revokeLease`** — lookup provider via lease, then check:
   ```go
   userID := auth.UserIDFromContext(r.Context())
   providerID, err := h.service.GetLeaseProviderID(r.Context(), leaseID)
   if err != nil { apierror.NotFound(w, "lease not found"); return }
   if !h.checkProviderAccess(r.Context(), w, providerID, userID, rbac.ActionWrite) { return }
   ```

9. **Update main.go** — pass pool:
   ```go
   dynHandler := dynsecret.NewHandler(dynSvc, pool)
   ```

## Todo List
- [ ] Add `ResourceDynSecret` to rbac.go + role perms
- [ ] Add `db` field to dynsecret Handler + update NewHandler
- [ ] Apply rbac.Middleware to project-scoped routes
- [ ] Add `checkProviderAccess` helper
- [ ] Add `GetLeaseProviderID` to service
- [ ] Update createLease with auth check
- [ ] Update listLeases with auth check
- [ ] Update revokeLease with auth check
- [ ] Update main.go dynsecret handler init
- [ ] Verify compilation

## Success Criteria
- All 5 dynsecret routes enforce project membership
- Non-members get 403
- Viewer role can list but not create/revoke

## Risk Assessment
- **Extra queries**: provider-scoped routes add 1-2 queries. Acceptable.
- **revokeLease**: 3 queries (lease->provider->project->membership). Still fast.

## Security Considerations
- Default-deny on missing membership
- Lease creation is a write op (generates real DB credentials)

## Next Steps
- Independent of other phases
