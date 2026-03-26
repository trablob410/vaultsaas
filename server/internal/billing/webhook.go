package billing

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// HandleWebhook processes Stripe webhook events. POST /webhooks/stripe
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("Stripe-Signature")
	evt, err := webhook.ConstructEvent(body, sig, h.svc.webhookSecret)
	if err != nil {
		log.Printf("[billing] webhook signature verification failed: %v", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	switch evt.Type {
	case "checkout.session.completed":
		h.handleCheckoutCompleted(r, evt)
	case "customer.subscription.updated":
		h.handleSubscriptionUpdated(r, evt)
	case "customer.subscription.deleted":
		h.handleSubscriptionDeleted(r, evt)
	case "invoice.payment_failed":
		log.Printf("[billing] payment failed for event %s", evt.ID)
	}

	// Always 200 — Stripe retries on non-2xx
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCheckoutCompleted(r *http.Request, evt stripe.Event) {
	var sess stripe.CheckoutSession
	if err := json.Unmarshal(evt.Data.Raw, &sess); err != nil {
		log.Printf("[billing] unmarshal checkout session: %v", err)
		return
	}

	orgID := sess.ClientReferenceID
	if orgID == "" {
		log.Printf("[billing] checkout missing client_reference_id")
		return
	}

	plan := "pro"
	if sess.Metadata != nil {
		if p, ok := sess.Metadata["plan"]; ok {
			plan = p
		}
	}

	subID := ""
	if sess.Subscription != nil {
		subID = sess.Subscription.ID
	}

	seats, _ := h.svc.SeatCount(r.Context(), orgID)
	if seats < 1 {
		seats = 1
	}

	if err := h.svc.UpdateOrgBilling(r.Context(), orgID, sess.Customer.ID, subID, plan, seats); err != nil {
		log.Printf("[billing] update org after checkout: %v", err)
	}
}

func (h *Handler) handleSubscriptionUpdated(r *http.Request, evt stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
		log.Printf("[billing] unmarshal subscription: %v", err)
		return
	}

	custID := ""
	if sub.Customer != nil {
		custID = sub.Customer.ID
	}
	if custID == "" {
		return
	}

	orgID, err := h.svc.FindOrgByCustomerID(r.Context(), custID)
	if err != nil {
		log.Printf("[billing] find org for subscription update: %v", err)
		return
	}

	// Determine plan from price ID
	plan := h.planFromSubscription(&sub)
	seats := int(sub.Items.Data[0].Quantity)
	if seats < 1 {
		seats = 1
	}

	if err := h.svc.UpdateOrgBilling(r.Context(), orgID, custID, sub.ID, plan, seats); err != nil {
		log.Printf("[billing] update org billing: %v", err)
	}
}

func (h *Handler) handleSubscriptionDeleted(r *http.Request, evt stripe.Event) {
	var sub stripe.Subscription
	if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
		log.Printf("[billing] unmarshal subscription deletion: %v", err)
		return
	}

	custID := ""
	if sub.Customer != nil {
		custID = sub.Customer.ID
	}
	if custID == "" {
		return
	}

	orgID, err := h.svc.FindOrgByCustomerID(r.Context(), custID)
	if err != nil {
		log.Printf("[billing] find org for subscription delete: %v", err)
		return
	}

	if err := h.svc.DowngradeToFree(r.Context(), orgID); err != nil {
		log.Printf("[billing] downgrade to free: %v", err)
	}
}

// planFromSubscription derives the plan name from subscription items' price ID.
func (h *Handler) planFromSubscription(sub *stripe.Subscription) string {
	if len(sub.Items.Data) == 0 {
		return "free"
	}
	priceID := sub.Items.Data[0].Price.ID
	switch priceID {
	case h.svc.priceProID:
		return "pro"
	case h.svc.priceTeamID:
		return "team"
	default:
		return "pro" // fallback for unknown price
	}
}
