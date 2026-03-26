---
phase: "4.12"
title: "Soft Launch Preparation"
priority: P2
effort: 3h
status: pending
depends_on: ["4.1-4.11"]
---

# Phase 4.12: Soft Launch Preparation

## Overview

Final polish and checklist before inviting first external users. Audit error handling, loading states, performance, and documentation.

## Checklist

### Error Handling Audit

- [ ] Replace generic 500 errors with actionable messages across all handlers
- [ ] Verify all API endpoints return consistent error format: `{ "error": { "message": "..." } }`
- [ ] Check error boundaries in dashboard (React error boundaries for component crashes)
- [ ] Add global error boundary in dashboard root layout
- [ ] Test: invalid UUID params return 400 not 500
- [ ] Test: missing required fields return descriptive errors

### Loading States

- [ ] All dashboard pages show skeleton/spinner during data fetch
- [ ] API calls show loading indicators (buttons disabled during submit)
- [ ] Empty states for: no secrets, no projects, no requests, no audit logs
- [ ] Network error states with retry button

### Rate Limiting Review

- [ ] Verify auth endpoints rate limited (login, register, forgot-password)
- [ ] Review agent rate limits per plan tier
- [ ] Add rate limiting to public endpoints (health excluded)
- [ ] Test rate limit responses return 429 with Retry-After header

### Performance Check

- [ ] Dashboard page load < 3s on 3G
- [ ] API response times < 500ms for common endpoints
- [ ] Check server memory usage under normal load
- [ ] Verify Docker resource limits appropriate (server: 1G, postgres: 2G)
- [ ] Check for N+1 queries in list endpoints

### API Documentation

- [ ] Add OpenAPI/Swagger spec for key endpoints (optional, low priority)
- [ ] Or: create a simple API reference page in docs/
- [ ] Document authentication methods (JWT, agent tokens)
- [ ] Document rate limits and error codes

### Audit Log Export

- [ ] Add CSV export endpoint: `GET /audit/logs/export?format=csv`
- [ ] Dashboard: "Export" button on audit log page
- [ ] Include: timestamp, user, action, resource, IP

### Security Final Check

- [ ] Verify all cookies are httpOnly + Secure + SameSite=Lax
- [ ] Verify CORS origins correctly configured for production domain
- [ ] Verify no debug/development endpoints exposed
- [ ] Check for sensitive data in error responses
- [ ] Verify HTTPS enforced (Caddy redirects HTTP -> HTTPS)
- [ ] Run basic security headers check (X-Frame-Options, CSP, etc.)

### Documentation

- [ ] Update `docs/deployment-guide.md` with all Phase 4 changes
- [ ] Update `docs/system-architecture.md` with new components
- [ ] Update `docs/project-roadmap.md` marking Phase 4 complete
- [ ] Update `README.md` with current feature list

### Smoke Test Checklist

Full end-to-end test on production:
- [ ] Register new account (email/password)
- [ ] Verify email
- [ ] Complete onboarding wizard
- [ ] Create a secret
- [ ] Create an agent + issue token
- [ ] Use MCP server to request secret access
- [ ] Approve request via dashboard
- [ ] Check audit log
- [ ] Invite team member
- [ ] Upgrade to Pro via Stripe
- [ ] Use CLI to list secrets
- [ ] Reset password flow
- [ ] Google OAuth login

## Success Criteria

- All smoke tests pass on production
- No generic 500 errors in common flows
- Loading states present on all pages
- Performance within acceptable bounds
- Documentation up to date

## Notes

- This phase is a sweep, not feature development
- Prioritize items that block user experience
- API documentation can be minimal (endpoint list + auth guide)
- Consider recording a demo video for marketing
