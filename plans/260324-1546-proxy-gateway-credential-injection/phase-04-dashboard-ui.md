---
phase: 4
title: "Dashboard UI — Proxy Route Management"
status: pending
priority: P2
effort: 6h
---

# Phase 4: Dashboard UI

## Overview

Add dashboard pages for managing proxy routes and endpoint limits per agent. Extends the existing agent detail page.

## Context Links
- [Agent detail page](../../dashboard/src/app/(dashboard)/agents/[id]/page.tsx)
- [API client](../../dashboard/src/lib/api-client.ts)
- [API types](../../dashboard/src/types/api.ts)

## UI Components

### Agent Detail Page — New "Proxy Routes" Tab

Location: `dashboard/src/app/(dashboard)/agents/[id]/page.tsx` (extend existing)

```
┌─ Agent: my-openai-agent ──────────────────────────┐
│ [Overview] [Tokens] [Proxy Routes] [Rate Limits]  │
│                                                     │
│ Proxy Routes                          [+ Add Route] │
│ ┌─────────────────────────────────────────────────┐ │
│ │ api.openai.com /v1/*                            │ │
│ │ Injection: Bearer header                        │ │
│ │ Placeholder: valt_pk_7Kx9mPqR2sT4vW6y  [Copy]  │ │
│ │ Secret: openai-prod-key                         │ │
│ │ Status: ● Enabled          [Edit] [Delete]      │ │
│ └─────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────┐ │
│ │ api.stripe.com /*                               │ │
│ │ Injection: Header (x-api-key)                   │ │
│ │ Placeholder: valt_pk_9Ax3bCd5eF7g  [Copy]       │ │
│ │ Secret: stripe-secret-key                       │ │
│ │ Status: ● Enabled          [Edit] [Delete]      │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

### Add/Edit Route Dialog

```
┌─ Add Proxy Route ──────────────────────┐
│                                         │
│ Host Pattern:  [api.openai.com      ]  │
│ Path Pattern:  [/v1/*               ]  │
│ Secret:        [▾ Select secret     ]  │
│ Injection Type: ○ Bearer  ○ Header  ○ Query │
│ Header/Param Name: [Authorization   ]  │
│ Format:        [Bearer {value}      ]  │
│                                         │
│              [Cancel]  [Save]           │
└─────────────────────────────────────────┘
```

### Agent Quick Setup Card

Show after first route is created:

```
┌─ Quick Setup ───────────────────────────┐
│ Configure your agent with these values: │
│                                         │
│ HTTPS_PROXY=http://valt:10256           │
│ OPENAI_API_KEY=valt_pk_7Kx9mPqR2sT4vW6y│
│                                         │
│ Your agent will never see real keys.    │
│                          [Copy All]     │
└─────────────────────────────────────────┘
```

### Endpoint Rate Limits Tab

```
┌─ Rate Limits ─────────────────────────────────────┐
│                                    [+ Add Limit]   │
│ ┌───────────────────────────────────────────────┐  │
│ │ api.openai.com /v1/*     30 rpm    [Edit][Del]│  │
│ │ api.stripe.com /*        ■ BLOCKED [Edit][Del]│  │
│ └───────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────┘
```

## Files to Create

| File | Purpose |
|------|---------|
| `dashboard/src/app/(dashboard)/agents/[id]/proxy-routes.tsx` | Proxy routes tab component |
| `dashboard/src/app/(dashboard)/agents/[id]/endpoint-limits.tsx` | Rate limits tab component |

## Files to Modify

| File | Change |
|------|--------|
| `dashboard/src/app/(dashboard)/agents/[id]/page.tsx` | Add tabs for proxy routes + rate limits |
| `dashboard/src/lib/api-client.ts` | Add proxy route + endpoint limit API functions |
| `dashboard/src/types/api.ts` | Add `ProxyRoute`, `EndpointLimit` types |

## API Client Functions

```typescript
// api-client.ts additions
export async function listProxyRoutes(agentId: string): Promise<ProxyRoute[]>
export async function createProxyRoute(data: CreateProxyRouteInput): Promise<ProxyRoute>
export async function updateProxyRoute(id: string, data: UpdateProxyRouteInput): Promise<ProxyRoute>
export async function deleteProxyRoute(id: string): Promise<void>
export async function listEndpointLimits(agentId: string): Promise<EndpointLimit[]>
export async function createEndpointLimit(data: CreateEndpointLimitInput): Promise<EndpointLimit>
export async function deleteEndpointLimit(id: string): Promise<void>
```

## Success Criteria
- [ ] Proxy routes tab on agent detail page
- [ ] CRUD for proxy routes via dashboard
- [ ] Placeholder key visible with copy button
- [ ] Quick setup card with proxy config
- [ ] Endpoint rate limits tab with CRUD
- [ ] Blocked endpoints shown with visual indicator
