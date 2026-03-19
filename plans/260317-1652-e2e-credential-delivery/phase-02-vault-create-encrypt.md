# Phase 2: Vault Create -- Encrypt + Store

## Context Links
- [plan.md](./plan.md)
- [phase-01](./phase-01-server-crypto-layer.md)
- `server/internal/vault/handler.go` -- current handler (lines 44-102)
- `server/internal/vault/service.go` -- CreateSecret (lines 63-115)
- `dashboard/src/components/secrets/secret-form.tsx` -- sends `{name, value, ...}`

## Overview
- **Priority:** P0
- **Status:** pending
- **Description:** Accept plaintext `value` in `POST /secrets`, server-side encrypt via envelope encryption, store blob in MinIO + wrapped DEK in Postgres.

## Key Insights
- Current `createSecretRequest` has `encrypted_blob` and `encrypted_dek` (base64 strings). Dashboard sends `value` instead -- these fields arrive empty.
- Handler currently base64-decodes empty strings (succeeds, yields empty `[]byte`). Service stores empty DEK in Postgres, skips MinIO put (`len(blob) == 0`).
- Need backward compat: if `value` provided, do server-side encryption. If `encrypted_blob`+`encrypted_dek` provided (future client-side encryption), use those directly.

## Requirements

### Functional
- `POST /secrets` accepts `value` string field
- When `value` present: generate DEK, encrypt value, wrap DEK with master key, store both
- When `encrypted_blob`+`encrypted_dek` present: use as-is (existing path)
- Error if neither `value` nor `encrypted_blob` provided

### Non-Functional
- Handler does all crypto; service remains storage-only (clean separation)
- `value` must not be logged

## Architecture

```
POST /secrets { name, value, credential_type, ... }
         |
         v
   [vault handler]
   if value != "":
     dek = GenerateDEK()
     blob = EncryptAES256GCM(dek, value)
     wrappedDEK = EncryptAES256GCM(masterKey, dek)
   else:
     blob = base64decode(encrypted_blob)
     wrappedDEK = base64decode(encrypted_dek)
         |
         v
   [vault service] CreateSecret(blob, wrappedDEK)
         |
         v
   Postgres: encrypted_dek    MinIO: encrypted blob
```

## Related Code Files

### Modify
- `server/internal/vault/handler.go` -- add `Value` to request, add `masterKey` to Handler, encrypt logic
- `server/cmd/server/main.go` -- pass masterKey to `vault.NewHandler`

### No changes needed
- `server/internal/vault/service.go` -- already accepts `EncryptedBlob`/`EncryptedDEK` bytes

## Implementation Steps

1. Modify `vault.Handler` struct to hold master key:
   ```go
   type Handler struct {
       service   *Service
       masterKey []byte
   }
   ```
   Update `NewHandler(service *Service, masterKey []byte) *Handler`

2. Add `Value` field to `createSecretRequest`:
   ```go
   type createSecretRequest struct {
       Name           string `json:"name"`
       Description    string `json:"description"`
       CredentialType string `json:"credential_type"`
       Source         string `json:"source"`
       Value          string `json:"value"`          // plaintext (server encrypts)
       EncryptedBlob  string `json:"encrypted_blob"` // base64 (pre-encrypted)
       EncryptedDEK   string `json:"encrypted_dek"`  // base64 (pre-encrypted)
       Policy         string `json:"policy"`
   }
   ```

3. Modify `createSecret` handler logic (after validation):
   ```
   if req.Value != "":
     dek, _ = crypto.GenerateDEK()
     blob, _ = crypto.EncryptAES256GCM(dek, []byte(req.Value))
     wrappedDEK, _ = crypto.EncryptAES256GCM(h.masterKey, dek)
   else if req.EncryptedBlob != "":
     blob = base64decode(req.EncryptedBlob)
     wrappedDEK = base64decode(req.EncryptedDEK)
   else:
     return 400 "value or encrypted_blob required"
   ```

4. Update `main.go`:
   - Change `vault.NewHandler(vaultService)` to `vault.NewHandler(vaultService, masterKey)`

5. Same pattern for `updateSecretRequest` -- add `Value` field, encrypt if present.

## Todo List
- [ ] Add `masterKey` to `vault.Handler`
- [ ] Add `Value` to `createSecretRequest`
- [ ] Implement encrypt-on-create logic in handler
- [ ] Add `Value` to `updateSecretRequest` + update handler
- [ ] Update `NewHandler` call in `main.go`
- [ ] Verify `go build ./...` compiles
- [ ] Test via curl: `POST /secrets` with `value` field

## Success Criteria
- `curl -X POST /api/v1/secrets -d '{"name":"test","value":"sk-123","credential_type":"api_key"}' -H "Authorization: Bearer $TOKEN"` returns 201
- MinIO has a non-empty blob at `secrets/{userID}/{secretID}`
- Postgres `secrets` row has non-empty `encrypted_dek`
- Blob is not plaintext (encrypted)

## Risk Assessment
- **Empty value** -- should we allow secrets with no value? For now, require either `value` or `encrypted_blob`.
- **Large values** -- AES-GCM has no practical size limit for secrets (< 64GB). Not a concern for credential strings.

## Security Considerations
- Never log `req.Value`
- Clear `dek` bytes after wrapping (best-effort in Go)
- Validate `value` is non-empty string when provided

## Next Steps
- Phase 3 reads back the stored blob and decrypts for credential delivery
