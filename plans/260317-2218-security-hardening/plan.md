---
title: "Security Hardening — Authorization & Encryption Fixes"
description: "Fix 4 critical and 5 high-priority security issues from code review (score 7.0/10)"
status: completed
priority: P0
effort: 3h
branch: master
tags: [security, authorization, encryption, rbac]
created: 2026-03-17
completed: 2026-03-24
---

# Security Hardening Plan

## Summary
Code review scored 7.0/10. 4 critical issues (plaintext secrets, missing auth on 9 routes, global usage counts) and 5 high issues (restricted GetRequest, unused rate limiter, swallowed errors, MCP path traversal, discarded rejection reason).

## Phases

| # | Phase | Issues | Effort | Status |
|---|-------|--------|--------|--------|
| 1 | [Encrypt DynSecret Storage](phase-01-encrypt-dynsecret-storage.md) | C1 | 40min | pending |
| 2 | [Scanner Authorization](phase-02-scanner-authorization.md) | C2 | 30min | pending |
| 3 | [DynSecret Authorization](phase-03-dynsecret-authorization.md) | C3 | 30min | pending |
| 4 | [Usage Org-Scoped Counts](phase-04-usage-org-scoped-counts.md) | C4 | 20min | pending |
| 5 | [High-Priority Fixes](phase-05-high-priority-fixes.md) | H1-H5 | 60min | pending |

## Dependencies
- `crypto.EncryptAES256GCM` / `DecryptAES256GCM` in `server/pkg/crypto/aes.go` — ready
- `rbac.Middleware` in `server/internal/rbac/middleware.go` — ready
- `config.MasterKey()` — already loaded in `main.go`
- Latest migration: `000023` -> new migration will be `000024`

## Medium Issues (Deferred)
M1-M8 noted but not implemented this cycle. Track in next review.

## Execution Order
Phases 1-4 are independent (parallelizable). Phase 5 has no deps on 1-4.
Recommended: sequential 1 -> 2 -> 3 -> 4 -> 5 for cleaner commits.
