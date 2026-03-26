---
phase: "3.3"
title: "Generic Webhooks API"
priority: P2
status: pending
effort: 5h
---

# Phase 3.3: Generic Webhooks API

## Context Links
- [Notify service](../../server/internal/notify/service.go) — existing notification dispatch
- [Org service](../../server/internal/org/service.go) — org scoping pattern
- [Config](../../server/internal/config/config.go)
- [Main](../../server/cmd/server/main.go)

## Overview

Allow orgs to register webhook URLs that receive HMAC-signed HTTP POST payloads on approval events. Enables third-party integrations (PagerDuty, custom dashboards, CI/CD).

## Requirements

### Functional
- CRUD: register, list, delete webhook URLs per org
- Each webhook has: URL, secret_key (auto-generated), event filter, active toggle
- Events: `request.created`, `request.approved`, `request.rejected`, `credential.issued`, `credential.revoked`
- On event: POST JSON payload to all active webhooks matching the event
- Payload signed: `X-Valt-Signature: sha256=HMAC(secret_key, body)`
- 5s timeout per webhook call, fire-and-forget (async)

### Non-Functional
- Max 10 webhooks per org
- Secret key auto-generated (32-byte hex)
- Webhook URLs must be HTTPS
- Response status logged but not blocking

## Architecture

```
Approval event → notify/service.go → webhooks.Dispatch(ctx, orgID, event, payload)
  → for each active webhook matching event:
      → goroutine: HMAC-sign body, POST to URL, log result
```

## Related Code Files

### Create
- `server/internal/webhooks/store.go` — DB CRUD
- `server/internal/webhooks/handler.go` — REST endpoints
- `server/internal/webhooks/dispatcher.go` — async event dispatch
- `server/internal/database/migrations/000032_org_webhooks.up.sql`
- `server/internal/database/migrations/000032_org_webhooks.down.sql`

### Modify
- `server/internal/notify/service.go` — call dispatcher on events
- `server/cmd/server/main.go` — wire routes

## Implementation Steps

### 1. DB Migration (000032)

```sql
-- 000032_org_webhooks.up.sql
CREATE TABLE org_webhooks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    secret_key  VARCHAR(64) NOT NULL,
    events      TEXT[] NOT NULL DEFAULT '{}',
    active      BOOLEAN NOT NULL DEFAULT true,
    description VARCHAR(255) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_org_webhooks_org ON org_webhooks(org_id);
```

### 2. webhooks/store.go (~90 lines)

```go
package webhooks

type Webhook struct {
    ID          string   `json:"id"`
    OrgID       string   `json:"org_id"`
    URL         string   `json:"url"`
    SecretKey   string   `json:"secret_key"` // shown only on create
    Events      []string `json:"events"`
    Active      bool     `json:"active"`
    Description string   `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}

type Store struct { pool *pgxpool.Pool }

func NewStore(pool) *Store
func (s *Store) Create(ctx, orgID, url, description string, events []string) (*Webhook, error)
  // Validate: URL starts with "https://", len(events) > 0
  // Generate secret_key: hex.EncodeToString(random 32 bytes)
  // Enforce max 10 per org: SELECT COUNT(*) ... < 10
func (s *Store) List(ctx, orgID string) ([]Webhook, error)
  // Return all webhooks; MASK secret_key (return only last 8 chars)
func (s *Store) Delete(ctx, orgID, webhookID string) error
func (s *Store) ListActive(ctx, orgID string, event string) ([]Webhook, error)
  // WHERE org_id=$1 AND active=true AND $2 = ANY(events)
  // Returns full secret_key (internal use only)
```

### 3. webhooks/handler.go (~70 lines)

```go
type Handler struct { store *Store }

func (h *Handler) Create(w, r)  // POST /orgs/{org_id}/webhooks
func (h *Handler) List(w, r)    // GET /orgs/{org_id}/webhooks
func (h *Handler) Delete(w, r)  // DELETE /orgs/{org_id}/webhooks/{id}
```

Auth: require org owner/admin via org membership check in handler.

Create body: `{url, events, description}`. Return full webhook with secret_key (only time it's shown).

### 4. webhooks/dispatcher.go (~60 lines)

```go
type Dispatcher struct {
    store  *Store
    client *http.Client // Timeout: 5s
}

func NewDispatcher(store) *Dispatcher
func (d *Dispatcher) Dispatch(ctx, orgID, event string, payload interface{})
```

`Dispatch` logic:
1. `store.ListActive(ctx, orgID, event)`
2. For each webhook, spawn goroutine:
   - Marshal payload to JSON
   - Compute `HMAC-SHA256(secret_key, body)`
   - POST to URL with headers:
     - `Content-Type: application/json`
     - `X-Valt-Signature: sha256=<hex>`
     - `X-Valt-Event: <event>`
     - `X-Valt-Delivery: <uuid>` (unique per delivery)
   - Log response status (ignore errors — fire-and-forget)

### 5. Wire dispatcher into notify service

In `notify/service.go`, add `webhookDispatcher *webhooks.Dispatcher` field.

Call `webhookDispatcher.Dispatch` at the end of:
- `NotifyApprovalNeeded` → event `request.created`

Add new methods or hook points for:
- After approve → `request.approved`
- After reject → `request.rejected`

Simplest approach: add `DispatchWebhookEvent(orgID, event, payload)` method to Service, called from `workflow/handler.go` after approve/reject/create.

### 6. Webhook event payload shape

```json
{
  "event": "request.approved",
  "timestamp": "2026-03-19T12:00:00Z",
  "data": {
    "request_id": "uuid",
    "secret_id": "uuid",
    "secret_name": "prod-db-password",
    "requester": "user@example.com",
    "status": "approved",
    "decided_by": "admin@example.com"
  }
}
```

### 7. Wire in main.go

```go
webhookStore := webhooks.NewStore(pool)
webhookDispatcher := webhooks.NewDispatcher(webhookStore)
// Pass dispatcher to notify service or workflow handler

// Authenticated routes:
r.Post("/orgs/{org_id}/webhooks", webhookHandler.Create)
r.Get("/orgs/{org_id}/webhooks", webhookHandler.List)
r.Delete("/orgs/{org_id}/webhooks/{id}", webhookHandler.Delete)
```

### 8. Dashboard — Webhook Management UI

New section in org settings page:
- List registered webhooks (URL, events, active status, masked secret)
- "Add Webhook" form: URL input, event checkboxes, description
- On create: show secret_key once with copy button + warning
- Delete button per webhook

## Todo

- [ ] Create migration 000032
- [ ] Implement webhooks/store.go
- [ ] Implement webhooks/handler.go
- [ ] Implement webhooks/dispatcher.go
- [ ] Wire dispatcher into notify service / workflow handler
- [ ] Wire routes in main.go
- [ ] Dashboard: webhook management UI
- [ ] Unit tests: HMAC signature, dispatch, store CRUD

## Success Criteria
- Org admin can register/delete webhook URLs
- Events fire HTTP POST with HMAC signature on approval actions
- Signature verifiable by consumer using `X-Valt-Signature` header
- Max 10 webhooks enforced per org
- Secret key shown only once on creation

## Security Considerations
- Webhook URLs must be HTTPS (reject HTTP)
- HMAC-SHA256 signature prevents payload tampering
- Secret key shown once, masked in list responses
- 5s timeout prevents slow-loris blocking server goroutines
- Org owner/admin authorization on all endpoints
- No SSRF risk: only POST to external URLs (no internal network access unless self-hosted)
