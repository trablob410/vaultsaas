# Researcher 02 — MCP get_credential & Dashboard UI Bugs

Date: 2026-03-17

---

## Area 1: MCP Server `get_credential` Tool

### What it does now

`tool_get_credential` (tools.rs:130-143):
1. Calls `client.get_credential(request_id)` → `GET /credentials/{request_id}` (client.rs:110-112)
2. Returns JSON with metadata fields only: `credential_id`, `request_id`, `issued_at`, `expires_at`, `status`
3. Tells agent: _"Access data via vault://requests/{request_id} resource"_

No decryption happens. The raw `credential_data` field from the API response is never read or returned.

### What "decrypt stub" means

There is **no decrypt stub in tools.rs** — the decrypt step is simply absent. The tool returns metadata and redirects the agent to an MCP resource URI instead of returning the actual secret value.

The full intended flow is:
1. Go backend returns `credential_data` = base64-encoded ciphertext (AES-256-GCM envelope-encrypted blob)
2. MCP server must: fetch ciphertext → decrypt with agent's key material → return plaintext to agent

Currently step 2 is unimplemented. `crypto.rs` exists in the source tree (seen in git status) but is not called from `tool_get_credential`. The resource handler at `vault://requests/{request_id}` presumably has the same gap — no file reads it from `tools.rs`.

### HTTP calls made in client.rs

| Method | Path | Used by |
|--------|------|---------|
| GET | `/secrets` | `list_secrets` |
| POST | `/secrets/{id}/access-requests` | `create_access_request` |
| GET | `/access-requests[?status=]` | `get_access_requests` |
| GET | `/credentials/{request_id}` | `get_credential` |
| POST | `/credentials/{request_id}/revoke` | `revoke_credential` |
| GET | `/audit/logs[?date=]` | `get_audit_logs` |

No endpoint exists for fetching the DEK or performing server-side decryption assist — any decryption must be fully client-side in the MCP server using `crypto.rs`.

---

## Area 2: Dashboard Approvals Table Bugs

### Bug 1 — Duration shows "m" instead of full label

**Location:** `approval-list.tsx:84`
```tsx
<TableCell className="text-sm">{req.duration_minutes}m</TableCell>
```
`duration_minutes` is a raw integer (e.g. `30`). Output is `30m`. This is intentional shorthand but arguably insufficient UX — no unit label ("mins", "minutes"). If `duration_minutes` is `null`/`undefined` (possible if field is missing from API response), it renders as `undefinedm` or `m`.

`AccessRequest` type (api.ts:18) declares `duration_minutes: number` (non-optional), so type is correct but runtime API could omit it.

### Bug 2 — Secret column shows truncated UUID instead of name

**Location:** `approval-list.tsx:82`
```tsx
<TableCell className="font-mono text-xs">{req.secret_id.slice(0, 8)}…</TableCell>
```
`AccessRequest` type (api.ts:14-27) has `secret_id: string` but **no `secret_name` field**. The API response for `/access-requests` returns only the UUID reference, not the joined secret name.

To show a name, either:
- The Go backend must join `secrets.name` into the access_requests response, or
- The frontend must do a secondary fetch per row (expensive), or
- A `secret_name` field must be added to `AccessRequest` type and populated by the API

### API response fields available vs accessed

| Field in `AccessRequest` type | Used in table | Notes |
|-------------------------------|---------------|-------|
| `id` | Yes (row key) | |
| `secret_id` | Yes (sliced to 8 chars) | Name not available |
| `reason` | Yes | |
| `duration_minutes` | Yes (appended "m") | |
| `status` | Yes (Badge) | |
| `created_at` | Yes (formatDate) | |
| `approved_at` | No | |
| `expires_at` | No | |
| `rejection_reason` | No | Not shown on rejected rows |
| `requester_id` | No | Not shown |
| `policy_tier` | No | Not shown |

---

## Summary

| Issue | Location | Fix needed |
|-------|----------|------------|
| `get_credential` returns no plaintext | tools.rs:130-143 | Call `crypto.rs` decrypt on `credential_data` from API response |
| No DEK fetch endpoint | client.rs | May need new Go endpoint or DEK bundled in credential response |
| Duration unit ambiguous / brittle | approval-list.tsx:84 | Guard against null; use "min" label or full string |
| Secret name missing | approval-list.tsx:82 | Add `secret_name` to Go API join + `AccessRequest` type |
| Rejection reason not displayed | approval-list.tsx | Surface `rejection_reason` on rejected rows |

## Unresolved Questions

1. Does `GET /credentials/{request_id}` currently return `credential_data` as ciphertext, or is the Go handler also a stub? Need to check `server/internal/` handler.
2. What key material does the MCP agent hold — is there a per-agent keypair stored via `crypto.rs` / `keychain.rs`, and what is the expected decryption flow (DEK wrapped by agent public key)?
3. Is `resources.rs` (seen in git status) the handler for `vault://requests/{request_id}` — does it perform decryption or is it also a stub?
