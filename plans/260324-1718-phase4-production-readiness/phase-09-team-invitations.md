---
phase: "4.9"
title: "Team Invitations"
priority: P2
effort: 4h
status: pending
depends_on: ["4.5"]
---

# Phase 4.9: Team Invitations

## Context Links
- `server/internal/org/` -- org service, membership CRUD
- `server/internal/auth/handler.go` -- registration flow
- `server/internal/notify/email.go` -- EmailSender
- `dashboard/src/lib/api-client.ts:86-91` -- org members API
- `server/internal/rbac/` -- role-based access control

## Overview

No way to invite team members to an organization. Currently must manually add by user_id after they register. Need email-based invitation flow.

## DB Migration: 000040

```sql
-- 000040_org_invitations.up.sql
CREATE TABLE org_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    invited_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, accepted, expired
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_at TIMESTAMPTZ
);

CREATE INDEX idx_org_invitations_token ON org_invitations(token_hash);
CREATE INDEX idx_org_invitations_org ON org_invitations(org_id);
CREATE INDEX idx_org_invitations_email ON org_invitations(email);
CREATE UNIQUE INDEX idx_org_invitations_org_email_pending
    ON org_invitations(org_id, email) WHERE status = 'pending';
```

```sql
-- 000040_org_invitations.down.sql
DROP TABLE IF EXISTS org_invitations;
```

## Related Code Files

### Modify
- `server/internal/org/handler.go` -- add invitation endpoints
- `server/internal/org/service.go` -- add invitation business logic
- `dashboard/src/lib/api-client.ts` -- add invitation API methods

### Create
- `server/internal/database/migrations/000040_org_invitations.up.sql`
- `server/internal/database/migrations/000040_org_invitations.down.sql`
- `dashboard/src/app/(auth)/accept-invite/page.tsx`
- `dashboard/src/components/org/invite-member-dialog.tsx`

## Implementation Steps

### Step 1: Create migration
As specified above.

### Step 2: Backend invitation endpoints

```go
// POST /orgs/{id}/invitations
// Body: { "email": "user@example.com", "role": "member" }
func (h *Handler) createInvitation(w http.ResponseWriter, r *http.Request) {
    orgID := chi.URLParam(r, "id")
    userID := auth.UserIDFromContext(r.Context())

    var req struct {
        Email string `json:"email"`
        Role  string `json:"role"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Validate role (member, admin only -- not owner)
    // Check caller is admin/owner of org
    // Generate invitation token
    raw, hash, _ := generateToken()

    // Insert invitation
    _, err := h.pool.Exec(r.Context(),
        `INSERT INTO org_invitations (org_id, email, role, token_hash, invited_by, expires_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
        orgID, req.Email, req.Role, hash, userID, time.Now().Add(7*24*time.Hour))

    // Send invitation email
    inviteURL := fmt.Sprintf("%s/accept-invite?token=%s", cfg.DashboardURL, raw)
    emailSender.Send(ctx, req.Email, "You're invited to join [OrgName] on Valt", body)

    // Return invitation object
}

// GET /orgs/{id}/invitations -- list pending invitations
// DELETE /orgs/{id}/invitations/{inviteId} -- cancel invitation

// POST /invitations/accept?token=X -- accept invitation (authenticated)
func (h *Handler) acceptInvitation(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    userID := auth.UserIDFromContext(r.Context())

    // Lookup invitation by token hash
    // Check not expired, not already accepted
    // Check user email matches invitation email
    // Add user to org_memberships with specified role
    // Mark invitation as accepted
}
```

### Step 3: Dashboard invite dialog

In org settings page, add "Invite Member" button:
```tsx
// invite-member-dialog.tsx
// Dialog with email input + role select (member/admin)
// POST /api/proxy/orgs/{orgId}/invitations
```

### Step 4: Accept invitation page

```tsx
// dashboard/src/app/(auth)/accept-invite/page.tsx
// Read ?token= from URL
// If not logged in: redirect to /login?redirect=/accept-invite?token=X
// If logged in: POST /api/proxy/invitations/accept?token=X
// On success: redirect to org dashboard
// On error: show "Invalid or expired invitation"
```

### Step 5: List invitations in org settings

Show pending invitations with:
- Email, role, invited by, expires at
- "Cancel" button to revoke invitation

### Step 6: API client updates

```typescript
// api-client.ts additions
orgs: {
    ...existing,
    listInvitations: (orgId: string) =>
        apiFetch<{ invitations: OrgInvitation[] }>(`/orgs/${orgId}/invitations`),
    createInvitation: (orgId: string, body: { email: string; role: string }) =>
        apiFetch<OrgInvitation>(`/orgs/${orgId}/invitations`, { method: 'POST', body: JSON.stringify(body) }),
    cancelInvitation: (orgId: string, inviteId: string) =>
        apiFetch<void>(`/orgs/${orgId}/invitations/${inviteId}`, { method: 'DELETE' }),
    acceptInvitation: (token: string) =>
        apiFetch<void>(`/invitations/accept?token=${token}`, { method: 'POST' }),
}
```

## Todo Checklist

- [ ] Create migration 000040
- [ ] Add `POST /orgs/{id}/invitations` endpoint
- [ ] Add `GET /orgs/{id}/invitations` endpoint
- [ ] Add `DELETE /orgs/{id}/invitations/{inviteId}` endpoint
- [ ] Add `POST /invitations/accept` endpoint
- [ ] Send invitation email with link
- [ ] Create invite-member-dialog component
- [ ] Create accept-invite page
- [ ] Show pending invitations in org settings
- [ ] Add invitation API methods to api-client.ts
- [ ] Test: invite -> email received -> accept -> member added
- [ ] Test: expired invitation rejected
- [ ] Test: email mismatch rejected
- [ ] Test: cancel invitation
- [ ] Test: duplicate invitation for same email prevented

## Success Criteria

- Org admin/owner can invite members by email
- Invitee receives email with accept link
- Accepting adds user to org with correct role
- Invitations expire after 7 days
- Duplicate pending invitations prevented
- Cancel invitation works

## Security Considerations

- Token hashed in DB, raw only in email
- 7-day expiry
- Email must match: invitation email == accepting user's email
- Only admin/owner can invite (RBAC check)
- Only member/admin roles can be assigned (not owner)
- Unique constraint prevents invitation spam to same email
