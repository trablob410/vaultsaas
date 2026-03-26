---
phase: "3.2"
title: "Slack OAuth per-org"
priority: P1
status: pending
effort: 6h
---

# Phase 3.2: Slack OAuth per-org

## Context Links
- [Slack adapter](../../server/internal/notify/slack.go) — current SLACK_BOT_TOKEN env var pattern
- [Slack webhook](../../server/internal/notify/slack_webhook.go) — signing secret verification
- [Notify service](../../server/internal/notify/service.go) — dispatches to SlackAdapter
- [Config](../../server/internal/config/config.go) — current SlackBotToken/SlackSigningSecret
- [Notification settings page](../../dashboard/src/app/(dashboard)/settings/notifications/page.tsx)

## Overview

Replace single SLACK_BOT_TOKEN env var with per-org OAuth install. Each org connects their own Slack workspace. Bot token stored encrypted in `org_integrations` table. Notify service resolves token per org at send time.

## Key Insights
- Current SlackAdapter takes a single botToken in constructor — needs to accept token per-call
- Slack OAuth v2 flow: redirect to Slack → callback with `code` → exchange for `access_token`
- Required scope: `chat:write` (send DMs), `users:read` (resolve user IDs)
- `SLACK_SIGNING_SECRET` remains global (one Slack app, multiple workspace installs)
- Remove per-user Slack handle from notification channels — becomes org-level install

## Requirements

### Functional
- `org_integrations` table: stores encrypted access_token per org per provider
- GET /oauth/slack?org_id= — redirects to Slack OAuth authorize URL
- GET /oauth/slack/callback — exchanges code for token, stores encrypted
- Notify service reads Slack token from org_integrations per org (fallback to env var)
- Dashboard: "Connect Slack" button in org settings
- Dashboard: "Disconnect Slack" button when connected

### Non-Functional
- Access token encrypted at rest (AES-256-GCM with masterKey)
- Only org owner/admin can initiate OAuth flow
- OAuth state parameter prevents CSRF

## Architecture

```
Org admin clicks "Connect Slack"
  → GET /oauth/slack?org_id=XXX
  → Go generates state token (orgID + random), stores in DB/cache
  → redirect 302 to https://slack.com/oauth/v2/authorize?client_id=X&scope=chat:write,users:read&state=STATE&redirect_uri=CALLBACK
Slack → user authorizes
  → GET /oauth/slack/callback?code=CODE&state=STATE
  → Go verifies state, exchanges code for token via Slack API
  → Encrypt token, store in org_integrations
  → Redirect to dashboard org settings
```

## Related Code Files

### Create
- `server/internal/integration/store.go` — org_integrations CRUD
- `server/internal/integration/slack_oauth.go` — OAuth flow handlers
- `server/internal/database/migrations/000031_org_integrations.up.sql`
- `server/internal/database/migrations/000031_org_integrations.down.sql`

### Modify
- `server/internal/config/config.go` — add SLACK_CLIENT_ID, SLACK_CLIENT_SECRET
- `server/internal/notify/slack.go` — accept token per-call instead of constructor
- `server/internal/notify/service.go` — resolve Slack token per org
- `server/cmd/server/main.go` — wire OAuth routes + integration store
- `dashboard/src/app/(dashboard)/settings/notifications/page.tsx` — or new org settings section

## Implementation Steps

### 1. DB Migration (000031)

```sql
-- 000031_org_integrations.up.sql
CREATE TABLE org_integrations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider        VARCHAR(50) NOT NULL, -- 'slack'
    access_token_enc BYTEA NOT NULL,      -- AES-256-GCM encrypted
    workspace_id    VARCHAR(100),          -- Slack team ID
    team_name       VARCHAR(255),          -- Slack workspace name
    bot_user_id     VARCHAR(100),          -- Slack bot user ID
    metadata_json   JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, provider)
);
CREATE INDEX idx_org_integrations_org ON org_integrations(org_id);
```

### 2. Config — Add Slack OAuth env vars

```go
SlackClientID     string `envconfig:"SLACK_CLIENT_ID" default:""`
SlackClientSecret string `envconfig:"SLACK_CLIENT_SECRET" default:""`
```

Keep `SLACK_BOT_TOKEN` as fallback for single-tenant deployments.

### 3. integration/store.go (~90 lines)

```go
package integration

type OrgIntegration struct {
    ID            string
    OrgID         string
    Provider      string
    AccessToken   string // decrypted in-memory only
    WorkspaceID   string
    TeamName      string
    BotUserID     string
}

type Store struct {
    pool      *pgxpool.Pool
    masterKey []byte
}

func NewStore(pool, masterKey) *Store
func (s *Store) Upsert(ctx, orgID, provider, accessToken, workspaceID, teamName, botUserID string) error
func (s *Store) Get(ctx, orgID, provider string) (*OrgIntegration, error) // decrypt token
func (s *Store) Delete(ctx, orgID, provider string) error
func (s *Store) GetSlackToken(ctx, orgID string) (string, error) // convenience
```

Encryption: use `crypto.EncryptAES256GCM(masterKey, []byte(accessToken))` for storage. Decrypt on read.

### 4. integration/slack_oauth.go (~100 lines)

```go
type SlackOAuthHandler struct {
    store        *Store
    clientID     string
    clientSecret string
    redirectURI  string // BASE_URL + "/api/v1/oauth/slack/callback"
    dashboardURL string
    pool         *pgxpool.Pool
}

func (h *SlackOAuthHandler) Authorize(w, r) // GET /oauth/slack
func (h *SlackOAuthHandler) Callback(w, r)  // GET /oauth/slack/callback
```

`Authorize`:
- Extract `org_id` from query param
- Verify caller is org owner/admin (from JWT via AuthMiddleware)
- Generate random state, store in a temp table or short-lived DB row: `INSERT INTO oauth_states (state, org_id, expires_at)`
  - Alternative: encode orgID + HMAC in state param (stateless, simpler). HMAC = HMAC-SHA256(masterKey, orgID + timestamp). State = base64(orgID|timestamp|hmac).
- Redirect to `https://slack.com/oauth/v2/authorize?client_id=X&scope=chat:write,users:read&state=STATE&redirect_uri=CALLBACK`

`Callback`:
- Validate state (decode + verify HMAC + check timestamp < 10 min)
- Extract `code` from query
- POST `https://slack.com/api/oauth.v2.access` with `{client_id, client_secret, code, redirect_uri}`
- Parse response: `access_token`, `team.id`, `team.name`, `bot_user_id`
- `store.Upsert(orgID, "slack", access_token, team_id, team_name, bot_user_id)`
- Redirect to `dashboardURL + "/settings/notifications?slack=connected"`

### 5. Update SlackAdapter — Token per-call

Change `SlackAdapter` from holding a fixed token to accepting token per method:

```go
// Before: s.botToken used in every call
// After: token passed as parameter

func (s *SlackAdapter) SendApprovalRequest(ctx, token, slackUserID, requestID, ...) error
func (s *SlackAdapter) UpdateMessage(ctx, token, channelID, ts, outcome string) error
func (s *SlackAdapter) post(ctx, token, url string, payload interface{}) error
```

Constructor becomes: `NewSlackAdapter() *SlackAdapter` (no token arg). Returns nil only if explicitly disabled. The adapter is just an HTTP client wrapper now.

Fallback: if `SLACK_BOT_TOKEN` is set, use it as default when no org token found.

### 6. Update notify/service.go

Add integration store to Service:

```go
type Service struct {
    // ...existing fields...
    integrations *integration.Store
    fallbackSlackToken string
}
```

In `NotifyApprovalNeeded`, resolve Slack token:
1. Find org_id from secret → project → workspace → org chain
2. `integrations.GetSlackToken(ctx, orgID)`
3. If not found, fall back to `fallbackSlackToken` (from env var)
4. Pass resolved token to `slack.SendApprovalRequest(ctx, token, ...)`

### 7. Wire in main.go

```go
integrationStore := integration.NewStore(pool, masterKey)
slackOAuth := integration.NewSlackOAuthHandler(integrationStore, cfg.SlackClientID, cfg.SlackClientSecret, cfg.BaseURL+"/api/v1/oauth/slack/callback", cfg.DashboardURL, pool)

// Public OAuth callback (no JWT — user redirected from Slack):
r.Get("/oauth/slack/callback", slackOAuth.Callback)

// Authenticated:
r.Get("/oauth/slack", slackOAuth.Authorize)
r.Delete("/integrations/slack", integrationStore.DeleteHandler) // disconnect
r.Get("/integrations", integrationStore.ListHandler) // list connected integrations
```

### 8. Dashboard — Org Slack Integration

In org settings or notification settings page:
- Fetch `GET /api/proxy/integrations?org_id=X`
- If Slack connected: show workspace name + "Disconnect" button
- If not connected: show "Connect Slack" button → `window.location = "/api/proxy/oauth/slack?org_id=X"`
- Handle `?slack=connected` query param to show success toast

## Todo

- [ ] Create migration 000031
- [ ] Add SLACK_CLIENT_ID/SECRET config vars
- [ ] Implement integration/store.go
- [ ] Implement integration/slack_oauth.go
- [ ] Refactor SlackAdapter to accept token per-call
- [ ] Update notify/service.go to resolve org token
- [ ] Wire OAuth routes in main.go
- [ ] Dashboard: Connect/Disconnect Slack UI
- [ ] Unit tests: OAuth state HMAC, token encryption roundtrip
- [ ] Update .env.example

## Success Criteria
- Org admin can OAuth-connect Slack workspace
- Approval notifications use org-specific bot token
- Fallback to env var SLACK_BOT_TOKEN when no org integration exists
- Disconnect removes token from DB
- Token encrypted at rest

## Security Considerations
- OAuth state param HMAC-verified (CSRF prevention)
- Access token AES-256-GCM encrypted at rest (same masterKey as secrets)
- Only org owner/admin can initiate OAuth or disconnect
- Slack signing secret remains global (verifies webhooks from Slack)
- Token never returned in API responses (only workspace_id, team_name visible)
