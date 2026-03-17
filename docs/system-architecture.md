# System Architecture

## Overview

```
Client Layer:  MCP Server (Rust, stdio) + Dashboard (Next.js)
    ↓ HTTPS
Backend:       Go Monolith (chi/v5 router)
               - Auth, Vault, Workflow, Audit, Notify, Org, Agent modules
    ↓
Data Layer:    PostgreSQL 16 (metadata + audit) + MinIO (encrypted blobs)
    ↑
Proxy:         Caddy (reverse proxy + auto TLS)
```

## Components

| Component | Tech | Purpose |
|-----------|------|---------|
| server/ | Go 1.22+ | REST API monolith |
| dashboard/ | Next.js 15 | Web UI (SSR, client-side crypto) |
| mcp-server/ | Rust | Local MCP server for AI agents |
| PostgreSQL | v16 | Metadata, audit logs (partitioned) |
| MinIO | Latest | S3-compatible encrypted blob storage |
| Caddy | v2 | Reverse proxy, auto TLS |

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Go monolith over microservices | Simpler deploy/debug for MVP |
| Caddy over Kong | Auto TLS, simple config |
| PostgreSQL for audit logs | Partitioned tables sufficient for MVP |
| JWT RS256 over Keycloak | Fewer infra dependencies |
| State machine over Temporal | Approval workflow is simple enough |
| Docker Compose over K8s | Single-region MVP |

## Dashboard Architecture (Phase 1.5)

**Pattern:** Next.js App Router with BFF (Backend For Frontend) proxy.

```
Browser
  └─▶ Next.js Server Components (SSR)
        └─▶ api/proxy/[...path]  ──HTTPS──▶  Go backend (JWT in Authorization header)
              ▲
              │ httpOnly cookie (token) set by Go OAuth callback
```

- All Go API calls route through `api/proxy/[...path]` — the BFF reads the httpOnly cookie and forwards `Authorization: Bearer <token>` to Go.
- `api/auth/logout` clears the httpOnly cookie server-side.
- Google OAuth: browser → `GET /api/v1/auth/google` (Go redirects to Google) → Google callback → `GET /api/v1/auth/google/callback` (Go upserts user, sets cookie, redirects to dashboard).

### Route Groups

| Group | Routes | Description |
|-------|--------|-------------|
| `(auth)` | `/login` | Google sign-in landing |
| `(dashboard)` | `/secrets`, `/approvals`, `/audit`, `/settings`, `/orgs`, `/projects`, `/agents`, `/agents/[id]` | Authenticated pages with Sidebar + Header layout |

## MCP Server Architecture (Phase 1.6)

**Pattern:** Async tokio process reading JSON-RPC 2.0 from stdin, writing to stdout.

```
AI Agent (Claude, etc.)
  └─▶ stdio JSON-RPC 2.0
        └─▶ mcp-server (Rust, tokio)
              ├─ tools.rs   (5 tools)
              ├─ resources.rs (3 resources)
              └─ client.rs  ──HTTPS──▶  Go backend
```

### MCP Tools

| Tool | Description |
|------|-------------|
| `request_secret_access` | Submit approval request for a secret |
| `check_approval_status` | Poll approval state by request ID |
| `get_credential` | Retrieve decrypted credential after approval |
| `revoke_credential` | Revoke an active credential session |
| `list_my_secrets` | List secrets owned by the authenticated agent |

### MCP Resources

| Resource URI | Description |
|-------------|-------------|
| `vault://secrets` | Index of accessible secrets |
| `vault://requests/{id}` | Approval request detail |
| `vault://audit/today` | Today's audit log entries |

### MCP Tools (Phase 9 addition)

| Tool | Description |
|------|-------------|
| `authenticate_agent` | Accept an agent token, persist to OS keychain, clear env var |

### Auth Token Storage
1. OS keychain via `keyring` crate (primary)
2. `VALT_TOKEN` env var / `VALT_AGENT_TOKEN` env var (fallback)

## Test Architecture (Phase 1.7)

| Layer | Runner | Count | Location |
|-------|--------|-------|----------|
| Go unit | `go test ./...` | 74 | `server/internal/*/`, `server/pkg/*/` |
| Dashboard unit | `vitest` | 19 | `dashboard/src/lib/__tests__/` |
| Rust unit | `cargo test` | 11 | `mcp-server/src/*.rs` |

**Make targets:** `test-unit`, `test-integration`, `test-dashboard`, `test-mcp`, `security`

Go linting via `.golangci.yml` (govet shadow, errcheck, staticcheck, unused).

## Go Backend Packages

### Phase 1.3

| Package | Purpose |
|---------|---------|
| `internal/auth/` | Argon2id hashing, JWT RS256 issuance/validation, auth middleware, register/login/refresh handlers |
| `internal/vault/` | MinIO storage layer, vault service (CRUD), REST handler |
| `internal/middleware/` | Security headers, sliding-window rate limiter |
| `internal/database/audit.go` | Audit log writer |
| `pkg/apierror/` | Standard API error JSON response format |
| `pkg/validator/` | Input validation helpers |
| `pkg/crypto/` | Storage key generation |

### Phase 1.4

| Package | Purpose |
|---------|---------|
| `internal/policy/engine.go` | Risk tier classification — `api_key`=Tier1, `db_credential`/`ssh_key`/`oauth_token`=Tier2, `cloud_credential`/`personal_session`=Tier3, unknown=Tier4 |
| `internal/audit/logger.go` | Structured audit log writer |
| `internal/audit/hash-chain.go` | SHA-256 hash chain — links each audit record to its predecessor |
| `internal/audit/handler.go` | `GET /api/v1/audit/logs` handler |
| `internal/notify/service.go` | Notification dispatcher |
| `internal/notify/email.go` | SMTP email delivery; no-op when SMTP not configured |
| `internal/consent/service.go` | Consent record creation and lookup |
| `internal/consent/handler.go` | `POST /api/v1/consent` handler |
| `internal/workflow/service.go` | Approval state machine: pending → approved/rejected → active → expired/revoked |
| `internal/workflow/credential.go` | Temporary credential lifecycle (issue, expire, revoke) |
| `internal/workflow/handler.go` | 6 workflow endpoints |

## Workflow State Machine (Phase 1.4)

```
                  ┌─────────┐
    create ──────▶│ pending │
                  └────┬────┘
           approve ────┤──── reject
                  ┌────▼────┐   ┌──────────┐
                  │approved │   │ rejected │
                  └────┬────┘   └──────────┘
              activate │
                  ┌────▼────┐
                  │ active  │
                  └────┬────┘
          expire / ────┤──── revoke
         natural TTL   │
              ┌────────┴──────────┐
         ┌────▼────┐        ┌─────▼────┐
         │ expired │        │ revoked  │
         └─────────┘        └──────────┘
```

### Policy Rules (enforced per tier)

| Rule | Tier 1 | Tier 2 | Tier 3 |
|------|--------|--------|--------|
| Auto-approve | Yes | No | No |
| Single-use session | No | No | Yes |
| Duration cap | 24h | 8h | 1h |
| Daily request limit | 100 | 20 | 5 |
| Cool-down between requests | None | 1h | 4h |
| Reason min length | 0 | 20 chars | 50 chars |

### Phase 1.5 (OAuth additions)

| Package | Purpose |
|---------|---------|
| `internal/auth/oauth.go` | Google OAuth2 flow — redirect, callback, user upsert, httpOnly cookie issuance |
| `internal/config/config.go` | `GoogleClientID`, `GoogleClientSecret`, `GoogleRedirectURL`, `DashboardURL` config fields |

### Phase 8 (Org hierarchy)

| Package | Purpose |
|---------|---------|
| `internal/org/` | Organization CRUD, membership management, org-scoped JWT claims |
| `internal/workspace/` | Workspace CRUD scoped to org |
| `internal/project/` | Project CRUD scoped to workspace, project membership |

### Phase 9 (Agent identity)

| Package | Purpose |
|---------|---------|
| `internal/agent/service.go` | Agent identity lifecycle — create, rotate tokens, deactivate |
| `internal/agent/handler.go` | REST handlers for agent and token endpoints |
| `internal/agent/middleware.go` | Agent token auth middleware (validates `agent_tokens` table) |

## API Routes

### Phase 1.3

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| POST | `/api/v1/auth/register` | — | Register user |
| POST | `/api/v1/auth/login` | — | Login, return tokens |
| POST | `/api/v1/auth/refresh` | — | Refresh access token |
| GET | `/api/v1/secrets` | JWT | List secrets (paginated) |
| POST | `/api/v1/secrets` | JWT | Create secret |
| GET | `/api/v1/secrets/{id}` | JWT | Get secret by ID |
| PUT | `/api/v1/secrets/{id}` | JWT | Update secret |
| DELETE | `/api/v1/secrets/{id}` | JWT | Delete secret |

### Phase 1.4

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| POST | `/api/v1/requests` | JWT | Create access request |
| GET | `/api/v1/requests` | JWT | List access requests |
| GET | `/api/v1/requests/{id}` | JWT | Get request by ID |
| POST | `/api/v1/requests/{id}/approve` | JWT | Approve request |
| POST | `/api/v1/requests/{id}/reject` | JWT | Reject request |
| POST | `/api/v1/requests/{id}/revoke` | JWT | Revoke active session |
| GET | `/api/v1/audit/logs` | JWT | List audit log entries |
| POST | `/api/v1/consent` | JWT | Record user consent |

### Phase 1.5

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/v1/auth/google` | — | Initiate Google OAuth2 flow |
| GET | `/api/v1/auth/google/callback` | — | Google OAuth callback; upsert user, set httpOnly cookie |

### Phase 8 (Org hierarchy)

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/v1/orgs` | JWT | List orgs for current user |
| POST | `/api/v1/orgs` | JWT | Create org |
| GET | `/api/v1/orgs/{id}` | JWT | Get org |
| PUT | `/api/v1/orgs/{id}` | JWT | Update org |
| DELETE | `/api/v1/orgs/{id}` | JWT | Delete org |
| GET | `/api/v1/orgs/{id}/members` | JWT | List org members |
| POST | `/api/v1/orgs/{id}/members` | JWT | Add member to org |
| DELETE | `/api/v1/orgs/{id}/members/{uid}` | JWT | Remove member from org |
| GET | `/api/v1/orgs/{id}/workspaces` | JWT | List workspaces in org |
| POST | `/api/v1/orgs/{id}/workspaces` | JWT | Create workspace |
| GET | `/api/v1/workspaces/{id}` | JWT | Get workspace |
| PUT | `/api/v1/workspaces/{id}` | JWT | Update workspace |
| DELETE | `/api/v1/workspaces/{id}` | JWT | Delete workspace |
| GET | `/api/v1/workspaces/{id}/projects` | JWT | List projects in workspace |
| POST | `/api/v1/workspaces/{id}/projects` | JWT | Create project |
| GET | `/api/v1/projects/{id}` | JWT | Get project |
| PUT | `/api/v1/projects/{id}` | JWT | Update project |
| DELETE | `/api/v1/projects/{id}` | JWT | Delete project |
| GET | `/api/v1/projects/{id}/members` | JWT | List project members |
| POST | `/api/v1/projects/{id}/members` | JWT | Add member to project |
| DELETE | `/api/v1/projects/{id}/members/{uid}` | JWT | Remove member from project |

### Phase 9 (Agent identity)

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/v1/projects/{id}/agents` | JWT | List agents in project |
| POST | `/api/v1/projects/{id}/agents` | JWT | Create agent identity |
| GET | `/api/v1/agents/{id}` | JWT | Get agent |
| PUT | `/api/v1/agents/{id}` | JWT | Update agent |
| DELETE | `/api/v1/agents/{id}` | JWT | Delete agent |
| GET | `/api/v1/agents/{id}/tokens` | JWT | List agent tokens |
| POST | `/api/v1/agents/{id}/tokens` | JWT | Issue new agent token |
| DELETE | `/api/v1/agents/{id}/tokens/{tid}` | JWT | Revoke agent token |

## Agent Auth Flow (Phase 9)

```
MCP Server (Rust)
  └─▶ VALT_AGENT_TOKEN env var (or OS keychain via authenticate_agent tool)
        └─▶ Go backend — agent token middleware
              └─▶ validates token hash against agent_tokens table
                    └─▶ attaches agent identity to request context
```

- Agent tokens are stored as hashed values in `agent_tokens` table; plaintext returned only at creation.
- `authenticate_agent` MCP tool accepts a token, stores it in the OS keychain, and clears `VALT_AGENT_TOKEN` env usage.

## Security Architecture
- Zero-knowledge: server never sees plaintext
- Envelope encryption: secret → DEK → user master key
- Master key derived client-side from password (Argon2id)
- JWT RS256 (15min access, 7day refresh)
- Sliding-window rate limiting on all routes
- Security headers enforced via middleware
- Audit hash chain (SHA-256)
