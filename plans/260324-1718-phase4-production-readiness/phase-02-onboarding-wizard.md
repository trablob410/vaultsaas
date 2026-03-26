---
phase: "4.2"
title: "Onboarding Wizard"
priority: P0
effort: 4h
status: pending
depends_on: ["4.1"]
---

# Phase 4.2: Onboarding Wizard

## Context Links
- `server/internal/auth/handler.go:102-144` -- register handler, creates user only
- `server/internal/org/` -- org CRUD
- `server/internal/workspace/` -- workspace CRUD
- `server/internal/project/` -- project CRUD
- `dashboard/src/app/(dashboard)/layout.tsx` -- dashboard layout, redirect to /login if no session
- `dashboard/src/lib/api-client.ts` -- all API calls

## Overview

New users register -> land on empty dashboard -> confused. Need: auto-create default org on registration + guided onboarding wizard for first-time users.

## Key Insights

- Registration currently only creates a user row. No org/workspace/project.
- Dashboard shows empty state for orgs list.
- Org creation requires name + slug. Can derive from email domain.
- Workspace is nested under org, project under workspace.

## Requirements

### Functional
- After registration, backend auto-creates: default org (from email domain) + default workspace ("Default")
- Dashboard detects "no org" state and redirects to `/onboarding`
- Onboarding wizard steps: 1) Confirm org name, 2) Create first project, 3) Add first secret (optional, skip-able), 4) Done
- Skip button on every step (user can do it later)

### Non-functional
- Wizard must work for both email/password and Google OAuth registrations
- Idempotent: visiting /onboarding with existing org skips to dashboard

## Architecture

```
Registration flow:
  1. POST /auth/register or Google OAuth callback
  2. Backend: create user -> auto-create org + workspace (transactional)
  3. Frontend: redirect to /onboarding
  4. Wizard: confirm org -> create project -> (optional) add secret -> redirect to /

Detection:
  Dashboard layout -> fetch /orgs -> if empty -> redirect /onboarding
  (After auto-create, orgs won't be empty, but no projects yet)
  Better: check if user has any projects. If 0 projects -> onboarding.
```

## Related Code Files

### Modify
- `server/internal/auth/handler.go` -- `register()` and `googleCallback()`: auto-create org+workspace after user creation
- `dashboard/src/app/(dashboard)/layout.tsx` -- add onboarding redirect check

### Create
- `dashboard/src/app/(onboarding)/layout.tsx` -- minimal layout for wizard
- `dashboard/src/app/(onboarding)/onboarding/page.tsx` -- multi-step wizard component

## Implementation Steps

### Step 1: Backend auto-create org + workspace on registration

In `auth/handler.go` register(), after user INSERT:
```go
// Auto-create default org
domain := extractDomain(req.Email) // e.g., "acme" from "user@acme.com"
orgName := strings.Title(domain) + "'s Org"
orgSlug := strings.ToLower(domain) + "-org"

var orgID string
err = h.pool.QueryRow(r.Context(),
    `INSERT INTO organizations (name, slug, owner_id) VALUES ($1, $2, $3) RETURNING id`,
    orgName, orgSlug, userID,
).Scan(&orgID)
// Handle duplicate slug by appending random suffix

// Add owner membership
h.pool.Exec(r.Context(),
    `INSERT INTO org_memberships (org_id, user_id, role) VALUES ($1, $2, 'owner')`,
    orgID, userID)

// Create default workspace
h.pool.Exec(r.Context(),
    `INSERT INTO workspaces (name, slug, org_id) VALUES ('Default', 'default', $1)`,
    orgID)
```

### Step 2: Same for Google OAuth callback
Apply identical org+workspace creation in `oauth.go` `googleCallback()` for new users only (not existing).

### Step 3: Onboarding detection in dashboard layout
```typescript
// dashboard/src/app/(dashboard)/layout.tsx
const session = await getSession()
if (!session) redirect('/login')

// Check if user needs onboarding
const projectsRes = await fetch(`${BACKEND}/api/v1/projects/mine`, {
  headers: { Authorization: `Bearer ${token}` },
})
// If 0 projects, redirect to onboarding
```

Alternative simpler approach: add `onboarded` flag to users table, check in auth response.

### Step 4: Create onboarding wizard page

Multi-step client component:
- Step 1: "Welcome! Confirm your organization name" (pre-filled from auto-created org, editable)
- Step 2: "Create your first project" (name + slug input)
- Step 3: "Add your first secret" (optional, with skip)
- Step 4: "You're all set!" with link to dashboard

Each step calls existing API endpoints (org update, project create, secret create).

### Step 5: Add `needs_onboarding` to auth response

Add a boolean to the auth response (login/register/refresh):
```go
type authResponse struct {
    AccessToken     string `json:"access_token"`
    RefreshToken    string `json:"refresh_token"`
    ExpiresIn       int    `json:"expires_in"`
    NeedsOnboarding bool   `json:"needs_onboarding,omitempty"`
}
```

Check by querying if user has any projects. If not, set true.

## Todo Checklist

- [ ] Add `extractDomain()` helper to auth package
- [ ] Update `register()` to auto-create org + workspace
- [ ] Update `googleCallback()` to auto-create org + workspace for new users
- [ ] Create `(onboarding)` route group with minimal layout
- [ ] Build multi-step onboarding wizard component
- [ ] Add onboarding detection in dashboard layout (redirect if no projects)
- [ ] Add "needs_onboarding" to auth response (optional optimization)
- [ ] Test: new email registration -> auto org+workspace -> onboarding wizard
- [ ] Test: new Google OAuth -> auto org+workspace -> onboarding wizard
- [ ] Test: existing user -> skips onboarding

## Success Criteria

- New user sees guided onboarding, not empty dashboard
- Default org + workspace auto-created on registration
- User can complete wizard or skip at any step
- Existing users unaffected

## Security Considerations

- Org slug uniqueness: handle collisions with random suffix
- Onboarding page still requires authentication
- Rate limit org creation (already covered by registration rate limit)
