# Approval Channels Implementation Plan

> **For agentic workers:** REQUIRED: Use `/ck:plan` in execute mode (subagent-driven or sequential) to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let approvers act on access requests directly from Slack, Telegram, or email — without opening the dashboard.

**Architecture:** Extend `notify.Service` with channel adapters (email-links, Slack, Telegram). Add `request_action_tokens` table for secure one-click approve/reject links. Add `user_notification_channels` table for user channel preferences. Public action endpoint handles token redemption.

**Tech Stack:** Go (chi router, pgx, crypto/hmac), Slack Block Kit API, Telegram Bot API, existing SMTP infra

**Context:** Brainstorm → `plans/reports/brainstorm-260319-1407-phase2-roadmap.md`

---

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| 01 | Email action links | COMPLETED — 2026-03-19 |
| 02 | Notification channel settings (DB + API + UI) | COMPLETED — 2026-03-19 |
| 03 | Slack bot | COMPLETED — 2026-03-19 |
| 04 | Telegram bot | COMPLETED — 2026-03-19 |

## Key Files

**Existing (modify):**
- `server/internal/notify/service.go` — add token generation, channel routing
- `server/internal/notify/email.go` — embed approve/reject links
- `server/internal/workflow/handler.go` — add `ActionViaToken` public handler; fix empty `to` in notify call
- `server/cmd/server/main.go` — wire new routes + env vars

**New:**
- `server/internal/database/migrations/000026_request_action_tokens.up.sql`
- `server/internal/database/migrations/000027_user_notification_channels.up.sql`
- `server/internal/notify/action_token.go` — token create/verify/consume
- `server/internal/notify/channel_store.go` — CRUD for user_notification_channels
- `server/internal/notify/channel_handler.go` — REST handler
- `server/internal/notify/slack.go` — Slack Block Kit adapter
- `server/internal/notify/slack_webhook.go` — Slack interactivity callback
- `server/internal/notify/telegram.go` — Telegram bot adapter
- `server/internal/notify/telegram_webhook.go` — Telegram update handler
- `dashboard/src/app/(dashboard)/settings/notifications/page.tsx`
- `dashboard/src/lib/api/notifications.ts`

## Phases

- [Phase 01](phase-01-email-action-links.md) — Email action links
- [Phase 02](phase-02-notification-channels.md) — Channel settings (DB + API + UI)
- [Phase 03](phase-03-slack-bot.md) — Slack bot
- [Phase 04](phase-04-telegram-bot.md) — Telegram bot
