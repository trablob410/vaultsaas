# Phase 5 — High-Priority Fixes (H1-H5)

## Context Links
- `server/internal/workflow/handler.go:216` — GetRequest restricts to requester only (H1)
- `server/cmd/server/main.go:132` — `_ = agentRateLimiter` unused (H2)
- `server/internal/workflow/handler.go:183-187` — IssueCredential error swallowed (H3)
- `mcp-server/src/scanner_tools.rs:8-11` — path from args used directly (H4)
- `server/internal/workflow/approval-chain.go:84` — 6th param `_ string` discarded (H5)
- `server/internal/database/migrations/000021_create_approval_steps.up.sql` — no rejection_reason column

---

## H1 — GetRequest Excludes Approvers

### Overview
`GetRequest` at handler.go:216 only allows the requester to view. Secret owners and assigned approvers are blocked.

### Implementation Steps

1. **Expand access check** in `GetRequest` (handler.go ~line 216):
   ```go
   if req.RequesterUserID != userID {
       // Check if user is the secret owner
       secret, _ := h.vaultSvc.GetSecretByID(r.Context(), req.SecretID)
       isOwner := secret != nil && secret.OwnerUserID == userID

       // Check if user is an approver for this request
       isApprover := false
       if !isOwner {
           var cnt int
           _ = h.service.pool.QueryRow(r.Context(),
               `SELECT COUNT(*) FROM approval_steps WHERE request_id = $1 AND approver_user_id = $2`,
               requestID, userID).Scan(&cnt)
           isApprover = cnt > 0
       }

       if !isOwner && !isApprover {
           apierror.Forbidden(w, "not your request")
           return
       }
   }
   ```

2. **Alternative (cleaner):** Add `IsApproverOrOwner(ctx, requestID, userID)` method to workflow Service. Returns bool. Handler calls it when requester check fails.

   Add to `server/internal/workflow/service.go`:
   ```go
   func (s *Service) IsApproverOrOwner(ctx context.Context, requestID, userID, secretID string) (bool, error) {
       // Check approval_steps
       var cnt int
       err := s.pool.QueryRow(ctx,
           `SELECT COUNT(*) FROM approval_steps WHERE request_id = $1 AND approver_user_id = $2`,
           requestID, userID).Scan(&cnt)
       if err != nil { return false, err }
       return cnt > 0, nil
   }
   ```

   Handler checks: requester OR isApprover OR isOwner (via GetSecretByID).

### Files to Modify
- `server/internal/workflow/handler.go` (GetRequest method)
- `server/internal/workflow/service.go` (add helper)

---

## H2 — Redis Rate Limiter Unused

### Overview
`agentRateLimiter` is initialized at main.go:122-131 but assigned to `_` at line 132. Never applied to any route.

### Implementation Steps

1. **Create a middleware wrapper** in `server/internal/ratelimit/middleware.go`:
   ```go
   package ratelimit

   func (l *RedisLimiter) Middleware(rpm int) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               // Extract agent ID from context (set by agent auth middleware)
               agentID := r.Header.Get("X-Agent-ID")
               if agentID == "" {
                   next.ServeHTTP(w, r)
                   return
               }
               allowed, _ := l.Allow(r.Context(), agentID, rpm)
               if !allowed {
                   http.Error(w, `{"error":"rate limited"}`, http.StatusTooManyRequests)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

2. **Apply in main.go** — remove `_ = agentRateLimiter` and apply to the authenticated route group:
   ```go
   if agentRateLimiter != nil {
       r.Use(agentRateLimiter.Middleware(60)) // 60 rpm per agent
   }
   ```

   Note: Since there's no separate agent route group currently, apply to the main auth group. The middleware is a no-op if X-Agent-ID header is missing (human users pass through).

### Files to Modify
- `server/internal/ratelimit/middleware.go` (new file, ~30 lines)
- `server/cmd/server/main.go` (remove `_ =`, apply middleware)

---

## H3 — IssueCredential Failure Swallowed on Approve

### Overview
At handler.go:184-187, if `IssueCredential` fails after approval, error is logged but response is 200 with the approved request. Client thinks credential is ready but it isn't.

### Implementation Steps

1. **Return 500 on issuance failure** (handler.go ~line 184):
   ```go
   _, issueErr := h.credMgr.IssueCredential(r.Context(), req.ID, credType, req.RequestedDurationMinutes)
   if issueErr != nil {
       log.Printf("Failed to issue credential after approval: %v", issueErr)
       apierror.InternalError(w, "approved but failed to issue credential")
       return
   }
   ```

### Files to Modify
- `server/internal/workflow/handler.go` (Approve method, ~line 184-187)

---

## H4 — MCP scan_secrets Accepts Arbitrary Paths

### Overview
`scanner_tools.rs:8` takes `path` from JSON-RPC args and passes directly to `scanner::scan_directory`. An attacker could scan `/etc/shadow`, `C:\Windows\System32`, etc.

### Implementation Steps

1. **Reject absolute paths** in `scanner_tools.rs`:
   ```rust
   let path = args["path"].as_str()
       .ok_or_else(|| crate::error::ValtError::Protocol("path required".into()))?;

   // Validate path safety
   if path.starts_with('/') || path.starts_with('\\') || path.contains("..") {
       return Err(crate::error::ValtError::Protocol(
           "path must be relative and cannot contain '..'".into()
       ));
   }
   if path.len() > 500 {
       return Err(crate::error::ValtError::Protocol("path too long".into()));
   }
   ```

2. **Limit scan depth** — pass max_depth to scan_directory if supported, or add it. Default 5 levels.

### Files to Modify
- `mcp-server/src/scanner_tools.rs` (tool_scan_secrets function)

---

## H5 — AdvanceChain Discards Rejection Reason

### Overview
`AdvanceChain` at approval-chain.go:84 takes 6th param `_ string` (rejection reason) but never stores it. The `approval_steps` table has no `rejection_reason` column.

### Implementation Steps

1. **Create migration `000024_add_rejection_reason.up.sql`:**
   ```sql
   ALTER TABLE approval_steps ADD COLUMN rejection_reason TEXT;
   ```

2. **Create `000024_add_rejection_reason.down.sql`:**
   ```sql
   ALTER TABLE approval_steps DROP COLUMN IF EXISTS rejection_reason;
   ```

3. **Update `AdvanceChain` signature** — rename `_ string` to `reason string`:
   ```go
   func AdvanceChain(ctx context.Context, db *pgxpool.Pool, requestID, approverID string, approve bool, reason string) (bool, error) {
   ```

4. **Update the UPDATE query** (line 119-121):
   ```go
   if _, err = tx.Exec(ctx,
       `UPDATE approval_steps SET status = $1, decided_at = NOW(), rejection_reason = $2 WHERE id = $3`,
       newStatus, reason, step.ID,
   ); err != nil {
   ```

5. **Update callers** of AdvanceChain — grep for all call sites, ensure they pass the reason string.

6. **Add `RejectionReason` to ApprovalStep struct:**
   ```go
   RejectionReason *string `json:"rejection_reason,omitempty"`
   ```
   Update all Scan calls that read approval_steps to include the new column.

### Files to Modify
- `server/internal/database/migrations/000024_add_rejection_reason.up.sql` (new)
- `server/internal/database/migrations/000024_add_rejection_reason.down.sql` (new)
- `server/internal/workflow/approval-chain.go`

---

## Combined Todo List
- [ ] H1: Add IsApproverOrOwner helper to workflow service
- [ ] H1: Update GetRequest to allow owner/approver access
- [ ] H2: Create ratelimit/middleware.go
- [ ] H2: Apply rate limiter in main.go
- [ ] H3: Return 500 on IssueCredential failure in Approve handler
- [ ] H4: Add path validation in scanner_tools.rs
- [ ] H5: Create migration 000024
- [ ] H5: Update AdvanceChain to store rejection_reason
- [ ] H5: Update ApprovalStep struct + scan calls
- [ ] Verify Go compilation
- [ ] Verify Rust compilation (`cargo check`)

## Success Criteria
- H1: Approver and secret owner can view access request
- H2: Agent requests rate-limited at 60rpm; human requests unaffected
- H3: Failed credential issuance returns 500
- H4: Absolute paths and `..` rejected by MCP tool
- H5: Rejection reasons persisted and queryable

## Risk Assessment
- H2: Rate limiter middleware is new code path. Fail-open on Redis errors (existing behavior in Allow).
- H5: Migration adds nullable column — no impact on existing rows.

## Security Considerations
- H4 is a path traversal prevention — defense in depth since MCP runs on client machine, but still prevents scanning sensitive system dirs
- H2 prevents agent API abuse (DoS from compromised agent tokens)
- H3 prevents clients from assuming credential is ready when it isn't
