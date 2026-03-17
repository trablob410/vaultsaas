# Phase 5: Next.js Dashboard

## Context Links
- Backend routes: `server/cmd/server/main.go` (lines 115-137)
- Auth handler: `server/internal/auth/handler.go`
- Vault service types: `server/internal/vault/service.go` (Secret, CreateSecretInput, etc.)
- Workflow handler: `server/internal/workflow/handler.go`
- Existing scaffold: `dashboard/src/app/layout.tsx`, `page.tsx`, `globals.css`
- Current deps: Next.js 15.1.0, React 19, Tailwind v4, standalone output

## Overview
- **Priority:** P1
- **Status:** pending
- **Effort:** 12h
- Build full dashboard: auth flow, secrets CRUD, approval workflow, audit log viewer, settings
- BFF pattern for auth (httpOnly cookies), client-side AES-256-GCM encryption

## Key Insights
- Tailwind v4 uses `@import "tailwindcss"` (already in globals.css) -- no tailwind.config.js needed
- shadcn/ui supports Tailwind v4 via `npx shadcn@latest init` -- uses CSS variables for theming
- Next.js 15 uses React 19 -- Server Components default, `"use client"` only for interactive components
- `tsconfig.json` already has `@/*` path alias pointing to `./src/*`
- **Google OAuth via Go backend** -- adds 2 endpoints, reuses existing `issueTokens()`, sets httpOnly cookies on callback redirect
- **No client-side crypto for MVP** -- secrets managed via backend API, encrypted at rest in MinIO
- Use `ui-ux-pro-max` skill for polished, production-quality dashboard design

## Requirements

### Functional
- Google OAuth login ("Sign in with Google" button)
- Secrets list with search/filter, create/edit/delete modals
- Approval queue: list pending, approve/reject with reason
- Audit log table with pagination and date filtering
- Settings page: profile info, avatar from Google

### Non-Functional
- Server Components by default; `"use client"` only for forms, modals, interactive components
- All pages behind auth (redirect to /login if no session)
- Responsive layout (mobile-friendly)
- Dark mode support via CSS variables

## Architecture

### Route Structure
```
src/app/
  (auth)/
    login/page.tsx            # "Sign in with Google" button
    layout.tsx                # Centered card layout
  (dashboard)/
    layout.tsx                # Sidebar + header + auth guard
    secrets/page.tsx          # Secrets list
    secrets/[id]/page.tsx     # Secret detail
    approvals/page.tsx        # Approval queue
    audit/page.tsx            # Audit log viewer
    settings/page.tsx         # User settings
  api/
    auth/logout/route.ts      # BFF: clear cookies
    proxy/[...path]/route.ts  # BFF: forward all /api/v1/* with JWT from cookie
```

**Note:** Google OAuth flow handled entirely by Go backend:
- Dashboard redirects to `BACKEND_URL/api/v1/auth/google`
- Go handles Google consent → callback → sets httpOnly cookies → redirects to dashboard

### Component Structure
```
src/
  components/
    ui/                       # shadcn/ui primitives (button, input, card, dialog, table, etc.)
    layout/
      sidebar.tsx             # Nav sidebar
      header.tsx              # Top bar with user menu
    secrets/
      secret-list.tsx         # Table of secrets
      secret-form.tsx         # Create/edit form with crypto
      secret-card.tsx         # Summary card
    approvals/
      approval-list.tsx       # Pending requests table
      approval-actions.tsx    # Approve/reject buttons + dialog
    audit/
      audit-table.tsx         # Paginated audit log
  lib/
    api-client.ts             # Fetch wrapper (calls BFF proxy routes)
    auth.ts                   # Session helpers (check cookie, redirect)
    constants.ts              # API base URL, credential types enum
  types/
    api.ts                    # TypeScript interfaces matching backend JSON
```

### Auth Flow (Google OAuth via Go Backend)
```
Browser                     Go Backend                    Google
  |-- Click "Sign in" -------->|                            |
  |                            |-- Redirect to Google ----->|
  |                            |                            |-- User consents
  |                            |<-- Callback with code -----|
  |                            |  Exchange code, upsert user, issueTokens()
  |                            |  Set httpOnly cookies (access+refresh token)
  |<-- Redirect to /secrets ---|                            |
  |                            |                            |
  |-- GET /api/proxy/secrets ->|  (Next.js BFF)             |
  |                            |  Read JWT from cookie      |
  |                            |-- GET /api/v1/secrets (Bearer) -->|
  |                            |<-- {secrets: [...]} -------|
  |<-- {secrets: [...]} -------|                            |
```

## Related Code Files

### Files to Create (Dashboard)
- `src/app/(auth)/layout.tsx`
- `src/app/(auth)/login/page.tsx` (Google sign-in button)
- `src/app/(dashboard)/layout.tsx`
- `src/app/(dashboard)/secrets/page.tsx`
- `src/app/(dashboard)/secrets/[id]/page.tsx`
- `src/app/(dashboard)/approvals/page.tsx`
- `src/app/(dashboard)/audit/page.tsx`
- `src/app/(dashboard)/settings/page.tsx`
- `src/app/api/auth/logout/route.ts`
- `src/app/api/proxy/[...path]/route.ts`
- `src/components/layout/sidebar.tsx`
- `src/components/layout/header.tsx`
- `src/components/secrets/secret-list.tsx`
- `src/components/secrets/secret-form.tsx`
- `src/components/approvals/approval-list.tsx`
- `src/components/approvals/approval-actions.tsx`
- `src/components/audit/audit-table.tsx`
- `src/lib/api-client.ts`
- `src/lib/auth.ts`
- `src/lib/constants.ts`
- `src/types/api.ts`

### Files to Create (Go Backend - Google OAuth)
- `server/internal/database/migrations/000012_alter_users_oauth.up.sql`
- `server/internal/database/migrations/000012_alter_users_oauth.down.sql`

### Files to Modify (Go Backend)
- `server/internal/auth/handler.go` -- add GoogleLogin + GoogleCallback handlers
- `server/internal/config/config.go` -- add Google OAuth env vars
- `server/go.mod` -- add golang.org/x/oauth2

### Files to Modify (Dashboard)
- `dashboard/package.json` -- add shadcn/ui, lucide-react, class-variance-authority, clsx, tailwind-merge
- `dashboard/src/app/globals.css` -- add CSS variables for shadcn theming
- `dashboard/src/app/layout.tsx` -- add font, theme provider
- `dashboard/src/app/page.tsx` -- update landing links
- `dashboard/next.config.ts` -- add env var for backend URL

### New Dependencies
```
# Dashboard
shadcn/ui (via npx shadcn@latest init)
lucide-react              # Icons
class-variance-authority  # Variant styling
clsx + tailwind-merge     # Class merging (cn() helper)

# Go Backend
golang.org/x/oauth2       # Google OAuth2 flow
```

## Implementation Steps

### Step 1: Project Setup (1h)
1. `cd dashboard && npx shadcn@latest init` -- select New York style, CSS variables
2. Install: `lucide-react`
3. Add shadcn components: `npx shadcn@latest add button input card dialog table badge dropdown-menu sheet toast tabs`
4. Create `src/lib/utils.ts` with `cn()` helper (shadcn generates this)
5. Update `globals.css` with CSS variables for light/dark theming
6. Update `next.config.ts`: add `env` block with `BACKEND_URL` (default `http://localhost:8080`)
7. Run `npm run build` to verify setup

### Step 2: Google OAuth Backend (2h)
**Go backend changes (server/):**
1. `go get golang.org/x/oauth2`
2. Create migration `000012_alter_users_oauth.up.sql`:
   - `ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;`
   - `ADD COLUMN google_id VARCHAR(255) UNIQUE;`
   - `ADD COLUMN avatar_url TEXT;`
   - `ADD COLUMN display_name VARCHAR(255);`
   - `ADD COLUMN auth_provider VARCHAR(20) NOT NULL DEFAULT 'local';`
3. Add to `auth/handler.go`:
   - `GET /auth/google` -- build OAuth2 URL with state, redirect to Google
   - `GET /auth/google/callback` -- exchange code, get user info, upsert user by email/google_id, call `issueTokens()`, set httpOnly cookies, redirect to dashboard URL
4. Config: add `GoogleClientID`, `GoogleClientSecret`, `GoogleRedirectURL` to config.go
5. Update login to also check `auth_provider` (local users can still email/password login)

### Step 3: Types & API Client (1h)
1. Create `src/types/api.ts` -- interfaces: `Secret`, `AccessRequest`, `Credential`, `AuditLog`, `User`, `PaginatedResponse<T>`
2. Create `src/lib/constants.ts` -- `CREDENTIAL_TYPES`, `REQUEST_STATUSES`, `RISK_TIERS`
3. Create `src/lib/api-client.ts`:
   - `apiFetch(path, options)` -- wraps fetch, adds `/api/proxy` prefix, handles errors
   - Export typed functions: `listSecrets()`, `createSecret()`, `getSecret()`, etc.
   - Handle 401 by redirecting to /login

### Step 4: BFF Proxy + Auth (1h)
1. Create `src/lib/auth.ts`:
   - `getSession()` -- reads cookie, decodes JWT payload, returns user info or null
   - Cookie names: `valt_access_token`, `valt_refresh_token`
2. `app/api/auth/logout/route.ts` -- clear cookies, redirect to /login
3. `app/api/proxy/[...path]/route.ts`:
   - Read access token from cookie
   - If expired, auto-refresh using refresh token cookie against backend `/auth/refresh`
   - Forward request to `BACKEND_URL/api/v1/{path}` with Bearer token
   - Return backend response

### Step 5: Auth Page (0.5h)
1. `(auth)/layout.tsx` -- centered card layout, no sidebar
2. `(auth)/login/page.tsx` -- "Sign in with Google" button, redirects to `BACKEND_URL/api/v1/auth/google`
3. Branding: Valt logo, tagline, clean design via `ui-ux-pro-max` skill

### Step 6: Dashboard Layout (1.5h)
1. `(dashboard)/layout.tsx` -- auth guard (check session, redirect if none), sidebar + header
2. `components/layout/sidebar.tsx` -- nav links: Secrets, Approvals, Audit, Settings. Active state. User avatar from Google
3. `components/layout/header.tsx` -- breadcrumb, user dropdown (logout, profile)
4. Mobile: sheet-based sidebar (hamburger menu)
5. Apply `ui-ux-pro-max` skill for professional layout design

### Step 7: Secrets Pages (2h)
1. `secrets/page.tsx` -- server component, fetch secrets list
2. `components/secrets/secret-list.tsx` -- client component, table with name/type/created/actions
3. `components/secrets/secret-form.tsx` -- client component, create/edit dialog
   - Form fields: name, description, credential_type (select), source, value (sent to backend for encryption)
   - On submit: POST to backend
4. `secrets/[id]/page.tsx` -- detail view with metadata, version info
5. Delete confirmation dialog

### Step 8: Approvals Page (1h)
1. `approvals/page.tsx` -- list pending access requests
2. `components/approvals/approval-list.tsx` -- table: secret name, requester, reason, status, date
3. `components/approvals/approval-actions.tsx` -- approve/reject buttons with confirmation dialog
4. Status filter tabs: All, Pending, Approved, Rejected

### Step 9: Audit & Settings (1h)
1. `audit/page.tsx` -- paginated audit log table
2. `components/audit/audit-table.tsx` -- columns: timestamp, action, resource, actor, IP
3. Pagination controls (prev/next, page size)
4. `settings/page.tsx` -- display user email, region. Placeholder for password change

## Todo List
- [ ] shadcn/ui init + component installation
- [ ] TypeScript types matching backend JSON
- [ ] API client with BFF proxy
- [ ] BFF auth routes (login/register/refresh/logout/proxy)
- [ ] Auth pages (login, register)
- [ ] Dashboard layout (sidebar, header, auth guard)
- [ ] Secrets CRUD pages
- [ ] Client-side AES-256-GCM crypto module
- [ ] Approvals page with approve/reject
- [ ] Audit log viewer
- [ ] Settings page
- [ ] `next lint` passes clean
- [ ] `npm run build` succeeds

## Success Criteria
- Login -> see secrets list -> create secret (encrypted) -> view in list
- Approval flow: create access request -> approve -> view credential (decrypted)
- Audit log shows all actions
- All pages responsive, dark mode toggle works
- `next lint` and `npm run build` pass
- No `any` types, strict mode clean

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| shadcn/ui + Tailwind v4 compatibility | Medium | Test init step first; fall back to manual CSS vars if needed |
| Web Crypto API browser support | Low | All modern browsers support SubtleCrypto; no IE support needed |
| Cookie-based auth CSRF | Medium | Use SameSite=Lax, verify Origin header in BFF proxy |
| Master key UX (re-prompt) | Medium | Cache in sessionStorage; clear on logout. Warn user |

## Security Considerations
- httpOnly + Secure + SameSite=Lax cookies for JWT tokens
- BFF proxy validates and strips cookies before forwarding
- CSRF protection via SameSite + Origin header check
- No plaintext secrets in localStorage/sessionStorage
- Master key derived client-side, never sent to server
- Auto-clear decrypted credentials from DOM after 30s timeout
- CSP headers in next.config.ts

## Next Steps
- Phase 7 will add vitest tests for crypto.ts, api-client.ts, and component rendering
- Future: WebSocket for real-time approval notifications
- Future: TOTP/WebAuthn 2FA on settings page
