---
phase: "4.5"
title: "Email Verification"
priority: P1
effort: 4h
status: pending
depends_on: ["SMTP configured"]
---

# Phase 4.5: Email Verification

## Context Links
- `server/internal/auth/handler.go:102-144` -- register handler
- `server/internal/notify/email.go` -- EmailSender already exists
- `server/internal/config/config.go:35-39` -- SMTP config
- `server/internal/auth/jwt.go` -- JWT signing for tokens
- `server/internal/database/migrations/` -- currently at 000037

## Overview

Users can register without verifying email. Need email verification to prevent fake accounts and enable email-dependent features (password reset, invitations).

## Key Insights

- `notify.EmailSender` already supports sending plaintext email via SMTP
- Need HTML email support for better UX (or keep plaintext, KISS)
- Verification token: signed JWT with `typ: "email_verify"`, 24h expiry -- simpler than DB tokens
- Alternative: opaque token in DB table -- more secure (single-use guaranteed)
- Choosing DB token approach: simpler to invalidate, single-use by DELETE

## Requirements

### Functional
- After registration, send verification email with link
- `GET /auth/verify-email?token=X` validates token, sets `email_verified=true`
- Dashboard shows "Verify your email" banner until verified
- Resend verification: `POST /auth/resend-verification`
- Block sensitive actions until verified: create secret, invite members

### Non-functional
- Token expires in 24 hours
- Single-use: token deleted after verification
- Rate limit resend to 1 per 5 minutes

## DB Migration: 000038

```sql
-- 000038_email_verification.up.sql
ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_verification_tokens_hash ON email_verification_tokens(token_hash);
CREATE INDEX idx_email_verification_tokens_user ON email_verification_tokens(user_id);

-- Existing users are considered verified (grandfathered)
UPDATE users SET email_verified = true WHERE created_at < now();
```

```sql
-- 000038_email_verification.down.sql
DROP TABLE IF EXISTS email_verification_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
```

## Related Code Files

### Modify
- `server/internal/auth/handler.go` -- register: send verification email; add verify + resend endpoints
- `server/internal/auth/handler.go` Routes() -- mount new endpoints
- `server/internal/notify/email.go` -- add verification email builder
- `dashboard/src/app/(dashboard)/layout.tsx` -- add email verification banner

### Create
- `server/internal/database/migrations/000038_email_verification.up.sql`
- `server/internal/database/migrations/000038_email_verification.down.sql`
- `dashboard/src/components/layout/email-verification-banner.tsx`

## Implementation Steps

### Step 1: Create migration
Create 000038 migration files as specified above.

### Step 2: Add verification token helpers to auth handler

```go
func generateVerificationToken() (raw string, hash string, err error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", "", err
    }
    raw = hex.EncodeToString(b)
    sum := sha256.Sum256([]byte(raw))
    hash = hex.EncodeToString(sum[:])
    return raw, hash, nil
}

func (h *Handler) sendVerificationEmail(ctx context.Context, userID, email string) error {
    raw, hash, err := generateVerificationToken()
    if err != nil {
        return err
    }

    _, err = h.pool.Exec(ctx,
        `INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
         VALUES ($1, $2, $3)`,
        userID, hash, time.Now().Add(24*time.Hour))
    if err != nil {
        return err
    }

    verifyURL := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", h.dashboardURL, raw)
    body := fmt.Sprintf("Welcome to Valt!\n\nPlease verify your email:\n%s\n\nThis link expires in 24 hours.", verifyURL)

    return h.emailSender.Send(ctx, email, "Verify your Valt email", body)
}
```

### Step 3: Update register handler
After user creation and org auto-create, call `sendVerificationEmail`.

### Step 4: Verify email endpoint

```go
// GET /auth/verify-email?token=X
func (h *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
    rawToken := r.URL.Query().Get("token")
    if rawToken == "" {
        apierror.BadRequest(w, "token is required")
        return
    }

    sum := sha256.Sum256([]byte(rawToken))
    tokenHash := hex.EncodeToString(sum[:])

    var userID string
    var expiresAt time.Time
    err := h.pool.QueryRow(r.Context(),
        `SELECT user_id, expires_at FROM email_verification_tokens WHERE token_hash = $1`,
        tokenHash).Scan(&userID, &expiresAt)

    if err != nil || time.Now().After(expiresAt) {
        apierror.BadRequest(w, "invalid or expired token")
        return
    }

    // Set verified + delete token (single-use)
    h.pool.Exec(r.Context(), `UPDATE users SET email_verified = true WHERE id = $1`, userID)
    h.pool.Exec(r.Context(), `DELETE FROM email_verification_tokens WHERE token_hash = $1`, tokenHash)

    // Redirect to dashboard with success message
    http.Redirect(w, r, h.dashboardURL+"/?verified=true", http.StatusFound)
}
```

### Step 5: Resend verification endpoint

```go
// POST /auth/resend-verification (authenticated)
func (h *Handler) resendVerification(w http.ResponseWriter, r *http.Request) {
    userID := auth.UserIDFromContext(r.Context())
    // Check rate limit (1 per 5 min)
    // Delete old tokens for user
    // Generate and send new token
}
```

### Step 6: Dashboard verification banner

```tsx
// email-verification-banner.tsx
// Shown in dashboard layout when user.email_verified is false
// "Please verify your email. [Resend verification email]"
```

Need to expose `email_verified` in a `/auth/me` or `/users/me` endpoint (or include in JWT claims).

### Step 7: Add email_verified to JWT claims (optional)
Simpler: add `GET /auth/me` endpoint that returns user profile including `email_verified`.

## Todo Checklist

- [ ] Create migration 000038 (email_verified column + verification_tokens table)
- [ ] Add `emailSender` field to auth.Handler
- [ ] Add `sendVerificationEmail()` helper
- [ ] Update `register()` to send verification email
- [ ] Add `GET /auth/verify-email` endpoint
- [ ] Add `POST /auth/resend-verification` endpoint (authenticated)
- [ ] Add `GET /auth/me` endpoint returning user profile with email_verified
- [ ] Create email verification banner component
- [ ] Add banner to dashboard layout
- [ ] Test: register -> receive email -> click link -> verified
- [ ] Test: resend verification
- [ ] Test: expired token rejection
- [ ] Test: existing users grandfathered as verified

## Success Criteria

- New users receive verification email after registration
- Clicking verification link sets email_verified=true
- Dashboard shows banner until verified
- Existing users unaffected (grandfathered)
- Resend works with rate limiting

## Security Considerations

- Token stored as SHA-256 hash (raw token only in email link)
- Single-use: deleted after verification
- 24-hour expiry prevents stale tokens
- Rate limit on resend prevents email bombing
- Grandfathering existing users avoids locking out current users
