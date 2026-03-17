package workflow

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/valt-dev/valt/server/internal/audit"
	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/internal/notify"
	"github.com/valt-dev/valt/server/internal/policy"
	"github.com/valt-dev/valt/server/internal/vault"
	"github.com/valt-dev/valt/server/pkg/apierror"
	"github.com/valt-dev/valt/server/pkg/crypto"
	"github.com/valt-dev/valt/server/pkg/validator"
)

// Handler serves workflow HTTP endpoints.
type Handler struct {
	service   *Service
	credMgr   *CredentialManager
	vaultSvc  *vault.Service
	auditLog  *audit.Logger
	notifySvc *notify.Service
	masterKey []byte
}

// NewHandler creates a workflow Handler.
func NewHandler(svc *Service, credMgr *CredentialManager, vaultSvc *vault.Service, auditLog *audit.Logger, notifySvc *notify.Service, masterKey []byte) *Handler {
	return &Handler{
		service:   svc,
		credMgr:   credMgr,
		vaultSvc:  vaultSvc,
		auditLog:  auditLog,
		notifySvc: notifySvc,
		masterKey: masterKey,
	}
}

type createRequestBody struct {
	RequesterType   string `json:"requester_type"`
	AIAgentID       string `json:"ai_agent_id,omitempty"`
	Reason          string `json:"reason"`
	DurationMinutes int    `json:"duration_minutes"`
}

// CreateRequest handles POST /secrets/{secret_id}/access-requests
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "secret_id")

	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret_id")
		return
	}

	var body createRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	if body.RequesterType == "" {
		body.RequesterType = "human"
	}
	if body.RequesterType != "human" && body.RequesterType != "ai_agent" {
		apierror.BadRequest(w, "requester_type must be 'human' or 'ai_agent'")
		return
	}

	// Look up the secret to get credential_type
	secret, err := h.vaultSvc.GetSecret(r.Context(), userID, secretID)
	if err != nil {
		log.Printf("Failed to get secret for access request: %v", err)
		apierror.InternalError(w, "failed to validate secret")
		return
	}
	if secret == nil {
		apierror.NotFound(w, "secret not found")
		return
	}

	req, err := h.service.CreateRequest(r.Context(), CreateRequestInput{
		SecretID:        secretID,
		RequesterUserID: userID,
		RequesterType:   body.RequesterType,
		AIAgentID:       body.AIAgentID,
		Reason:          body.Reason,
		DurationMinutes: body.DurationMinutes,
		CredentialType:  secret.CredentialType,
	})
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	h.auditLog.LogFromRequest(r, userID, "access_request.create", "access_request", req.ID)

	// Auto-approve for Tier 1: issue credential immediately
	if req.Status == "approved" {
		_, issueErr := h.credMgr.IssueCredential(r.Context(), req.ID, secret.CredentialType, req.RequestedDurationMinutes)
		if issueErr != nil {
			log.Printf("Failed to auto-issue credential: %v", issueErr)
		}
	}

	// Notify if policy requires it
	p := policy.ForCredentialType(secret.CredentialType)
	if p.NotifyOnAccess && h.notifySvc != nil {
		_ = h.notifySvc.NotifyApprovalNeeded(r.Context(), "", secret.Name, userID, body.Reason)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// ListPending handles GET /access-requests?status=
func (h *Handler) ListPending(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	status := r.URL.Query().Get("status")

	pg, err := validator.ValidatePagination(
		r.URL.Query().Get("page"),
		r.URL.Query().Get("limit"),
	)
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	requests, total, err := h.service.ListPending(r.Context(), userID, status, pg.Limit, pg.Offset)
	if err != nil {
		log.Printf("Failed to list access requests: %v", err)
		apierror.InternalError(w, "failed to list access requests")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"total":    total,
		"page":     pg.Page,
		"limit":    pg.Limit,
	})
}

type approveRejectBody struct {
	RejectionReason string `json:"rejection_reason,omitempty"`
}

// Approve handles POST /access-requests/{request_id}/approve
func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requestID := chi.URLParam(r, "request_id")

	if _, err := validator.ValidateUUID(requestID); err != nil {
		apierror.BadRequest(w, "invalid request_id")
		return
	}

	req, err := h.service.Approve(r.Context(), requestID, userID)
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	// C2/H1 fix: use GetSecretByID (no owner constraint) so approver ≠ owner works.
	// Error is surfaced instead of silently swallowed.
	secret, err := h.vaultSvc.GetSecretByID(r.Context(), req.SecretID)
	if err != nil {
		log.Printf("Failed to fetch secret for credential issuance: %v", err)
		apierror.InternalError(w, "failed to issue credential")
		return
	}
	credType := ""
	if secret != nil {
		credType = secret.CredentialType
	}

	// Issue credential
	_, issueErr := h.credMgr.IssueCredential(r.Context(), req.ID, credType, req.RequestedDurationMinutes)
	if issueErr != nil {
		log.Printf("Failed to issue credential after approval: %v", issueErr)
	}

	h.auditLog.LogFromRequest(r, userID, "access_request.approve", "access_request", requestID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// GetRequest handles GET /access-requests/{request_id}
func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requestID := chi.URLParam(r, "request_id")

	if _, err := validator.ValidateUUID(requestID); err != nil {
		apierror.BadRequest(w, "invalid request_id")
		return
	}

	req, err := h.service.GetRequestByID(r.Context(), requestID)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		apierror.InternalError(w, "failed to get request")
		return
	}
	if req == nil {
		apierror.NotFound(w, "request not found")
		return
	}
	// Only the requester or the secret owner can view the request
	if req.RequesterUserID != userID {
		apierror.Forbidden(w, "not your request")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// Reject handles POST /access-requests/{request_id}/reject
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requestID := chi.URLParam(r, "request_id")

	if _, err := validator.ValidateUUID(requestID); err != nil {
		apierror.BadRequest(w, "invalid request_id")
		return
	}

	var body approveRejectBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	req, err := h.service.Reject(r.Context(), requestID, userID, body.RejectionReason)
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	h.auditLog.LogFromRequest(r, userID, "access_request.reject", "access_request", requestID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// GetCredential handles GET /credentials/{request_id}
func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requestID := chi.URLParam(r, "request_id")

	if _, err := validator.ValidateUUID(requestID); err != nil {
		apierror.BadRequest(w, "invalid request_id")
		return
	}

	// Verify the requester owns this request
	req, err := h.service.GetRequestByID(r.Context(), requestID)
	if err != nil {
		log.Printf("Failed to get request: %v", err)
		apierror.InternalError(w, "failed to get request")
		return
	}
	if req == nil {
		apierror.NotFound(w, "request not found")
		return
	}
	if req.RequesterUserID != userID {
		apierror.Forbidden(w, "not your request")
		return
	}

	session, err := h.credMgr.GetCredential(r.Context(), requestID)
	if err != nil {
		apierror.NotFound(w, err.Error())
		return
	}

	h.auditLog.LogFromRequest(r, userID, "credential.access", "credential_session", session.ID)

	// Fetch secret by ID (no owner constraint — requester may not be owner)
	secret, err := h.vaultSvc.GetSecretByID(r.Context(), req.SecretID)
	if err != nil {
		log.Printf("Failed to fetch secret for credential delivery: %v", err)
	}

	// Decrypt and attach value if secret has an encrypted DEK
	if secret != nil && len(secret.EncryptedDEK) > 0 {
		blob, blobErr := h.vaultSvc.GetBlob(r.Context(), secret.StorageKey)
		if blobErr != nil {
			log.Printf("Failed to get blob for credential delivery: %v", blobErr)
		} else {
			dek, dekErr := crypto.DecryptAES256GCM(h.masterKey, secret.EncryptedDEK)
			if dekErr != nil {
				log.Printf("Failed to decrypt DEK for credential delivery: %v", dekErr)
			} else {
				plaintext, ptErr := crypto.DecryptAES256GCM(dek, blob)
				if ptErr != nil {
					log.Printf("Failed to decrypt blob for credential delivery: %v", ptErr)
				} else {
					session.Value = string(plaintext)
				}
				// Zero out DEK (best-effort)
				for i := range dek {
					dek[i] = 0
				}
			}
		}
	}

	// Auto-revoke if single-use
	if secret != nil {
		p := policy.ForCredentialType(secret.CredentialType)
		_ = h.credMgr.AutoRevokeIfSingleUse(r.Context(), requestID, p.SingleUse)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// RevokeCredential handles POST /credentials/{request_id}/revoke
func (h *Handler) RevokeCredential(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	requestID := chi.URLParam(r, "request_id")

	if _, err := validator.ValidateUUID(requestID); err != nil {
		apierror.BadRequest(w, "invalid request_id")
		return
	}

	// C1 fix: verify caller owns this access request before revoking.
	req, err := h.service.GetRequestByID(r.Context(), requestID)
	if err != nil {
		log.Printf("Failed to get request for revoke: %v", err)
		apierror.InternalError(w, "failed to get request")
		return
	}
	if req == nil {
		apierror.NotFound(w, "request not found")
		return
	}
	if req.RequesterUserID != userID {
		apierror.Forbidden(w, "not your request")
		return
	}

	if err := h.credMgr.RevokeCredential(r.Context(), requestID); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	h.auditLog.LogFromRequest(r, userID, "credential.revoke", "credential_session", requestID)

	w.WriteHeader(http.StatusNoContent)
}
