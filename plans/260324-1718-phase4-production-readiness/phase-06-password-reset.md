---
phase: "4.6"
title: "Password Reset"
priority: P1
effort: 3h
status: pending
depends_on: ["SMTP configured"]
---

# Phase 4.6: Password Reset

## Context Links
- `server/internal/auth/handler.go` -- login, register handlers
- `server/internal/auth/handler.go:122` -- HashPassword / VerifyPassword
- `server/internal/notify/email.go` -- EmailSender
- `dashboard/src/app/(auth)/login/page.tsx` -- login page (needs "Forgot password?" link)

## Overview

No password reset flow exists. Users who forget passwords are locked out. Need forgot-password email flow.

## DB Migration: 000039

```sql
-- 000039_password_reset_tokens.up.sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_tokens_hash ON password_reset_tokens(token_hash);
```

```sql
-- 000039_password_reset_tokens.down.sql
DROP TABLE IF EXISTS password_reset_tokens;
```

## Related Code Files

### Modify
- `server/internal/auth/handler.go` -- add forgotPassword + resetPassword handlers, mount routes
- `dashboard/src/app/(auth)/login/page.tsx` -- add "Forgot password?" link

### Create
- `server/internal/database/migrations/000039_password_reset_tokens.up.sql`
- `server/internal/database/migrations/000039_password_reset_tokens.down.sql`
- `dashboard/src/app/(auth)/forgot-password/page.tsx`
- `dashboard/src/app/(auth)/reset-password/page.tsx`

## Implementation Steps

### Step 1: Create migration
As specified above.

### Step 2: Forgot password endpoint

```go
// POST /auth/forgot-password
// Body: { "email": "user@example.com" }
func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
    var req struct{ Email string `json:"email"` }
    json.NewDecoder(r.Body).Decode(&req)

    // Always return 200 (don't leak whether email exists)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "If the email exists, a reset link has been sent"})

    // Background: check if user exists
    var userID string
    err := h.pool.QueryRow(r.Context(),
        `SELECT id FROM users WHERE email = $1 AND password_hash != ''`, req.Email,
    ).Scan(&userID)
    if err != nil { return } // user doesn't exist or is OAuth-only

    // Generate token
    raw, hash, _ := generateVerificationToken() // reuse same helper
    h.pool.Exec(r.Context(),
        `INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
         VALUES ($1, $2, $3)`,
        userID, hash, time.Now().Add(1*time.Hour))

    resetURL := fmt.Sprintf("%s/reset-password?token=%s", h.dashboardURL, raw)
    body := fmt.Sprintf("Reset your Valt password:\n%s\n\nThis link expires in 1 hour.", resetURL)
    h.emailSender.Send(r.Context(), req.Email, "Reset your Valt password", body)
}
```

### Step 3: Reset password endpoint

```go
// POST /auth/reset-password
// Body: { "token": "xxx", "password": "new_password" }
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Token    string `json:"token"`
        Password string `json:"password"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    if err := validator.ValidatePassword(req.Password); err != nil {
        apierror.BadRequest(w, err.Error())
        return
    }

    sum := sha256.Sum256([]byte(req.Token))
    tokenHash := hex.EncodeToString(sum[:])

    var userID string
    var expiresAt time.Time
    var used bool
    err := h.pool.QueryRow(r.Context(),
        `SELECT user_id, expires_at, used FROM password_reset_tokens WHERE token_hash = $1`,
        tokenHash).Scan(&userID, &expiresAt, &used)

    if err != nil || used || time.Now().After(expiresAt) {
        apierror.BadRequest(w, "invalid or expired token")
        return
    }

    hash, _ := HashPassword(req.Password)
    h.pool.Exec(r.Context(), `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, userID)
    h.pool.Exec(r.Context(), `UPDATE password_reset_tokens SET used = true WHERE token_hash = $1`, tokenHash)

    // Invalidate all refresh tokens for security
    h.pool.Exec(r.Context(), `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successful"})
}
```

### Step 4: Mount routes
In `Routes()`:
```go
r.Post("/forgot-password", h.forgotPassword)
r.Post("/reset-password", h.resetPassword)
```

### Step 5: Dashboard forgot-password page

```tsx
// dashboard/src/app/(auth)/forgot-password/page.tsx
// Simple form: email input -> POST /api/v1/auth/forgot-password
// Show success message regardless of email existence
```

### Step 6: Dashboard reset-password page

```tsx
// dashboard/src/app/(auth)/reset-password/page.tsx
// Read ?token= from URL
// Form: new password + confirm password
// POST /api/v1/auth/reset-password with { token, password }
// On success: redirect to /login with "Password reset" toast
```

### Step 7: Add "Forgot password?" link to login page

Add below the password field in `login/page.tsx`:
```tsx
<button onClick={() => router.push('/forgot-password')}
  className="text-xs text-primary hover:underline">
  Forgot password?
</button>
```

## Todo Checklist

- [ ] Create migration 000039
- [ ] Add `POST /auth/forgot-password` endpoint
- [ ] Add `POST /auth/reset-password` endpoint
- [ ] Mount new routes in auth handler
- [ ] Create forgot-password page
- [ ] Create reset-password page
- [ ] Add "Forgot password?" link to login page
- [ ] Test: forgot password -> email received -> reset link works
- [ ] Test: expired token (>1 hour) rejected
- [ ] Test: used token rejected (single-use)
- [ ] Test: OAuth-only user (no password_hash) doesn't receive email
- [ ] Test: non-existent email returns same 200 (no leak)

## Success Criteria

- Users can reset password via email link
- Token is single-use with 1-hour expiry
- All refresh tokens invalidated after password reset (security)
- No email existence leak (always 200)
- OAuth-only users excluded from password reset

## Security Considerations

- Always return 200 on forgot-password (prevent email enumeration)
- Token hashed in DB (SHA-256), raw only in email
- 1-hour expiry, single-use (marked as used, not deleted for audit trail)
- All refresh tokens revoked on password change
- Password validation reused from registration
- OAuth-only users (empty password_hash) excluded
