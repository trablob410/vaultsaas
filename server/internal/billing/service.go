package billing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
)

// Service wraps Stripe API calls and org billing state.
type Service struct {
	pool           *pgxpool.Pool
	priceProID     string
	priceTeamID    string
	webhookSecret  string
}

// NewService creates a billing Service. Sets the Stripe API key globally.
func NewService(pool *pgxpool.Pool, secretKey, webhookSecret, priceProID, priceTeamID string) *Service {
	stripe.Key = secretKey
	return &Service{pool: pool, webhookSecret: webhookSecret, priceProID: priceProID, priceTeamID: priceTeamID}
}

// GetOrCreateCustomer returns the Stripe customer ID for an org, creating one if needed.
func (s *Service) GetOrCreateCustomer(ctx context.Context, orgID, orgName, ownerEmail string) (string, error) {
	var custID *string
	err := s.pool.QueryRow(ctx,
		`SELECT stripe_customer_id FROM organizations WHERE id = $1`, orgID,
	).Scan(&custID)
	if err != nil {
		return "", fmt.Errorf("querying org stripe customer: %w", err)
	}
	if custID != nil && *custID != "" {
		return *custID, nil
	}

	// Create new Stripe customer
	params := &stripe.CustomerParams{
		Name:  stripe.String(orgName),
		Email: stripe.String(ownerEmail),
	}
	params.AddMetadata("org_id", orgID)
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("creating stripe customer: %w", err)
	}

	// Persist customer ID
	_, err = s.pool.Exec(ctx,
		`UPDATE organizations SET stripe_customer_id = $1 WHERE id = $2`,
		c.ID, orgID,
	)
	if err != nil {
		return "", fmt.Errorf("saving stripe customer id: %w", err)
	}
	return c.ID, nil
}

// SeatCount returns the number of members in an org.
func (s *Service) SeatCount(ctx context.Context, orgID string) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM org_memberships WHERE org_id = $1`, orgID,
	).Scan(&count)
	return count, err
}

// CreateCheckoutSession creates a Stripe Checkout session for a plan upgrade.
func (s *Service) CreateCheckoutSession(ctx context.Context, orgID, orgName, ownerEmail, plan, successURL, cancelURL string) (string, error) {
	priceID := s.priceForPlan(plan)
	if priceID == "" {
		return "", fmt.Errorf("unknown plan: %s", plan)
	}

	custID, err := s.GetOrCreateCustomer(ctx, orgID, orgName, ownerEmail)
	if err != nil {
		return "", err
	}

	seats, err := s.SeatCount(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("counting seats: %w", err)
	}
	if seats < 1 {
		seats = 1
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(custID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceID), Quantity: stripe.Int64(int64(seats))},
		},
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		ClientReferenceID: stripe.String(orgID),
	}
	params.AddMetadata("org_id", orgID)
	params.AddMetadata("plan", plan)

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("creating checkout session: %w", err)
	}
	return sess.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session.
func (s *Service) CreatePortalSession(ctx context.Context, orgID, returnURL string) (string, error) {
	var custID *string
	err := s.pool.QueryRow(ctx,
		`SELECT stripe_customer_id FROM organizations WHERE id = $1`, orgID,
	).Scan(&custID)
	if err != nil || custID == nil || *custID == "" {
		return "", fmt.Errorf("org has no stripe customer")
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  custID,
		ReturnURL: stripe.String(returnURL),
	}
	ps, err := portalsession.New(params)
	if err != nil {
		return "", fmt.Errorf("creating portal session: %w", err)
	}
	return ps.URL, nil
}

// UpdateOrgBilling updates the org's Stripe billing state after a webhook event.
func (s *Service) UpdateOrgBilling(ctx context.Context, orgID, customerID, subscriptionID, plan string, seats int) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE organizations
		 SET stripe_customer_id = $1, stripe_subscription_id = $2, plan = $3, plan_seats = $4, updated_at = NOW()
		 WHERE id = $5`,
		customerID, subscriptionID, plan, seats, orgID,
	)
	return err
}

// DowngradeToFree resets the org to the free plan.
func (s *Service) DowngradeToFree(ctx context.Context, orgID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE organizations
		 SET plan = 'free', stripe_subscription_id = NULL, plan_seats = 1, updated_at = NOW()
		 WHERE id = $1`,
		orgID,
	)
	return err
}

// FindOrgByCustomerID looks up an org by its Stripe customer ID.
func (s *Service) FindOrgByCustomerID(ctx context.Context, customerID string) (string, error) {
	var orgID string
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM organizations WHERE stripe_customer_id = $1`, customerID,
	).Scan(&orgID)
	if err != nil {
		return "", fmt.Errorf("org not found for customer %s: %w", customerID, err)
	}
	return orgID, nil
}

func (s *Service) priceForPlan(plan string) string {
	switch plan {
	case "pro":
		return s.priceProID
	case "team":
		return s.priceTeamID
	default:
		return ""
	}
}
