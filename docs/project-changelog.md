# Project Changelog

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
