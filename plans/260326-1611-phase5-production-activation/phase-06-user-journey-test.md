# Phase 6: Full User Journey Test

**Priority:** P1 | **Effort:** 2h | **Status:** pending

End-to-end test of complete user flow: register → verify email → login → create secret → request access → approve → use via CLI.

**Blocked by:** Phases 1-4 (Stripe, SMTP, Redis, OAuth)

## Full E2E Checklist

### Account Setup (15min)

- [ ] Open https://valt.turbo.ai.vn
- [ ] Click "Sign up"
- [ ] Enter email, password, name
- [ ] Submit → check email for verification link
- [ ] Click verification link
- [ ] Verify email marked as verified
- [ ] Login with credentials
- [ ] Onboarding wizard appears (create org name)
- [ ] Complete wizard → dashboard loads

### Project & Secret Setup (15min)

- [ ] Click "Create Project"
- [ ] Enter project name, description
- [ ] Create project
- [ ] Click "Create Secret"
- [ ] Enter secret name (e.g., "DB_PASSWORD")
- [ ] Enter secret value (e.g., "super-secret-123")
- [ ] Set policy: 2 approvers, 24hr duration
- [ ] Create secret

### Agent & Token Setup (15min)

- [ ] Go to Agents page
- [ ] Click "Create Agent"
- [ ] Enter agent name (e.g., "my-dev-agent")
- [ ] Create agent
- [ ] Click "Generate Token" on agent detail
- [ ] Copy agent token to clipboard

### CLI Setup (15min)

- [ ] Install valt CLI: `bash scripts/install-valt-cli.sh`
- [ ] Run `valt setup` (use agent token from above)
- [ ] Verify MCP server installed
- [ ] Run `valt list` → should show your secret

### Access Request Flow (20min)

- [ ] Run `valt get DB_PASSWORD` → returns "access request created"
- [ ] Go to dashboard → Access Requests page
- [ ] New request appears for DB_PASSWORD
- [ ] Click "Approve"
- [ ] Request moves to "active" status
- [ ] Run `valt get DB_PASSWORD` → returns "super-secret-123"
- [ ] Run `valt run echo $DB_PASSWORD` → prints secret value

### Dashboard Settings (10min)

- [ ] Go to Settings → Notifications
- [ ] Verify notification channel options exist
- [ ] Go to Settings → Upgrade
- [ ] Click "Upgrade to Pro"
- [ ] Verify Stripe checkout redirects
- [ ] Go to Settings → Team Invitations
- [ ] Invite team member via email
- [ ] Check if email sent (SMTP)

### Audit & Security (10min)

- [ ] Go to Audit Log
- [ ] Verify all actions logged: create secret, access request, approval, credential issued
- [ ] Verify no plaintext secrets in logs (only encrypted)
- [ ] Go to Agent detail page
- [ ] Verify token displayed once, never again

## Success Criteria

- All steps complete without errors
- No 500 errors in server logs
- Email notifications arrive (if SMTP enabled)
- Stripe checkout works (test mode)
- CLI correctly injects secrets
- Audit log is complete and accurate

## Troubleshooting

If any step fails:
1. Check server logs: `docker compose logs server | tail -100`
2. Check dashboard browser console for JS errors
3. Check database migration status: `docker compose exec postgres psql -U valt -c "\dt"`
4. Verify all env vars set correctly
