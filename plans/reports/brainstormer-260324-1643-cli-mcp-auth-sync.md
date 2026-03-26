# Brainstorm: CLI and MCP Server Authentication and Sync

**Date:** 2026-03-24
**Status:** Complete

---

## 1. Current State Assessment

### CLI (valt-cli/ -- Go, Cobra)

**Auth flow already implemented via valt setup:**
1. User enters API URL (default: https://api.valt.dev)
2. CLI calls GET /auth/cli-start -> gets session_id + login_url
3. Opens browser to login_url (Google OAuth with cli_session param)
4. Polls GET /auth/cli-poll?session={id} every 2s for up to 3 min
5. On OAuth success, server stores access JWT in cli_auth_sessions table
6. CLI receives JWT, stores in OS keychain (valt-cli/auth-token via go-keyring)
7. User picks org -> workspace -> project, saved to ~/.valt/config.toml
8. Final output suggests: valt mcp install --ide claude

**Existing commands:**

| Command | Status | What it does |
|---------|--------|-------------|
| setup | Working | Browser OAuth -> keychain token + project selection |
| list | Working | Lists secrets in selected project |
| get NAME | Working | Gets secret value; auto-creates access request if none |
| request NAME | Working | Creates access request with reason + duration |
| status ID | Working | Checks approval status |
| run -- CMD | Working | Injects active credentials as env vars |
| mcp install | Working | Writes MCP server config to IDE (Claude/Cursor) |
| completion | Listed but no implementation found | |

**Config storage:**
- Token: OS keychain (service=valt-cli, user=auth-token)
- Config: ~/.valt/config.toml (api_url, project_id)
- Env overrides: VALT_AUTH_TOKEN, VALT_API_URL, VALT_PROJECT_ID

### MCP Server (mcp-server/ -- Rust)

**Auth token resolution (priority order):**
1. VALT_AGENT_TOKEN env var
2. OS keychain valt-agent / valt-agent-token
3. VALT_TOKEN env var (legacy user JWT)
4. OS keychain valt-mcp-server / auth-token (falls back to VALT_AUTH_TOKEN env)

**Config:** ~/.valt/config.toml (same path as CLI -- api_url, project_id)

**Key observation:** MCP server prefers agent tokens but falls back to user tokens. Both CLI and MCP read from the same config file.

### Server-Side Auth

**Token types:**
- **User JWT** (typ: access) -- RS256, 15-min expiry, contains sub: userID
- **Agent token** -- opaque, stored as SHA-256 hash in agent_tokens table, scoped per-agent
- **TOTP token** -- short-lived JWT for 2FA challenge
- **Refresh token** -- 64-char hex, stored server-side

**CLI session flow:**
- cli_auth_sessions table: UUID PK, nullable token TEXT, 5-min expiry
- cli-start: inserts row, returns session_id + login_url
- Google OAuth callback: if cli_session cookie present, writes JWT to session row
- cli-poll: returns token if populated, deletes row after retrieval

---

## 2. Gap Analysis

### Critical Issues

| Gap | Impact | Severity |
|-----|--------|----------|
| **15-min JWT expiry, no refresh in CLI** | CLI token expires fast; user must re-run valt setup frequently | HIGH |
| **No valt logout** | Cannot clear keychain token | LOW |
| **No valt whoami** | Cannot verify current auth state | LOW |
| **MCP and CLI use different keychain entries** | valt setup stores to valt-cli/auth-token; MCP reads from valt-mcp-server/auth-token or valt-agent -- they do not share tokens | HIGH |
| **No token refresh mechanism in CLI** | Only server-side cookie refresh exists (dashboard proxy) | HIGH |
| **No device code flow fallback** | SSH/headless users cannot authenticate | MEDIUM |
| **valt mcp install does not pass token** | MCP server must find its own token independently | MEDIUM |
| **No org/workspace context switching** | Once set in config, cannot easily switch | LOW |

### Token Sharing Problem (The Big One)

CLI stores: keychain("valt-cli", "auth-token")
MCP reads: keychain("valt-mcp-server", "auth-token") OR keychain("valt-agent", "valt-agent-token")

**These are different keychain entries.** After valt setup, the MCP server has NO token unless the user manually sets VALT_AUTH_TOKEN env var. The valt mcp install command only writes the binary path to IDE config -- it does NOT propagate the auth token.

---

## 3. Competitor Analysis

### GitHub CLI (gh auth login)
- Browser OAuth with device code flow (user enters code on github.com)
- Polls for completion; 15-min window
- Token stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Supports --with-token for piped PAT input (CI/headless)
- gh auth status shows current auth state
- gh auth token prints token to stdout (for piping to other tools)

### Doppler CLI (doppler login)
- Opens browser, copies auth code to clipboard automatically
- Token stored in OS keychain
- doppler run -- CMD injects secrets as env vars (same as valt run)
- Service tokens for CI (non-interactive)
- Project/config scoping via doppler setup (similar to valt setup)

### HashiCorp Vault CLI (vault login)
- Multiple auth methods: token, OIDC, userpass, LDAP, GitHub, etc.
- vault login -method=oidc opens browser
- Token cached in ~/.vault-token (file, not keychain)
- Token can be passed via VAULT_TOKEN env var

### Key Patterns Across Competitors

1. **Single auth command** that handles browser + fallback (PAT/token paste)
2. **Token sharing** -- one auth session used by CLI + integrations
3. **whoami/status** command to verify auth state
4. **logout** command to clean up
5. **Token refresh** handled transparently
6. **CI mode** -- env var or --with-token flag for non-interactive

---

## 4. Recommended Auth Flow

### valt setup (Enhanced -- Backward Compatible)

```
$ valt setup

? Valt API URL [https://api.valt.dev]:

? Authentication method:
  > Login with browser (recommended)
    Paste a token

Opening browser for login...
Waiting for login... Done. Logged in as user@example.com

? Select organization:
  > Acme Corp
    Personal

? Select project:
  > production-api
    staging-api

Config saved to ~/.valt/config.toml
Auth token stored in system keychain

Next steps:
  valt list              # view your secrets
  valt mcp install       # set up MCP for your IDE
```

### Token Architecture (Recommended)

**Single shared token location for CLI + MCP:**

```
Keychain: "valt" / "auth-token"  <-- BOTH CLI and MCP read from here
Config:   ~/.valt/config.toml    <-- already shared
```

CLI valt setup stores the user JWT to the shared keychain entry. MCP server reads from the same entry. No extra setup needed.

**Priority chain (both CLI and MCP):**
1. VALT_AGENT_TOKEN env var (agent token for CI/production)
2. VALT_AUTH_TOKEN env var (user token for CI)
3. OS keychain "valt" / "auth-token" (user token from valt setup)

### Token Refresh Strategy

The current 15-min JWT is too short for CLI use. Two options:

**Option A: Long-lived CLI token (Recommended -- KISS)**
- Server issues a separate typ: "cli" token with 30-day expiry when completing CLI auth
- Token revocable via dashboard (Active Sessions page)
- No refresh logic in CLI needed
- Same pattern as gh (GitHub PATs last until revoked)

**Option B: Transparent refresh in CLI**
- Store refresh token alongside access token in keychain
- CLI auto-refreshes before each request
- More complex; requires refresh endpoint accessible without cookies

**Recommendation:** Option A. A 30-day CLI token is standard practice. The security model is: token in OS keychain = as secure as the machine. Revocation via dashboard covers the lost device scenario.

### Headless / CI Authentication

```bash
# Option 1: Environment variable
export VALT_AUTH_TOKEN=<token>
valt list

# Option 2: Pipe token
echo "<token>" | valt setup --with-token

# Option 3: Agent token (for MCP in production)
export VALT_AGENT_TOKEN=<agent-token>
```

---

## 5. Feature Parity Table

| Capability | Dashboard | CLI | MCP Server | Notes |
|-----------|-----------|-----|------------|-------|
| Login (Google OAuth) | Yes | Yes | N/A (uses CLI token) | |
| View secrets | Yes | Yes (list) | Yes (list_secrets tool) | |
| Get secret value | Yes | Yes (get) | Yes (get_credential tool) | |
| Create access request | Yes | Yes (request) | Yes (create_access_request) | |
| Check request status | Yes | Yes (status) | Yes (get_access_requests) | |
| Approve/reject request | Yes | **No** | **No** | Dashboard-only is fine (human approval) |
| Create secret | Yes | **No** | Yes (create_secret tool) | CLI should add this |
| Delete secret | Yes | **No** | **No** | Dashboard-only is fine |
| Inject secrets as env | N/A | Yes (run) | N/A | CLI-unique feature |
| Secret scanning | N/A | **No** | Yes (scan_project tool) | |
| Dynamic secrets | Yes | **No** | Yes (create_lease tool) | |
| Audit logs | Yes | **No** | Yes (get_audit_logs tool) | |
| Org/project management | Yes | **No** | **No** | Dashboard-only is fine |
| MCP config install | N/A | Yes (mcp install) | N/A | CLI-unique feature |
| Switch context | Yes (sidebar) | **No** | **No** | CLI needs valt switch |
| Auth status | Implicit | **No** | **No** | Need valt whoami |
| Logout | Implicit | **No** | **No** | Need valt logout |

### Recommended CLI Additions (Priority Order)

1. **valt whoami** -- show current user, org, project, token expiry
2. **valt logout** -- clear keychain + config
3. **valt switch** -- change active org/project without re-auth
4. **valt create NAME --type TYPE** -- create a secret from CLI
5. **Shell completion** (valt completion bash/zsh/fish/powershell)

---

## 6. Architecture: Shared Auth/Data Layer

```
+---------------+     +----------------+     +-----------------+
|   Dashboard   |     |    Valt CLI    |     |   MCP Server    |
|   (Next.js)   |     |     (Go)       |     |    (Rust)       |
+-------+-------+     +-------+--------+     +-------+---------+
        |                      |                      |
        | cookie-based         | Bearer JWT            | Bearer JWT
        | (proxy->Bearer)      | (keychain)            | (keychain)
        |                      |                      |
        +----------------------+----------------------+
                               |
                    +----------v----------+
                    |    Go API Server    |
                    |    (chi router)     |
                    |                     |
                    |  /auth/*            |
                    |  /secrets/*         |
                    |  /access-requests/* |
                    |  /credentials/*     |
                    +----------+----------+
                               |
                    +----------v----------+
                    |  PostgreSQL + MinIO  |
                    +---------------------+
```

**Key insight:** CLI, MCP, and Dashboard already share the same backend API. The only gap is token sharing between CLI and MCP on the local machine.

### Shared Keychain Contract

```
OS Keychain:
  "valt" / "auth-token"              <-- user JWT (written by valt setup)
                                          read by: valt CLI + valt-mcp-server

  "valt-agent" / "valt-agent-token"  <-- agent token (optional)
                                          written by: manual / dashboard copy-paste
                                          read by: valt-mcp-server (priority 1)

~/.valt/config.toml:
  api_url = "https://api.valt.dev"
  project_id = "uuid"
```

---

## 7. Implementation Plan

### Phase 1: Fix Token Sharing (HIGH priority, ~2h)

| Task | Effort | Files |
|------|--------|-------|
| Unify keychain service name to "valt" in both CLI and MCP | 30min | valt-cli/internal/keychain/keychain.go, mcp-server/src/keychain.rs |
| Add typ: "cli" long-lived token (30d) to JWTManager | 30min | server/internal/auth/jwt.go |
| Update cli_session.go to issue CLI token instead of access token | 15min | server/internal/auth/cli_session.go, oauth.go |
| Update ValidateAccessToken to accept both access and cli types | 15min | server/internal/auth/jwt.go |
| Migration: check old keychain name, move to new, delete old | 15min | valt-cli/internal/keychain/keychain.go |
| Test: valt setup -> valt list -> MCP server reads same token | 30min | Integration test |

### Phase 2: Missing CLI Commands (~2h)

| Task | Effort | Files |
|------|--------|-------|
| valt whoami -- GET /auth/me, print user + org + project | 30min | valt-cli/cmd/whoami.go |
| valt logout -- delete keychain + optionally delete config | 20min | valt-cli/cmd/logout.go |
| valt switch -- re-pick org/project without re-auth | 30min | valt-cli/cmd/switch.go |
| --with-token flag on setup for CI | 20min | valt-cli/cmd/setup.go |
| Shell completions (cobra built-in) | 20min | valt-cli/cmd/root.go |

### Phase 3: Polish (~1h)

| Task | Effort | Files |
|------|--------|-------|
| Error message when token expired (run valt setup to re-authenticate) | 15min | valt-cli/cmd/root.go |
| valt mcp install injects env hint if token not found | 15min | valt-cli/cmd/mcp.go |
| Add /auth/me endpoint if not exists | 30min | server/internal/auth/handler.go |

### Phase 4: Optional Enhancements (Future)

- Device code flow (for SSH/headless -- like gh)
- valt create command
- Token revocation from dashboard (Active CLI Sessions)
- MCP server auto-refresh (detect 401 -> prompt user)

---

## 8. Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Keychain name change breaks existing users | Medium | Migration logic: check old name, move to new, delete old |
| Long-lived CLI token security concern | Low | 30-day is industry standard; revocable via dashboard; keychain = device-level security |
| Windows keychain compatibility | Low | Both go-keyring and keyring crate already support Windows Credential Manager |
| Token shared between CLI and MCP -- logout affects both | Medium | Expected behavior; document it; valt logout warns user |

---

## 9. Unresolved Questions

1. **Should cli token type be separate JWT type or just a longer-lived access token?** Separate type is cleaner for auditing and revocation, but adds validation complexity.
2. **Does the server have a /auth/me endpoint already?** If not, it is needed for valt whoami.
3. **Should valt setup create an agent token in addition to user token?** This would let MCP server use a proper agent token with scoped permissions, but adds complexity to the setup flow.
4. **Token rotation strategy?** Should the CLI silently refresh the 30-day token on each use (sliding window), or let it hard-expire?
5. **Multi-profile support?** Like gh auth login supports multiple GitHub accounts -- probably YAGNI for now.

---

## Sources

- [gh auth login -- GitHub CLI](https://cli.github.com/manual/gh_auth_login)
- [Doppler CLI Guide](https://docs.doppler.com/docs/cli)
- [How to build browser-based OAuth into your CLI -- WorkOS](https://workos.com/blog/how-to-build-browser-based-oauth-into-your-cli-with-workos)
- [OAuth Device Flow -- GitHub Docs](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps)
