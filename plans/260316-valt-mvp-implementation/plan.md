# Valt - AI Secret Vault SaaS MVP Implementation Plan

## Status: IN_PROGRESS
## Started: 2026-03-16

## Phases

| Phase | Name | Status | Priority | Depends |
|-------|------|--------|----------|---------|
| 1 | Project Scaffolding & Infrastructure | DONE (2026-03-16) | P0 | - |
| 2 | Database Layer & Migrations | Planned | P0 | Phase 1 |
| 3 | Backend Core - Auth + Vault + Middleware | Planned | P0 | Phase 2 |
| 4 | Backend - Workflow + Audit + Notify | Planned | P0 | Phase 3 |
| 5 | Next.js Dashboard | Planned | P1 | Phase 3+4 |
| 6 | Rust MCP Server | Planned | P1 | Phase 3+4 |
| 7 | Testing & Hardening | Planned | P0 | Phase 3-6 |

## Dependency Graph
```
Phase 1 -> Phase 2 -> Phase 3 -> Phase 4 -> Phase 5 (parallel)
                                          -> Phase 6 (parallel)
                                 Phase 3-6 -> Phase 7
```

## Spec Reference
- `./idea-saas-valt.md`
