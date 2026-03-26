---
phase: "3.6"
title: "Plan Limits UI"
priority: P1
status: pending
effort: 4h
---

# Phase 3.6: Plan Limits UI

## Context Links
- [Upgrade page](../../dashboard/src/app/(dashboard)/settings/upgrade/page.tsx) — static placeholder
- [Usage handler](../../server/internal/usage/handler.go) — GET /orgs/{id}/usage API
- [Usage tracker](../../server/internal/usage/tracker.go) — limits + current counts
- [API client](../../dashboard/src/lib/api-client.ts) — proxy fetch pattern
- [Proxy route](../../dashboard/src/app/api/proxy/[...path]/route.ts)

## Overview

Wire the existing static /settings/upgrade page to the real GET /orgs/{org_id}/usage API. Show live usage bars, plan badges, and enforce paywall in UI. After Phase 3.1 (Stripe), add Checkout button; this phase can start independently using the existing usage API.

## Requirements

### Functional
- Fetch real usage data from GET /api/proxy/orgs/{orgID}/usage on page mount
- Display live usage bars: secrets, agents, requests/day
- Show plan badge (free/pro/team) from API response
- Plan-specific limits shown (not hardcoded free tier)
- "Upgrade to Pro" button: if Phase 3.1 done, triggers Stripe Checkout; otherwise mailto fallback
- "Manage Billing" button visible for paid plans (triggers Stripe Portal)
- Disable "Create Secret" button when at limit (UI enforcement)
- Show warning banner when usage > 80% of limit

### Non-Functional
- Page loads within 1s (single API call)
- Graceful fallback if usage API fails

## Architecture

```
/settings/upgrade (client component)
  → useEffect: fetch /api/proxy/orgs/{orgID}/usage
  → response: { plan, usage: {secrets_count, agents_count, requests_today}, limits: {...} }
  → render UsageBars with real data
  → render plan badge
  → render CTA buttons based on plan
```

## Related Code Files

### Modify
- `dashboard/src/app/(dashboard)/settings/upgrade/page.tsx` — main rewrite
- `dashboard/src/lib/api-client.ts` — add usage API method
- `server/internal/usage/handler.go` — return plan-specific limits (depends on 3.1)
- `server/internal/usage/tracker.go` — add limitsForPlan function (depends on 3.1)

### Create
- `dashboard/src/components/usage-bar.tsx` — extracted reusable component
- `dashboard/src/components/plan-badge.tsx` — plan badge component

## Implementation Steps

### 1. Backend: Plan-specific limits in usage response

Update `handler.go` `getUsage` to return plan-specific limits:

```go
func (h *Handler) getUsage(w http.ResponseWriter, r *http.Request) {
    orgID := chi.URLParam(r, "org_id")
    plan, _ := h.tracker.GetOrgPlan(r.Context(), orgID)
    secrets, _ := h.tracker.GetCurrent(r.Context(), orgID, "secrets_count")
    agents, _ := h.tracker.GetCurrent(r.Context(), orgID, "agents_count")
    requests, _ := h.tracker.GetCurrent(r.Context(), orgID, "requests_today")
    limits := limitsForPlan(plan)
    // Return limits map with -1 meaning unlimited
    resp := map[string]interface{}{
        "plan": plan,
        "usage": map[string]int{...},
        "limits": limits,
    }
}
```

This depends on `limitsForPlan` from Phase 3.1. If 3.1 not done yet, use free-tier-only constants.

### 2. API client — Add usage method

In `api-client.ts`:
```typescript
usage: {
  get: (orgId: string) => fetchAPI<{
    plan: string
    usage: { secrets_count: number; agents_count: number; requests_today: number }
    limits: { secrets_count: number; agents_count: number; requests_today: number }
  }>(`/orgs/${orgId}/usage`),
},
```

### 3. Extract UsageBar component

Move the existing `UsageBar` from upgrade page to `components/usage-bar.tsx`:

```tsx
interface UsageBarProps {
  label: string
  current: number
  limit: number // -1 = unlimited
  warningThreshold?: number // default 0.8
}

export function UsageBar({ label, current, limit, warningThreshold = 0.8 }: UsageBarProps) {
  const isUnlimited = limit < 0
  const pct = isUnlimited ? 0 : Math.min((current / limit) * 100, 100)
  const isWarning = !isUnlimited && current / limit >= warningThreshold
  // Render with warning color when isWarning
}
```

### 4. PlanBadge component

```tsx
const PLAN_COLORS: Record<string, string> = {
  free: 'bg-secondary text-secondary-foreground',
  pro: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  team: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
}

export function PlanBadge({ plan }: { plan: string }) {
  return <span className={cn('px-2 py-0.5 rounded-full text-xs font-medium', PLAN_COLORS[plan] ?? PLAN_COLORS.free)}>
    {plan.charAt(0).toUpperCase() + plan.slice(1)}
  </span>
}
```

### 5. Rewrite upgrade page

Replace static content with data-fetching client component:

```tsx
'use client'

export default function UpgradePage() {
  const [data, setData] = useState<UsageData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Get orgID from context/cookie/first-org-fetch
    api.usage.get(orgId).then(setData).finally(() => setLoading(false))
  }, [])

  if (loading) return <Skeleton />
  if (!data) return <ErrorState />

  return (
    <div className="space-y-6 max-w-2xl">
      <div className="flex items-center justify-between">
        <h1>Plan & Usage</h1>
        <PlanBadge plan={data.plan} />
      </div>

      {/* Usage bars */}
      <Card>
        <UsageBar label="Secrets" current={data.usage.secrets_count} limit={data.limits.secrets_count} />
        <UsageBar label="Agents" current={data.usage.agents_count} limit={data.limits.agents_count} />
        <UsageBar label="API Requests (today)" current={data.usage.requests_today} limit={data.limits.requests_today} />
      </Card>

      {/* CTA */}
      {data.plan === 'free' && <UpgradeCard onUpgrade={handleCheckout} />}
      {data.plan !== 'free' && <ManageBillingCard onManage={handlePortal} />}
    </div>
  )
}
```

### 6. OrgID resolution

The upgrade page needs the current org ID. Two options:
- **Option A**: Read from URL/context if org selector exists in sidebar
- **Option B**: Fetch user's first org via `GET /api/proxy/orgs` and use the first result

Use whichever pattern the existing dashboard already follows. The sidebar likely has org context.

### 7. Paywall enforcement in secrets page

In the secrets list page, conditionally disable "Create Secret" button:
- Fetch usage on page load (or use a shared context/hook)
- If `usage.secrets_count >= limits.secrets_count` and `limits.secrets_count > 0`: disable button, show tooltip "Upgrade to create more secrets"

## Todo

- [ ] Update usage handler with plan-specific limits
- [ ] Add usage API method to api-client.ts
- [ ] Extract UsageBar component
- [ ] Create PlanBadge component
- [ ] Rewrite upgrade page with real data
- [ ] Add Stripe Checkout integration (post Phase 3.1)
- [ ] Add paywall enforcement on secrets creation
- [ ] Add warning banners at 80% usage

## Success Criteria
- Upgrade page shows real usage numbers from API
- Plan badge reflects actual org plan
- Usage bars show correct current/limit ratios
- "Upgrade" button works (Stripe or mailto fallback)
- UI disables creation when at limit

## Security Considerations
- Usage data scoped to org via org_id in API
- No sensitive data exposed in usage response
- Paywall is UI-only — backend enforcement already exists in usage middleware
