# Codebase Summary

## Structure

```
valt/
├── server/              # Go 1.22+ backend monolith
│   ├── cmd/server/      # Main entry point
│   ├── cmd/migrate/     # Migration runner
│   ├── cmd/seed/        # Data seeder
│   ├── internal/        # Private packages
│   │   ├── auth/        # Authentication (JWT, Argon2id)
│   │   ├── vault/       # Secret management + encryption
│   │   ├── workflow/    # Approval state machine
│   │   ├── audit/       # Audit logging + hash chain
│   │   ├── notify/      # Email notifications
│   │   ├── middleware/  # Rate limit, CORS, security headers
│   │   ├── config/      # Configuration loader
│   │   └── database/    # PostgreSQL connection + migrations
│   └── pkg/             # Shared utilities (crypto, validator)
│
├── dashboard/           # Next.js 15 + TypeScript + Tailwind
│   └── src/
│       ├── app/         # App Router pages
│       ├── components/  # UI components (shadcn/ui)
│       └── lib/         # API client, crypto, auth utils
│
├── mcp-server/          # Rust MCP server (stdio)
│   └── src/
│       ├── mcp/         # Protocol, tools, resources
│       ├── client/      # Valt API client
│       └── keychain/    # OS keychain integration
│
├── docs/                # Project documentation
├── scripts/             # Dev/ops scripts
└── docker-compose.yml   # Dev environment
```

## Tech Stack
- Go 1.22+ (chi/v5, pgx/v5, golang-jwt/v5)
- Next.js 15 (React 19, TypeScript, shadcn/ui, Tailwind)
- Rust (serde, tokio, reqwest)
- PostgreSQL 16, MinIO, Caddy
