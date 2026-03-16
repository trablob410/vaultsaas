# Development Roadmap

## Phase 1: MVP (Current)

### Phase 1.1: Scaffolding & Infrastructure [DONE - 2026-03-16]
- [x] Monorepo structure
- [x] Docker Compose (dev + prod)
- [x] Service stubs (Go, Next.js, Rust)
- [x] Documentation

### Phase 1.2: Database Layer & Migrations [Planned]
- [ ] PostgreSQL connection pool
- [ ] All 6 tables (users, secrets, access_requests, credential_sessions, audit_logs, refresh_tokens)
- [ ] Migration runner
- [ ] Seed script

### Phase 1.3: Backend Core [Planned]
- [ ] Auth (register, login, refresh, JWT RS256, Argon2id)
- [ ] Vault (CRUD, MinIO storage, envelope encryption)
- [ ] Middleware (rate limit, CORS, security headers)

### Phase 1.4: Backend Workflow + Audit [Planned]
- [ ] Access request CRUD + approval state machine
- [ ] Temporary credential management
- [ ] Audit logging with hash chain
- [ ] Email notifications

### Phase 1.5: Next.js Dashboard [Planned]
- [ ] Auth pages (login, register)
- [ ] Secret management (list, create, reveal)
- [ ] Approval management
- [ ] Audit log viewer
- [ ] Client-side crypto (Web Crypto API)

### Phase 1.6: Rust MCP Server [Planned]
- [ ] MCP Protocol 1.0 over stdio
- [ ] 5 tools, 3 resources
- [ ] OS keychain for token storage

### Phase 1.7: Testing & Hardening [Planned]
- [ ] 80% Go test coverage
- [ ] Integration tests (testcontainers)
- [ ] E2E tests (Playwright)
- [ ] CI/CD (GitHub Actions)

## Phase 2: Product-Market Fit [Future]
- Zalo/Slack notifications
- VSCode extension
- Team management + RBAC
- TOTP 2FA

## Phase 3: Scale [Future]
- Kubernetes migration
- Multi-region
- SSO (OIDC)
