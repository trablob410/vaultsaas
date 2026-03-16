# Codebase Summary

## Structure

```
vaultsaas/
├── server/                  # Go 1.22 backend
│   ├── cmd/
│   │   ├── server/          # Main entry point (main.go)
│   │   ├── migrate/         # Migration runner (stub)
│   │   └── seed/            # Data seeder (stub)
│   ├── internal/            # Private packages (stubs, Phase 2+)
│   │   ├── auth/
│   │   ├── vault/
│   │   ├── workflow/
│   │   ├── audit/
│   │   ├── notify/
│   │   ├── middleware/
│   │   ├── config/
│   │   └── database/
│   ├── pkg/                 # Shared utilities (stubs, Phase 2+)
│   │   ├── crypto/
│   │   └── validator/
│   ├── go.mod               # github.com/go-chi/chi/v5, github.com/go-chi/cors
│   ├── go.sum
│   └── Dockerfile
│
├── dashboard/               # Next.js 15 + TypeScript + Tailwind v4
│   ├── src/app/
│   │   ├── (auth)/          # Auth route group (stub)
│   │   ├── (dashboard)/     # Dashboard route group (stub)
│   │   ├── api/             # API routes (stub)
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── globals.css
│   ├── package.json         # next@15.1.0, react@19, tailwindcss@4
│   ├── next.config.ts
│   ├── tsconfig.json
│   ├── postcss.config.mjs
│   └── Dockerfile
│
├── mcp-server/              # Rust MCP server (stdio transport)
│   ├── src/
│   │   ├── main.rs          # JSON-RPC 2.0 entry point
│   │   ├── mcp/             # Protocol, tools, resources (stub)
│   │   ├── client/          # Valt API client (stub)
│   │   └── keychain/        # OS keychain integration (stub)
│   ├── Cargo.toml
│   └── Dockerfile
│
├── docs/                    # Project documentation
├── plans/                   # Implementation plans
├── scripts/
│   └── setup-dev.sh         # Dev environment setup
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

## Tech Stack (Phase 1 — scaffold only)

| Layer | Technology |
|---|---|
| Backend | Go 1.22, chi/v5 v5.1.0, go-chi/cors v1.2.1 |
| Frontend | Next.js 15.1.0, React 19, TypeScript 5.7, Tailwind v4 |
| MCP Server | Rust (tokio, serde — see Cargo.toml) |
| Proxy | Caddy (CSP headers, HSTS, security headers) |
| Infra | Docker Compose (dev + prod), Makefile |

## Phase 1 Implementation Status

- **server**: Health endpoint (`GET /health`) live; `/api/v1` router stub; all `internal/` and `pkg/` packages are empty directory stubs for Phase 2+
- **dashboard**: Scaffold only — layout, landing page, globals.css; no components or lib yet
- **mcp-server**: JSON-RPC 2.0 handler stub; all sub-packages are stubs
- **infra**: Full Docker Compose (dev/prod), Caddy with CSP/security headers, Makefile, setup script
