---
phase: "4.3"
title: "Stripe Activation"
priority: P0
effort: 3h
status: pending
---

# Phase 4.3: Stripe Activation

## Context Links
- `server/internal/billing/` -- existing Stripe integration (handler.go, service.go)
- `server/internal/config/config.go:73-77` -- Stripe env vars already defined
- `docker-compose.prod.yml:94-98` -- Stripe env vars passed to server
- `dashboard/src/app/(dashboard)/settings/upgrade/page.tsx` -- upgrade UI exists
- `dashboard/src/lib/api-client.ts:186-191` -- billing API client exists

## Overview

Stripe code is fully implemented but not activated. Need to create Stripe account, configure products/prices, set env vars on VPS, and test end-to-end.

## Key Insights

- `billing/service.go` handles checkout session creation, customer portal, webhook processing
- `billing/handler.go` exposes: `POST /billing/checkout-session`, `POST /billing/portal`, `POST /billing/webhook`
- Usage tracking in `server/internal/usage/` enforces plan limits
- Plans: `free` (default), `pro`, `team` -- prices set via `STRIPE_PRO_PRICE_ID`, `STRIPE_TEAM_PRICE_ID`
- Webhook handles: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`

## Requirements

### Functional
- Create Stripe account with products matching free/pro/team tiers
- Configure webhook endpoint pointing to `https://valt.turbo.ai.vn/api/v1/billing/webhook`
- Test full checkout flow: click Upgrade -> Stripe checkout -> subscription created -> plan updated in DB
- Test Customer Portal: manage billing, cancel subscription
- Test webhook events fire correctly

### Non-functional
- Stripe test mode first, then switch to live mode
- Webhook signature verification must work (STRIPE_WEBHOOK_SECRET)

## Implementation Steps

### Step 1: Stripe Dashboard Setup
1. Create Stripe account at stripe.com
2. Create 2 products:
   - **Valt Pro** -- monthly recurring, set price (e.g., $19/mo)
   - **Valt Team** -- monthly recurring, set price (e.g., $49/mo)
3. Note the Price IDs (e.g., `price_xxx`)
4. Configure webhook:
   - URL: `https://valt.turbo.ai.vn/api/v1/billing/webhook`
   - Events: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`
5. Note the Webhook Signing Secret

### Step 2: VPS Environment Configuration
SSH into VPS, update `.env`:
```bash
STRIPE_SECRET_KEY=sk_test_xxx      # or sk_live_xxx for production
STRIPE_WEBHOOK_SECRET=whsec_xxx
STRIPE_PRO_PRICE_ID=price_xxx
STRIPE_TEAM_PRICE_ID=price_xxx
```

### Step 3: Test with Stripe CLI (optional, local)
```bash
stripe listen --forward-to localhost:8080/api/v1/billing/webhook
stripe trigger checkout.session.completed
```

### Step 4: End-to-End Testing on Production
1. Login to dashboard
2. Go to Settings > Upgrade
3. Click "Upgrade to Pro"
4. Complete Stripe checkout with test card (4242 4242 4242 4242)
5. Verify: webhook received, subscription created in DB, usage limits updated
6. Go to Settings > Manage Billing (Customer Portal)
7. Cancel subscription
8. Verify: webhook received, plan reverted to free

### Step 5: Update Deployment Guide
Document Stripe setup steps in `docs/deployment-guide.md`.

## Todo Checklist

- [ ] Create Stripe account
- [ ] Create Pro and Team products with prices
- [ ] Configure webhook endpoint in Stripe dashboard
- [ ] Set env vars on VPS (.env file)
- [ ] Redeploy server with new env vars
- [ ] Test checkout flow (test card)
- [ ] Test webhook delivery (check Stripe dashboard for 200 responses)
- [ ] Test Customer Portal access
- [ ] Test subscription cancellation flow
- [ ] Test usage limit enforcement after plan change
- [ ] Switch to live mode when ready
- [ ] Document setup in deployment guide

## Success Criteria

- Checkout flow works end-to-end on production
- Webhooks process correctly (verified in Stripe dashboard)
- Plan changes reflect in `organizations.plan` column
- Usage limits enforce correctly per plan
- Customer Portal accessible

## Security Considerations

- STRIPE_SECRET_KEY must never be committed to git
- Webhook signature verification prevents spoofed events
- Stripe secret key should be sk_test_ during testing, sk_live_ for production
- Customer email from Stripe must match org owner email
