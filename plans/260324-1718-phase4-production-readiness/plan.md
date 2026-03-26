---
title: "Phase 4: Production Readiness"
description: "JWT refresh, onboarding, Stripe activation, email verification, password reset, landing page, monitoring, team invitations, CI/CD, legal pages, and launch prep"
status: completed
priority: P1
effort: 37h
branch: master
tags: [production, auth, billing, onboarding, ci-cd, launch]
created: 2026-03-24
---

# Phase 4: Production Readiness

## Timeline: 4 Weeks (~37h total)

## Phase Overview

| # | Phase | Priority | Effort | Status | Depends On |
|---|-------|----------|--------|--------|------------|
| 4.1 | [JWT Auto-Refresh](./phase-01-jwt-auto-refresh.md) | P0 | 3h | Done | - |
| 4.2 | [Onboarding Wizard](./phase-02-onboarding-wizard.md) | P0 | 4h | Done | 4.1 |
| 4.3 | [Stripe Activation](./phase-03-stripe-activation.md) | P0 | 3h | Done | - |
| 4.4 | [Google OAuth Verification](./phase-04-google-oauth-verification.md) | P1 | 1h | Done | - |
| 4.5 | [Email Verification](./phase-05-email-verification.md) | P1 | 4h | Done | SMTP configured |
| 4.6 | [Password Reset](./phase-06-password-reset.md) | P1 | 3h | Done | SMTP configured |
| 4.7 | [Landing Page](./phase-07-landing-page.md) | P1 | 4h | Done | - |
| 4.8 | [Monitoring + Backups](./phase-08-monitoring-backups.md) | P1 | 2h | Done | - |
| 4.9 | [Team Invitations](./phase-09-team-invitations.md) | P2 | 4h | Done | 4.5 |
| 4.10 | [CI/CD Pipeline](./phase-10-cicd-pipeline.md) | P2 | 4h | Done | - |
| 4.11 | [Legal Pages](./phase-11-legal-pages.md) | P2 | 2h | Done | - |
| 4.12 | [Soft Launch Prep](./phase-12-soft-launch-prep.md) | P2 | 3h | Done | All above |

## Dependency Graph

```
Week 1 (P0):
  4.1 JWT Refresh ──> 4.2 Onboarding Wizard
  4.3 Stripe Activation (independent)
  4.4 Google OAuth Verification (independent)

Week 2 (P1):
  SMTP ──> 4.5 Email Verification ──> 4.9 Team Invitations
  SMTP ──> 4.6 Password Reset
  4.7 Landing Page (independent)
  4.8 Monitoring + Backups (independent)

Week 3 (P2):
  4.9 Team Invitations (needs 4.5)
  4.10 CI/CD Pipeline (independent)
  4.11 Legal Pages (independent)

Week 4:
  4.12 Soft Launch Prep (after all above)
```

## DB Migrations

| Migration | Table | Phase |
|-----------|-------|-------|
| 000038 | `email_verification_tokens` + `users.email_verified` | 4.5 |
| 000039 | `password_reset_tokens` | 4.6 |
| 000040 | `org_invitations` | 4.9 |

## Key Architecture Decisions

- **No new env vars needed** -- SMTP, Stripe, Google OAuth already in config
- **Single VPS** -- all solutions Docker Compose compatible, no K8s
- **Refresh via httpOnly cookie** -- refresh_token already stored in DB (migration 000006), cookie already set client-side but not used by proxy
- **Email infra** -- `notify.EmailSender` already exists, reuse for verification/reset/invitations

## Current State (as of 2026-03-24)

- Backend: Go monolith, chi router, 37 migrations applied
- Dashboard: Next.js 15, cookies for auth, no silent refresh
- JWT: 15min access, 7-day refresh (DB), 30-day CLI
- Billing: Stripe code exists (`server/internal/billing/`) but not activated
- SMTP: EmailSender exists in `notify/email.go`, operational if SMTP_HOST set
- Google OAuth: handler exists in `auth/oauth.go`, needs production verification
