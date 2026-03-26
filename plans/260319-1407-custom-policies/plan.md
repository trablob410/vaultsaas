# Custom Policies Implementation Plan

> **For agentic workers:** REQUIRED: Use `/ck:plan` in execute mode (subagent-driven or sequential) to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let teams define approval rules per project and per secret — custom approvers, duration caps, business-hours windows, auto-approve overrides — without touching code.

**Architecture:** Add `policy_config JSONB` to `projects` and `secrets` tables. `policy.Resolver` merges secret-level → project-level → global tier defaults at request time. REST endpoints + dashboard UI for editing policy per secret and per project.

**Tech Stack:** Go (pgx, encoding/json), Next.js 15 (shadcn/ui form, drawer), existing `policy.MergePolicy`

**Context:** Brainstorm → `plans/reports/brainstorm-260319-1407-phase2-roadmap.md`

---

## Status: COMPLETED

| Phase | Description | Status |
|-------|-------------|--------|
| 01 | DB migrations + policy resolver | COMPLETED — 2026-03-19 |
| 02 | Policy API endpoints | COMPLETED — 2026-03-19 |
| 03 | Dashboard UI | COMPLETED — 2026-03-19 |

## Key Files

**Existing (modify):**
- `server/internal/policy/engine.go` — extend `MergePolicy` to handle full custom config
- `server/internal/workflow/service.go` — call `Resolver` instead of `ForCredentialType` directly
- `server/internal/vault/handler.go` — add GET/PUT `/secrets/{id}/policy`
- `server/internal/project/handler.go` — add GET/PUT `/projects/{id}/policy`
- `server/cmd/server/main.go` — register policy routes
- `dashboard/src/app/(dashboard)/secrets/[id]/page.tsx` — add policy drawer

**New:**
- `server/internal/database/migrations/000029_policy_config_columns.up.sql`
- `server/internal/policy/resolver.go` — DB-aware policy resolution
- `server/internal/policy/custom_policy.go` — CustomPolicy struct + validation
- `dashboard/src/lib/api/policy.ts`
- `dashboard/src/app/(dashboard)/projects/[id]/policy/page.tsx`
- `dashboard/src/components/policy-editor.tsx`

---

## Phases

- [Phase 01](phase-01-db-and-resolver.md) — DB migrations + policy resolver
- [Phase 02](phase-02-policy-api.md) — Policy API endpoints
- [Phase 03](phase-03-policy-ui.md) — Dashboard UI
