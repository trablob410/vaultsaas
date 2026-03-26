# Phase 2: SMTP Configuration (SendGrid)

**Priority:** P0 | **Effort:** 1h | **Status:** pending

Set up email delivery via SendGrid for account verification, password reset, notifications.

## Steps

1. Go to https://sendgrid.com and create account (free tier: 100 emails/day)
2. Verify sender domain or email address:
   - Prefer: Add custom domain `noreply@valt.turbo.ai.vn` (requires DNS CNAME)
   - Fallback: Verify single email `noreply@valt.turbo.ai.vn`
3. Go to API Keys section, create new API key
4. Copy API key

## VPS Setup

SSH into VPS, edit `.env`:

```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=SG.xxx...
SMTP_FROM=noreply@valt.turbo.ai.vn
```

Restart server:

```bash
docker compose restart server
```

## Verification

1. Register new user in dashboard at https://valt.turbo.ai.vn
2. Check email inbox (may take 1-2min)
3. Verify email verification link arrives
4. Click link, verify email marked as verified in dashboard
5. Test password reset: login, forgot password, check email

## Notes

- SendGrid free tier sufficient for early beta (100 emails/day = ~3k users)
- Monitor SendGrid dashboard for bounce rate (aim <1%)
- Enable click tracking in SendGrid settings for better metrics
