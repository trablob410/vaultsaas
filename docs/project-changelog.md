# Project Changelog

## [1.4.0] - 2026-03-17 — Phase 16: Security Hardening

### Added

**Encryption at Rest — Dynamic Secrets (Phase 16.1)**
- `dynsecret.Service` now holds `masterKey []byte`; `NewService(db, masterKey)` updated
- `CreateProvider`: AES-256-GCM encrypts provider config before INSERT
- `CreateLease`: AES-256-GCM encrypts lease credentials before INSERT
- `ListProviders`, `GetProvider`, `RevokeLease`: decrypt with plaintext fallback for pre-migration rows
- `VAULT_MASTER_KEY` env var documented in `.env.example` and `docker-compose.yml`

**Scanner RBAC (Phase 16.2)**
- `ResourceScans Resource = "scans"` added to `rbac/rbac.go` with full role permission matrix
- `rbac.Middleware` applied to project-scoped scanner routes
- `checkScanAccess` helper for scan-scoped routes
- `GetScanProjectID` added to scanner service
- `listFindings`, `importFinding`, `dismissFinding` replaced `_ = auth.UserIDFromContext(...)` stubs with real RBAC checks
- `scanner.NewHandler` now accepts `db *pgxpool.Pool`

**DynSecret RBAC (Phase 16.3)**
- `ResourceDynSecret Resource = "dynsecret"` added to `rbac/rbac.go`
- `rbac.Middleware` applied to project-scoped dynsecret routes
- `checkProviderAccess` helper for provider-scoped routes
- `GetLeaseProviderID` added to dynsecret service
- Auth checks applied to `createLease`, `listLeases`, `revokeLease`
- `dynsecret.NewHandler` now accepts `db *pgxpool.Pool`

**Org-Scoped Usage Counts (Phase 16.4)**
- Fixed global `SELECT COUNT(*) FROM secrets` → 3-table JOIN: `secrets → projects → workspaces WHERE w.org_id = $1`
- Fixed `agents_count` same way via `agent_identities → projects → workspaces`
- Removed stale TODO comments from `tracker.go`

**High-Priority Security Fixes (Phase 16.5)**
- H1: `GetRequest`, `Approve`, `Reject` now authorize assigned approvers and secret owners, not requester only
- H2: `ratelimit/middleware.go` — `Middleware(rpm)` applied to auth route group (60rpm per agent via X-Agent-ID); Redis ZADD member includes `rand.Int63()` to prevent millisecond-collision deduplication
- H3: `Approve` handler returns HTTP 500 if `IssueCredential` fails (was silently swallowed)
- H4: MCP `scanner_tools.rs` rejects absolute paths, Windows drive letters (`C:`), `..` sequences, paths > 500 chars
- H5: Migration `000024` adds `rejection_reason TEXT` to `approval_steps`; `AdvanceChain` persists rejection reason; `ApprovalStep` struct updated with `RejectionReason *string`

**RBAC Middleware Hardening**
- `rbac/middleware.go` returns HTTP 400 on missing `project_id` (was silent pass-through)

### Security Impact
| Finding | Severity | Status |
|---------|----------|--------|
| DynSecret provider/lease config stored plaintext | CRITICAL | Fixed |
| Scanner routes unauthenticated | CRITICAL | Fixed |
| DynSecret routes unauthenticated | CRITICAL | Fixed |
| Usage counts not org-scoped (data leak) | CRITICAL | Fixed |
| Approval access limited to requester only | HIGH | Fixed |
| Rate limiter Redis ZADD dedup collision | HIGH | Fixed |
| IssueCredential failure silently swallowed | HIGH | Fixed |
| MCP path traversal / absolute path injection | HIGH | Fixed |
| Rejection reason not persisted | MEDIUM | Fixed |
| RBAC middleware silent on missing project_id | MEDIUM | Fixed |

### Known Follow-ups (Non-blocking)
- `ListProviders`: silently continues with empty config when both decrypt paths fail (minor UX)
- `ListPending`: non-owner approvers cannot see approvals assigned to them

---

## [1.3.0] - 2026-03-17 — Phases 10-15: Scanner, Dynamic Secrets, Gateway, RBAC, CLI, Free Tier

### Added

**Phase 10 — Secret Scanner**
- `mcp-server/src/scanner.rs` — 17 regex patterns (AWS, GitHub, Stripe, OpenAI, generic high-entropy keys)
- 2 new MCP tools: `scan_secrets`, `store_secret` (total: 8 tools)
- `server/internal/scanner/` — scan result + finding CRUD, 5 HTTP endpoints
- Migration `000018`: `scan_results` + `scan_findings` tables
- Dashboard `/scans` (list) and `/scans/[id]` (findings detail) pages

**Phase 11 — Dynamic Secrets**
- `server/internal/dynsecret/` — `Provider` interface, `PostgresProvider` (CREATE ROLE LOGIN ... VALID UNTIL), lease management, auto-expiry background worker
- Migrations `000019-000020`: `dynamic_providers` + `dynamic_leases` tables
- New MCP tool: `request_dynamic_secret`
- Dashboard `/providers` page

**Phase 12 — MCP Gateway (HTTP/SSE)**
- `mcp-server/src/http.rs` — axum HTTP server, `POST /mcp` and `GET /mcp/sse` (SSE) endpoints, Bearer token auth
- `--transport stdio|http` and `--port` CLI flags via clap; stdio remains default (backward-compatible)

**Phase 13 — Enhanced RBAC + Policies**
- `server/internal/rbac/` — role permission matrix (owner/admin/member/viewer), RBAC middleware
- `server/internal/ratelimit/` — Redis sliding-window per-agent rate limiting (go-redis/v9); optional via `REDIS_URL`
- `server/internal/workflow/approval-chain.go` — multi-step sequential approval chains
- Migrations `000021-000022`: `approval_steps` table, `rate_limit_rpm` column on `agent_identities`

**Phase 14 — CLI + SDKs**
- `server/cmd/valt/` — `valt` CLI (cobra): `login`, `secrets list/create`, `agents list/create`, `scan`, `run`; config at `~/.valt/config.json`
- `sdk/go/` — Go SDK (stdlib only, module `github.com/valt-dev/valt-go`)
- `sdk/python/` — Python SDK (urllib only, zero dependencies)

**Phase 15 — Cloud Free Tier**
- `server/internal/usage/` — usage tracking, free tier enforcement middleware (50 secrets, 3 agents, 1000 req/day)
- Migration `000023`: `usage_metrics` table
- `GET /orgs/{id}/usage` endpoint
- Dashboard `/settings/upgrade` page with `UsageBar` components; Sidebar Upgrade link

---

## [0.7.0] - 2026-03-17 — Phases 1.5-1.7: Dashboard + MCP Server + Testing

### Added

**Google OAuth (Go backend)**
- Migration `000012_alter_users_oauth` — adds `google_id`, `avatar_url`, `display_name`, `auth_provider` columns; `password_hash` nullable
- `internal/auth/oauth.go` — Google OAuth2 flow: `GET /api/v1/auth/google`, `GET /api/v1/auth/google/callback`; upserts user by `google_id`/email; issues httpOnly cookie tokens
- `internal/config/config.go` — added `GoogleClientID`, `GoogleClientSecret`, `GoogleRedirectURL`, `DashboardURL` fields
- `golang.org/x/oauth2` dependency added

**Next.js Dashboard (`dashboard/`)**
- Route groups: `(auth)/login` (Google sign-in), `(dashboard)/secrets`, `approvals`, `audit`, `settings`
- BFF routes: `api/proxy/[...path]` (JWT forwarding proxy), `api/auth/logout`
- shadcn/ui primitives: button, input, card, badge, dialog, table, dropdown-menu, select, label, textarea, separator, avatar
- Layout: Sidebar nav + Header with user dropdown (lucide-react icons)
- Secrets CRUD: list table, create/edit dialog, detail page
- Approvals: tabbed list (All/Pending/Approved/Rejected), approve/reject dialog
- Audit: paginated log table
- Settings: user profile display + sign-out
- Dark mode default via CSS variables (zinc/slate palette), Geist Sans + Geist Mono fonts
- TypeScript strict mode, Server Components by default

**Rust MCP Server (`mcp-server/`)**
- Converted from sync to async tokio runtime
- New modules: `error.rs`, `protocol.rs`, `config.rs`, `keychain.rs`, `client.rs`, `crypto.rs`, `tools.rs`, `resources.rs`
- 5 MCP tools: `request_secret_access`, `check_approval_status`, `get_credential`, `revoke_credential`, `list_my_secrets`
- 3 MCP resources: `vault://secrets`, `vault://requests/{id}`, `vault://audit/today`
- OS keychain auth token storage (keyring crate), env var fallback
- AES-256-GCM decrypt (aes-gcm crate), reqwest HTTP client to Go backend

**Testing (Phase 1.7)**
- 74 Go unit tests across: `pkg/validator`, `pkg/apierror`, `pkg/crypto`, `internal/policy`, `internal/audit`, `internal/auth` (password, jwt, middleware), `internal/middleware` (rate-limit, security-headers)
- 19 dashboard tests via vitest + happy-dom: `lib/utils.test.ts` (7), `lib/api-client.test.ts` (12)
- 11 Rust unit tests passing, `cargo clippy -- -D warnings` clean
- Makefile targets added: `test-unit`, `test-integration`, `test-dashboard`, `test-mcp`, `security`
- `.golangci.yml`: govet shadow, errcheck, staticcheck, unused enabled
- `github.com/stretchr/testify@v1.9.0` added to go.mod

### API Routes Added

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/v1/auth/google` | — | Redirect to Google OAuth consent |
| GET | `/api/v1/auth/google/callback` | — | Google OAuth callback, issue tokens |

---

## [0.4.0] - 2026-03-17 — Phase 1.4: Workflow + Audit + Policy + Notify + Consent

### Added

**DB Migrations (000007-000011)**
- `000007`: alter `secrets` — add `credential_type`, `source`, `version` columns
- `000008`: alter `access_requests` — rename `requester_id` → `requester_user_id`, add `ai_agent_id`, `rejection_reason`, `requested_duration_minutes`
- `000009`: alter `credential_sessions` — add `credential_type`, `revoked_at`
- `000010`: alter `audit_logs` — add `user_id`, `resource_type`, `event_type`, `status`, `ip_address`, `user_agent`, `region_code`
- `000011`: new `user_consent_logs` table

**Policy (`internal/policy/`)**
- `engine.go` — risk tier classification: `api_key`=Tier1, `db_credential`/`ssh_key`/`oauth_token`=Tier2, `cloud_credential`/`personal_session`=Tier3
- Policy enforcement rules: duration cap, reason length, daily request limit, cool-down period, auto-approve Tier 1, single-use sessions for Tier 3

**Audit (`internal/audit/`)**
- `logger.go` — structured audit log writer
- `hash-chain.go` — SHA-256 hash chain integrity for audit records
- `handler.go` — `GET /api/v1/audit/logs` endpoint

**Notify (`internal/notify/`)**
- `service.go` + `email.go` — SMTP email notification service; no-op when SMTP not configured

**Consent (`internal/consent/`)**
- `service.go` + `handler.go` — user consent recording, `POST /api/v1/consent`

**Workflow (`internal/workflow/`)**
- `service.go` — full approval state machine: pending → approved/rejected → active → expired/revoked
- `credential.go` — temporary credential lifecycle management
- `handler.go` — 6 endpoints (see API Routes below)

**Routes wired in `main.go`**
- `POST /api/v1/requests` — create access request
- `GET /api/v1/requests` — list access requests
- `GET /api/v1/requests/{id}` — get request by ID
- `POST /api/v1/requests/{id}/approve` — approve request
- `POST /api/v1/requests/{id}/reject` — reject request
- `POST /api/v1/requests/{id}/revoke` — revoke active session
- `GET /api/v1/audit/logs` — list audit log entries
- `POST /api/v1/consent` — record user consent

---

## [0.3.0] - 2026-03-16 — Phase 1.3: Backend Core

### Added

**Auth (`internal/auth/`)**
- Password hashing with Argon2id
- JWT RS256 token issuance and validation (15min access / 7day refresh)
- Auth middleware for protected routes
- Handlers: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`

**Vault (`internal/vault/`)**
- MinIO storage layer for encrypted blobs
- Vault service with full CRUD operations
- REST handler: `GET /api/v1/secrets` (paginated), `POST /api/v1/secrets`, `GET /api/v1/secrets/{id}`, `PUT /api/v1/secrets/{id}`, `DELETE /api/v1/secrets/{id}`

**Middleware (`internal/middleware/`)**
- Security headers middleware
- Sliding-window rate limiter

**Audit (`internal/database/audit.go`)**
- Audit log writer

**Shared packages**
- `pkg/apierror/`: standard API error JSON response format
- `pkg/validator/`: input validation helpers
- `pkg/crypto/`: storage key generation

**Infrastructure**
- `scripts/gen-keys.sh`: RSA key pair generation script
- `docker-compose.yml`: added JWT key paths and volume mount for `keys/`

**Dependencies added**
- `golang-jwt/jwt/v5`
- `google/uuid`
- `minio-go/v7`

---

## [0.1.0] - 2026-03-16 — MVP Phase 1: Project Scaffold

### Added

**Backend (server/)**
- `cmd/server/main.go`: Go 1.22 HTTP server with chi/v5, configurable CORS via `CORS_ORIGINS` env var, `GET /health` endpoint returning JSON status, `/api/v1` router stub
- chi/v5 middleware stack: Logger, Recoverer, RequestID, RealIP, 30s Timeout
- `go.mod`/`go.sum`: dependencies — `github.com/go-chi/chi/v5 v5.1.0`, `github.com/go-chi/cors v1.2.1`
- `internal/` stubs: auth, vault, workflow, audit, notify, middleware, config, database
- `pkg/` stubs: crypto, validator
- `cmd/migrate/`, `cmd/seed/` runner stubs
- `server/Dockerfile`

**Dashboard (dashboard/)**
- Next.js 15.1.0 + React 19 + TypeScript 5.7 + Tailwind v4 scaffold
- `src/app/layout.tsx`, `page.tsx`, `globals.css`
- Route group stubs: `(auth)/`, `(dashboard)/`, `api/`
- `next.config.ts`, `tsconfig.json`, `postcss.config.mjs`
- `dashboard/Dockerfile`

**MCP Server (mcp-server/)**
- Rust MCP server with JSON-RPC 2.0 entry point (`src/main.rs`)
- Sub-package stubs: `mcp/`, `client/`, `keychain/`
- `Cargo.toml`, `mcp-server/Dockerfile`

**Infrastructure**
- `docker-compose.yml`: dev environment (server, dashboard, mcp-server)
- `docker-compose.prod.yml`: production environment
- `Caddyfile`: dev reverse proxy with CSP headers, HSTS, security headers
- `Caddyfile.prod`: production Caddy config
- `Makefile`: common dev/build/deploy commands
- `scripts/setup-dev.sh`: dev environment bootstrap script
- `.env.example`: environment variable template

**Project**
- `README.md`, `CLAUDE.md`, `SECURITY.md`, `LICENSE`
- `.gitignore` (excludes `.env`, `keys/`, binaries)
- `docs/`: architecture, code standards, deployment, system architecture, PDR
