# Phase 1: Stripe Account + Products Setup

**Priority:** P0 | **Effort:** 2h | **Status:** pending

Manual steps to create Stripe account and payment products.

## Steps

1. Go to https://stripe.com and create account (or use existing)
2. Verify email and complete KYC (if new account)
3. In Stripe Dashboard, create 2 products:
   - **Product Name:** "Valt Pro"
   - **Product Name:** "Valt Team"
4. For each product, create a price:
   - Pro: $19/month (USD), recurring monthly
   - Team: $39/month (USD), recurring monthly
5. Copy **Price IDs** for both (format: `price_...`)
6. Create webhook endpoint:
   - URL: `https://valt.turbo.ai.vn/api/v1/webhooks/stripe`
   - Events to listen: `charge.succeeded`, `charge.failed`, `customer.subscription.updated`, `customer.subscription.deleted`
7. Copy **Webhook Secret** (starts with `whsec_`)

## VPS Setup

SSH into VPS, edit `.env`:

```bash
STRIPE_SECRET_KEY=sk_live_... (or sk_test_...)
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRO_PRICE_ID=price_...
STRIPE_TEAM_PRICE_ID=price_...
```

Restart server:

```bash
docker compose restart server
```

## Verification

- Dashboard loads `/settings/upgrade` page without errors
- Clicking "Upgrade to Pro" redirects to Stripe checkout
- Webhook logs show incoming events in Stripe Dashboard

## Notes

- Use **test mode** first to validate flow (test card: 4242 4242 4242 4242)
- Switch to **live mode** only when ready for beta users
- Store webhook secret in `.env`, never in code
