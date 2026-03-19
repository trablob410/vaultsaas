# Phase 01: Email Action Links

**Priority:** P0 — fastest win, zero new infra
**Status:** COMPLETED — 2026-03-19

## Context
- `server/internal/notify/service.go` — current notify service (email only, `to` address is always `""`)
- `server/internal/notify/email.go` — SMTP sender
- `server/internal/workflow/handler.go:159-174` — CreateRequest calls NotifyApprovalNeeded with empty `to`
- Latest migration: `000025_nullable_requester_user_id`

## Requirements
- Signed one-time tokens for approve/reject embedded in email
- Token stored in DB (single-use, TTL 72h)
- Public endpoint (no JWT) — `POST /api/v1/action-tokens/{token}/redeem?action=approve|reject`
- On redemption: call existing `Approve`/`Reject` service methods, return HTML confirmation
- Fix: fetch secret owner email so `to` is populated

## Architecture

```
CreateRequest
  └─▶ fetch owner user email (users table)
  └─▶ actionToken.Create(requestID, "approve", 72h) → approve_token
  └─▶ actionToken.Create(requestID, "reject", 72h)  → reject_token
  └─▶ NotifyApprovalNeeded(ctx, to=owner_email, ..., approveURL, rejectURL)
       └─▶ email.Send with HTML body containing clickable links

Approver clicks link in email
  └─▶ POST /api/v1/action-tokens/{token}/redeem?action=approve
       └─▶ validate token (exists, not used, not expired)
       └─▶ call service.Approve or service.Reject
       └─▶ call credMgr.IssueCredential if approved
       └─▶ return HTML confirmation page (200)
```

## Files

| File | Change |
|------|--------|
| `server/internal/database/migrations/000026_request_action_tokens.up.sql` | Create |
| `server/internal/database/migrations/000026_request_action_tokens.down.sql` | Create |
| `server/internal/notify/action_token.go` | Create |
| `server/internal/notify/service.go` | Modify — add tokenStore, baseURL, approveURL/rejectURL params |
| `server/internal/notify/email.go` | Modify — HTML email with action buttons |
| `server/internal/workflow/handler.go` | Modify — fetch owner email, pass URLs to notify; add ActionToken handler |
| `server/cmd/server/main.go` | Modify — pass BASE_URL to notify.Service; register public action-token route |

---

## Task 1: DB migration — request_action_tokens

**Files:**
- Create: `server/internal/database/migrations/000026_request_action_tokens.up.sql`
- Create: `server/internal/database/migrations/000026_request_action_tokens.down.sql`

- [ ] Write migration up:

```sql
CREATE TABLE request_action_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id  UUID NOT NULL REFERENCES access_requests(id) ON DELETE CASCADE,
    action      TEXT NOT NULL CHECK (action IN ('approve', 'reject')),
    token_hash  TEXT NOT NULL UNIQUE,       -- SHA-256 hex of random token bytes
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_action_tokens_request_id ON request_action_tokens(request_id);
```

- [ ] Write migration down:

```sql
DROP TABLE IF EXISTS request_action_tokens;
```

- [ ] Run: `make migrate-up` — verify no errors

- [ ] Commit:
```bash
git add server/internal/database/migrations/000026_*
git commit -m "feat(db): add request_action_tokens table for email approve/reject links"
```

---

## Task 2: action_token.go — token create/verify/consume

**Files:**
- Create: `server/internal/notify/action_token.go`
- Test: `server/internal/notify/action_token_test.go`

- [ ] Write `action_token.go`:

```go
package notify

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ActionToken represents a single-use approve/reject token.
type ActionToken struct {
	ID        string
	RequestID string
	Action    string
	ExpiresAt time.Time
}

// ActionTokenStore manages action tokens.
type ActionTokenStore struct {
	pool *pgxpool.Pool
}

// NewActionTokenStore creates an ActionTokenStore.
func NewActionTokenStore(pool *pgxpool.Pool) *ActionTokenStore {
	return &ActionTokenStore{pool: pool}
}

// Create generates a random token, stores its hash, returns the raw token.
func (s *ActionTokenStore) Create(ctx context.Context, requestID, action string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	token := hex.EncodeToString(raw)
	hash := sha256Token(token)
	expiresAt := time.Now().Add(ttl)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO request_action_tokens (request_id, action, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		requestID, action, hash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("storing token: %w", err)
	}
	return token, nil
}

// Consume validates the token and marks it used. Returns requestID + action.
func (s *ActionTokenStore) Consume(ctx context.Context, rawToken string) (*ActionToken, error) {
	hash := sha256Token(rawToken)
	var t ActionToken
	err := s.pool.QueryRow(ctx,
		`UPDATE request_action_tokens
		 SET used_at = NOW()
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		 RETURNING id, request_id, action, expires_at`,
		hash,
	).Scan(&t.ID, &t.RequestID, &t.Action, &t.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("token invalid or expired")
	}
	return &t, nil
}

func sha256Token(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
```

- [ ] Write failing tests for Create + Consume (happy path, expired, already-used):

```go
package notify_test
// Use a real test DB or mock pgxpool — follow existing test patterns in server/internal/
```

- [ ] Run: `cd server && go test ./internal/notify/... -run TestActionToken -v`

- [ ] Commit:
```bash
git add server/internal/notify/action_token.go server/internal/notify/action_token_test.go
git commit -m "feat(notify): add ActionTokenStore for single-use approve/reject tokens"
```

---

## Task 3: Update notify.Service — accept tokenStore + baseURL

**Files:**
- Modify: `server/internal/notify/service.go`
- Modify: `server/internal/notify/email.go`

- [ ] Update `service.go`:

```go
package notify

import (
	"context"
	"fmt"
)

type Service struct {
	email     Notifier
	tokens    *ActionTokenStore
	baseURL   string
}

func NewService(email Notifier, tokens *ActionTokenStore, baseURL string) *Service {
	return &Service{email: email, tokens: tokens, baseURL: baseURL}
}

// NotifyApprovalNeeded sends email with approve/reject action links.
func (s *Service) NotifyApprovalNeeded(ctx context.Context, to, requestID, secretName, requester, reason string) error {
	if s.email == nil || to == "" {
		return nil
	}
	approveToken, err := s.tokens.Create(ctx, requestID, "approve", 72*time.Hour)
	if err != nil {
		return fmt.Errorf("creating approve token: %w", err)
	}
	rejectToken, err := s.tokens.Create(ctx, requestID, "reject", 72*time.Hour)
	if err != nil {
		return fmt.Errorf("creating reject token: %w", err)
	}
	approveURL := s.baseURL + "/api/v1/action-tokens/" + approveToken + "/redeem?action=approve"
	rejectURL  := s.baseURL + "/api/v1/action-tokens/" + rejectToken + "/redeem?action=reject"

	subject := "Valt: Access Request Needs Your Approval"
	body := buildApprovalEmail(secretName, requester, reason, approveURL, rejectURL)
	return s.email.Send(ctx, to, subject, body)
}

// NotifyAccessGranted — unchanged
func (s *Service) NotifyAccessGranted(ctx context.Context, to, secretName string, durationMin int) error {
	if s.email == nil {
		return nil
	}
	subject := "Valt: Access Granted"
	body := "Access to secret '" + secretName + "' has been granted.\nDuration: " + intToStr(durationMin) + " minutes."
	return s.email.Send(ctx, to, subject, body)
}
```

- [ ] Add `buildApprovalEmail` to `email.go`:

```go
func buildApprovalEmail(secretName, requester, reason, approveURL, rejectURL string) string {
	return fmt.Sprintf(`Secret access request requires your approval.

Secret:    %s
Requester: %s
Reason:    %s

Approve: %s
Reject:  %s

This link expires in 72 hours and can only be used once.
`, secretName, requester, reason, approveURL, rejectURL)
}
```

- [ ] Run: `cd server && go build ./...` — must compile clean

- [ ] Commit:
```bash
git add server/internal/notify/service.go server/internal/notify/email.go
git commit -m "feat(notify): embed approve/reject action links in approval emails"
```

---

## Task 4: Add ActionToken handler to workflow

**Files:**
- Modify: `server/internal/workflow/handler.go`

- [ ] Add `RedeemActionToken` handler — resolves token → calls Approve/Reject → returns HTML:

```go
// RedeemActionToken handles POST /action-tokens/{token}/redeem?action=approve|reject
// This is a PUBLIC endpoint (no JWT required).
func (h *Handler) RedeemActionToken(w http.ResponseWriter, r *http.Request) {
	rawToken := chi.URLParam(r, "token")
	action := r.URL.Query().Get("action")
	if action != "approve" && action != "reject" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	tok, err := h.tokenStore.Consume(r.Context(), rawToken)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "<h2>Link expired or already used.</h2>")
		return
	}
	if tok.Action != action {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "<h2>Invalid action for this link.</h2>")
		return
	}

	switch action {
	case "approve":
		req, err := h.service.Approve(r.Context(), tok.RequestID, "email-action")
		if err != nil {
			http.Error(w, "Failed to approve: "+err.Error(), http.StatusBadRequest)
			return
		}
		secret, _ := h.vaultSvc.GetSecretByID(r.Context(), req.SecretID)
		credType := ""
		if secret != nil { credType = secret.CredentialType }
		_, _ = h.credMgr.IssueCredential(r.Context(), req.ID, credType, req.RequestedDurationMinutes)
		h.auditLog.LogFromRequest(r, "email-action", "access_request.approve", "access_request", req.ID)
	case "reject":
		_, err := h.service.Reject(r.Context(), tok.RequestID, "email-action", "Rejected via email link")
		if err != nil {
			http.Error(w, "Failed to reject: "+err.Error(), http.StatusBadRequest)
			return
		}
		h.auditLog.LogFromRequest(r, "email-action", "access_request.reject", "access_request", tok.RequestID)
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h2>Request %sd successfully.</h2><p>You can close this tab.</p>", action)
}
```

- [ ] Add `tokenStore *notify.ActionTokenStore` field to `Handler` struct and `NewHandler` signature

- [ ] Fix `CreateRequest` — fetch owner email and pass requestID to NotifyApprovalNeeded:

```go
// After fetching secret, fetch owner email:
var ownerEmail string
_ = h.pool.QueryRow(r.Context(),
    `SELECT email FROM users WHERE id = $1`, secret.UserID,
).Scan(&ownerEmail)

// Replace the notify call:
if p.RequireApproval && h.notifySvc != nil {
    _ = h.notifySvc.NotifyApprovalNeeded(r.Context(), ownerEmail, req.ID, secret.Name, callerDesc, body.Reason)
}
```

- [ ] Run: `cd server && go build ./...`

- [ ] Commit:
```bash
git add server/internal/workflow/handler.go
git commit -m "feat(workflow): add email action token redemption endpoint; fix notify recipient"
```

---

## Task 5: Wire in main.go

**Files:**
- Modify: `server/cmd/server/main.go`

- [ ] Add `BASE_URL` env var to config (`server/internal/config/config.go`):

```go
BaseURL string `envconfig:"BASE_URL" default:"http://localhost:8080"`
```

- [ ] Wire ActionTokenStore and update NewService + NewHandler calls:

```go
tokenStore := notify.NewActionTokenStore(pool)
notifySvc := notify.NewService(emailNotifier, tokenStore, cfg.BaseURL)
workflowHandler := workflow.NewHandler(workflowSvc, credMgr, vaultService, auditLogger, notifySvc, masterKey, pool, tokenStore)
```

- [ ] Register **public** route (outside the authenticated group):

```go
r.Post("/api/v1/action-tokens/{token}/redeem", workflowHandler.RedeemActionToken)
```

- [ ] Run: `cd server && go build ./cmd/server` — must compile

- [ ] Run: `make test-unit` — all existing tests must pass

- [ ] Commit:
```bash
git add server/cmd/server/main.go server/internal/config/config.go
git commit -m "feat(server): wire email action tokens into notify service and public route"
```

---

## Task 6: Integration test

- [ ] Write test in `server/tests/integration/` (if integration test suite exists):
  - Create request → check email body contains approve/reject URLs
  - Hit redeem endpoint with valid token → status 200 HTML
  - Hit redeem endpoint again → status 410 Gone
  - Hit redeem endpoint with expired token → status 410 Gone

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/tests/
git commit -m "test(workflow): email action token integration tests"
```

---

## Success Criteria
- Email sent on new access request contains two clickable links
- Clicking Approve link → request transitions to `approved`, credential issued, HTML confirmation
- Clicking Reject link → request transitions to `rejected`, HTML confirmation
- Link cannot be reused after first click (409 or 410 response)
- Link expires after 72h

## Risks
- `to` email may be empty if user has no email (Google OAuth users always have email — low risk)
- Double-click race condition: `UPDATE ... WHERE used_at IS NULL` is atomic — safe
