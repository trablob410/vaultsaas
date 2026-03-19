# Phase 3: Credential Delivery -- Decrypt + Return

## Context Links
- [plan.md](./plan.md)
- [phase-01](./phase-01-server-crypto-layer.md)
- `server/internal/workflow/handler.go` -- GetCredential (lines 212-254)
- `server/internal/workflow/credential.go` -- CredentialSession struct
- `server/internal/vault/service.go` -- GetSecret (lines 155-173)
- `server/internal/vault/storage.go` -- Storage.Get

## Overview
- **Priority:** P0
- **Status:** pending
- **Description:** `GET /credentials/{request_id}` fetches encrypted blob from MinIO, decrypts envelope, returns plaintext `value` in response.

## Key Insights
- Current `GetCredential` returns only `CredentialSession` metadata (id, status, expires_at, usage_count). No blob fetch, no decryption.
- `GetSecret` requires `userID` ownership check. For credential delivery, the access request is already approved -- need a method without owner constraint.
- `workflow.Handler` already has `vaultSvc` field. Need to add `masterKey` and access to `storage`.
- The `Storage` interface is on `vault` package. Options: (a) expose storage via vault service method, (b) pass storage to workflow handler. Option (a) is cleaner.

## Requirements

### Functional
- `GET /credentials/{request_id}` returns `{"id","status","expires_at",..., "value":"the-secret-value"}`
- Only returns `value` when session status is `active` and not expired
- Increments usage_count (existing behavior preserved)
- Auto-revoke for single-use (existing behavior preserved)

### Non-Functional
- Decryption happens in-memory, plaintext never hits disk
- Audit log entry for every credential access (existing)

## Architecture

```
GET /credentials/{request_id}
         |
         v
   [workflow handler]
   1. Ownership check (existing)
   2. credMgr.GetCredential (increments usage, returns session)
   3. vaultSvc.GetSecretByID(req.SecretID)  <-- NEW method
   4. vaultSvc.GetBlob(storageKey)           <-- NEW method
   5. crypto.DecryptAES256GCM(masterKey, encryptedDEK) --> dek
   6. crypto.DecryptAES256GCM(dek, encryptedBlob) --> plaintext
   7. Return session + value
```

## Related Code Files

### Modify
- `server/internal/vault/service.go` -- add `GetSecretByID(ctx, secretID)` (no userID), add `GetBlob(ctx, storageKey)`
- `server/internal/workflow/handler.go` -- add `masterKey` to Handler, decrypt in GetCredential
- `server/internal/workflow/credential.go` -- add `Value` to CredentialSession response
- `server/cmd/server/main.go` -- pass masterKey to workflow handler

## Implementation Steps

1. Add `GetSecretByID` to `vault.Service`:
   ```go
   func (s *Service) GetSecretByID(ctx context.Context, secretID string) (*Secret, error)
   ```
   - Same query as `GetSecret` but without `user_id = $2` constraint
   - Returns `storage_key` and `encrypted_dek` (needed for decryption)

2. Add `GetBlob` to `vault.Service`:
   ```go
   func (s *Service) GetBlob(ctx context.Context, storageKey string) ([]byte, error)
   ```
   - Delegates to `s.storage.Get(ctx, storageKey)`
   - Exposes storage access without leaking Storage interface

3. Add `Value` field to `CredentialSession`:
   ```go
   type CredentialSession struct {
       // ... existing fields ...
       Value string `json:"value,omitempty"`
   }
   ```

4. Add `masterKey` to `workflow.Handler`:
   ```go
   type Handler struct {
       service    *Service
       credMgr    *CredentialManager
       vaultSvc   *vault.Service
       auditLog   *audit.Logger
       notifySvc  *notify.Service
       masterKey  []byte
   }
   ```
   Update `NewHandler` signature to accept `masterKey []byte`.

5. Modify `GetCredential` handler (after existing ownership check + credMgr.GetCredential):
   ```
   secret, err := h.vaultSvc.GetSecretByID(ctx, req.SecretID)
   if secret != nil && len(secret.EncryptedDEK) > 0 {
       blob, err := h.vaultSvc.GetBlob(ctx, secret.StorageKey)
       dek, err := crypto.DecryptAES256GCM(h.masterKey, secret.EncryptedDEK)
       plaintext, err := crypto.DecryptAES256GCM(dek, blob)
       session.Value = string(plaintext)
   }
   ```
   - Handle errors: if decryption fails, log error but still return session metadata (degraded response)

6. Update `main.go`:
   - Change `workflow.NewHandler(workflowSvc, credMgr, vaultService, auditLogger, notifySvc)` to also pass `masterKey`

7. Update the policy/auto-revoke section: currently calls `h.vaultSvc.GetSecret(ctx, userID, req.SecretID)` which does ownership check. The requester may not be the secret owner (AI agent requesting access). Change to `GetSecretByID` here too.

## Todo List
- [ ] Add `GetSecretByID` to vault service
- [ ] Add `GetBlob` to vault service
- [ ] Add `Value` to `CredentialSession` struct
- [ ] Add `masterKey` to workflow handler
- [ ] Update `NewHandler` signature
- [ ] Implement decrypt logic in `GetCredential`
- [ ] Fix policy lookup to use `GetSecretByID` (bug: current code uses owner-scoped query)
- [ ] Update `main.go` wiring
- [ ] Verify `go build ./...` compiles

## Success Criteria
- After creating a secret with `value` (Phase 2), creating an access request, approving it:
  ```
  curl GET /api/v1/credentials/{request_id} -H "Authorization: Bearer $TOKEN"
  ```
  Response includes `"value": "sk-123"` (the original plaintext)
- Expired/revoked sessions return metadata only, no `value`
- Audit log records `credential.access` event

## Risk Assessment
- **Requester != owner bug**: Current `GetCredential` calls `GetSecret(ctx, userID, req.SecretID)` where `userID` is the requester. If requester != owner, `GetSecret` returns nil. The `GetSecretByID` fix resolves this.
- **Blob missing**: If MinIO blob was never stored (empty value), return session without value gracefully.

## Security Considerations
- Plaintext value only in HTTP response body, never logged
- Value only returned for `active` sessions (enforced by `credMgr.GetCredential` query)
- TLS required in production (Caddy handles this)
- Auto-revoke for single-use happens AFTER value is returned

## Next Steps
- Phase 4 reads `value` from this response in MCP tools.rs
