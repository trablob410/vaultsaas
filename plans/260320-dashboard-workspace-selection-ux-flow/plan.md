---
title: "Dashboard org/workspace selection UX fix"
description: "Add clear org→workspace selection flow so users can set current workspace and use project-scoped pages"
status: pending
priority: P1
effort: 7h
branch: feat/custom-policy
tags: [dashboard, ux, workspace, organization, navigation]
created: 2026-03-20
---

# Dashboard org/workspace selection UX fix plan

## Executive summary

Current dashboard asks user to “navigate to organization/workspace”, but `/orgs` has no interaction to pick org/workspace. Result: dead-end, high confusion, project pages blocked. Fix should be small-scope, no backend schema change, use existing APIs (`/orgs`, `/orgs/:id/workspaces`, `/workspaces/:id/projects`), and persist selection in localStorage.

Recommended approach: add explicit org→workspace selection flow, plus lightweight current-context indicator/switch in sidebar. Keep simple, avoid global state library.

## Research findings (from current codebase)

1. `projects/page.tsx` requires `localStorage['valt_current_workspace']`; without it shows blocking empty state.
2. `orgs/page.tsx` only lists org cards; no click/select action; no workspace list.
3. API client already supports required endpoints:
   - `api.orgs.list()`
   - `api.workspaces.list(orgId)`
   - `api.projects.list(workspaceId)`
4. Sidebar has “Organizations” button but only hard-nav to `/orgs`; no current org/workspace context shown.
5. Existing context pattern already used for project:
   - `valt_current_project`
   - custom event `valt:project-changed`

Conclusion: mostly UX + state wiring gap, not API gap.

## Goals

- Let user select active organization and workspace in <=2 clicks from org list.
- Remove dead-end state in Projects flow.
- Keep interaction predictable across refresh/session.
- Minimize code churn; no new backend endpoints.

## Non-goals (YAGNI)

- No org/workspace role management redesign.
- No cross-tab synced global state framework.
- No major IA rewrite or multi-tenant settings overhaul.
- No server-persisted “last workspace” preference (future).

## UX principles

- KISS: one obvious path to set context.
- Visibility: always show current org/workspace in sidebar.
- Feedback: immediate success state after selection.
- Safe defaults: clear dependent project selection when workspace changes.
- A11y baseline: keyboard reachable buttons, visible focus, aria-label on icon-only controls.

## Proposed UX flow

## Flow A — Organizations page (primary entry)

1. User opens `/orgs`.
2. Each org card has clear action: **“View workspaces”** (row click or button).
3. User lands on org detail screen with workspace list.
4. Workspace row has **“Set current”** CTA.
5. On set:
   - store `valt_current_org`
   - store `valt_current_workspace`
   - clear `valt_current_project`
   - dispatch events (`valt:workspace-changed`, `valt:project-changed`)
   - show “Current” badge on selected workspace
   - optional secondary CTA: “Go to Projects”

## Flow B — Sidebar quick switching (secondary)

1. Sidebar context section shows current workspace name (fallback: “No workspace selected”).
2. Click opens `/orgs` (MVP) or compact popover (optional stretch).
3. If no workspace selected, show inline subtle warning style.

MVP recommendation: keep sidebar simple link + current label; do not build complex dropdown yet.

## Information architecture

- Keep `/orgs` as index page.
- Add nested route: `/orgs/[orgId]` for workspace listing and selection.
- Keep `/projects` behavior mostly same, but empty-state copy can now reference concrete action: “Choose workspace in Organizations”.

## State model

LocalStorage keys:
- `valt_current_org` (new)
- `valt_current_workspace` (existing, now set via UI)
- `valt_current_project` (existing; reset on workspace switch)

Events:
- `valt:workspace-changed` (new)
- `valt:project-changed` (existing)

Rules:
1. Workspace switch always invalidates current project.
2. If workspace no longer accessible or fetch returns 403/404, clear workspace + project, show recovery UI.
3. If org list empty, guide to create org (existing behavior).

## Implementation plan (no code yet)

### Phase 1 — UX scaffolding and route structure (2h)

Deliverables:
- Add org card interaction on `/orgs`.
- Create `/orgs/[orgId]/page.tsx` showing workspaces list + loading/error/empty states.

Acceptance criteria:
- User can navigate org list → org workspace list.
- Keyboard and pointer interaction both work.

### Phase 2 — Workspace selection behavior (2h)

Deliverables:
- Add “Set current” action per workspace.
- Persist localStorage keys and dispatch events.
- Clear `valt_current_project` on workspace change.

Acceptance criteria:
- After refresh, selected workspace remains current.
- `/projects` loads projects for selected workspace.

### Phase 3 — Sidebar context visibility (1.5h)

Deliverables:
- Show current workspace label in sidebar context block.
- Update label on `storage` and `valt:workspace-changed` events.

Acceptance criteria:
- Sidebar immediately reflects selection changes.
- When no workspace selected, clear fallback text visible.

### Phase 4 — UX polish + copy + guardrails (1h)

Deliverables:
- Improve empty-state copy in `/projects` and `/orgs/[orgId]`.
- Add error toasts/messages for failed workspace selection fetches.
- Ensure touch target >=44px for action controls.

Acceptance criteria:
- No dead-end message without actionable CTA.
- Error states have next-step guidance.

### Phase 5 — Test and validation (0.5h)

Test checklist:
- Manual flow: org list → workspace set → projects load.
- Workspace switch clears current project.
- Reload/tab reopen preserves workspace.
- 401/403 from workspace APIs handled cleanly.
- Basic a11y smoke: tab order, focus visibility, semantic button/link usage.

## Files likely impacted

- `dashboard/src/app/(dashboard)/orgs/page.tsx`
- `dashboard/src/app/(dashboard)/orgs/[orgId]/page.tsx` (new)
- `dashboard/src/components/layout/sidebar.tsx`
- `dashboard/src/app/(dashboard)/projects/page.tsx` (copy/guardrails)
- optional: small helper in `dashboard/src/lib/` for context read/write (if needed to keep DRY)

## Risks and mitigations

1. **State drift across tabs**
   - Mitigation: listen to `storage` event and custom event.
2. **Stale workspace id (deleted/inaccessible)**
   - Mitigation: clear invalid keys on API failure + route user to `/orgs`.
3. **Overengineering sidebar switcher**
   - Mitigation: keep MVP as label + navigate button; postpone dropdown.

## Rollout strategy

1. Ship route + selection flow first.
2. Verify internal usage with seeded data.
3. Then ship sidebar context indicator.
4. Monitor user feedback for quick-switch demand before building popover selector.

## Done definition

- User can always discover where to set workspace.
- User can set workspace from UI without devtools.
- “No workspace selected” state has direct, working path to resolution.
- Selection persists and drives project-scoped pages.

## Unresolved questions

1. Should default workspace auto-select when org has exactly 1 workspace? (recommended yes, but needs product decision)
2. Should we store and display current org globally now, or infer from workspace only?
3. Should sidebar support inline workspace switching in MVP, or only link to `/orgs`?
4. Repo instruction referenced `docs/development-rules.md`; file not found. Need canonical rules location confirmation.
