# Phase 02: Policy API Endpoints

**Priority:** P1 — requires Phase 01 complete
**Status:** COMPLETED — 2026-03-19

## Endpoints

```
GET  /api/v1/secrets/{id}/policy    → returns current secret policy_config (or {})
PUT  /api/v1/secrets/{id}/policy    → save custom policy for secret
GET  /api/v1/projects/{id}/policy   → returns current project policy_config (or {})
PUT  /api/v1/projects/{id}/policy   → save custom policy for project
```

Auth: JWT required. RBAC: `secret write` for secret policy; `project admin` for project policy.

## Files

| File | Change |
|------|--------|
| `server/internal/vault/handler.go` | Modify — add GetPolicy / PutPolicy handlers |
| `server/internal/project/handler.go` | Modify — add GetPolicy / PutPolicy handlers |
| `server/cmd/server/main.go` | Modify — register policy routes |

---

## Task 1: Secret policy endpoints

**Files:**
- Modify: `server/internal/vault/handler.go`

- [ ] Add `GetSecretPolicy` handler:

```go
// GetSecretPolicy handles GET /secrets/{id}/policy
func (h *Handler) GetSecretPolicy(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "id")
	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret_id")
		return
	}

	var policyJSON []byte
	err := h.pool.QueryRow(r.Context(),
		`SELECT COALESCE(policy_config, '{}') FROM secrets WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		secretID, userID,
	).Scan(&policyJSON)
	if err != nil {
		apierror.NotFound(w, "secret not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(policyJSON)
}
```

- [ ] Add `PutSecretPolicy` handler:

```go
// PutSecretPolicy handles PUT /secrets/{id}/policy
func (h *Handler) PutSecretPolicy(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "id")
	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret_id")
		return
	}

	var cp policy.CustomPolicy
	if err := json.NewDecoder(r.Body).Decode(&cp); err != nil {
		apierror.BadRequest(w, "invalid policy body")
		return
	}
	if err := cp.Validate(); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	policyJSON, _ := json.Marshal(cp)
	tag, err := h.pool.Exec(r.Context(),
		`UPDATE secrets SET policy_config = $1 WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL`,
		policyJSON, secretID, userID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		apierror.NotFound(w, "secret not found or not authorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(policyJSON)
}
```

- [ ] Register routes in `main.go` (inside authenticated + RBAC group for secrets):

```go
r.Get("/secrets/{id}/policy",  vaultHandler.GetSecretPolicy)
r.Put("/secrets/{id}/policy",  vaultHandler.PutSecretPolicy)
```

- [ ] Run: `cd server && go build ./cmd/server`

- [ ] Write unit tests covering: not-found secret, invalid JSON, validation failure, success:

```go
// server/internal/vault/handler_policy_test.go
func TestPutSecretPolicy_ValidationError(t *testing.T) { ... }
func TestPutSecretPolicy_NotFound(t *testing.T) { ... }
func TestPutSecretPolicy_Success(t *testing.T) { ... }
```

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/vault/handler.go server/internal/vault/handler_policy_test.go server/cmd/server/main.go
git commit -m "feat(vault): add GET/PUT /secrets/{id}/policy endpoints"
```

---

## Task 2: Project policy endpoints

**Files:**
- Modify: `server/internal/project/handler.go`

- [ ] Add `GetProjectPolicy` handler (same pattern as secret, scoped to project owner/admin):

```go
// GetProjectPolicy handles GET /projects/{id}/policy
// Auth: must be project member (RBAC middleware handles it).
func (h *Handler) GetProjectPolicy(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	var policyJSON []byte
	err := h.pool.QueryRow(r.Context(),
		`SELECT COALESCE(policy_config, '{}') FROM projects WHERE id = $1`,
		projectID,
	).Scan(&policyJSON)
	if err != nil {
		apierror.NotFound(w, "project not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(policyJSON)
}
```

- [ ] Add `PutProjectPolicy` handler (admin only, validated by RBAC middleware):

```go
// PutProjectPolicy handles PUT /projects/{id}/policy
func (h *Handler) PutProjectPolicy(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	var cp policy.CustomPolicy
	if err := json.NewDecoder(r.Body).Decode(&cp); err != nil {
		apierror.BadRequest(w, "invalid policy body")
		return
	}
	if err := cp.Validate(); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}
	policyJSON, _ := json.Marshal(cp)
	tag, err := h.pool.Exec(r.Context(),
		`UPDATE projects SET policy_config = $1 WHERE id = $2`,
		policyJSON, projectID,
	)
	if err != nil || tag.RowsAffected() == 0 {
		apierror.NotFound(w, "project not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(policyJSON)
}
```

- [ ] Register in `main.go` with RBAC admin guard:

```go
// Inside project routes group — wrap with rbac.Middleware(pool, "id", rbac.ResourceProject, "admin")
r.With(rbac.Middleware(pool, "id", rbac.ResourceProject, "admin")).Get("/projects/{id}/policy", projectHandler.GetProjectPolicy)
r.With(rbac.Middleware(pool, "id", rbac.ResourceProject, "admin")).Put("/projects/{id}/policy", projectHandler.PutProjectPolicy)
```

- [ ] Run: `cd server && go build ./cmd/server`

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/project/handler.go server/cmd/server/main.go
git commit -m "feat(project): add GET/PUT /projects/{id}/policy endpoints with RBAC admin guard"
```

---

## Task 3: Wire custom approvers into approval_steps

**Files:**
- Modify: `server/internal/workflow/service.go`

Currently `CreateRequest` inserts into `access_requests` but doesn't create `approval_steps` entries. Custom approvers need to be inserted so `IsAssignedApprover` works.

- [ ] After inserting the access request, if `customApprovers` is non-empty, insert approval_steps:

```go
if len(customApprovers) > 0 {
	for i, approverID := range customApprovers {
		_, err := s.pool.Exec(ctx,
			`INSERT INTO approval_steps (request_id, approver_user_id, step_order, status)
			 VALUES ($1, $2, $3, 'pending')`,
			req.ID, approverID, i+1,
		)
		if err != nil {
			return nil, fmt.Errorf("creating approval step: %w", err)
		}
	}
}
```

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/workflow/service.go
git commit -m "feat(workflow): insert custom approvers into approval_steps from policy config"
```

---

## Success Criteria
- `GET /secrets/{id}/policy` returns `{}` for new secrets, saved config for updated ones
- `PUT /secrets/{id}/policy` validates and persists; rejects invalid business_hours format
- `PUT /projects/{id}/policy` returns 403 for non-admin project members
- Custom approvers in secret policy appear as `approval_steps` for new requests
