package vault

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
	"github.com/valt-dev/valt/server/pkg/validator"
)

var validCredentialTypes = map[string]bool{
	"api_key":          true,
	"db_credential":    true,
	"ssh_key":          true,
	"oauth_token":      true,
	"cloud_credential": true,
	"personal_session": true,
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.createSecret)
	r.Get("/", h.listSecrets)
	r.Get("/{id}", h.getSecret)
	r.Put("/{id}", h.updateSecret)
	r.Delete("/{id}", h.deleteSecret)
	return r
}

type createSecretRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	CredentialType string `json:"credential_type"`
	Source         string `json:"source"`
	EncryptedBlob  string `json:"encrypted_blob"` // base64
	EncryptedDEK   string `json:"encrypted_dek"`  // base64
	Policy         string `json:"policy"`
}

func (h *Handler) createSecret(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req createSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		apierror.BadRequest(w, "name is required")
		return
	}

	if req.CredentialType != "" && !validCredentialTypes[req.CredentialType] {
		apierror.BadRequest(w, "invalid credential_type")
		return
	}

	blob, err := base64.StdEncoding.DecodeString(req.EncryptedBlob)
	if err != nil {
		apierror.BadRequest(w, "invalid encrypted_blob encoding")
		return
	}
	dek, err := base64.StdEncoding.DecodeString(req.EncryptedDEK)
	if err != nil {
		apierror.BadRequest(w, "invalid encrypted_dek encoding")
		return
	}

	secret, err := h.service.CreateSecret(r.Context(), userID, CreateSecretInput{
		Name:           req.Name,
		Description:    req.Description,
		CredentialType: req.CredentialType,
		Source:         req.Source,
		EncryptedBlob:  blob,
		EncryptedDEK:   dek,
		Policy:         req.Policy,
	})
	if err != nil {
		log.Printf("Failed to create secret: %v", err)
		apierror.InternalError(w, "failed to create secret")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(secret)
}

func (h *Handler) listSecrets(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	pg, err := validator.ValidatePagination(
		r.URL.Query().Get("page"),
		r.URL.Query().Get("limit"),
	)
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	result, err := h.service.ListSecrets(r.Context(), userID, pg.Page, pg.Limit, pg.Offset)
	if err != nil {
		log.Printf("Failed to list secrets: %v", err)
		apierror.InternalError(w, "failed to list secrets")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) getSecret(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "id")

	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret ID")
		return
	}

	secret, err := h.service.GetSecret(r.Context(), userID, secretID)
	if err != nil {
		log.Printf("Failed to get secret: %v", err)
		apierror.InternalError(w, "failed to get secret")
		return
	}
	if secret == nil {
		apierror.NotFound(w, "secret not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

type updateSecretRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	EncryptedBlob string `json:"encrypted_blob"` // base64
	EncryptedDEK  string `json:"encrypted_dek"`  // base64
}

func (h *Handler) updateSecret(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "id")

	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret ID")
		return
	}

	var req updateSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	var blob, dek []byte
	if req.EncryptedBlob != "" {
		var err error
		blob, err = base64.StdEncoding.DecodeString(req.EncryptedBlob)
		if err != nil {
			apierror.BadRequest(w, "invalid encrypted_blob encoding")
			return
		}
	}
	if req.EncryptedDEK != "" {
		var err error
		dek, err = base64.StdEncoding.DecodeString(req.EncryptedDEK)
		if err != nil {
			apierror.BadRequest(w, "invalid encrypted_dek encoding")
			return
		}
	}

	secret, err := h.service.UpdateSecret(r.Context(), userID, secretID, UpdateSecretInput{
		Name:          req.Name,
		Description:   req.Description,
		EncryptedBlob: blob,
		EncryptedDEK:  dek,
	})
	if err != nil {
		log.Printf("Failed to update secret: %v", err)
		apierror.InternalError(w, "failed to update secret")
		return
	}
	if secret == nil {
		apierror.NotFound(w, "secret not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func (h *Handler) deleteSecret(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	secretID := chi.URLParam(r, "id")

	if _, err := validator.ValidateUUID(secretID); err != nil {
		apierror.BadRequest(w, "invalid secret ID")
		return
	}

	err := h.service.DeleteSecret(r.Context(), userID, secretID)
	if err != nil {
		if err == pgx.ErrNoRows {
			apierror.NotFound(w, "secret not found")
			return
		}
		log.Printf("Failed to delete secret: %v", err)
		apierror.InternalError(w, "failed to delete secret")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
