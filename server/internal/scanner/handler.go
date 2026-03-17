package scanner

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
)

// Handler handles HTTP routes for the scanner package.
type Handler struct {
	service *Service
}

// NewHandler creates a new scanner Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Routes registers all scanner routes.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/projects/{project_id}/scans", h.createScan)
	r.Get("/projects/{project_id}/scans", h.listScans)
	r.Get("/scans/{scan_id}/findings", h.listFindings)
	r.Post("/scans/{scan_id}/findings/{finding_id}/import", h.importFinding)
	r.Post("/scans/{scan_id}/findings/{finding_id}/dismiss", h.dismissFinding)
	return r
}

type createScanRequest struct {
	ScanPath      string        `json:"scan_path"`
	FindingsCount int           `json:"findings_count"`
	Findings      []ScanFinding `json:"findings"`
}

func (h *Handler) createScan(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	userID := auth.UserIDFromContext(r.Context())

	var req createScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}
	if req.ScanPath == "" {
		apierror.BadRequest(w, "scan_path is required")
		return
	}

	scan, err := h.service.CreateScan(r.Context(), projectID, userID, req.ScanPath, req.Findings)
	if err != nil {
		log.Printf("Failed to create scan: %v", err)
		apierror.InternalError(w, "failed to create scan")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scan) //nolint:errcheck
}

func (h *Handler) listScans(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "project_id")
	_ = auth.UserIDFromContext(r.Context()) // TODO(Phase13): add RBAC authorization

	scans, err := h.service.ListScans(r.Context(), projectID)
	if err != nil {
		log.Printf("Failed to list scans: %v", err)
		apierror.InternalError(w, "failed to list scans")
		return
	}
	if scans == nil {
		scans = []ScanResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"scans": scans}) //nolint:errcheck
}

func (h *Handler) listFindings(w http.ResponseWriter, r *http.Request) {
	scanID := chi.URLParam(r, "scan_id")
	_ = auth.UserIDFromContext(r.Context()) // TODO(Phase13): add RBAC authorization

	findings, err := h.service.ListFindings(r.Context(), scanID)
	if err != nil {
		log.Printf("Failed to list findings: %v", err)
		apierror.InternalError(w, "failed to list findings")
		return
	}
	if findings == nil {
		findings = []ScanFinding{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"findings": findings}) //nolint:errcheck
}

type importFindingRequest struct {
	SecretID string `json:"secret_id"`
}

func (h *Handler) importFinding(w http.ResponseWriter, r *http.Request) {
	findingID := chi.URLParam(r, "finding_id")
	_ = auth.UserIDFromContext(r.Context()) // TODO(Phase13): add RBAC authorization

	var req importFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}
	if req.SecretID == "" {
		apierror.BadRequest(w, "secret_id is required")
		return
	}

	if err := h.service.ImportFinding(r.Context(), findingID, req.SecretID); err != nil {
		log.Printf("Failed to import finding: %v", err)
		apierror.InternalError(w, "failed to import finding")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "imported"}) //nolint:errcheck
}

func (h *Handler) dismissFinding(w http.ResponseWriter, r *http.Request) {
	findingID := chi.URLParam(r, "finding_id")
	_ = auth.UserIDFromContext(r.Context()) // TODO(Phase13): add RBAC authorization

	if err := h.service.DismissFinding(r.Context(), findingID); err != nil {
		log.Printf("Failed to dismiss finding: %v", err)
		apierror.InternalError(w, "failed to dismiss finding")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "dismissed"}) //nolint:errcheck
}
