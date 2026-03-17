---
phase: 4
title: "Tests"
status: pending
priority: P0
created: 2026-03-18
---

# Phase 4: Tests

## Context

- [Existing test patterns](../../server/internal/) -- check for existing test files
- Phases 1-2 introduce dual-auth middleware + revised CreateRequest handler

## Key Insights

- Go unit tests should use mock interfaces (AgentValidator, AgentGetter)
- Integration tests need Docker (Postgres) -- verify with `make test-integration`
- MCP Rust tests in `mcp-server/` -- `cargo test`

## Requirements

### Functional
- Cover all authorization paths in CreateRequest
- Cover dual-auth middleware (JWT success, agent token success, both fail)

### Non-functional
- Tests runnable via `make test-unit` (no Docker dependency for unit tests)

## Related Code Files

**Create:**
- `server/internal/auth/dual_auth_middleware_test.go`
- `server/internal/workflow/handler_test.go` (or extend existing)

## Implementation Steps

### 1. Dual-auth middleware unit tests (`auth/dual_auth_middleware_test.go`)

| Test case | Input | Expected |
|-----------|-------|----------|
| Valid JWT | Bearer <jwt> | 200, userID in ctx |
| Valid agent token | Bearer <agent-token> | 200, agentID in ctx |
| Invalid token (both fail) | Bearer <garbage> | 401 |
| No Authorization header | (none) | 401 |

Mock `AgentValidator` interface -- return `*Token` or `nil`.
Mock `JWTManager.ValidateAccessToken` -- return userID or error.

### 2. CreateRequest handler unit tests (`workflow/handler_test.go`)

| Test case | Caller | Secret | Expected |
|-----------|--------|--------|----------|
| Agent, same project | agentID (project=P1) | secret (project=P1) | 201 |
| Agent, different project | agentID (project=P1) | secret (project=P2) | 403 |
| Agent, legacy secret (no project) | agentID | secret (project=nil) | 403 |
| User, project member | userID (member of P1) | secret (project=P1) | 201 |
| User, non-member | userID (not in P1) | secret (project=P1) | 403 |
| User, owner, legacy secret | userID=ownerID | secret (project=nil) | 201 |
| User, non-owner, legacy | userID!=ownerID | secret (project=nil) | 403 |
| No identity (both empty) | - | - | 401 |

Use `httptest.NewRecorder` + mock services.

### 3. Service-level daily limit tests

| Test case | Expected |
|-----------|----------|
| Agent daily limit by `ai_agent_id` | Rejects after N requests/day |
| User daily limit by `requester_user_id` | Existing behavior preserved |

### 4. Rust MCP client test

Verify `create_access_request` JSON body includes `requester_type: "ai_agent"`.

## Todo

- [ ] Write dual-auth middleware tests
- [ ] Write CreateRequest handler tests (7+ cases)
- [ ] Write daily-limit tests for agent path
- [ ] Run `make test-unit`
- [ ] Run `cargo test`
- [ ] All tests pass

## Success Criteria

- All test cases above pass
- `make test` green
- No regressions in existing tests

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Handler tests need many mocks | Define minimal interfaces, mock only what's needed |
| DB-dependent tests | Unit tests use mocks; integration tests use Docker Postgres |
