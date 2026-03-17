# Auth Methods Research: Next.js 15 + Go Backend

**Date:** 2026-03-17
**Context:** Valt has a Go backend with JWT RS256 auth (register/login/refresh), Argon2id passwords, refresh token rotation. Need to add Google OAuth for SaaS dashboard.

---

## Existing Auth Stack (Go Backend)

- `auth.JWTManager`: RS256 signing with `golang-jwt/jwt/v5`
- 15min access tokens, 7d refresh tokens (SHA-256 hashed, stored in `refresh_tokens` table)
- `auth.AuthMiddleware`: validates Bearer token, extracts `sub` (user ID)
- Users table: `id`, `email`, `password_hash`, `region_code`, `status`
- Password hashing: Argon2id

**Key constraint:** The MCP server and API consumers already depend on this JWT flow. Any auth solution must produce the same JWT RS256 access tokens.

---

## Option 1: Auth.js v5 (NextAuth)

**What:** Frontend auth library, handles OAuth flows, session management.

| Aspect | Assessment |
|--------|-----------|
| Google OAuth | Built-in Google provider, works well |
| Next.js 15 compat | v5 supports App Router natively |
| Go backend integration | **Poor fit.** Auth.js wants to own the session/token lifecycle. Would create dual auth systems -- Auth.js sessions for dashboard, Go JWTs for API. Requires bridging: Auth.js callback must call Go backend to create user + get JWT. |
| Pricing | Free, open source |
| Complexity | Medium-high. Two auth systems to maintain. Session sync issues. |

**Verdict:** Over-engineered for this case. Creates unnecessary dual-auth complexity when Go already handles JWT issuance.

---

## Option 2: Clerk

| Aspect | Assessment |
|--------|-----------|
| Google OAuth | Built-in, zero config |
| Next.js 15 compat | First-class support, App Router middleware |
| Go backend integration | **Bad fit.** Clerk owns user management entirely. Would need to replace Go JWT auth with Clerk JWT verification, migrate user table, depend on external service for core security function. |
| Pricing | Free: 10k MAU. Pro: $25/mo (100k MAU). |
| Complexity | Low to start, high to integrate with existing Go auth |

**Verdict:** Wrong choice. Valt is a security product -- externalizing auth to a third party contradicts the zero-knowledge architecture. Also creates vendor lock-in.

---

## Option 3: Custom Google OAuth via Go Backend (golang.org/x/oauth2)

| Aspect | Assessment |
|--------|-----------|
| Google OAuth | `golang.org/x/oauth2/google` -- well-maintained, standard library quality |
| Next.js 15 compat | N/A (backend handles OAuth). Dashboard just redirects to `/api/v1/auth/google` and receives JWT back. |
| Go backend integration | **Perfect fit.** Adds 2 endpoints to existing auth handler: `GET /auth/google` (redirect to Google) and `GET /auth/google/callback` (exchange code, create/find user, issue JWT). Returns same JWT RS256 tokens. |
| Pricing | Free. Google OAuth is free. |
| Complexity | Low. ~100-150 lines of Go code. Reuses existing `issueTokens()`. |

**Flow:**
1. Dashboard: `window.location = "/api/v1/auth/google"`
2. Go redirects to Google consent screen
3. Google redirects back to Go callback endpoint
4. Go exchanges code for Google user info (email, name, avatar)
5. Go upserts user in `users` table (null `password_hash` for OAuth-only users)
6. Go calls existing `issueTokens()` -- same JWT RS256 access + refresh tokens
7. Go redirects to dashboard with tokens (via URL fragment or sets httpOnly cookies)

**Changes needed:**
- Add `google_id`, `avatar_url`, `display_name` columns to `users` table
- Make `password_hash` nullable (OAuth users have no password)
- Add `auth_provider` column (enum: `local`, `google`)
- 2 new routes in `auth/handler.go`
- Config: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`

---

## Option 4: Better Auth

| Aspect | Assessment |
|--------|-----------|
| Google OAuth | Supported via plugin |
| Next.js 15 compat | App Router support, newer library |
| Go backend integration | **Same problem as Auth.js.** TypeScript-only, wants to own auth. Creates dual system. |
| Pricing | Free, open source |
| Complexity | Medium. Less mature ecosystem than Auth.js. Smaller community. |

**Verdict:** Same dual-auth problem as Auth.js, with less community support.

---

## Comparison Matrix

| Criteria | Auth.js v5 | Clerk | Custom Go OAuth | Better Auth |
|----------|-----------|-------|-----------------|-------------|
| Fits existing JWT flow | No | No | **Yes** | No |
| Implementation effort | 2-3 days | 1-2 days | **0.5-1 day** | 2-3 days |
| Ongoing maintenance | Medium | Low | **Low** | Medium |
| Security control | Partial | None | **Full** | Partial |
| Vendor dependency | None | High | **None** | None |
| Google OAuth quality | Good | Excellent | Good | Decent |

---

## Recommendation: Option 3 -- Custom Go OAuth

**Why:** Simplest path. Reuses 100% of existing auth infrastructure. No new dependencies on the frontend. No dual-auth complexity. Full security control (critical for a vault product).

**Implementation summary:**
1. `go get golang.org/x/oauth2` (one dependency)
2. Add migration: alter `users` table (nullable `password_hash`, add `google_id`, `auth_provider`)
3. Add `GoogleLogin` + `GoogleCallback` to `auth/handler.go` (~100 lines)
4. Dashboard: single "Sign in with Google" button that redirects to Go endpoint
5. Callback redirects to dashboard with JWT in httpOnly secure cookie

**Token delivery to dashboard:** Use httpOnly cookies instead of URL params for security. Set `access_token` and `refresh_token` as httpOnly, Secure, SameSite=Lax cookies. Dashboard reads them automatically on API calls.

---

## Unresolved Questions

1. Should existing email/password users be able to link their Google account later? (Suggest: yes, match by email)
2. Cookie-based vs Authorization header for dashboard API calls? (Suggest: cookies for dashboard, Bearer header for MCP/API clients)
3. Should Google-only users be able to set a password later? (Suggest: defer, YAGNI for MVP)
