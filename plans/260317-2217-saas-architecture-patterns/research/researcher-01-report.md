# Research Report: SaaS Architecture Patterns for Valt

Date: 2026-03-17
Sources: 5 (crunchydata.com, oneuptime.com, runreveal.com, techbuddies.io, medium.com/one2n.io)

---

## 1. Per-Tenant Resource Counting (PostgreSQL)

### Pattern: Scoped COUNT with composite index

Never issue `COUNT(*) FROM secrets WHERE org_id = $1` without an index on `(org_id)`.
Best practice is a partial composite index covering the most common filter:

```sql
CREATE INDEX idx_secrets_org ON secrets (org_id);
-- or for filtered counts (e.g., active only):
CREATE INDEX idx_secrets_org_active ON secrets (org_id) WHERE deleted_at IS NULL;
```

Query:
```sql
SELECT COUNT(*) FROM secrets WHERE org_id = $1 AND deleted_at IS NULL;
```

### Pattern: Batch counts to avoid N+1

When displaying usage across multiple resources in one request, use GROUP BY:

```sql
SELECT resource_type, COUNT(*)
FROM usage_counters
WHERE org_id = $1
GROUP BY resource_type;
```

Never loop and issue a COUNT per resource type — that is the N+1 anti-pattern.

### Pattern: Materialized usage counters (counter table)

For high-write workloads, maintain a `org_usage` summary table updated via trigger or in the write path:

```sql
-- On INSERT into secrets:
INSERT INTO org_usage (org_id, resource, count)
VALUES ($1, 'secrets', 1)
ON CONFLICT (org_id, resource) DO UPDATE
  SET count = org_usage.count + 1;
```

Read path becomes `O(1)`:
```sql
SELECT count FROM org_usage WHERE org_id = $1 AND resource = 'secrets';
```

Trade-off: counter drift on rollback. Mitigate with periodic reconciliation job.

### Pattern: Window function for paginated lists with count

Avoid a separate `SELECT COUNT(*)` + `SELECT * LIMIT/OFFSET` pair:

```sql
SELECT *, COUNT(*) OVER () AS total
FROM secrets
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT 50 OFFSET 0;
```

---

## 2. Free Tier Enforcement in Go Middleware

### Core decision: check-before-write vs. async tracking

| Approach | Pros | Cons |
|---|---|---|
| Check-before-write (synchronous) | Hard enforcement, simple reasoning | Adds latency to every write, race condition under high concurrency |
| Async usage tracking | No write-path latency | Soft enforcement only, complexity in reconciliation |
| Transactional counter increment | Consistent + enforces hard limit | Advisory lock or SELECT FOR UPDATE needed |

### Recommended: check-before-write with advisory lock in transaction

```go
// middleware/quota.go
func QuotaMiddleware(db *pgxpool.Pool, resource string, limit int) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            orgID := orgFromCtx(r.Context())

            tx, _ := db.Begin(r.Context())
            defer tx.Rollback(r.Context())

            var count int
            err := tx.QueryRow(r.Context(),
                `SELECT count FROM org_usage WHERE org_id = $1 AND resource = $2 FOR UPDATE`,
                orgID, resource,
            ).Scan(&count)

            if err != nil || count >= limit {
                http.Error(w, "quota exceeded", http.StatusPaymentRequired)
                return
            }
            // increment is done by the handler after successful write
            next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), txKey, tx)))
        })
    }
}
```

### Soft vs. Hard limits

- **Hard limit**: block the write entirely (return 402/429). Use for free tier secret count caps.
- **Soft limit**: allow write, flag org for notification (e.g., 80% usage warning). Use for storage quotas where exact enforcement is less critical.

Pattern for soft limit — store threshold state:

```sql
ALTER TABLE orgs ADD COLUMN quota_warned_at timestamptz;
```

Check in background worker, email at 80%, block at 100%.

### Anti-patterns to avoid

- `SELECT COUNT(*) ... ` in middleware without `FOR UPDATE` — race condition allows burst over limit under concurrent requests.
- Checking quota in application layer only, no DB-level guard — a bug bypasses middleware.

---

## 3. Dynamic Secret Provider Authorization (Go)

### Pattern: membership check in handler, not just JWT

JWT proves identity, not membership. After authenticating user, always verify org membership for the requested resource:

```go
// internal/middleware/require_project_member.go
func RequireProjectMember(db *pgxpool.Pool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := userFromCtx(r.Context())
            projectID := chi.URLParam(r, "projectID")

            var role string
            err := db.QueryRow(r.Context(),
                `SELECT role FROM project_members
                 WHERE project_id = $1 AND user_id = $2 AND deleted_at IS NULL`,
                projectID, userID,
            ).Scan(&role)

            if errors.Is(err, pgx.ErrNoRows) {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            ctx := context.WithValue(r.Context(), projectRoleKey, role)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

Apply per route group (chi):
```go
r.Route("/projects/{projectID}/providers", func(r chi.Router) {
    r.Use(RequireProjectMember(db))
    r.Post("/", handlers.CreateProvider)
    r.Delete("/{providerID}", handlers.DeleteProvider)
})
```

### Pattern: org ownership check for provider itself

When mutating a provider by ID, verify it belongs to the org before acting:

```go
var ownerOrgID string
err := db.QueryRow(ctx,
    `SELECT org_id FROM secret_providers WHERE id = $1`,
    providerID,
).Scan(&ownerOrgID)

if ownerOrgID != orgIDFromCtx(ctx) {
    http.Error(w, "forbidden", http.StatusForbidden)
    return
}
```

This prevents IDOR (Insecure Direct Object Reference) — a critical OWASP vulnerability in multi-tenant systems.

### Defense-in-depth: PostgreSQL RLS

```sql
ALTER TABLE secret_providers ENABLE ROW LEVEL SECURITY;
CREATE POLICY providers_org_isolation ON secret_providers
  USING (org_id = current_setting('app.current_org_id')::uuid);
```

Set at query time: `SET LOCAL app.current_org_id = '...'` inside transaction.
RLS catches missing WHERE clauses in application code — defense layer 2.

---

## 4. Scanner Result Ownership Scoping

### Pattern: redundant org_id + project_id on result tables

Every scan-result table carries both `org_id` and `project_id` explicitly — never infer `org_id` via a join to `projects`:

```sql
CREATE TABLE scan_results (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      uuid NOT NULL REFERENCES orgs(id),
    project_id  uuid NOT NULL REFERENCES projects(id),
    scan_id     uuid NOT NULL REFERENCES scans(id),
    -- ...
    CONSTRAINT scan_results_org_matches_project
        CHECK (org_id = (SELECT org_id FROM projects WHERE id = project_id))
);

CREATE INDEX idx_scan_results_project ON scan_results (project_id, org_id);
CREATE INDEX idx_scan_results_org     ON scan_results (org_id);
```

The `CHECK` constraint prevents cross-org references from entering the table even if application code is buggy.

### Pattern: scan initiation ties result ownership at write time

When a scan is started, stamp `org_id` and `project_id` from the initiating request context, not from any user-supplied parameter:

```go
func (h *ScanHandler) CreateScan(w http.ResponseWriter, r *http.Request) {
    orgID    := orgFromCtx(r.Context())   // from JWT + membership middleware
    projectID := chi.URLParam(r, "projectID") // already validated by RequireProjectMember

    _, err := h.db.Exec(ctx,
        `INSERT INTO scans (org_id, project_id, ...) VALUES ($1, $2, ...)`,
        orgID, projectID,
    )
    // ...
}
```

Never trust `org_id` / `project_id` from the request body for ownership-stamping.

### Pattern: result retrieval always double-filters

```sql
SELECT * FROM scan_results
WHERE project_id = $1
  AND org_id     = $2   -- redundant but critical safety net
ORDER BY created_at DESC
LIMIT 50;
```

Double-filtering on both `project_id` and `org_id` costs nothing with a composite index and prevents IDOR if project_id is guessable/enumerable.

### Pattern: scan findings inherit parent ownership

Findings, vulnerabilities, or sub-results tables reference `scan_results(id)` and carry their own `org_id`:

```sql
CREATE TABLE scan_findings (
    id            uuid PRIMARY KEY,
    org_id        uuid NOT NULL REFERENCES orgs(id),
    scan_result_id uuid NOT NULL REFERENCES scan_results(id),
    -- findings data
);
```

This allows direct org-scoped queries on findings without joining up the parent chain.

---

## Summary: Applied to Valt

| Concern | Recommended Pattern |
|---|---|
| Count secrets per org | Composite index on `(org_id)`, counter table for high-write |
| Free tier enforcement | Check-before-write + `SELECT FOR UPDATE` on counter row |
| Soft limit warning | Background reconciliation job + `quota_warned_at` on orgs |
| Provider auth | `RequireProjectMember` middleware on route group + IDOR check in handler |
| Scanner ownership | Stamp `org_id`/`project_id` from context at write time, double-filter on read |
| Defense-in-depth | PostgreSQL RLS + CHECK constraints |

---

## Sources

- [How to Design Multi-Tenant APIs with Tenant Isolation in Go](https://oneuptime.com/blog/post/2026-01-25-multi-tenant-apis-tenant-isolation-go/view)
- [Designing Your Postgres Database for Multi-tenancy — Crunchy Data](https://www.crunchydata.com/blog/designing-your-postgres-database-for-multi-tenancy)
- [OWASP? O'Please. RBAC in Go — RunReveal](https://blog.runreveal.com/owasp-oplease-a-secure-design-pattern-for-role-based-authorization-in-go/)
- [PostgreSQL RLS for Multi-Tenant SaaS — TechBuddies](https://www.techbuddies.io/2026/02/04/how-to-implement-postgresql-row-level-security-for-multi-tenant-saas-2/)
- [Building multi-tenant authorization in Go with Permify — One2N](https://one2n.io/blog/building-multi-tenant-authorization-system-for-b2b-saas-in-go-using-permify)

---

## Unresolved Questions

1. Does Valt already maintain an `org_usage` counter table, or will counts always be live `COUNT(*)` queries? Choice affects free-tier enforcement architecture significantly.
2. Is PostgreSQL RLS already enabled on any tables? Enabling it mid-flight requires careful migration and performance testing.
3. Are scan initiations synchronous (HTTP) or async (queue)? Async path requires stamping ownership at enqueue time, not just at job execution time.
4. What is the exact free tier limit set for secrets/providers? This affects whether a soft-limit warning stage is needed or hard cutoff is sufficient.
