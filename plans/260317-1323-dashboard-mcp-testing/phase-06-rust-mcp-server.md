# Phase 6: Rust MCP Server

## Context Links
- Current scaffold: `mcp-server/src/main.rs` (99 lines, sync stdio loop)
- Cargo.toml: serde 1, serde_json 1, tracing 0.1 (no async runtime, no HTTP client)
- Backend API routes: `server/cmd/server/main.go` lines 115-137
- Workflow handler: `server/internal/workflow/handler.go`
- Credential types: `server/internal/policy/engine.go`
- MCP protocol version: `2024-11-05`

## Overview
- **Priority:** P1
- **Status:** pending
- **Effort:** 8h
- Implement 5 MCP tools + 3 resources over stdio JSON-RPC 2.0
- HTTP client to call Go backend, keychain for credential storage, local AES-256-GCM decryption

## Key Insights
- Current main.rs has working JSON-RPC loop (sync, blocking stdin). Must convert to async for reqwest
- MCP protocol is simple: `initialize`, `tools/list`, `tools/call`, `resources/list`, `resources/read`, `notifications/initialized`
- Only 8 methods total -- hand-rolling is simpler than pulling in an MCP SDK crate
- `keyring` crate provides cross-platform secret storage (macOS Keychain, Windows DPAPI, Linux Secret Service)
- Sampling explicitly disabled: no secret data ever sent to LLM context
- Config: API URL + auth token read from `~/.valt/config.toml` or env vars

## Requirements

### Functional
- **Tools (5):**
  - `request_secret_access` -- POST /secrets/{id}/access-requests
  - `check_approval_status` -- GET /access-requests (filter by ID)
  - `get_credential` -- GET /credentials/{id} + local decrypt
  - `revoke_credential` -- POST /credentials/{id}/revoke
  - `list_my_secrets` -- GET /secrets
- **Resources (3):**
  - `vault://secrets` -- list all user's secrets
  - `vault://requests/{id}` -- access request detail
  - `vault://audit/today` -- today's audit log entries

### Non-Functional
- `cargo clippy -- -D warnings` clean
- No `unsafe`
- Graceful error messages (no panics in tool handlers)
- Sampling disabled (capabilities.sampling not advertised)

## Architecture

### Module Structure
```
mcp-server/src/
  main.rs              # Async stdio loop, dispatcher
  config.rs            # Load ~/.valt/config.toml + env overrides
  client.rs            # HTTP client wrapper (reqwest -> Go backend)
  crypto.rs            # AES-256-GCM decryption (ring or aes-gcm crate)
  keychain.rs          # keyring crate wrapper (store/retrieve auth token)
  protocol.rs          # JSON-RPC types, MCP request/response structs
  tools.rs             # Tool definitions + handlers
  resources.rs         # Resource definitions + handlers
  error.rs             # Error types (thiserror)
```

### Data Flow
```
AI Agent (Claude, etc.)
    | stdio (JSON-RPC 2.0, one JSON object per line)
    v
valt-mcp-server
    | reqwest HTTPS
    v
Go Backend (/api/v1/*)
    |
    v
PostgreSQL + MinIO
```

### Config File (~/.valt/config.toml)
```toml
api_url = "https://localhost:8443/api/v1"
# auth_token stored in OS keychain, not in file
```

### Tool Schemas
```json
{
  "name": "request_secret_access",
  "description": "Request temporary access to a secret. Returns request ID for status checking.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "secret_id": {"type": "string", "description": "UUID of the secret"},
      "reason": {"type": "string", "description": "Why access is needed"},
      "duration_minutes": {"type": "integer", "description": "Requested duration (subject to policy caps)"}
    },
    "required": ["secret_id", "reason"]
  }
}
```

Similar schemas for other 4 tools (see implementation steps for details).

## Related Code Files

### Files to Create
- `mcp-server/src/config.rs`
- `mcp-server/src/client.rs`
- `mcp-server/src/crypto.rs`
- `mcp-server/src/keychain.rs`
- `mcp-server/src/protocol.rs`
- `mcp-server/src/tools.rs`
- `mcp-server/src/resources.rs`
- `mcp-server/src/error.rs`

### Files to Modify
- `mcp-server/Cargo.toml` -- add tokio, reqwest, toml, keyring, aes-gcm, dirs
- `mcp-server/src/main.rs` -- rewrite to async, use module dispatcher

### New Dependencies
```toml
tokio = { version = "1", features = ["full"] }
reqwest = { version = "0.12", features = ["json", "rustls-tls"], default-features = false }
toml = "0.8"
keyring = "3"
aes-gcm = "0.10"
dirs = "6"
thiserror = "2"
anyhow = "1"
base64 = "0.22"
chrono = { version = "0.4", features = ["serde"] }
```

## Implementation Steps

### Step 1: Project Restructure + Dependencies (1h)
1. Update `Cargo.toml` with all new dependencies
2. Create `src/error.rs`:
   - `ValtError` enum with variants: `Config`, `Http`, `Api`, `Crypto`, `Keychain`, `Protocol`
   - Derive `thiserror::Error`
3. Create `src/protocol.rs`:
   - `JsonRpcRequest`, `JsonRpcResponse`, `JsonRpcError` structs
   - `McpCapabilities`, `ServerInfo`, `ToolDef`, `ResourceDef` structs
   - Helper: `success_response(id, result)`, `error_response(id, code, msg)`
4. Convert `main.rs` to `#[tokio::main]` async, keep stdio loop but use `tokio::io::BufReader`
5. `cargo clippy` pass

### Step 2: Config + Keychain (1h)
1. `src/config.rs`:
   - `ValtConfig { api_url: String }`
   - Load from `~/.valt/config.toml` (using `dirs::home_dir()`)
   - Env overrides: `VALT_API_URL`, `VALT_AUTH_TOKEN`
   - If config file missing, create default
2. `src/keychain.rs`:
   - `store_token(token)` -- save to OS keychain via `keyring::Entry`
   - `get_token() -> Option<String>` -- retrieve from keychain
   - `delete_token()` -- remove from keychain
   - Service name: "valt-mcp-server", user: "auth-token"
   - Fallback: read `VALT_AUTH_TOKEN` env var if keychain unavailable

### Step 3: HTTP Client (1h)
1. `src/client.rs`:
   - `ValtClient { http: reqwest::Client, base_url: String, token: String }`
   - `new(config, token)` -- build reqwest client with timeout (30s)
   - Methods matching backend API:
     - `list_secrets(page, limit) -> Vec<Secret>`
     - `create_access_request(secret_id, reason, duration) -> AccessRequest`
     - `get_access_requests(status) -> Vec<AccessRequest>`
     - `get_credential(request_id) -> Credential`
     - `revoke_credential(request_id)`
     - `get_audit_logs(date) -> Vec<AuditLog>`
   - All methods return `Result<T, ValtError>`
   - Handle 401 -> clear keychain + prompt re-auth message

### Step 4: Tools Implementation (2h)
1. `src/tools.rs`:
   - `list_tools() -> Vec<ToolDef>` -- return 5 tool definitions with JSON schemas
   - `call_tool(name, args, client) -> Result<Value>` -- dispatcher
   - `request_secret_access(args, client)`:
     - Parse secret_id, reason, duration_minutes from args
     - Call `client.create_access_request()`
     - Return structured result with request_id and status
   - `check_approval_status(args, client)`:
     - Parse request_id
     - Call `client.get_access_requests()`, filter by ID
     - Return status, timestamps
   - `get_credential(args, client)`:
     - Parse request_id
     - Call `client.get_credential()`
     - If credential_data is encrypted, decrypt locally (see Step 5)
     - Return credential metadata (NOT the raw secret to LLM -- return "credential ready, use vault://")
   - `revoke_credential(args, client)`:
     - Parse request_id
     - Call `client.revoke_credential()`
     - Return confirmation
   - `list_my_secrets(args, client)`:
     - Call `client.list_secrets()`
     - Return list with id, name, type, created_at (no secret values)

### Step 5: Crypto Module (1h)
1. `src/crypto.rs`:
   - `decrypt_aes_gcm(ciphertext, key, nonce) -> Result<Vec<u8>>`
   - Use `aes-gcm` crate with Aes256Gcm
   - Parse encrypted blob format: `nonce (12 bytes) || ciphertext+tag`
   - `unwrap_dek(wrapped_dek, master_key) -> Result<Vec<u8>>`
   - Master key handling: read from keychain or derive from password
   - Note: For MVP, the MCP server receives already-decrypted credential_data from backend. Full E2E encryption where MCP decrypts locally is Phase 2. Stub the decrypt path

### Step 6: Resources Implementation (1h)
1. `src/resources.rs`:
   - `list_resources() -> Vec<ResourceDef>` -- 3 resource URIs
   - `read_resource(uri, client) -> Result<Value>` -- dispatcher
   - `vault://secrets`:
     - Call `client.list_secrets(1, 100)`
     - Return JSON array of {id, name, type, created_at}
   - `vault://requests/{id}`:
     - Parse request_id from URI
     - Call `client.get_access_requests()`, filter by ID
     - Return full request detail
   - `vault://audit/today`:
     - Call `client.get_audit_logs(today)`
     - Return JSON array of log entries

### Step 7: Main Loop Integration (1h)
1. Rewrite `main.rs`:
   - Load config, get token from keychain
   - Build `ValtClient`
   - Async stdin line reader
   - Dispatch: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`, `resources/list`, `resources/read`
   - Initialize response: advertise tools+resources capabilities, no sampling
   - Handle unknown methods with -32601
   - Handle parse errors with -32700
2. `cargo build && cargo clippy -- -D warnings`

## Todo List
- [ ] Add async deps (tokio, reqwest, etc.) to Cargo.toml
- [ ] Error types with thiserror
- [ ] JSON-RPC protocol types
- [ ] Config loading (~/.valt/config.toml)
- [ ] Keychain integration (store/retrieve auth token)
- [ ] HTTP client wrapping all backend endpoints
- [ ] 5 tool definitions + handlers
- [ ] 3 resource definitions + handlers
- [ ] AES-256-GCM decrypt stub
- [ ] Async main loop with dispatcher
- [ ] `cargo clippy -- -D warnings` clean
- [ ] `cargo build --release` succeeds
- [ ] Manual test: initialize -> tools/list -> list_my_secrets

## Success Criteria
- `echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | cargo run` returns valid capabilities
- `tools/list` returns 5 tools with correct schemas
- `resources/list` returns 3 resources
- `tools/call` with `list_my_secrets` returns secrets from backend (with running server)
- Full flow: request_secret_access -> check_approval_status -> get_credential works
- `cargo clippy -- -D warnings` passes
- No `unsafe` blocks

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| keyring crate fails on CI/headless Linux | Medium | Fallback to env var `VALT_AUTH_TOKEN`; skip keychain in CI |
| reqwest + rustls build issues on Windows | Low | Use `native-tls` feature as fallback |
| MCP protocol spec changes | Low | Pin to version 2024-11-05; only implement what's needed |
| Credential decryption key management | Medium | MVP: backend returns decrypted cred data; full E2E crypto in Phase 2 |

## Security Considerations
- Auth token stored in OS keychain, never in plaintext config files
- No secret values included in tool responses to LLM (sampling disabled)
- HTTPS required for backend communication (configurable TLS verification for dev)
- Credential data auto-cleared from memory after use (zeroize where possible)
- stdin/stdout only -- no network listeners, no attack surface

## Next Steps
- Phase 7 adds `cargo test` suite for protocol parsing, tool dispatch, crypto
- Future: MCP notifications for real-time approval status updates
- Future: Multi-user support (switch between keychain entries)
