DROP INDEX IF EXISTS idx_organizations_stripe_customer;

ALTER TABLE organizations
  DROP COLUMN IF EXISTS stripe_customer_id,
  DROP COLUMN IF EXISTS stripe_subscription_id,
  DROP COLUMN IF EXISTS plan_seats;
