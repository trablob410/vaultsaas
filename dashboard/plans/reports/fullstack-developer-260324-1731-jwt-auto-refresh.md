# Phase Implementation Report

### Executed Phase
- Phase: phase-01-jwt-auto-refresh
- Plan: D:/vaultsaas/plans/260324-1718-phase4-production-readiness/
- Status: completed

### Files Modified

| File | Action | Notes |
|------|--------|-------|
| `dashboard/src/app/api/auth/set-tokens/route.ts` | created | POST endpoint to set httpOnly auth cookies from login response |
| `dashboard/src/app/api/auth/refresh/route.ts` | created | POST endpoint to silently refresh via backend, sets new httpOnly cookies |
| `dashboard/src/app/api/proxy/[...path]/route.ts` | updated | Auto-refresh on 401: reads body once, retries with new token, sets rotated cookies |
| `dashboard/src/app/(auth)/login/page.tsx` | updated | Replaced `document.cookie` with `POST /api/auth/set-tokens` in both login and TOTP flows |
| `dashboard/src/lib/api-client.ts` | updated | `apiFetch` now calls `POST /api/auth/refresh` on 401 and retries once before redirecting to /login |

### Tasks Completed
- [x] Create `/api/auth/set-tokens` route for httpOnly cookie setting
- [x] Update login page to use set-tokens route instead of document.cookie
- [x] Update TOTP flow to use set-tokens route
- [x] Add refresh logic to BFF proxy (auto-retry on 401)
- [x] Create `/api/auth/refresh` client-side endpoint
- [x] Update `apiFetch` to retry via refresh before redirecting to login
- [ ] Update Google OAuth callback to set both cookies server-side (server/internal/auth/handler.go — out of scope for this phase, not in file ownership)

### Key Design Decisions

**Body caching in proxy**: `req.text()` is called once at the top of `proxyRequest` and the string passed explicitly to both the initial and retry `forwardToBackend` calls. This prevents the empty-body bug that would occur if `req.text()` were called twice on the same consumed ReadableStream.

**Two-layer refresh**:
1. BFF proxy intercepts backend 401s transparently — most cases handled here with no client awareness
2. `apiFetch` in api-client.ts calls `/api/auth/refresh` as a second layer for any 401s that reach the client (e.g. direct fetch calls not going through the proxy)

**Cookie security**: All auth cookies are now `httpOnly`, `SameSite=Lax`, `Secure` (production only, to allow local HTTP dev).

### Tests Status
- Type check: not run (Bash restricted) — no type-unsafe patterns used; all types explicit
- Unit tests: not run
- Integration tests: not run

### Issues Encountered
- `document.cookie` grep returned a false positive in `set-tokens/route.ts` — it was in a comment only; confirmed no JS-accessible cookie assignments remain in login flow
- Google OAuth callback (`server/internal/auth/handler.go`) still sets only the access_token cookie server-side — the refresh_token is not set. This is a Go server file outside dashboard file ownership. Needs a follow-up task.

### Next Steps
- Follow-up: update Google OAuth callback (`server/internal/auth/handler.go`) to set `valt_refresh_token` cookie on redirect (currently only sets access_token)
- Manual test: login -> wait 15min -> verify action succeeds silently
- Manual test: expire both tokens -> verify redirect to /login
