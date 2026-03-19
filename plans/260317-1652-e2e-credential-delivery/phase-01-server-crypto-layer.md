# Phase 1: Server Crypto Layer

## Context Links
- [plan.md](./plan.md)
- `server/pkg/crypto/storage-key.go` -- existing crypto package
- `server/internal/config/config.go` -- env config struct
- `server/cmd/server/main.go` -- service wiring

## Overview
- **Priority:** P0 (blocker for all other phases)
- **Status:** pending
- **Description:** Add AES-256-GCM encrypt/decrypt to `pkg/crypto/` and `VAULT_MASTER_KEY` env var to config.

## Key Insights
- `pkg/crypto/` currently only has `StorageKey()` helper. No encryption functions exist.
- `mcp-server/src/crypto.rs` already has AES-256-GCM decrypt with nonce-prepended format. Go side must match this format: `nonce(12) || ciphertext+tag`.
- Config uses `kelseyhightower/envconfig` -- simple struct tag addition.

## Requirements

### Functional
- `EncryptAES256GCM(key, plaintext []byte) ([]byte, error)` -- returns `nonce || ciphertext+tag`
- `DecryptAES256GCM(key, ciphertext []byte) ([]byte, error)` -- expects `nonce || ciphertext+tag`
- `GenerateDEK() ([]byte, error)` -- returns 32 random bytes
- Config loads `VAULT_MASTER_KEY` (base64-encoded 32 bytes)

### Non-Functional
- Nonce must be 12 bytes from `crypto/rand`
- Key must be exactly 32 bytes (AES-256)
- Fail loudly on startup if master key missing or wrong length

## Architecture

```
plaintext value
    |
    v
[GenerateDEK] --> 32-byte random DEK
    |
    v
[EncryptAES256GCM(DEK, plaintext)] --> encryptedBlob (nonce+ct)
[EncryptAES256GCM(masterKey, DEK)] --> encryptedDEK (nonce+ct)
    |
    v
Store encryptedBlob in MinIO, encryptedDEK in Postgres
```

## Related Code Files

### Create
- `server/pkg/crypto/aes.go` (~60 lines)

### Modify
- `server/internal/config/config.go` -- add `VaultMasterKey` field
- `server/cmd/server/main.go` -- decode master key, pass to vault handler

## Implementation Steps

1. Create `server/pkg/crypto/aes.go`:
   - Import `crypto/aes`, `crypto/cipher`, `crypto/rand`, `fmt`
   - `EncryptAES256GCM(key, plaintext []byte) ([]byte, error)`:
     - Validate `len(key) == 32`
     - Create `aes.NewCipher(key)` then `cipher.NewGCM(block)`
     - Generate 12-byte nonce via `rand.Read`
     - `ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)` -- prepends nonce
     - Return ciphertext, nil
   - `DecryptAES256GCM(key, data []byte) ([]byte, error)`:
     - Validate `len(key) == 32`, `len(data) >= 12+gcm.Overhead()`
     - Split `data[:12]` (nonce) and `data[12:]` (ciphertext)
     - `gcm.Open(nil, nonce, ciphertext, nil)`
   - `GenerateDEK() ([]byte, error)`:
     - 32 bytes from `crypto/rand`

2. Modify `server/internal/config/config.go`:
   - Add field: `VaultMasterKey string \`envconfig:"VAULT_MASTER_KEY" required:"true"\``

3. Modify `server/cmd/server/main.go`:
   - After `config.Load()`, decode `cfg.VaultMasterKey` from base64
   - Validate decoded length == 32, `log.Fatalf` if not
   - Store as `masterKey []byte` for passing to handlers

4. Add unit tests in `server/pkg/crypto/aes_test.go`:
   - Roundtrip encrypt/decrypt
   - Wrong key fails
   - Short ciphertext fails
   - GenerateDEK returns 32 bytes, unique each call

## Todo List
- [ ] Create `aes.go` with Encrypt/Decrypt/GenerateDEK
- [ ] Add `VaultMasterKey` to config struct
- [ ] Decode + validate master key in main.go
- [ ] Unit tests for crypto functions
- [ ] Verify `go build ./...` compiles

## Success Criteria
- `go test ./pkg/crypto/...` passes
- Encrypt then decrypt roundtrip produces original plaintext
- Wrong key returns error, not garbage
- Server refuses to start without `VAULT_MASTER_KEY`
- Format matches Rust `crypto.rs` (nonce-prepended, same AES-256-GCM)

## Risk Assessment
- **Key rotation not in scope** -- MVP uses single static KEK. Future: re-encrypt DEKs with new KEK.
- **Master key in env var** -- acceptable for MVP. Production: use KMS.

## Security Considerations
- Never log master key or DEK bytes
- Use `crypto/rand` only (not `math/rand`)
- Zeroize DEK after wrapping if Go allows (note: Go GC makes this best-effort)

## Next Steps
- Phase 2 uses these functions in vault handler
