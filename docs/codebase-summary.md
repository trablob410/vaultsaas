# Codebase Summary

## Structure

```
vaultsaas/
├── server/                  # Go 1.22 backend
│   ├── cmd/
│   │   ├── server/          # Main entry point (main.go)
│   │   ├── migrate/         # Migration runner (stub)
│   │   └── seed/            # Data seeder (stub)
│   ├── internal/
│   │   ├── auth/            # Argon2id, JWT RS256, middleware, handlers, Google OAuth2 (Phase 1.3/1.5)
│   │   ├── vault/           # MinIO storage, service CRUD, REST handler (Phase 1.3)
│   │   ├── middleware/      # Security headers, rate limiter (Phase 1.3)
│   │   ├── database/        # audit.go — audit log writer (Phase 1.3)
│   │   ├── policy/          # Risk tier engine (Tier 1-4 by credential type) (Phase 1.4)
│   │   ├── audit/           # Structured logger, SHA-256 hash chain, GET /audit/logs (Phase 1.4)
│   │   ├── notify/          # SMTP email notifications, no-op fallback (Phase 1.4)
│   │   ├── consent/         # User consent recording, POST /consent (Phase 1.4)
│   │   ├── workflow/        # Approval state machine, 6 endpoints, policy enforcement (Phase 1.4)
│   │   └── config/          # GoogleClientID/Secret/RedirectURL, DashboardURL (Phase 1.5)
│   ├── pkg/
│   │   ├── apierror/        # Standard API error JSON (Phase 1.3)
│   │   ├── crypto/          # Storage key generation (Phase 1.3)
│   │   └── validator/       # Input validation helpers (Phase 1.3)
│   ├── go.mod               # chi/v5, go-chi/cors, golang-jwt/jwt/v5, google/uuid, minio-go/v7,
│   │   │                   #   golang.org/x/oauth2, testify/v1.9.0
│   ├── go.sum
│   └── Dockerfile
│
├── dashboard/               # Next.js 15 + TypeScript + Tailwind v4 (Phase 1.5)
│   ├── src/
│   │   ├── app/
│   │   │   ├── (auth)/login/        # Google sign-in page
│   │   │   ├── (dashboard)/
│   │   │   │   ├── secrets/         # Secrets list + create/edit dialog + detail page
│   │   │   │   ├── approvals/       # Tabbed approval list + approve/reject dialog
│   │   │   │   ├── audit/           # Paginated audit log table
│   │   │   │   └── settings/        # User profile + sign-out
│   │   │   ├── api/
│   │   │   │   ├── proxy/[...path]/ # BFF JWT-forwarding proxy to Go backend
│   │   │   │   └── auth/logout/     # Logout route (clears httpOnly cookie)
│   │   │   ├── layout.tsx           # Sidebar nav + Header with user dropdown
│   │   │   ├── page.tsx
│   │   │   └── globals.css          # Dark mode CSS variables (zinc/slate palette)
│   │   ├── components/ui/           # shadcn/ui primitives (button, input, card, badge,
│   │   │   │                        #   dialog, table, dropdown-menu, select, label,
│   │   │   │                        #   textarea, separator, avatar)
│   │   └── lib/
│   │       ├── api-client.ts        # Typed fetch wrapper around BFF proxy
│   │       └── utils.ts             # cn() and shared helpers
│   ├── src/__tests__/               # vitest + happy-dom (19 tests)
│   │   ├── lib/utils.test.ts        # cn() helper (7 tests)
│   │   └── lib/api-client.test.ts   # fetch mocking, error handling (12 tests)
│   ├── package.json         # next@15.1.0, react@19, tailwindcss@4, lucide-react, vitest
│   ├── next.config.ts
│   ├── tsconfig.json
│   ├── postcss.config.mjs
│   └── Dockerfile
│
├── mcp-server/              # Rust MCP server (stdio transport) (Phase 1.6)
│   ├── src/
│   │   ├── main.rs          # Async tokio entry point, JSON-RPC 2.0 dispatcher
│   │   ├── error.rs         # Unified error type
│   │   ├── protocol.rs      # MCP protocol types
│   │   ├── config.rs        # Config loading (env vars)
│   │   ├── keychain.rs      # OS keychain auth token storage (keyring crate)
│   │   ├── client.rs        # reqwest HTTP client to Go backend
│   │   ├── crypto.rs        # AES-256-GCM decrypt (aes-gcm crate)
│   │   ├── tools.rs         # 5 MCP tools
│   │   └── resources.rs     # 3 MCP resources
│   ├── Cargo.toml
│   └── Dockerfile
│
├── docs/                    # Project documentation
├── plans/                   # Implementation plans
├── scripts/
│   ├── setup-dev.sh         # Dev environment setup
│   └── gen-keys.sh          # RSA key pair generation (Phase 1.3)
├── keys/                    # Local dev keys (gitignored)
├── docker-compose.yml       # Dev environment
├── docker-compose.prod.yml  # Production environment
├── Caddyfile                # Dev reverse proxy (CSP, security headers)
├── Caddyfile.prod           # Prod reverse proxy
├── Makefile                 # Common commands
├── .env.example             # Environment variable template
├── CLAUDE.md                # AI assistant instructions
├── SECURITY.md              # Security policy
└── README.md
```

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22, chi/v5 v5.1.0, go-chi/cors v1.2.1, golang-jwt/jwt/v5, google/uuid, minio-go/v7, golang.org/x/oauth2, testify v1.9.0 |
| Frontend | Next.js 15.1.0, React 19, TypeScript 5.7, Tailwind v4, shadcn/ui, lucide-react, Geist fonts, vitest |
| MCP Server | Rust (tokio async, serde, reqwest, keyring, aes-gcm — see Cargo.toml) |
| Proxy | Caddy (CSP headers, HSTS, security headers) |
| Infra | Docker Compose (dev + prod), Makefile |

## Implementation Status

| Phase | Status |
|---|---|
| 1.1 Scaffolding & Infrastructure | DONE — 2026-03-16 |
| 1.2 Database Layer & Migrations | DONE — 2026-03-16 |
| 1.3 Backend Core (auth, vault, middleware) | DONE — 2026-03-16 |
| 1.4 Backend Workflow + Audit + Policy + Notify + Consent | DONE — 2026-03-17 |
| 1.5 Next.js Dashboard + Google OAuth | DONE — 2026-03-17 |
| 1.6 Rust MCP Server | DONE — 2026-03-17 |
| 1.7 Testing & Hardening | DONE — 2026-03-17 |

- **server**: All Phase 1.3-1.5 backend live; Google OAuth2 flow added (`oauth.go`), migration 000012 adds OAuth columns to users table, `golang.org/x/oauth2` dep added; 74 unit tests passing
- **dashboard**: Full Next.js App Router implementation — Google sign-in, secrets CRUD, approvals (tabbed + dialogs), audit log viewer, settings/profile; BFF proxy pattern; 19 vitest tests passing
- **mcp-server**: Async tokio runtime; 5 tools (`request_secret_access`, `check_approval_status`, `get_credential`, `revoke_credential`, `list_my_secrets`); 3 resources (`vault://secrets`, `vault://requests/{id}`, `vault://audit/today`); OS keychain auth; 11 unit tests passing
- **testing**: Makefile targets `test-unit`, `test-integration`, `test-dashboard`, `test-mcp`, `security`; `.golangci.yml` enabled
- **infra**: Full Docker Compose (dev/prod) with JWT key volume mount, Caddy with CSP/security headers, Makefile, setup + key-gen scripts
