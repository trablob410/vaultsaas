---
title: "Agent cross-user access request"
description: "Enable AI agents to create access requests for secrets they don't own"
status: completed
priority: P0
effort: 6h
branch: master
tags: [workflow, agent-auth, security, p0-gap]
created: 2026-03-18
completed: 2026-03-24
---

# Agent Cross-User Access Request

## Problem

Two bugs block AI agents from requesting secrets owned by other users:
1. `CreateRequest` calls `GetSecret(userID, secretID)` with owner filter -- agents aren't owners
2. `agent.AuthMiddleware` is never wired to workflow routes -- agents can't auth at all

## Sub-problems discovered during research

- `Secret` struct missing `ProjectID` field; `GetSecretByID` SQL doesn't select `project_id`
- `requester_user_id` column is `NOT NULL REFERENCES users(id)` -- agent-only requests need nullable
- MCP client doesn't send `requester_type: "ai_agent"` in request body

## Phases

| # | Phase | Status | Effort |
|---|-------|--------|--------|
| 1 | [Wire agent auth middleware](./phase-01-wire-agent-auth.md) | done | 1.5h |
| 2 | [Fix CreateRequest handler](./phase-02-fix-create-request.md) | done | 2.5h |
| 3 | [MCP client fixes](./phase-03-mcp-validation.md) | done | 1h |
| 4 | [Tests](./phase-04-tests.md) | done | 1h |

## Dependencies

- Phase 2 depends on Phase 1 (dual-auth middleware must exist first)
- Phase 3 independent (Rust side)
- Phase 4 depends on Phases 1+2

## Key decisions

- **Dual-auth middleware** over separate route groups (avoids route duplication)
- **DB migration** to make `requester_user_id` nullable (vs. creating sentinel user for agents)
- **Project-scoped guard** using `agent_identities.project_id == secret.project_id` for agents (no `project_memberships` row for agents)

## Research reports

- [Researcher 1: CreateRequest flow](./research/researcher-1-report.md)
- [Researcher 2: RBAC & agent identity](./research/researcher-2-report.md)
