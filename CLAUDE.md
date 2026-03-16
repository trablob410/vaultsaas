# Valt - Project Instructions

## Project Overview
Valt is an MCP-native secret vault with human-in-the-loop approval for AI agents. Monorepo with Go backend, Next.js dashboard, Rust MCP server.

## Architecture
- **server/**: Go 1.22+ monolith (chi/v5 router, pgx/v5, JWT RS256 auth)
- **dashboard/**: Next.js 15 + TypeScript + shadcn/ui + Tailwind
- **mcp-server/**: Rust MCP server (stdio JSON-RPC 2.0)
- **PostgreSQL 16**: Metadata + audit logs (partitioned)
- **MinIO**: Encrypted blob storage (S3-compatible)
- **Caddy**: Reverse proxy + auto TLS

## Key Patterns
- Zero-knowledge encryption: client-side AES-256-GCM, envelope encryption (DEK wrapped by master key)
- JWT RS256 with 15min access / 7day refresh tokens
- Argon2id password hashing
- Approval state machine: pending -> approved/rejected -> active -> expired/revoked
- Audit log hash chain (SHA-256)

## Commands
```bash
make dev            # Start dev environment
make test           # Run all tests
make lint           # Run linters
make build          # Build all services
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make seed           # Seed test data
```

## Code Standards
- Go: Follow standard project layout, use internal/ for private packages
- TypeScript: Strict mode, no any
- Rust: cargo clippy clean, no unsafe unless justified
- All: kebab-case file names, files under 200 lines
- See `docs/code-standards.md` for details
