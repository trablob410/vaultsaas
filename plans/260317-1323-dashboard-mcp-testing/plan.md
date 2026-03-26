---
title: "Dashboard + MCP Server + Testing"
description: "Next.js dashboard, Rust MCP server, and comprehensive test suite for Valt MVP"
status: completed
priority: P1
effort: 28h
branch: master
tags: [dashboard, mcp, testing, nextjs, rust]
created: 2026-03-17
---

# Valt Phases 5-7: Dashboard, MCP Server, Testing

## Phases

| Phase | Name | Status | Effort | Depends |
|-------|------|--------|--------|---------|
| 5 | [Next.js Dashboard](./phase-05-nextjs-dashboard.md) | complete | 12h | Phase 4 |
| 6 | [Rust MCP Server](./phase-06-rust-mcp-server.md) | complete | 8h | Phase 4 |
| 7 | [Testing & Hardening](./phase-07-testing-hardening.md) | complete | 8h | Phase 5+6 |

## Execution Strategy

```
Phase 5 (dashboard) ──────┐
                           ├──> Phase 7 (testing)
Phase 6 (mcp-server) ─────┘
```

- Phase 5 and 6 run in parallel (zero file overlap)
- Phase 7 Go unit tests can start immediately (no dashboard/MCP dependency)
- Phase 7 E2E tests require Phase 5+6 complete

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Dashboard auth | BFF proxy (`app/api/auth/`) with httpOnly cookies | JWT never exposed to browser JS |
| shadcn/ui init | `npx shadcn@latest init` with Tailwind v4 CSS vars | Native dark mode via CSS custom properties |
| Client crypto | Web Crypto API (SubtleCrypto) | No extra deps, browser-native AES-256-GCM |
| MCP async runtime | tokio (multi-thread) | De facto standard, reqwest requires it |
| MCP HTTP client | reqwest | Mature, tokio-native, TLS built-in |
| MCP keychain | keyring crate | Cross-platform (macOS/Windows/Linux) |
| MCP protocol | Hand-roll JSON-RPC | Only 8 methods needed, no MCP SDK dependency |
| Go integration tests | testcontainers-go | Real Postgres+MinIO per test, no shared state |
| Dashboard tests | vitest + @testing-library/react | Fast, RSC-compatible |

## Backend API Reference (all implemented)

Auth: `POST /auth/register`, `/login`, `/refresh`
Secrets: `GET/POST /secrets`, `GET/PUT/DELETE /secrets/{id}`
Workflow: `POST /secrets/{id}/access-requests`, `GET /access-requests`, `POST /access-requests/{id}/approve|reject`
Credentials: `GET /credentials/{id}`, `POST /credentials/{id}/revoke`
Audit: `GET /audit/logs`
Consent: `POST /consent`
