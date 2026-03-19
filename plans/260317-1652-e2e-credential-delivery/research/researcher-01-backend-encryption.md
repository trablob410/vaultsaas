# Backend Encryption & Credential Delivery — Research Report
**Date:** 2026-03-17

---

## 1. What fields does `POST /secrets` accept and store?

### Input (`CreateSecretInput`)
| Field | Type | Notes |
|---|---|---|
| `name` | string | required |
| `description` | string | optional |
| `credential_type` | string | defaults to `api_key` |
| `source` | string | optional |
| `encrypted_blob` | `[]byte` | the actual secret payload |
| `encrypted_dek` | `[]byte` | envelope key wrapping the DEK |
| `policy` | string (JSON) | defaults to `{}` |

### DB columns written (`secrets` table, migrations 000002 + 000007)
`id, user_id, name, description, storage_key, encrypted_dek, policy, credential_type, source, version, created_at, updated_at, deleted_at`

`storage_key` is derived from `crypto.StorageKey(userID, secretID)` — a deterministic path, NOT random.

### MinIO blob
`storage_key` → raw `encrypted_blob` bytes. Put via `PutObject` with `application/octet-stream`. No server-side encryption layer added by the application — it relies entirely on the blob being pre-encrypted by the caller.

---

## 2. Is there any server-side encryption of the secret value?

**No.** The server is a pass-through. It:
- Accepts `encrypted_blob` (already encrypted by the client)
- Accepts `encrypted_dek` (already wrapped by the client)
- Stores both verbatim — `encrypted_dek` in Postgres, `encrypted_blob` in MinIO

There is no server-side decryption, re-encryption, or key-wrapping. The server never holds or derives a master key. This is the intended zero-knowledge design.

**Key gap:** `pkg/crypto` provides `StorageKey()` but there is no server-side `Unwrap(DEK)` or `Decrypt(blob)` logic present anywhere in the codebase. Without a server-managed master key (KEK), the server cannot decrypt blobs on behalf of the MCP client.

---

## 3. What does `GET /credentials/{request_id}` currently return?

Handler at `workflow/handler.go:212-254` returns a `CredentialSession` JSON object:

```json
{
  "id": "<uuid>",
  "access_request_id": "<uuid>",
  "credential_type": "api_key",
  "status": "active",
  "expires_at": "...",
  "usage_count": 1,
  "revoked_at": null,
  "created_at": "..."
}
```

**No secret value is returned.** The handler:
1. Validates request ownership via `req.RequesterUserID == userID`
2. Calls `credMgr.GetCredential()` — increments `usage_count` and returns session metadata only
3. Optionally auto-revokes if policy is single-use
4. Encodes and returns the `CredentialSession` struct — no blob fetch, no decryption

---

## 4. What is missing to return an actual decrypted value?

### Critical gap: decryption path does not exist
To deliver the plaintext secret to an approved MCP agent the server needs:

1. **Server-managed KEK (Key Encryption Key)**: A master key held by the server to unwrap the `encrypted_dek` stored in Postgres. Currently no KEK exists server-side.

2. **Blob fetch + decrypt in `GetCredential` handler**: After session validation, the handler must:
   - Call `vaultSvc.GetSecret()` to get `storage_key` + `encrypted_dek`
   - Unwrap DEK using server KEK → plaintext DEK
   - Call `storage.Get(storage_key)` → `encrypted_blob`
   - Decrypt blob with DEK (AES-256-GCM) → plaintext value
   - Return plaintext in response

3. **Response shape**: `CredentialSession` has no `value` / `secret_value` field. A new response type or an augmented struct is needed.

4. **Decision: server-side KEK vs. client-side only**: The current design expects the client (dashboard browser) to hold the DEK. For MCP agent delivery (no browser), the server must hold a KEK or use a hybrid approach (e.g., agent presents a one-time token that the approver has wrapped with an agent-specific key).

---

## 5. DB schema changes needed?

### `secrets` table — no changes required
All necessary columns exist: `storage_key`, `encrypted_dek`, `credential_type`, `source`, `version`.

### `credential_sessions` table — no changes required for basic delivery
Columns: `id, access_request_id, credential_type, status, expires_at, usage_count, revoked_at, created_at`.

### Potential addition (if value is cached in session)
A `encrypted_value BYTEA` column on `credential_sessions` could cache a one-time-use re-encrypted payload (wrapped for the requesting agent's public key). This is optional and depends on chosen key delivery model.

---

## Summary Table

| Concern | Current State | Gap |
|---|---|---|
| Secret storage (MinIO) | Blob stored verbatim, no server encryption | None if client pre-encrypts |
| DEK storage (Postgres) | `encrypted_dek` stored verbatim | Server needs KEK to unwrap |
| `GET /credentials` response | Returns session metadata only | Missing: decrypted value |
| Server-side decryption | Not implemented anywhere | Entire decrypt path missing |
| `CredentialSession` shape | No `value` field | Needs augmented response type |
| Schema | Complete for metadata | Optional: cache column on `credential_sessions` |

---

## Unresolved Questions

1. **Key delivery model**: Should the server hold a KEK (symmetric, from env/Vault/KMS) to unwrap DEKs? Or should approvals use agent-specific asymmetric wrapping (agent pubkey encrypts DEK at approval time)?
2. **Who encrypted the blob?**: Does the dashboard currently send a real `encrypted_blob` + `encrypted_dek` pair, or are these empty/placeholder in current frontend implementation?
3. **MCP agent authentication**: How does the MCP client identify itself to `GET /credentials` — does it use the same JWT auth as users, or a separate agent token?
4. **Single-use enforcement**: Auto-revoke after read is implemented, but usage count is incremented before auto-revoke check — race condition risk under concurrent reads.
