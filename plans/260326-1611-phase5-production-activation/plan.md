---
title: "Phase 5: Production Activation"
description: "Operational tasks to go from deployed code to revenue-generating SaaS: Stripe setup, SMTP, Redis, OAuth verification, git commit, full user journey test, QA, error audit, monitoring, API docs, deployment guide, and beta launch"
status: pending
priority: P0
effort: 18h
branch: master
tags: [production, operations, billing, launch]
created: 2026-03-26
---

# Phase 5: Production Activation

Operational tasks needed to go from "deployed code" to "revenue-generating SaaS". Most work is config/manual setup, not code.

## Phases

| # | Phase | Priority | Effort | Status |
|---|-------|----------|--------|--------|
| 1 | [Stripe Account + Products Setup](./phase-01-stripe-setup.md) | P0 | 2h | pending |
| 2 | [SMTP Configuration (SendGrid)](./phase-02-smtp-setup.md) | P0 | 1h | pending |
| 3 | [Redis for Rate Limiting](./phase-03-redis-setup.md) | P1 | 30min | pending |
| 4 | [Google OAuth E2E Verification](./phase-04-google-oauth-test.md) | P1 | 30min | pending |
| 5 | [Git Commit + Push to GitHub](./phase-05-git-commit.md) | P0 | 30min | pending |
| 6 | [Full User Journey Test](./phase-06-user-journey-test.md) | P1 | 2h | pending |
| 7 | [Mobile Responsive QA](./phase-07-mobile-qa.md) | P1 | 2h | pending |
| 8 | [Error Message Audit](./phase-08-error-audit.md) | P2 | 3h | pending |
| 9 | [UptimeRobot Monitoring](./phase-09-monitoring.md) | P1 | 15min | pending |
| 10 | [API Documentation (OpenAPI)](./phase-10-api-docs.md) | P2 | 4h | pending |
| 11 | [Deployment Guide](./phase-11-deployment-guide.md) | P2 | 2h | pending |
| 12 | [Beta Launch (5-10 users)](./phase-12-beta-launch.md) | P1 | 1h | pending |

**Total: ~18h**

## Dependencies

```
Phase 1-5 (independent, can parallelize)
  ↓
Phase 6 (blocked on 1-4)
Phase 7-11 (can start independently)
  ↓
Phase 12 (needs all above)
```

## Key Env Vars (to set on VPS)

| Variable | Source | Phase |
|----------|--------|-------|
| `STRIPE_SECRET_KEY` | Stripe Dashboard | 1 |
| `STRIPE_WEBHOOK_SECRET` | Stripe Dashboard | 1 |
| `STRIPE_PRO_PRICE_ID` | Stripe Dashboard | 1 |
| `STRIPE_TEAM_PRICE_ID` | Stripe Dashboard | 1 |
| `SMTP_HOST` | SendGrid | 2 |
| `SMTP_PORT` | SendGrid | 2 |
| `SMTP_USER` | SendGrid | 2 |
| `SMTP_PASSWORD` | SendGrid | 2 |
| `SMTP_FROM` | SendGrid | 2 |
| `REDIS_URL` | Docker Compose | 3 |

## Post-Phase 5

- Announce soft launch (email to interested developers)
- Monitor logs for errors
- Gather feedback from beta users
- Plan Phase 6: Scale & Compliance (multi-org support, SOC 2, GDPR)
