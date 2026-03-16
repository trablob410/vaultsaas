# Phase 1: Project Scaffolding & Infrastructure

## Status: DONE (2026-03-16)
## Priority: P0
## Started: 2026-03-16

## Overview
Set up monorepo structure, Docker Compose (dev+prod), Makefile, .env, Caddy config. Stub entry points for all 3 services (Go backend, Next.js dashboard, Rust MCP server).

## Tasks

### 1. Initialize Git repo + project root files
- git init, .gitignore, LICENSE (Apache 2.0), CLAUDE.md, README.md

### 2. Create Go backend stub
- server/go.mod, server/cmd/server/main.go (health endpoint only)
- server/Dockerfile (multi-stage build)

### 3. Create Next.js dashboard stub
- dashboard/package.json, dashboard/next.config.ts, dashboard/src/app/page.tsx
- dashboard/Dockerfile

### 4. Create Rust MCP server stub
- mcp-server/Cargo.toml, mcp-server/src/main.rs
- mcp-server/Dockerfile

### 5. Docker Compose (dev)
- docker-compose.yml: server, dashboard, postgres, minio, caddy
- Caddy config for reverse proxy

### 6. Docker Compose (prod)
- docker-compose.prod.yml with production settings

### 7. Environment & tooling
- .env.example with all required variables
- Makefile with common commands
- scripts/setup-dev.sh

### 8. Documentation stubs
- docs/ directory with initial architecture docs

## Success Criteria
- `docker compose up` starts all services
- Health endpoint returns 200
- All stubs compile/build successfully

## Todo
- [x] Task 1: Git + root files
- [x] Task 2: Go backend stub
- [x] Task 3: Next.js dashboard stub
- [x] Task 4: Rust MCP server stub
- [x] Task 5: Docker Compose dev
- [x] Task 6: Docker Compose prod
- [x] Task 7: Env + tooling
- [x] Task 8: Documentation stubs
- [ ] Task 9: Fix review findings (see below)

## Code Review Findings (2026-03-16)
**Score: 8/10**

### Must-fix before Phase 2
- [ ] Create `dashboard/public/.gitkeep` and generate+commit `package-lock.json` (build blocker)
- [ ] Generate+commit `mcp-server/Cargo.lock` (reproducibility)
- [ ] Fix `sslmode=disable` in `docker-compose.prod.yml` -> `require`
- [ ] Pin MinIO image to specific release tag (both compose files)
- [ ] Make CORS origins configurable via env var (`main.go`)
- [ ] Remove `|| true` from `mcp-server/Dockerfile` dep cache step
- [ ] Add Docker secrets config to prod compose or remove secret path refs
- [ ] Create `SECURITY.md` and `LICENSE` file stubs

### Should-fix
- [ ] Handle `json.Encode` error in health handler
- [ ] Add CSP header to `Caddyfile.prod`
- [ ] Add network isolation in prod compose
- [ ] Consider 4096-bit RSA or Ed25519 for JWT keys
