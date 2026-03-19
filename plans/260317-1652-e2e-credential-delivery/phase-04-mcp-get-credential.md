# Phase 4: MCP get_credential

## Context Links
- [plan.md](./plan.md)
- [phase-03](./phase-03-credential-delivery.md)
- `mcp-server/src/tools.rs` -- tool_get_credential (lines 130-143)
- `mcp-server/src/client.rs` -- ValtClient::get_credential
- `mcp-server/src/crypto.rs` -- has decrypt_aes_gcm (unused)

## Overview
- **Priority:** P0
- **Status:** pending
- **Description:** MCP `get_credential` tool reads `value` from server response, returns it to AI agent. No client-side decryption needed (server delivers plaintext).

## Key Insights
- `tool_get_credential` currently returns only metadata: `credential_id`, `request_id`, `issued_at`, `expires_at`, `status`. The `value` field from server response is ignored.
- `client.rs` uses `serde_json::Value` (dynamic JSON) -- no struct update needed. `cred.get("value")` just works once server returns it.
- `crypto.rs` has `decrypt_aes_gcm` but it is unused (`#[allow(dead_code)]`). In server-side encryption model, MCP client does NOT decrypt. Keep it for potential future client-side mode.
- Tool description says "Retrieve an approved credential" but message says "Access data via vault://requests/{request_id} resource." -- misleading since value should be returned directly.

## Requirements

### Functional
- `get_credential` tool returns `value` field when present in server response
- If `value` absent (expired session, not yet approved), return status only
- Update tool result message to reflect that value is directly returned

### Non-Functional
- Do not log the credential value in MCP server

## Architecture

```
AI Agent calls get_credential(request_id)
         |
         v
   [tools.rs] tool_get_credential
         |
         v
   [client.rs] GET /credentials/{request_id}
         |
         v
   Server returns { id, status, value, ... }
         |
         v
   Return to agent: { credential_id, status, value, expires_at }
```

## Related Code Files

### Modify
- `mcp-server/src/tools.rs` -- `tool_get_credential` function

### No changes needed
- `mcp-server/src/client.rs` -- already returns `serde_json::Value`
- `mcp-server/src/crypto.rs` -- not used in server-side model

## Implementation Steps

1. Modify `tool_get_credential` in `tools.rs` (lines 130-143):
   ```rust
   async fn tool_get_credential(args: &Value, client: &ValtClient) -> Result<Value> {
       let request_id = args["request_id"].as_str()
           .ok_or_else(|| crate::error::ValtError::Protocol("request_id required".into()))?;
       let cred = client.get_credential(request_id).await?;

       let has_value = cred.get("value")
           .and_then(|v| v.as_str())
           .map(|s| !s.is_empty())
           .unwrap_or(false);

       if has_value {
           Ok(json!({
               "credential_id": cred.get("id"),
               "request_id": request_id,
               "status": cred.get("status"),
               "value": cred.get("value"),
               "expires_at": cred.get("expires_at"),
               "message": "Credential value retrieved successfully. Handle securely."
           }))
       } else {
           Ok(json!({
               "credential_id": cred.get("id"),
               "request_id": request_id,
               "status": cred.get("status"),
               "expires_at": cred.get("expires_at"),
               "message": "No credential value available. Check status -- may be expired or not yet approved."
           }))
       }
   }
   ```

2. Update the `get_credential` tool description in `list_tools()`:
   - Change to: `"Retrieve an approved credential's value. Returns the secret value if the session is active."`

3. Verify `cargo check` passes.

## Todo List
- [ ] Update `tool_get_credential` to include `value` in response
- [ ] Update tool description in `list_tools()`
- [ ] Remove misleading "vault://requests/" resource URI message
- [ ] `cargo check` passes
- [ ] `cargo test` passes

## Success Criteria
- With an active credential session that has a value:
  - MCP tool returns JSON with `"value": "sk-123"` (the actual credential)
- With expired/revoked session:
  - MCP tool returns JSON with status info, no value, helpful message
- `cargo clippy` clean

## Risk Assessment
- **Value exposure in MCP logs**: AI agent runtimes may log tool responses. Document that credential values appear in tool output. This is by design -- the agent needs the value to use the credential.

## Security Considerations
- MCP tool output goes to AI agent -- this IS the intended delivery mechanism
- Do not add extra logging of the `value` field in tools.rs
- `crypto.rs` remains available for future client-side encryption mode

## Next Steps
- E2E test: create secret via dashboard -> request access via MCP -> approve in dashboard -> get_credential via MCP -> verify value returned
