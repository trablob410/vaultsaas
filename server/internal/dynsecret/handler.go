package dynsecret

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
)

// Handler exposes dynamic secret provider and lease HTTP endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a new Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Routes returns a chi.Router with all dynamic secret routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/projects/{project_id}/providers", h.createProvider)
	r.Get("/projects/{project_id}/providers", h.listProviders)
	r.Post("/providers/{provider_id}/leases", h.createLease)
	r.Get("/providers/{provider_id}/leases", h.listLeases)
	r.Delete("/leases/{lease_id}", h.revokeLease)
	return r
}

type createProviderRequest struct {
	Name         string            `json:"name"`
	ProviderType string            `json:"provider_type"`
	Config       map[string]string `json:"config"`
}

type createLeaseRequest struct {
	AgentID    string `json:"agent_id"`
	RequestID  string `json:"request_id,omitempty"`
	TTLSeconds int    `json:"ttl_seconds"`
}

type leaseResponse struct {
	ID          string            `json:"id"`
	Credentials map[string]string `json:"credentials"`
	ExpiresAt   time.Time         `json:"expires_at"`
	TTLSeconds  int               `json:"ttl_seconds"`
}

func (h *Handler) createProvider(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	userID := auth.UserIDFromContext(r.Context())

	var req createProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}
	if req.Name == "" {
		apierror.BadRequest(w, "name is required")
		return
	}
	if req.ProviderType == "" {
		apierror.BadRequest(w, "provider_type is required")
		return
	}

	pc, err := h.service.CreateProvider(r.Context(), projectID, req.Name, req.ProviderType, req.Config, userID)
	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		apierror.InternalError(w, "failed to create provider")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pc.ToResponse()) //nolint:errcheck
}

func (h *Handler) listProviders(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")

	providers, err := h.service.ListProviders(r.Context(), projectID)
	if err != nil {
		log.Printf("Failed to list providers: %v", err)
		apierror.InternalError(w, "failed to list providers")
		return
	}

	responses := make([]ProviderResponse, len(providers))
	for i := range providers {
		responses[i] = providers[i].ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"providers": responses}) //nolint:errcheck
}

func (h *Handler) createLease(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider_id")

	var req createLeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}
	if req.TTLSeconds <= 0 || req.TTLSeconds > 3600 {
		apierror.BadRequest(w, "ttl_seconds must be between 1 and 3600")
		return
	}

	lease, err := h.service.CreateLease(r.Context(), providerID, req.AgentID, req.RequestID, req.TTLSeconds)
	if err != nil {
		log.Printf("Failed to create lease: %v", err)
		apierror.InternalError(w, "failed to create lease")
		return
	}

	resp := leaseResponse{
		ID:          lease.ID,
		Credentials: lease.Credentials,
		ExpiresAt:   lease.ExpiresAt,
		TTLSeconds:  req.TTLSeconds,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (h *Handler) listLeases(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider_id")

	leases, err := h.service.ListActiveLeases(r.Context(), providerID)
	if err != nil {
		log.Printf("Failed to list leases: %v", err)
		apierror.InternalError(w, "failed to list leases")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"leases": leases}) //nolint:errcheck
}

func (h *Handler) revokeLease(w http.ResponseWriter, r *http.Request) {
	leaseID := chi.URLParam(r, "lease_id")

	if err := h.service.RevokeLease(r.Context(), leaseID); err != nil {
		log.Printf("Failed to revoke lease: %v", err)
		apierror.NotFound(w, "lease not found or already revoked")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
