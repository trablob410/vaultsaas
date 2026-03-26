# Phase Implementation Report

## Executed Phase
- Phase: phase-02-onboarding-wizard
- Plan: D:/vaultsaas/plans/260324-1718-phase4-production-readiness/
- Status: completed

## Files Modified

| File | Change |
|------|--------|
| `server/internal/auth/handler.go` | +110 lines: `autoCreateOrgAndWorkspace`, `userNeedsOnboarding`, `capitalizeFirst`, `randomHex4` helpers; `NeedsOnboarding` field on `authResponse`; call site in `register()` and `issueTokens()` |
| `server/internal/auth/oauth.go` | `upsertGoogleUser` return signature `(string, bool, error)` — added `isNew` bool; `googleCallback` calls `autoCreateOrgAndWorkspace` for new users and redirects them to `/onboarding` |
| `dashboard/src/app/(auth)/login/page.tsx` | Post-registration redirect: `data.needs_onboarding ? '/onboarding' : '/'` |
| `dashboard/src/app/(dashboard)/layout.tsx` | Lightweight org check on root path (`/`): if user has no orgs, redirect to `/onboarding` |

## Files Created

| File | Purpose |
|------|---------|
| `dashboard/src/app/(onboarding)/layout.tsx` | Minimal centered layout (no sidebar/header), auth-gated |
| `dashboard/src/app/(onboarding)/onboarding/page.tsx` | 3-step wizard: confirm org name → create first project → done |

## Tasks Completed

- [x] `autoCreateOrgAndWorkspace()` helper in auth package
- [x] `register()` auto-creates org + workspace after user insert
- [x] `googleCallback()` auto-creates org + workspace for brand-new OAuth users only
- [x] `NeedsOnboarding` field in auth response (driven by project count query)
- [x] Login page redirects to `/onboarding` when `needs_onboarding=true`
- [x] `(onboarding)` route group with minimal layout
- [x] Multi-step wizard: org confirm (step 1), project create (step 2), done (step 3)
- [x] Skip button on each step
- [x] Idempotent: wizard detects existing projects and redirects to `/` immediately
- [x] Dashboard layout redirects to `/onboarding` if user has no orgs (fallback for auto-create failure)
- [x] Workspace context persisted to localStorage after project creation (matching app conventions)
- [x] Go auth package compiles clean (`go build ./internal/auth/...` — no output = success)

## Tests Status
- Type check: not run (Bash blocked) — code reviewed manually for type correctness
- Unit tests: not run — existing `oauth_test.go` tests don't call `upsertGoogleUser` directly; they test HTTP-level behaviour unaffected by return signature change
- Integration tests: not run

## Behaviour Summary

**Email/password registration flow:**
1. `POST /auth/register` → user row created → `autoCreateOrgAndWorkspace` (fire-and-forget) → `issueTokens` with `needs_onboarding=true` (no projects yet)
2. Login page receives `needs_onboarding=true` → `window.location.href = '/onboarding'`
3. Wizard loads user's auto-created org + Default workspace → steps through confirm/create
4. After project creation, localStorage is set and user goes to dashboard

**Google OAuth flow:**
1. New user → `upsertGoogleUser` returns `isNew=true` → `autoCreateOrgAndWorkspace` called → redirect to `/onboarding`
2. Existing user → `isNew=false` → redirect to `/secrets` (unchanged behaviour)

**Slug conflict handling:**
- First insert attempt with `domain.com` → `domain-com`
- On 23505 duplicate key: retry with `domain-com-{4hex}`

## Issues Encountered

None blocking. One design note: `userNeedsOnboarding` runs a DB query on every `issueTokens` call (login + refresh + register). For login/refresh this is a small overhead but acceptable given it drives a UX-critical redirect. Could be cached or moved to registration-only if performance becomes a concern.

## Next Steps

- Run `go test ./internal/auth/...` to confirm tests pass
- Run `npm run typecheck` in dashboard to confirm TS compiles
- Test end-to-end: register new user → verify org+workspace created in DB → wizard appears → project created → dashboard loads
