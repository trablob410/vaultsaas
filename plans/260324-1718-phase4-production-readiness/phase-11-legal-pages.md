---
phase: "4.11"
title: "Terms of Service + Privacy Policy"
priority: P2
effort: 2h
status: pending
---

# Phase 4.11: Legal Pages (ToS + Privacy Policy)

## Context Links
- `dashboard/src/app/(auth)/login/page.tsx` -- login page (add footer links)
- `dashboard/src/app/(dashboard)/settings/` -- settings pages

## Overview

No legal pages. SaaS product handling secrets needs ToS and Privacy Policy for user trust and legal compliance.

## Requirements

### Functional
- `/terms` page with Terms of Service
- `/privacy` page with Privacy Policy
- Links in login page footer and dashboard settings
- Registration consent checkbox: "I agree to Terms and Privacy Policy"

### Non-functional
- Static content, server-rendered
- Dark theme consistent with rest of app

## Related Code Files

### Modify
- `dashboard/src/app/(auth)/login/page.tsx` -- add ToS/Privacy links in footer, consent checkbox on register

### Create
- `dashboard/src/app/(marketing)/terms/page.tsx`
- `dashboard/src/app/(marketing)/privacy/page.tsx`
- `dashboard/src/app/(marketing)/layout.tsx` -- minimal layout for legal pages

## Implementation Steps

### Step 1: Create (marketing) route group

Minimal layout without sidebar/header:
```tsx
// dashboard/src/app/(marketing)/layout.tsx
export default function MarketingLayout({ children }) {
  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b p-4">
        <a href="/" className="font-semibold">Valt</a>
      </nav>
      <main className="max-w-3xl mx-auto px-4 py-8">
        {children}
      </main>
    </div>
  )
}
```

### Step 2: Terms of Service page

Key sections:
1. **Acceptance** -- By using Valt, you agree to these terms
2. **Service Description** -- MCP-native secret vault for AI agents
3. **Account Responsibilities** -- Password security, authorized use
4. **Data & Encryption** -- Client-side encryption, zero-knowledge for secret values
5. **Prohibited Use** -- No illegal content, no abuse
6. **Service Availability** -- Best-effort uptime, no SLA guarantee for free tier
7. **Limitation of Liability** -- Standard limitation clause
8. **Termination** -- Right to terminate accounts
9. **Changes to Terms** -- May update with notice
10. **Contact** -- Support email

### Step 3: Privacy Policy page

Key sections:
1. **Information Collected** -- Email, name (from OAuth), usage metrics
2. **How We Use Data** -- Authentication, service operation, analytics
3. **Data Encryption** -- Secret values encrypted client-side (zero-knowledge)
4. **Data Storage** -- PostgreSQL + MinIO, hosted on [provider]
5. **Data Retention** -- Account data retained while active, deleted on request
6. **Third Parties** -- Stripe (billing), Google (OAuth), notification services
7. **Cookies** -- httpOnly auth cookies, no tracking cookies
8. **User Rights** -- Access, export, delete your data
9. **Security** -- AES-256 encryption, audit logging, HTTPS
10. **Contact** -- Privacy inquiries email

### Step 4: Add consent to registration

In login page register mode, add checkbox:
```tsx
{mode === 'register' && (
  <label className="flex items-center gap-2 text-sm text-muted-foreground">
    <input type="checkbox" required checked={agreed} onChange={...} />
    I agree to the <a href="/terms" className="text-primary hover:underline">Terms</a> and
    <a href="/privacy" className="text-primary hover:underline">Privacy Policy</a>
  </label>
)}
```

### Step 5: Footer links on login page

Add below the existing footer text:
```tsx
<div className="flex gap-4 text-xs text-muted-foreground">
  <a href="/terms" className="hover:underline">Terms</a>
  <a href="/privacy" className="hover:underline">Privacy</a>
</div>
```

## Todo Checklist

- [ ] Create (marketing) route group with minimal layout
- [ ] Write Terms of Service content
- [ ] Write Privacy Policy content
- [ ] Create /terms page
- [ ] Create /privacy page
- [ ] Add consent checkbox on registration form
- [ ] Add footer links on login page
- [ ] Add links in dashboard settings
- [ ] Review legal content (consider professional legal review later)

## Success Criteria

- /terms and /privacy pages accessible
- Registration requires consent checkbox
- Links visible on login page
- Content covers key SaaS legal requirements

## Notes

- Initial content can be template-based; professional legal review recommended before scale
- Consider adding a consent log (table exists: `user_consent_logs`, migration 000011)
- Update consent log on registration with ToS acceptance timestamp
