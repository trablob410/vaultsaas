package usage

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler serves usage endpoints.
type Handler struct {
	tracker *Tracker
}

// NewHandler creates a new Handler.
func NewHandler(tracker *Tracker) *Handler {
	return &Handler{tracker: tracker}
}

// Routes returns the chi router for usage endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	h.registerRoutes(r)
	return r
}

// RegisterRoutes registers usage routes onto an existing router group.
func (h *Handler) RegisterRoutes(r chi.Router) {
	h.registerRoutes(r)
}

func (h *Handler) registerRoutes(r chi.Router) {
	r.Get("/orgs/{org_id}/usage", h.GetUsage)
}

// getUsage returns usage stats for an org.
// Response: { plan, usage: {...}, limits: {...} }
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "org_id")

	plan, err := h.tracker.GetOrgPlan(r.Context(), orgID)
	if err != nil {
		plan = "unknown"
	}

	secrets, _ := h.tracker.GetCurrent(r.Context(), orgID, "secrets_count")
	agents, _ := h.tracker.GetCurrent(r.Context(), orgID, "agents_count")
	requests, _ := h.tracker.GetCurrent(r.Context(), orgID, "requests_today")

	resp := map[string]interface{}{
		"plan": plan,
		"usage": map[string]int{
			"secrets_count":  secrets,
			"agents_count":   agents,
			"requests_today": requests,
		},
		"limits": LimitsForPlan(plan),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("[usage] failed to encode response: %v", err)
	}
}
