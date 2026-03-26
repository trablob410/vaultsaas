# Phase 5: Git Commit + Push to GitHub

**Priority:** P0 | **Effort:** 30min | **Status:** pending

Commit all production-ready code and push to GitHub.

## Steps

1. Review all changes:

```bash
git status
```

2. Stage production files (server, dashboard, CLI, docker, scripts, CI/CD):

```bash
git add \
  server/ \
  dashboard/ \
  valt-cli/ \
  docker-compose*.yml \
  Caddyfile* \
  scripts/ \
  .github/
```

3. DO NOT commit (keep confidential):

```
.env                          # Local env vars
.env.production              # Production secrets
keys/                         # JWT keys
backups/                      # Database backups
node_modules/                 # Dependencies
server/.bin/                  # Built binaries
server/server.exe             # Windows binary
valt-cli/valt-cli.exe        # Windows binary
mcp-server/target/           # Rust build artifacts
```

4. Review staged changes:

```bash
git diff --cached | head -100
```

5. Commit with descriptive message:

```bash
git commit -m "feat: phase 4 complete - production-ready deployment

- JWT auto-refresh via httpOnly cookies
- Onboarding wizard for new orgs
- Stripe billing activation with webhook handler
- Google OAuth production verification
- Email verification flow
- Password reset via email
- Landing page with pricing
- UptimeRobot monitoring integration
- Team invitations system
- CI/CD pipeline (GitHub Actions)
- Legal pages (privacy, ToS)
- Soft launch prep & docs"
```

6. Push to GitHub:

```bash
git push origin master
```

## Verification

- https://github.com/vaultsaas/vault/commits/master shows new commit
- GitHub Actions CI/CD pipeline triggers
- All checks pass (tests, linting, build)

## Notes

- Use conventional commit format: `feat:`, `fix:`, `docs:`, `chore:`
- Keep message under 72 chars for title
- No AI references in commit message
