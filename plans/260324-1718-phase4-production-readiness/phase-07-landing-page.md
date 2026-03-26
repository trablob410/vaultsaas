---
phase: "4.7"
title: "Landing Page"
priority: P1
effort: 4h
status: pending
---

# Phase 4.7: Landing Page

## Context Links
- `dashboard/src/app/(auth)/login/page.tsx` -- current login page
- `dashboard/src/app/(dashboard)/layout.tsx` -- dashboard layout
- `dashboard/src/app/(dashboard)/settings/upgrade/page.tsx` -- pricing reference

## Overview

No marketing/landing page exists. Visitors hitting the root URL see dashboard (if logged in) or login. Need a landing page explaining the product for first-time visitors.

## Requirements

### Functional
- Landing page at `/` for unauthenticated users
- Sections: Hero, Features, How It Works, Pricing, CTA
- "Get Started" / "Sign In" buttons -> /login
- Logged-in users visiting `/` see dashboard (existing behavior)

### Non-functional
- Mobile responsive
- Dark theme consistent with dashboard
- Fast load (static content, no API calls)
- SEO-friendly (server-rendered)

## Architecture

```
Route structure:
  / (root)
    -> if authenticated: redirect to /dashboard (or show dashboard)
    -> if unauthenticated: show landing page

Option A: root page.tsx checks session, renders landing or redirects
Option B: (marketing) route group at / with middleware check

Going with Option A (simpler):
  - dashboard/src/app/page.tsx -- landing page (SSR)
  - Check session: if logged in, redirect to /secrets (first dashboard page)
  - If not, render landing page
```

## Related Code Files

### Modify
- `dashboard/src/app/page.tsx` -- replace current root page with landing/redirect logic

### Create
- `dashboard/src/app/page.tsx` -- landing page component (or refactor existing)
- `dashboard/src/components/landing/hero-section.tsx`
- `dashboard/src/components/landing/features-section.tsx`
- `dashboard/src/components/landing/pricing-section.tsx`
- `dashboard/src/components/landing/footer-section.tsx`

## Implementation Steps

### Step 1: Root page with auth check

```tsx
// dashboard/src/app/page.tsx
import { getSession } from '@/lib/auth'
import { redirect } from 'next/navigation'
import LandingPage from '@/components/landing/landing-page'

export default async function RootPage() {
  const session = await getSession()
  if (session) redirect('/secrets')
  return <LandingPage />
}
```

### Step 2: Hero section

```tsx
// Hero: "MCP-Native Secret Vault for AI Agents"
// Subtitle: "Human-in-the-loop approval for every secret access.
//           Your agents get credentials. You stay in control."
// CTA: "Get Started Free" -> /login
// Secondary: "View Documentation"
```

### Step 3: Features section

Key features to highlight:
1. **Approval Workflow** -- Every secret access requires human approval via email, Slack, or Telegram
2. **MCP Integration** -- Works natively with Claude, GPT, and any MCP-compatible agent
3. **Dynamic Secrets** -- Auto-rotating database credentials with TTL
4. **CLI + Dashboard** -- Manage secrets from terminal or web UI
5. **Audit Trail** -- Cryptographic hash chain audit log for compliance
6. **Team Collaboration** -- RBAC with org > workspace > project hierarchy

### Step 4: Pricing section

| | Free | Pro | Team |
|---|---|---|---|
| Secrets | 10 | 100 | Unlimited |
| Projects | 1 | 10 | Unlimited |
| Members | 1 | 5 | 25 |
| Notification channels | Email | Email, Slack, Telegram | All + Webhooks |
| Price | $0 | $19/mo | $49/mo |

(Adjust pricing to match actual Stripe products from Phase 4.3)

### Step 5: Footer

- Links: Documentation, Terms, Privacy, Contact
- Copyright
- Social links (GitHub if public)

### Step 6: Responsive + dark theme

- Use existing Tailwind config and shadcn/ui design tokens
- Test on mobile, tablet, desktop breakpoints
- Ensure dark theme matches dashboard aesthetic

## Todo Checklist

- [ ] Create root page.tsx with auth check + landing render
- [ ] Build hero section component
- [ ] Build features section component
- [ ] Build pricing section component
- [ ] Build footer section component
- [ ] Mobile responsive testing
- [ ] Dark theme consistency check
- [ ] Add "Get Started" -> /login navigation
- [ ] Add proper meta tags for SEO (title, description, og:image)
- [ ] Test: unauthenticated user sees landing page
- [ ] Test: authenticated user redirected to dashboard

## Success Criteria

- Unauthenticated users see professional landing page
- Authenticated users redirected to dashboard
- Mobile responsive, dark theme
- Clear value proposition and pricing
- CTA leads to registration

## Notes

- Keep content concise. Can iterate on copy later.
- Consider adding an animated demo/screenshot of the approval flow
- Pricing must match actual Stripe products (coordinate with Phase 4.3)
