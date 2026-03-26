# Phase 4: Google OAuth E2E Verification

**Priority:** P1 | **Effort:** 30min | **Status:** pending

Verify Google OAuth flow works end-to-end on production domain.

## Prerequisites

- Google OAuth app already configured in Google Cloud Console
- `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` set in VPS `.env`
- Server restarted after env changes

## Checklist

1. Open https://valt.turbo.ai.vn in **incognito window** (fresh session)
2. Click **"Sign in with Google"** button on login page
3. Complete Google authentication (select account)
4. Browser redirects to https://valt.turbo.ai.vn/dashboard (or onboarding if new user)
5. Verify user email from Google account matches dashboard profile
6. Verify organization auto-created with name from Google first/last name
7. Logout, try again with different Google account to ensure multi-user works

## Troubleshooting

- **Redirect error:** Check `GOOGLE_REDIRECT_URI` in Google Cloud Console matches `https://valt.turbo.ai.vn/api/auth/google/callback`
- **Token exchange fails:** Verify `GOOGLE_CLIENT_SECRET` is correct in `.env`
- **Org not created:** Check migration 000040 (org_invitations) ran; check server logs

## Notes

- Test with 2+ different Google accounts if possible
- Verify email notification sent if SMTP configured
