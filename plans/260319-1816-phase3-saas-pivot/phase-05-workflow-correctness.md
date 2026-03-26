---
phase: "3.5"
title: "Workflow Correctness"
priority: P1
status: pending
effort: 2h
---

# Phase 3.5: Workflow Correctness

## Context Links
- [Workflow service](../../server/internal/workflow/service.go) — ListPending query (line 169-216)
- [Telegram webhook handler](../../server/internal/notify/telegram_webhook.go) — no auto-registration
- [Config](../../server/internal/config/config.go) — TELEGRAM_BOT_TOKEN
- [Main](../../server/cmd/server/main.go) — startup sequence

## Overview

Two targeted fixes:
1. **ListPending bug**: query only shows requests where calling user owns the secret. Should also include requests where user is an assigned approver in `approval_steps`.
2. **Telegram webhook auto-registration**: call `setWebhook` Telegram Bot API on server startup if TELEGRAM_BOT_TOKEN is set.

## Requirements

### Functional
- ListPending returns requests where user is secret owner OR assigned approver
- Telegram webhook URL auto-registered on server startup

### Non-Functional
- No performance regression on ListPending (indexed join)
- Webhook registration is idempotent (Telegram ignores duplicate setWebhook)

## Related Code Files

### Modify
- `server/internal/workflow/service.go` — fix ListPending query
- `server/internal/notify/telegram.go` — add SetWebhook method
- `server/cmd/server/main.go` — call SetWebhook on startup

## Implementation Steps

### 1. Fix ListPending query

Current query (service.go line 176-195) only joins `secrets s ON s.user_id = $1`. Need to also include requests where user appears in `approval_steps`.

Replace the COUNT query:
```sql
SELECT COUNT(DISTINCT ar.id)
FROM access_requests ar
JOIN secrets s ON s.id = ar.secret_id AND s.deleted_at IS NULL
LEFT JOIN approval_steps ast ON ast.request_id = ar.id AND ast.approver_user_id = $1
WHERE (s.user_id = $1 OR ast.approver_user_id IS NOT NULL)
  AND ar.status = $2
```

Replace the data query:
```sql
SELECT DISTINCT ar.id, ar.secret_id, COALESCE(s.name, '') AS secret_name,
       COALESCE(ar.requester_user_id, '') AS requester_user_id, ar.requester_type, ar.ai_agent_id,
       ar.status, ar.reason, ar.requested_duration_minutes, ar.decided_by, ar.decided_at, ar.expires_at, ar.created_at
FROM access_requests ar
JOIN secrets s ON s.id = ar.secret_id AND s.deleted_at IS NULL
LEFT JOIN approval_steps ast ON ast.request_id = ar.id AND ast.approver_user_id = $1
WHERE (s.user_id = $1 OR ast.approver_user_id IS NOT NULL)
  AND ar.status = $2
ORDER BY ar.created_at DESC LIMIT $3 OFFSET $4
```

Key changes:
- LEFT JOIN approval_steps
- WHERE condition uses OR (owner OR approver)
- DISTINCT to avoid duplicates when user is both owner and approver

### 2. Telegram webhook auto-registration

Add to `TelegramAdapter`:

```go
func (t *TelegramAdapter) SetWebhook(ctx context.Context, webhookURL string) error {
    payload := map[string]interface{}{
        "url":            webhookURL,
        "allowed_updates": []string{"message", "callback_query"},
    }
    return t.post(ctx, "setWebhook", payload)
}
```

### 3. Call SetWebhook on startup in main.go

After `telegramAdapter` initialization:

```go
if telegramAdapter != nil && cfg.BaseURL != "" {
    webhookURL := cfg.BaseURL + "/api/v1/webhooks/telegram"
    if err := telegramAdapter.SetWebhook(context.Background(), webhookURL); err != nil {
        log.Printf("Warning: failed to register Telegram webhook: %v", err)
    } else {
        log.Printf("Telegram webhook registered: %s", webhookURL)
    }
}
```

This is idempotent — Telegram just overwrites the existing webhook URL.

## Todo

- [ ] Fix ListPending COUNT query
- [ ] Fix ListPending data query
- [ ] Add SetWebhook to TelegramAdapter
- [ ] Call SetWebhook in main.go startup
- [ ] Unit test: ListPending returns approver's requests
- [ ] Verify existing ListPending tests still pass

## Success Criteria
- User assigned as approver via approval_steps sees pending requests in /access-requests list
- Telegram webhook auto-registers on server start when TELEGRAM_BOT_TOKEN is set
- No regression on existing approval flow

## Security Considerations
- ListPending still scopes to authenticated user only (no data leak)
- Telegram setWebhook uses the bot token for auth (existing pattern)
- BaseURL must be HTTPS in production for Telegram webhook to work
