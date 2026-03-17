# Phase 1 — Encrypt DynSecret Storage (C1)

## Context Links
- `server/internal/dynsecret/service.go` — service with plaintext storage
- `server/pkg/crypto/aes.go` — AES-256-GCM encrypt/decrypt helpers
- `server/cmd/server/main.go:117` — `dynsecret.NewService(pool)` instantiation
- `server/internal/config/config.go` — `cfg.MasterKey()` returns `[]byte`

## Overview
- **Priority:** CRITICAL
- **Status:** pending
- **Description:** Provider config (contains admin_user/admin_password) and lease credentials stored as raw JSON in `config_enc` / `secret_data_enc` columns. Must encrypt with AES-256-GCM using master key.

## Key Insights
- TODOs already in code at lines 23 and 92 acknowledge this gap
- `config_enc` column is `BYTEA` — already compatible with ciphertext
- `crypto.EncryptAES256GCM` returns `[12-byte nonce || ciphertext+tag]`
- Master key already loaded in main.go line 80-83

## Requirements
**Functional:** All provider configs and lease credentials encrypted at rest; decrypted on read.
**Non-functional:** Zero downtime; existing data needs one-time migration or accept that old rows are plaintext (check at read time).

## Architecture
```
main.go -> masterKey -> dynsecret.NewService(pool, masterKey)
                            |
                CreateProvider: json.Marshal -> crypto.Encrypt -> INSERT
                ListProviders:  SELECT -> crypto.Decrypt -> json.Unmarshal
                GetProvider:    SELECT -> crypto.Decrypt -> json.Unmarshal
                CreateLease:    json.Marshal -> crypto.Encrypt -> INSERT
                RevokeLease:    SELECT -> crypto.Decrypt -> json.Unmarshal
```

## Related Code Files
| Action | File |
|--------|------|
| Modify | `server/internal/dynsecret/service.go` |
| Modify | `server/cmd/server/main.go` (line 117) |

## Implementation Steps

1. **Add `masterKey` to Service struct**
   ```go
   type Service struct {
       db        *pgxpool.Pool
       masterKey []byte
   }
   func NewService(db *pgxpool.Pool, masterKey []byte) *Service {
       return &Service{db: db, masterKey: masterKey}
   }
   ```

2. **Update `main.go`** — pass masterKey:
   ```go
   dynSvc := dynsecret.NewService(pool, masterKey)
   ```

3. **CreateProvider** (line 24-41) — encrypt `raw` before INSERT:
   ```go
   enc, err := crypto.EncryptAES256GCM(s.masterKey, raw)
   // use enc instead of raw in INSERT
   ```

4. **ListProviders** (line 44-63) — decrypt `raw` before unmarshal:
   ```go
   decrypted, err := crypto.DecryptAES256GCM(s.masterKey, raw)
   if err != nil {
       // Fallback: try plaintext JSON for pre-migration rows
       _ = json.Unmarshal(raw, &pc.Config)
   } else {
       _ = json.Unmarshal(decrypted, &pc.Config)
   }
   ```

5. **GetProvider** (line 67-78) — same decrypt + fallback pattern

6. **CreateLease** (line 93-131) — encrypt `credBytes` before INSERT

7. **RevokeLease** (line 135-165) — decrypt `credRaw` before unmarshal, with plaintext fallback

8. Add import `"github.com/valt-dev/valt/server/pkg/crypto"`

## Todo List
- [ ] Add masterKey field to Service struct
- [ ] Update NewService signature
- [ ] Update main.go instantiation
- [ ] Encrypt in CreateProvider
- [ ] Decrypt in ListProviders (with fallback)
- [ ] Decrypt in GetProvider (with fallback)
- [ ] Encrypt in CreateLease
- [ ] Decrypt in RevokeLease (with fallback)
- [ ] Add crypto import
- [ ] Verify compilation: `go build ./...`
- [ ] Manual test: create provider, list providers, create lease, revoke lease

## Success Criteria
- `config_enc` and `secret_data_enc` columns contain ciphertext (not readable JSON)
- Read paths successfully decrypt and return correct data
- Pre-migration plaintext rows still readable (graceful fallback)

## Risk Assessment
- **Backward compatibility**: Existing plaintext rows must still be readable. Fallback handles this.
- **Key rotation**: Not in scope. Master key is single point. Document as future work.
- **Performance**: AES-256-GCM is fast; negligible overhead for small payloads.

## Security Considerations
- Master key zeroing after use not practical (Go GC). Acceptable for now.
- Encrypted config still in memory during provider instantiation — unavoidable.
- Nonce uniqueness guaranteed by `crypto/rand` in EncryptAES256GCM.

## Next Steps
- Phase 2 (Scanner auth) can proceed independently
- Future: key rotation mechanism, per-project encryption keys
