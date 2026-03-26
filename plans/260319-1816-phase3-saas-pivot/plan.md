---
title: "Phase 3: SaaS Pivot"
description: "Stripe billing, Slack OAuth, webhooks, notification reliability, workflow fixes, plan limits UI, TOTP 2FA, Zalo notifications, VSCode extension"
status: completed
priority: P1
effort: 48h
branch: master
tags: [saas, billing, integrations, security, notifications, vscode]
created: 2026-03-19
---

# Phase 3: SaaS Pivot

Transforms Valt from self-hosted MVP to production SaaS. Nine sub-phases — can be parallelized in groups.

## Dependency Graph

```
3.1 Stripe Billing ──┐
3.5 Workflow Fix ─────┤
3.4 Notif Reliability─┼──▶ 3.6 Plan Limits UI (needs billing + usage API)
3.2 Slack OAuth ──────┘
3.3 Webhooks API ─────────── (independent)
3.7 TOTP 2FA ─────────────── (independent)
3.8 Zalo Notifications ──── (independent, after 3.4)
3.9 VSCode Extension ─────── (independent)
```

## Phases

| # | Phase | Priority | Effort | Status |
|---|-------|----------|--------|--------|
| 1 | [Stripe Billing](./phase-01-stripe-billing.md) | P1 | 8h | done |
| 2 | [Slack OAuth per-org](./phase-02-slack-oauth.md) | P1 | 6h | done |
| 3 | [Webhooks API](./phase-03-webhooks-api.md) | P2 | 5h | done |
| 4 | [Notification Reliability](./phase-04-notification-reliability.md) | P1 | 5h | done |
| 5 | [Workflow Correctness](./phase-05-workflow-correctness.md) | P1 | 2h | done |
| 6 | [Plan Limits UI](./phase-06-plan-limits-ui.md) | P1 | 4h | done |
| 7 | [TOTP 2FA](./phase-07-totp-2fa.md) | P2 | 6h | done |
| 8 | [Zalo Notifications](./phase-08-zalo-notifications.md) | P3 | 4h | done |
| 9 | [VSCode Extension](./phase-09-vscode-extension.md) | P2 | 8h | deferred |

## DB Migrations Plan

| Migration | Phase | Purpose |
|-----------|-------|---------|
| 000030 | 3.1 | Stripe columns on organizations |
| 000031 | 3.2 | org_integrations table |
| 000032 | 3.3 | org_webhooks table |
| 000033 | 3.4 | notification_jobs table |
| 000034 | 3.7 | TOTP secret + backup_codes on users |
| 000035 | 3.8 | Zalo OA user linking columns |

## New Go Dependencies

| Library | Phase | Purpose |
|---------|-------|---------|
| `github.com/stripe/stripe-go/v84` | 3.1 | Stripe API |
| `github.com/pquerna/otp` | 3.7 | TOTP generation/validation |

## Key Env Vars (new)

| Variable | Phase | Required |
|----------|-------|----------|
| `STRIPE_SECRET_KEY` | 3.1 | For billing |
| `STRIPE_WEBHOOK_SECRET` | 3.1 | Webhook HMAC |
| `STRIPE_PRO_PRICE_ID` | 3.1 | Stripe Price ID |
| `STRIPE_TEAM_PRICE_ID` | 3.1 | Stripe Price ID |
| `SLACK_CLIENT_ID` | 3.2 | OAuth app |
| `SLACK_CLIENT_SECRET` | 3.2 | OAuth app |
| `ZALO_OA_TOKEN` | 3.8 | Zalo OA API |
| `ZALO_OA_ID` | 3.8 | Zalo OA ID |
