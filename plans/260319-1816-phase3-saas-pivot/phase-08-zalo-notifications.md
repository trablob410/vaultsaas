---
phase: "3.8"
title: "Zalo OA Notifications"
priority: P3
status: pending
effort: 4h
---

# Phase 3.8: Zalo OA Notifications

## Context Links
- [Notify service](../../server/internal/notify/service.go) — notification dispatch
- [Telegram adapter](../../server/internal/notify/telegram.go) — similar pattern for Zalo
- [Telegram webhook](../../server/internal/notify/telegram_webhook.go) — linking flow pattern
- [Channel store](../../server/internal/notify/channel_store.go) — per-user channel storage
- [Notification settings](../../dashboard/src/app/(dashboard)/settings/notifications/page.tsx)

## Overview

Add Zalo Official Account (OA) one-way notifications. IMPORTANT: Zalo ZNS is **one-way only** — no interactive approve/reject from Zalo (unlike Slack/Telegram). Users link their Zalo account via phone number, and receive text notifications with a dashboard link when approval requests are created.

## Key Insights
- Zalo OA API uses REST: `POST https://openapi.zalo.me/v3.0/oa/message/cs` (customer service message)
- Requires user to have interacted with OA first (follow or message)
- User linking: user provides phone number → we send OTP via Zalo → user confirms
- Alternative simpler flow: user follows OA, sends their Valt email → we match and link
- ZNS (Zalo Notification Service) requires ZNS templates — more complex, skip for MVP
- Use CS (Customer Service) messages instead — free, no template needed, but requires user to have messaged OA within 7 days
- Simplest MVP: use the "send message by phone" endpoint if available, or require OA follow + /start pattern like Telegram

## Requirements

### Functional
- Per-user Zalo linking (similar to Telegram)
- Config: ZALO_OA_TOKEN + ZALO_OA_ID env vars
- Send notification when approval request created (text only, with dashboard link)
- Dashboard: "Connect Zalo" in notification settings

### Non-Functional
- One-way only: no interactive buttons in Zalo messages
- Graceful fallback: skip if ZALO_OA_TOKEN not configured
- Use notification_jobs queue (Phase 3.4) for reliability

## Architecture

```
User Linking Flow:
1. User clicks "Connect Zalo" in settings
2. POST /me/zalo-link → generate link token
3. User opens Zalo OA, sends "/start {token}" message
4. Zalo OA webhook receives message → match token → link user

Notification Flow:
Approval event → notify service
  → check user has zalo channel in user_notification_channels
  → enqueue notification_job (channel_type='zalo')
  → worker: POST to Zalo OA API with text message
```

## Related Code Files

### Create
- `server/internal/notify/zalo.go` — ZaloAdapter: send messages via Zalo OA API
- `server/internal/notify/zalo_webhook.go` — handle Zalo OA webhook callbacks
- `server/internal/database/migrations/000035_zalo_linking.up.sql`
- `server/internal/database/migrations/000035_zalo_linking.down.sql`

### Modify
- `server/internal/config/config.go` — add ZALO_OA_TOKEN, ZALO_OA_ID
- `server/internal/notify/service.go` — add Zalo dispatch path
- `server/internal/notify/worker.go` — add zalo delivery in processBatch
- `server/internal/notify/channel_handler.go` — accept 'zalo' channel_type
- `server/cmd/server/main.go` — wire Zalo adapter + webhook
- `dashboard/src/app/(dashboard)/settings/notifications/page.tsx` — add Zalo option

## Implementation Steps

### 1. Config env vars

```go
ZaloOAToken string `envconfig:"ZALO_OA_TOKEN" default:""`
ZaloOAID    string `envconfig:"ZALO_OA_ID" default:""`
```

### 2. DB Migration (000035)

```sql
-- 000035_zalo_linking.up.sql
CREATE TABLE zalo_link_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token      VARCHAR(64) NOT NULL UNIQUE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_zalo_link_tokens_token ON zalo_link_tokens(token) WHERE used_at IS NULL;
```

Pattern mirrors `telegram_link_tokens` table exactly.

### 3. notify/zalo.go (~70 lines)

```go
type ZaloAdapter struct {
    oaToken string
    oaID    string
    client  *http.Client
}

func NewZaloAdapter(oaToken, oaID string) *ZaloAdapter
  // Return nil if oaToken is empty

func (z *ZaloAdapter) SendMessage(ctx context.Context, userID, text string) error
  // POST https://openapi.zalo.me/v3.0/oa/message/cs
  // Headers: access_token=oaToken
  // Body: {recipient: {user_id: zaloUserID}, message: {text: text}}
```

Note: Zalo OA API uses `user_id` which is Zalo's internal user ID (obtained when user messages OA). This is stored in `user_notification_channels.handle` for channel_type='zalo'.

### 4. notify/zalo_webhook.go (~60 lines)

```go
type ZaloWebhookHandler struct {
    zalo     *ZaloAdapter
    channels *ChannelStore
    pool     *pgxpool.Pool
}

func (h *ZaloWebhookHandler) Handle(w, r) // POST /webhooks/zalo
```

Handle incoming Zalo OA events:
- Parse webhook payload (Zalo sends JSON with `event_name`, `sender.id`, `message.text`)
- If message starts with "/start ": extract token, validate against `zalo_link_tokens`
- Link user: upsert notification channel (type='zalo', handle=sender.id), mark verified
- Reply with confirmation message

### 5. Generate link URL

```go
func (h *ZaloWebhookHandler) GenerateLinkURL(ctx, userID string) (string, error)
  // Same pattern as Telegram: generate random token, store in zalo_link_tokens
  // Return deep link or instruction text (Zalo doesn't have universal deep links like Telegram)
  // Return: {token: "XXX", instructions: "Open Zalo, follow OA 'Valt', send: /start XXX"}
```

### 6. Update notify/service.go

Add Zalo dispatch path in `NotifyApprovalNeeded`:
```go
if s.zalo != nil && s.channels != nil && ownerUserID != "" {
    if zaloID, err := s.channels.GetPreferred(ctx, ownerUserID, "zalo"); err == nil && zaloID != "" {
        text := fmt.Sprintf("Valt: %s requests access to %s.\nReason: %s\nReview: %s/approvals",
            requester, secretName, reason, dashboardURL)
        // Enqueue via jobStore or direct send
        s.jobStore.Enqueue(ctx, "zalo", zaloID, "Approval Request", map[string]string{"text": text})
    }
}
```

### 7. Update worker.go

Add `zalo` case in `processBatch`:
```go
case "zalo":
    var p struct{ Text string `json:"text"` }
    json.Unmarshal(job.PayloadJSON, &p)
    err = w.zalo.SendMessage(ctx, job.Recipient, p.Text)
```

### 8. Update channel_handler.go

Add 'zalo' to accepted channel types:
```go
if body.ChannelType != "slack" && body.ChannelType != "telegram" && body.ChannelType != "email" && body.ChannelType != "zalo" {
```

### 9. Wire in main.go

```go
zaloAdapter := notify.NewZaloAdapter(cfg.ZaloOAToken, cfg.ZaloOAID)
zaloWebhookHandler := notify.NewZaloWebhookHandler(zaloAdapter, channelStore, pool)

// Public webhook:
r.Post("/webhooks/zalo", zaloWebhookHandler.Handle)

// Authenticated:
r.Post("/me/zalo-link", func(w, r) {
    userID := auth.UserIDFromContext(r.Context())
    result, err := zaloWebhookHandler.GenerateLinkURL(r.Context(), userID)
    // return result
})
```

### 10. Dashboard — Zalo option in notification settings

Add Zalo option alongside existing email/Slack/Telegram:
- SelectItem value="zalo" in channel type dropdown
- When selected: show "Connect via Zalo" button (similar to Telegram pattern)
- On click: POST /me/zalo-link → show instructions modal: "Open Zalo, follow our OA, send: /start TOKEN"
- After linking, show as verified channel

## Todo

- [ ] Add ZALO_OA_TOKEN/ZALO_OA_ID config vars
- [ ] Create migration 000035
- [ ] Implement notify/zalo.go
- [ ] Implement notify/zalo_webhook.go
- [ ] Update notify/service.go for Zalo dispatch
- [ ] Update notify/worker.go for Zalo delivery
- [ ] Update channel_handler.go to accept 'zalo'
- [ ] Wire in main.go
- [ ] Dashboard: add Zalo to notification settings
- [ ] Unit tests: message sending, link token flow

## Success Criteria
- User can link Zalo account via OA message flow
- Approval notifications sent to linked Zalo users
- Message includes dashboard link for review
- One-way only — no approve/reject from Zalo
- Graceful skip when ZALO_OA_TOKEN not configured

## Security Considerations
- Link tokens short-lived (10 min), single-use
- Zalo user ID stored in notification_channels (not sensitive)
- OA token never exposed in API responses
- Webhook endpoint should verify Zalo signature if available (check Zalo OA docs for HMAC)
- Messages contain secret name but no secret value
