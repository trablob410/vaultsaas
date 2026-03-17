---
phase: 3
title: "MCP server client fixes"
status: pending
priority: P0
created: 2026-03-18
---

# Phase 3: MCP Server Validation & Fixes

## Context

- [mcp-server/src/client.rs](../../mcp-server/src/client.rs) -- HTTP client, `create_access_request`
- [mcp-server/src/tools.rs](../../mcp-server/src/tools.rs) -- `tool_request_access`

## Key Insights

- `client.create_access_request` sends `{ reason, duration_minutes }` -- **missing `requester_type: "ai_agent"`**
- Backend defaults `requester_type` to `"human"` if absent (handler.go:65-67)
- Auth header `Bearer <agent_token>` is correct -- `bearer_auth(&self.token)` in client.rs
- Agent token priority in `get_auth_token()` is correct (agent token > user JWT)

## Requirements

### Functional
- MCP client must send `requester_type: "ai_agent"` in access request body
- Optionally: send `ai_agent_id` if available (currently not tracked client-side)

### Non-functional
- No breaking change for existing MCP tool interface

## Related Code Files

**Modify:**
- `mcp-server/src/client.rs` -- add `requester_type` to `create_access_request` JSON body

**Read only:**
- `mcp-server/src/tools.rs` -- verify tool invocation is correct

## Implementation Steps

1. In `client.rs:create_access_request`, update JSON body:
   ```rust
   serde_json::json!({
       "requester_type": "ai_agent",
       "reason": reason,
       "duration_minutes": duration_minutes
   })
   ```

2. Verify `tool_request_access` in `tools.rs` passes correct args -- confirmed OK, no changes needed.

3. Run `cargo clippy` and `cargo test` to verify no regressions.

## Todo

- [ ] Add `requester_type: "ai_agent"` to `create_access_request` body
- [ ] `cargo clippy -- -D warnings`
- [ ] `cargo test`

## Success Criteria

- MCP `request_secret_access` tool sends `requester_type: "ai_agent"` in POST body
- `cargo clippy` clean
- Existing tests pass

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| None significant | Single-line change, additive field |

## Security Considerations

- Agent tokens already handled securely via keychain/env var
- No secret values transmitted in this request
