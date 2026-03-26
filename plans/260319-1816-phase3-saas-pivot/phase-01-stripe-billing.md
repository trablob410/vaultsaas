---
phase: "3.1"
title: "Stripe Billing"
priority: P1
status: pending
effort: 8h
---

# Phase 3.1: Stripe Billing

## Context Links
- [Org service](../../server/internal/org/service.go) — existing org CRUD, `plan` column
- [Usage tracker](../../server/internal/usage/tracker.go) — existing free tier limits
- [Usage handler](../../server/internal/usage/handler.go) — GET /orgs/{id}/usage
- [Config](../../server/internal/config/config.go) — env var pattern
- [Main](../../server/cmd/server/main.go) — service wiring
- [Upgrade page](../../dashboard/src/app/(dashboard)/settings/upgrade/page.tsx) — static placeholder

## Overview

Add Stripe Checkout + Customer Portal + webhook handling. Per-seat pricing: quantity = org member count. Three plans: free (default), pro ($19/user/mo), team ($39/user/mo). Paid plans unlock higher usage limits.

## Key Insights
- `organizations.plan` column already exists (VARCHAR(50), default 'free')
- Usage tracker already gates on plan != 'free' (returns allowed=true for paid plans)
- Need 3 new columns on orgs: `stripe_customer_id`, `stripe_subscription_id`, `plan_seats`
- Dashboard upgrade page exists as static placeholder — wire to real API

## Requirements

### Functional
- POST /billing/checkout-session: create Stripe Checkout Session, return URL
- POST /billing/portal: create Stripe Customer Portal session, return URL
- POST /webhooks/stripe: handle Stripe webhook events
- Webhook events: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`, `invoice.payment_failed`, `invoice.paid`
- On checkout complete: update org plan + stripe IDs
- On subscription updated: sync plan + seat count
- On subscription deleted: downgrade to free
- Per-seat: quantity = COUNT of org_memberships for the org

### Non-Functional
- Stripe webhook signature verification (HMAC)
- Idempotent webhook processing (check current state before update)
- No plaintext Stripe keys in logs

## Architecture

```
Dashboard → POST /api/proxy/billing/checkout-session → Go billing handler
  → stripe.CheckoutSession.New({mode: subscription, line_items: [{price: PRO_PRICE_ID, quantity: seat_count}]})
  → return {url: session.URL}
Dashboard redirects to Stripe Checkout
Stripe → POST /webhooks/stripe → Go billing webhook handler
  → verify signature → update org plan/stripe_customer_id/stripe_subscription_id
```

## Related Code Files

### Create
- `server/internal/billing/service.go` — Stripe API wrapper
- `server/internal/billing/handler.go` — HTTP handlers
- `server/internal/billing/webhook.go` — webhook event processing
- `server/internal/database/migrations/000030_stripe_billing.up.sql`
- `server/internal/database/migrations/000030_stripe_billing.down.sql`

### Modify
- `server/internal/config/config.go` — add Stripe env vars
- `server/cmd/server/main.go` — wire billing routes
- `server/internal/usage/tracker.go` — add plan-specific limits (pro/team)
- `server/internal/usage/handler.go` — return plan-specific limits in response
- `server/internal/org/service.go` — add UpdatePlan method
- `server/go.mod` — add stripe-go/v84
- `dashboard/src/app/(dashboard)/settings/upgrade/page.tsx` — wire to real API

## Implementation Steps

### 1. DB Migration (000030)

```sql
-- 000030_stripe_billing.up.sql
ALTER TABLE organizations
  ADD COLUMN stripe_customer_id VARCHAR(255),
  ADD COLUMN stripe_subscription_id VARCHAR(255),
  ADD COLUMN plan_seats INTEGER NOT NULL DEFAULT 1;

CREATE UNIQUE INDEX idx_organizations_stripe_customer ON organizations(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;
```

```sql
-- 000030_stripe_billing.down.sql
ALTER TABLE organizations
  DROP COLUMN IF EXISTS stripe_customer_id,
  DROP COLUMN IF EXISTS stripe_subscription_id,
  DROP COLUMN IF EXISTS plan_seats;
```

### 2. Config — Add Stripe env vars

Add to `config.Config`:
```go
StripeSecretKey    string `envconfig:"STRIPE_SECRET_KEY" default:""`
StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET" default:""`
StripePriceProID   string `envconfig:"STRIPE_PRO_PRICE_ID" default:""`
StripePriceTeamID  string `envconfig:"STRIPE_TEAM_PRICE_ID" default:""`
```

### 3. Go dependency

```bash
cd server && go get github.com/stripe/stripe-go/v84
```

### 4. billing/service.go (~80 lines)

```go
package billing

type Service struct {
    pool       *pgxpool.Pool
    priceProID string
    priceTeamID string
}

func NewService(pool *pgxpool.Pool, secretKey, priceProID, priceTeamID string) *Service
func (s *Service) GetOrCreateCustomer(ctx, orgID, orgName, ownerEmail string) (string, error)
func (s *Service) CreateCheckoutSession(ctx, orgID, plan, successURL, cancelURL string) (string, error)
func (s *Service) CreatePortalSession(ctx, orgID, returnURL string) (string, error)
func (s *Service) SeatCount(ctx, orgID string) (int, error) // COUNT org_memberships
func (s *Service) UpdateOrgBilling(ctx, orgID, customerID, subscriptionID, plan string, seats int) error
func (s *Service) DowngradeToFree(ctx, orgID string) error
```

Key logic in `CreateCheckoutSession`:
- Call `SeatCount` to get quantity
- Map plan string to price ID (pro→priceProID, team→priceTeamID)
- `stripe.CheckoutSession.New` with `mode=subscription`, `line_items=[{price, quantity}]`
- Store `orgID` in `metadata` and `client_reference_id`
- Return `session.URL`

`GetOrCreateCustomer`:
- Check if org has `stripe_customer_id`; if so, return it
- Otherwise call `stripe.Customer.New` with org name + owner email
- UPDATE organizations SET stripe_customer_id WHERE id = orgID
- Return customer ID

### 5. billing/handler.go (~80 lines)

```go
type Handler struct { svc *Service }

func (h *Handler) CreateCheckout(w, r) // POST /billing/checkout-session
func (h *Handler) CreatePortal(w, r)   // POST /billing/portal
```

`CreateCheckout`:
- Read `{plan, success_url, cancel_url}` from body
- Validate plan is "pro" or "team"
- Get orgID from user's first org (query org_memberships WHERE user_id AND role='owner')
- Call `svc.CreateCheckoutSession`
- Return `{url: sessionURL}`

`CreatePortal`:
- Get orgID from user's org ownership
- Call `svc.CreatePortalSession`
- Return `{url: portalURL}`

### 6. billing/webhook.go (~100 lines)

```go
func (h *Handler) HandleWebhook(w, r) // POST /webhooks/stripe
```

Logic:
- Read raw body, verify signature via `webhook.ConstructEvent(body, sig, webhookSecret)`
- Switch on `event.Type`:
  - `checkout.session.completed`: extract `client_reference_id` (orgID), subscription ID → `UpdateOrgBilling`
  - `customer.subscription.updated`: extract customer ID → find org → update plan + seats from subscription items
  - `customer.subscription.deleted`: extract customer ID → find org → `DowngradeToFree`
  - `invoice.payment_failed`: log warning (future: notify org owner)
  - `invoice.paid`: no-op (subscription.updated handles plan sync)
- Always return 200 (Stripe retries on non-2xx)

### 7. Wire in main.go

```go
// After usageTracker init:
billingSvc := billing.NewService(pool, cfg.StripeSecretKey, cfg.StripePriceProID, cfg.StripePriceTeamID)
billingHandler := billing.NewHandler(billingSvc)

// Public route (Stripe webhook — verified via signing secret):
r.Post("/webhooks/stripe", billingHandler.HandleWebhook)

// Inside authenticated group:
r.Post("/billing/checkout-session", billingHandler.CreateCheckout)
r.Post("/billing/portal", billingHandler.CreatePortal)
```

### 8. Update usage tracker limits

In `tracker.go`, update `limitForMetric` to accept plan:

```go
func limitsForPlan(plan string) map[string]int {
    switch plan {
    case "pro":
        return map[string]int{"secrets_count": 500, "agents_count": 25, "requests_today": 10000}
    case "team":
        return map[string]int{"secrets_count": -1, "agents_count": -1, "requests_today": 50000} // -1 = unlimited
    default: // free
        return map[string]int{"secrets_count": 50, "agents_count": 3, "requests_today": 1000}
    }
}
```

Update `CheckLimit` to use plan-specific limits. Update `handler.go` `getUsage` to return plan-specific limits.

### 9. Update org service

Add `UpdatePlan(ctx, orgID, plan string) error` — simple UPDATE query.

### 10. Dashboard — Wire upgrade page

Replace static upgrade page:
- Fetch `GET /api/proxy/orgs/{orgID}/usage` on mount
- Display real usage bars with live numbers
- Plan badge from API response
- "Upgrade to Pro" button → `POST /api/proxy/billing/checkout-session` with `{plan: "pro", success_url, cancel_url}`
- Redirect to returned Stripe URL
- "Manage Billing" button (visible when plan != free) → `POST /api/proxy/billing/portal`

## Todo

- [ ] Create migration 000030
- [ ] Add Stripe config env vars
- [ ] `go get stripe-go/v84`
- [ ] Implement billing/service.go
- [ ] Implement billing/handler.go
- [ ] Implement billing/webhook.go
- [ ] Wire routes in main.go
- [ ] Update usage tracker with plan-specific limits
- [ ] Add UpdatePlan to org service
- [ ] Wire dashboard upgrade page to real API
- [ ] Add .env.example entries for STRIPE_* vars
- [ ] Unit tests: webhook signature verification, plan mapping, seat counting

## Success Criteria
- `POST /billing/checkout-session` returns Stripe Checkout URL
- Stripe webhook updates org plan on subscription changes
- Usage limits reflect plan tier (pro/team/free)
- Dashboard shows real usage + Stripe Checkout redirect
- Downgrade to free on subscription cancellation

## Security Considerations
- Stripe webhook signature MUST be verified before processing
- `STRIPE_SECRET_KEY` never logged or returned in API responses
- Checkout session uses `client_reference_id` = orgID (tamper-proof — set server-side)
- Only org owners can create checkout/portal sessions
- Webhook endpoint is public but signature-verified (same pattern as Slack webhook)

## Risk Assessment
- **Stripe API rate limits**: unlikely at current scale; no mitigation needed
- **Webhook replay**: Stripe sends unique event IDs; could add idempotency table later (YAGNI for now)
- **Seat count drift**: reconcile on `subscription.updated` webhook; also on member add/remove (future enhancement)
