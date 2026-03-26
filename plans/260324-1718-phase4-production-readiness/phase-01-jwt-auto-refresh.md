---
phase: "4.1"
title: "Dashboard JWT Auto-Refresh"
priority: P0
effort: 3h
status: pending
---

# Phase 4.1: Dashboard JWT Auto-Refresh

## Context Links
- `server/internal/auth/jwt.go` -- 15min access token, GenerateRefreshToken (opaque hex)
- `server/internal/auth/handler.go:200` -- `POST /auth/refresh` already exists, rotates token
- `dashboard/src/app/api/proxy/[...path]/route.ts` -- BFF proxy, reads `valt_access_token` cookie
- `dashboard/src/lib/api-client.ts:36` -- client `apiFetch`, redirects to /login on 401
- `dashboard/src/lib/auth.ts` -- SSR session decode, checks exp
- `dashboard/src/app/(auth)/login/page.tsx:47-48` -- sets both cookies client-side

## Overview

Dashboard access token expires in 15min. Users get logged out constantly. Need silent refresh: proxy intercepts 401, uses refresh_token to get new access_token, retries original request.

## Key Insights

- `POST /api/v1/auth/refresh` already exists and works (handler.go:200-240)
- Refresh token stored in `refresh_tokens` table (migration 000006), 7-day expiry
- Login page already sets `valt_refresh_token` cookie client-side (max-age=604800)
- Problem: cookie is NOT httpOnly (set via `document.cookie`), and proxy never reads it
- Google OAuth callback sets cookie server-side but only sets access_token

## Requirements

### Functional
- On 401 from backend, proxy auto-refreshes using refresh_token cookie
- On successful refresh, proxy sets new cookies and retries original request
- On refresh failure, return 401 to client (triggers login redirect)
- Client-side `apiFetch` retries once on 401 via dedicated refresh endpoint

### Non-functional
- No visible interruption to user during refresh
- Refresh token rotation on each use (already implemented server-side)
- Cookies must be httpOnly, Secure, SameSite=Lax

## Architecture

```
Browser request
  -> Next.js proxy (api/proxy/[...path])
  -> Backend returns 401
  -> Proxy calls POST /api/v1/auth/refresh with refresh_token
  -> Gets new access_token + refresh_token
  -> Sets new cookies on response
  -> Retries original request with new access_token
  -> Returns result to browser
```

## Related Code Files

### Modify
- `dashboard/src/app/api/proxy/[...path]/route.ts` -- add refresh logic
- `dashboard/src/app/(auth)/login/page.tsx` -- set httpOnly cookies via API route instead of document.cookie
- `dashboard/src/lib/api-client.ts` -- add client-side retry with refresh
- `dashboard/src/lib/auth.ts` -- read refresh_token for SSR redirect logic
- `server/internal/auth/handler.go` -- Google OAuth callback: set refresh_token cookie too

### Create
- `dashboard/src/app/api/auth/set-tokens/route.ts` -- server route to set httpOnly cookies from login response
- `dashboard/src/app/api/auth/refresh/route.ts` -- dedicated refresh endpoint for client-side calls

## Implementation Steps

### Step 1: Server-side cookie helper route (set-tokens)
Create `dashboard/src/app/api/auth/set-tokens/route.ts`:
```typescript
// POST /api/auth/set-tokens
// Body: { access_token, refresh_token, expires_in }
// Sets httpOnly cookies and returns 200
export async function POST(req: NextRequest) {
  const { access_token, refresh_token, expires_in } = await req.json()
  const res = NextResponse.json({ ok: true })
  res.cookies.set('valt_access_token', access_token, {
    httpOnly: true, secure: true, sameSite: 'lax',
    path: '/', maxAge: expires_in,
  })
  res.cookies.set('valt_refresh_token', refresh_token, {
    httpOnly: true, secure: true, sameSite: 'lax',
    path: '/', maxAge: 7 * 24 * 60 * 60,
  })
  return res
}
```

### Step 2: Update login page
Replace `document.cookie` with fetch to `/api/auth/set-tokens`:
```typescript
const res = await fetch(endpoint, { method: 'POST', ... })
const data = await res.json()
if (data.access_token) {
  await fetch('/api/auth/set-tokens', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  window.location.href = '/'
}
```

### Step 3: Add refresh logic to BFF proxy
In `dashboard/src/app/api/proxy/[...path]/route.ts`:
```typescript
async function proxyRequest(req, params) {
  const cookieStore = await cookies()
  const token = cookieStore.get('valt_access_token')?.value
  const refreshToken = cookieStore.get('valt_refresh_token')?.value

  if (!token && !refreshToken) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // Try with current access token
  if (token) {
    const res = await forwardToBackend(req, params, token)
    if (res.status !== 401) return res
  }

  // Attempt refresh
  if (!refreshToken) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const refreshRes = await fetch(`${BACKEND}/api/v1/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })

  if (!refreshRes.ok) {
    // Clear cookies on refresh failure
    const errRes = NextResponse.json({ error: 'Session expired' }, { status: 401 })
    errRes.cookies.delete('valt_access_token')
    errRes.cookies.delete('valt_refresh_token')
    return errRes
  }

  const tokens = await refreshRes.json()
  // Retry original request
  const retryRes = await forwardToBackend(req, params, tokens.access_token)
  // Set new cookies on response
  const response = new NextResponse(await retryRes.text(), {
    status: retryRes.status,
    headers: { 'Content-Type': retryRes.headers.get('Content-Type') ?? 'application/json' },
  })
  response.cookies.set('valt_access_token', tokens.access_token, {
    httpOnly: true, secure: true, sameSite: 'lax', path: '/', maxAge: tokens.expires_in,
  })
  response.cookies.set('valt_refresh_token', tokens.refresh_token, {
    httpOnly: true, secure: true, sameSite: 'lax', path: '/', maxAge: 604800,
  })
  return response
}
```

### Step 4: Client-side refresh endpoint
Create `dashboard/src/app/api/auth/refresh/route.ts` for client-side `apiFetch` retry:
```typescript
// POST /api/auth/refresh -- reads refresh_token cookie, calls backend, sets new cookies
```

### Step 5: Update apiFetch in api-client.ts
Replace immediate login redirect with refresh attempt:
```typescript
if (res.status === 401) {
  const refreshRes = await fetch('/api/auth/refresh', { method: 'POST' })
  if (refreshRes.ok) {
    // Retry original request
    const retryRes = await fetch(`${BASE}${path}`, { ...options, headers: { ... } })
    if (retryRes.ok) return retryRes.json()
  }
  window.location.href = '/login'
  throw new Error('Session expired')
}
```

### Step 6: Update Google OAuth callback
In `server/internal/auth/oauth.go`, set `valt_refresh_token` cookie alongside access token in the redirect response.

## Todo Checklist

- [ ] Create `/api/auth/set-tokens` route for httpOnly cookie setting
- [ ] Update login page to use set-tokens route instead of document.cookie
- [ ] Update TOTP flow to use set-tokens route
- [ ] Add refresh logic to BFF proxy (auto-retry on 401)
- [ ] Create `/api/auth/refresh` client-side endpoint
- [ ] Update `apiFetch` to retry via refresh before redirecting to login
- [ ] Update Google OAuth callback to set both cookies server-side
- [ ] Test: login -> wait 15min -> action succeeds via silent refresh
- [ ] Test: expired refresh token -> redirected to login
- [ ] Test: Google OAuth login -> both cookies set

## Success Criteria

- User stays logged in for 7 days without manual re-login
- All cookies are httpOnly (not accessible via document.cookie)
- Token rotation works (each refresh invalidates old refresh_token)
- Google OAuth flow sets both cookies

## Security Considerations

- httpOnly cookies prevent XSS token theft
- Secure flag ensures HTTPS-only (production uses Caddy TLS)
- SameSite=Lax prevents CSRF on state-changing requests
- Refresh token rotation prevents token reuse attacks
- Failed refresh clears all auth cookies
