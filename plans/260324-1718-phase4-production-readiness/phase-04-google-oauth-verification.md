---
phase: "4.4"
title: "Google OAuth E2E Verification"
priority: P1
effort: 1h
status: pending
---

# Phase 4.4: Google OAuth E2E Verification

## Context Links
- `server/internal/auth/oauth.go` -- Google OAuth handler
- `server/internal/auth/handler.go:36-41` -- OAuth config setup
- `server/internal/config/config.go:42-45` -- GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL
- `docker-compose.prod.yml:75-77` -- Google OAuth env vars

## Overview

Google OAuth code exists. Need to verify it works end-to-end on production domain with HTTPS.

## Testing Checklist

### Step 1: Google Cloud Console Configuration
- [ ] Verify OAuth consent screen is configured (production, not testing)
- [ ] Verify authorized redirect URI: `https://valt.turbo.ai.vn/api/v1/auth/google/callback`
- [ ] Verify authorized JavaScript origins: `https://valt.turbo.ai.vn`
- [ ] Check OAuth client type is "Web application"

### Step 2: VPS Environment
- [ ] Verify env vars set: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- [ ] `GOOGLE_REDIRECT_URL` must be `https://valt.turbo.ai.vn/api/v1/auth/google/callback`
- [ ] `DASHBOARD_URL` must be `https://valt.turbo.ai.vn`

### Step 3: Flow Testing
- [ ] Click "Sign in with Google" on login page
- [ ] Google consent screen appears
- [ ] After consent, redirected back to dashboard
- [ ] User created in DB (for new users) or logged in (existing)
- [ ] Both `valt_access_token` and `valt_refresh_token` cookies set (after 4.1)
- [ ] Dashboard loads correctly after redirect

### Step 4: Edge Cases
- [ ] Test with existing email/password user -> Google OAuth links to same account
- [ ] Test with new Google account -> creates user
- [ ] Test canceling at Google consent screen -> redirected to login with error
- [ ] Test with invalid redirect URI -> proper error handling

## Success Criteria

- Full Google OAuth flow works on `https://valt.turbo.ai.vn`
- New and existing users handled correctly
- Cookies set properly (httpOnly after 4.1 is implemented)

## Notes

- If Google Console shows "Testing" status, need to submit for verification or add test users
- May need to publish OAuth consent screen for external users
- Consider adding Google OAuth app verification if user count > 100
