---
phase: "3.7"
title: "TOTP 2FA"
priority: P2
status: pending
effort: 6h
---

# Phase 3.7: TOTP 2FA

## Context Links
- [Auth handler](../../server/internal/auth/handler.go) — login/register flow
- [Auth middleware](../../server/internal/auth/middleware.go) — JWT validation
- [JWT manager](../../server/internal/auth/jwt.go) — token issuance
- [Config](../../server/internal/config/config.go)
- [Dashboard settings](../../dashboard/src/app/(dashboard)/settings/page.tsx)

## Overview

Add TOTP-based two-factor authentication using `pquerna/otp` library. Users can enable/disable 2FA via settings. Login flow demands TOTP code after password when 2FA is enabled. Backup codes for recovery.

## Key Insights
- `pquerna/otp` is maintained, Apache-2.0, pure Go — no CGO
- TOTP secret stored encrypted in users table (AES-256-GCM with masterKey)
- Backup codes: 10 single-use codes, stored hashed (SHA-256)
- Login becomes two-step: password → intermediate token → TOTP → full JWT
- Google OAuth users can also enable 2FA (checked on subsequent logins)

## Requirements

### Functional
- POST /me/totp/setup — generate TOTP secret, return QR URI + secret string
- POST /me/totp/verify — verify TOTP code, enable 2FA (first-time activation)
- DELETE /me/totp — disable 2FA (requires current TOTP code)
- POST /me/totp/backup-codes — regenerate backup codes
- Login flow: if totp_enabled → return `{requires_totp: true, totp_token: "..."}` instead of JWT
- POST /auth/totp/validate — accept totp_token + code → return JWT
- Backup code accepted in place of TOTP code

### Non-Functional
- TOTP secret encrypted at rest
- Backup codes hashed (not reversible)
- TOTP window: 1 step (30s before/after)
- Rate limit TOTP validation: 5 attempts per totp_token

## Architecture

```
Login Flow (2FA enabled):
POST /auth/login {email, password}
  → verify password → check totp_enabled
  → if true: return {requires_totp: true, totp_token: JWT(sub=userID, purpose="totp", exp=5min)}
  → NOT return access/refresh tokens yet

POST /auth/totp/validate {totp_token, code}
  → validate totp_token JWT (purpose="totp")
  → validate TOTP code against stored secret
  → OR validate against backup_codes (mark used)
  → return {access_token, refresh_token} (normal login response)

Setup Flow:
POST /me/totp/setup (authenticated)
  → generate random TOTP secret via otp.GenerateOpts
  → store temporarily (return to user, NOT yet enabled)
  → return {secret, qr_uri} (otpauth://totp/Valt:user@email?secret=XX&issuer=Valt)

POST /me/totp/verify {code} (authenticated)
  → validate code against pending secret
  → if valid: encrypt secret, store in users.totp_secret_enc, set totp_enabled=true
  → generate 10 backup codes, hash and store
  → return {backup_codes: [...]} (plaintext, shown once)
```

## Related Code Files

### Create
- `server/internal/auth/totp.go` — TOTP setup/verify/validate handlers
- `server/internal/auth/totp_store.go` — DB operations for TOTP
- `server/internal/database/migrations/000034_totp_2fa.up.sql`
- `server/internal/database/migrations/000034_totp_2fa.down.sql`
- `dashboard/src/app/(dashboard)/settings/security/page.tsx` — 2FA setup UI

### Modify
- `server/internal/auth/handler.go` — modify login to check totp_enabled
- `server/internal/auth/handler.go` — add Routes for TOTP endpoints
- `server/cmd/server/main.go` — wire TOTP routes
- `server/go.mod` — add pquerna/otp
- `dashboard/src/app/(dashboard)/settings/layout.tsx` or sidebar — add Security nav link

## Implementation Steps

### 1. Go dependency

```bash
cd server && go get github.com/pquerna/otp
```

### 2. DB Migration (000034)

```sql
-- 000034_totp_2fa.up.sql
ALTER TABLE users
  ADD COLUMN totp_secret_enc BYTEA,
  ADD COLUMN totp_enabled BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN totp_setup_secret VARCHAR(64); -- temp: pending verification

CREATE TABLE backup_codes (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(64) NOT NULL, -- SHA-256 hex
    used      BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_backup_codes_user ON backup_codes(user_id) WHERE used = false;
```

```sql
-- 000034_totp_2fa.down.sql
DROP TABLE IF EXISTS backup_codes;
ALTER TABLE users
  DROP COLUMN IF EXISTS totp_secret_enc,
  DROP COLUMN IF EXISTS totp_enabled,
  DROP COLUMN IF EXISTS totp_setup_secret;
```

### 3. auth/totp_store.go (~80 lines)

```go
type TOTPStore struct {
    pool      *pgxpool.Pool
    masterKey []byte
}

func NewTOTPStore(pool, masterKey) *TOTPStore
func (s *TOTPStore) SetSetupSecret(ctx, userID, secret string) error
  // UPDATE users SET totp_setup_secret = $2 WHERE id = $1
func (s *TOTPStore) GetSetupSecret(ctx, userID string) (string, error)
func (s *TOTPStore) EnableTOTP(ctx, userID, secret string) error
  // Encrypt secret, UPDATE users SET totp_secret_enc=$2, totp_enabled=true, totp_setup_secret=NULL
func (s *TOTPStore) DisableTOTP(ctx, userID string) error
  // UPDATE users SET totp_secret_enc=NULL, totp_enabled=false
  // DELETE FROM backup_codes WHERE user_id=$1
func (s *TOTPStore) GetTOTPSecret(ctx, userID string) (string, bool, error)
  // Return decrypted secret + totp_enabled flag
func (s *TOTPStore) StoreBackupCodes(ctx, userID string, codes []string) error
  // DELETE old codes, INSERT hashed codes
func (s *TOTPStore) UseBackupCode(ctx, userID, code string) (bool, error)
  // Hash code, UPDATE backup_codes SET used=true WHERE user_id=$1 AND code_hash=$2 AND used=false
  // Return true if a row was updated
```

### 4. auth/totp.go (~120 lines)

```go
type TOTPHandler struct {
    store  *TOTPStore
    jwtMgr *JWTManager
}

func (h *TOTPHandler) Setup(w, r)         // POST /me/totp/setup
func (h *TOTPHandler) Verify(w, r)        // POST /me/totp/verify
func (h *TOTPHandler) Disable(w, r)       // DELETE /me/totp
func (h *TOTPHandler) RegenerateBackup(w, r) // POST /me/totp/backup-codes
func (h *TOTPHandler) Validate(w, r)      // POST /auth/totp/validate
```

`Setup`:
- Generate via `totp.Generate(totp.GenerateOpts{Issuer: "Valt", AccountName: userEmail})`
- Store secret temporarily: `store.SetSetupSecret(ctx, userID, key.Secret())`
- Return `{secret: key.Secret(), qr_uri: key.URL()}`

`Verify`:
- Read `{code}` from body
- Get pending secret: `store.GetSetupSecret`
- Validate code: `totp.Validate(code, secret)`
- If valid: `store.EnableTOTP(ctx, userID, secret)`
- Generate 10 backup codes (crypto/rand, 8 chars each)
- `store.StoreBackupCodes(ctx, userID, codes)`
- Return `{backup_codes: codes}` (plaintext, shown once)

`Disable`:
- Read `{code}` from body (require current TOTP to disable)
- Validate against stored secret
- `store.DisableTOTP(ctx, userID)`

`Validate` (public, called during login):
- Read `{totp_token, code}` from body
- Validate totp_token JWT (purpose claim = "totp")
- Extract userID from token
- Get stored secret: `store.GetTOTPSecret(ctx, userID)`
- Try TOTP: `totp.Validate(code, secret)`
- If TOTP fails, try backup code: `store.UseBackupCode(ctx, userID, code)`
- If either succeeds: issue normal JWT (access + refresh)
- If both fail: return 401

### 5. Modify login flow in auth/handler.go

In `login` handler, after password verification:

```go
// Check if TOTP enabled
var totpEnabled bool
_ = h.pool.QueryRow(ctx, `SELECT totp_enabled FROM users WHERE id = $1`, userID).Scan(&totpEnabled)

if totpEnabled {
    // Issue short-lived TOTP challenge token
    totpToken, _ := h.jwtMgr.IssueTOTPToken(userID) // 5min expiry, purpose="totp"
    json.NewEncoder(w).Encode(map[string]interface{}{
        "requires_totp": true,
        "totp_token":    totpToken,
    })
    return
}
// else: normal JWT issuance (existing code)
```

Add `IssueTOTPToken` and `ValidateTOTPToken` to JWTManager:
- Same RS256 signing, but with custom claim `purpose: "totp"` and 5min expiry
- `ValidateTOTPToken` checks purpose claim

### 6. Wire routes

In auth `Routes()`:
```go
r.Post("/totp/validate", h.totpHandler.Validate) // public (no JWT needed)
```

In authenticated group in main.go:
```go
r.Post("/me/totp/setup", totpHandler.Setup)
r.Post("/me/totp/verify", totpHandler.Verify)
r.Delete("/me/totp", totpHandler.Disable)
r.Post("/me/totp/backup-codes", totpHandler.RegenerateBackup)
```

### 7. Dashboard — /settings/security page

New page at `dashboard/src/app/(dashboard)/settings/security/page.tsx`:

**When 2FA is disabled:**
- "Enable Two-Factor Authentication" card
- Click → POST /me/totp/setup → show QR code (use `qrcode.react` or simple `<img>` from QR URI)
- Input field for TOTP code verification
- Submit → POST /me/totp/verify → show backup codes once
- "Download backup codes" button (txt file download)

**When 2FA is enabled:**
- Status: "2FA is enabled" with green badge
- "Regenerate Backup Codes" button → POST /me/totp/backup-codes
- "Disable 2FA" button → prompt for current TOTP code → DELETE /me/totp

**Login page update:**
- After login returns `{requires_totp: true}`, show TOTP code input
- Submit → POST /auth/totp/validate → receive JWT → redirect to dashboard

### 8. Add settings/security nav link

Add "Security" to settings navigation (sidebar or tab bar), linking to `/settings/security`.

## Todo

- [ ] `go get pquerna/otp`
- [ ] Create migration 000034
- [ ] Implement auth/totp_store.go
- [ ] Implement auth/totp.go (5 handlers)
- [ ] Add IssueTOTPToken/ValidateTOTPToken to JWTManager
- [ ] Modify login handler for TOTP challenge
- [ ] Wire routes in main.go + auth routes
- [ ] Dashboard: /settings/security page
- [ ] Dashboard: login page TOTP step
- [ ] QR code rendering (qrcode.react or img)
- [ ] Unit tests: TOTP validation, backup code usage, setup flow
- [ ] Rate limit TOTP validate endpoint (5 attempts per token)

## Success Criteria
- User can enable 2FA: see QR, scan, enter code, get backup codes
- Login with 2FA: password → TOTP code → JWT
- Backup code works as TOTP alternative (single-use)
- User can disable 2FA with current TOTP code
- Google OAuth users can also enable 2FA

## Security Considerations
- TOTP secret encrypted at rest (AES-256-GCM)
- Backup codes stored as SHA-256 hashes (irreversible)
- TOTP challenge token: short-lived (5min), purpose-scoped
- Rate limit on /auth/totp/validate to prevent brute force
- Setup secret cleared after verification (no lingering plaintext)
- TOTP window = 1 step (30s tolerance) — standard security/usability tradeoff
