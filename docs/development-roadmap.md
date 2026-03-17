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

### Phase 1.3: Backend Core [DONE - 2026-03-16]
- [x] Auth (register, login, refresh, JWT RS256, Argon2id)
- [x] Vault (CRUD, MinIO storage, envelope encryption)
- [x] Middleware (rate limit, CORS, security headers)

### Phase 1.4: Backend Workflow + Audit + Policy + Notify + Consent [DONE - 2026-03-17]
- [x] DB migrations 000007-000011 (secrets, access_requests, credential_sessions, audit_logs, user_consent_logs)
- [x] Policy engine — risk tier system (Tier 1-4 by credential type)
- [x] Audit logging with SHA-256 hash chain, GET /audit/logs endpoint
- [x] SMTP email notification service
- [x] User consent recording, POST /consent endpoint
- [x] Approval state machine (pending→approved/rejected→active→expired/revoked), 6 workflow endpoints
- [x] Policy enforcement (duration cap, reason length, daily limit, cool-down, auto-approve Tier 1, single-use Tier 3)

### Phase 1.5: Next.js Dashboard + Google OAuth [DONE - 2026-03-17]
- [x] Google OAuth2 flow (Go backend: `oauth.go`, migration 000012, `golang.org/x/oauth2`)
- [x] Auth page — Google sign-in (`(auth)/login`)
- [x] Secrets CRUD — list table, create/edit dialog, detail page
- [x] Approvals — tabbed list (All/Pending/Approved/Rejected), approve/reject dialog
- [x] Audit log viewer — paginated table
- [x] Settings — user profile + sign-out
- [x] BFF proxy (`api/proxy/[...path]`) for JWT forwarding
- [x] shadcn/ui primitives, Sidebar + Header layout, dark mode

### Phase 1.6: Rust MCP Server [DONE - 2026-03-17]
- [x] Async tokio runtime
- [x] 5 MCP tools: `request_secret_access`, `check_approval_status`, `get_credential`, `revoke_credential`, `list_my_secrets`
- [x] 3 MCP resources: `vault://secrets`, `vault://requests/{id}`, `vault://audit/today`
- [x] OS keychain auth token storage (keyring crate), env var fallback
- [x] reqwest HTTP client to Go backend, AES-256-GCM decrypt stub
- [x] `cargo clippy -- -D warnings` clean

### Phase 1.7: Testing & Hardening [DONE - 2026-03-17]
- [x] 74 Go unit tests passing (auth, vault, policy, audit, middleware, validator, apierror, crypto)
- [x] 19 dashboard vitest tests passing (utils, api-client)
- [x] 11 Rust unit tests passing
- [x] Makefile targets: `test-unit`, `test-integration`, `test-dashboard`, `test-mcp`, `security`
- [x] `.golangci.yml`: govet shadow, errcheck, staticcheck, unused enabled

### Phase 8: Organization Hierarchy [DONE - 2026-03-17]
- [x] DB migrations 013-015: organizations, workspaces, projects, memberships, secrets.project_id
- [x] Go packages: `internal/org/`, `internal/workspace/`, `internal/project/`
- [x] API routes: /orgs, /orgs/{id}/members, /orgs/{id}/workspaces, /workspaces/{id}/projects, /projects/{id}, /projects/{id}/members
- [x] Dashboard: org context in Sidebar, /orgs page, /projects page

### Phase 9: AI Agent Identity [DONE - 2026-03-17]
- [x] DB migrations 016-017: agent_identities, agent_tokens tables
- [x] Go package: `internal/agent/` (service, handler, middleware)
- [x] API routes: /projects/{id}/agents, /agents/{id}, /agents/{id}/tokens, /agents/{id}/tokens/{tid}
- [x] MCP server: `authenticate_agent` tool, VALT_AGENT_TOKEN env, agent token keychain storage
- [x] Dashboard: /agents page (list + create), /agents/[id] page (detail + token management)

## Phase 2: Product-Market Fit [Future]
- Zalo/Slack notifications
- VSCode extension
- Team management + RBAC
- TOTP 2FA

## Phase 3: Scale [Future]
- Kubernetes migration
- Multi-region
- SSO (OIDC)
