package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/valt-dev/valt/server/internal/config"
	"github.com/valt-dev/valt/server/pkg/apierror"
	"github.com/valt-dev/valt/server/pkg/validator"
)

// Handler holds the DB pool and JWT manager for auth endpoints.
type Handler struct {
	pool         *pgxpool.Pool
	jwtMgr       *JWTManager
	oauthConfig  *oauth2.Config
	dashboardURL string
}

// NewHandler constructs a Handler with required dependencies.
func NewHandler(pool *pgxpool.Pool, jwtMgr *JWTManager, cfg *config.Config) *Handler {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &Handler{pool: pool, jwtMgr: jwtMgr, oauthConfig: oauthCfg, dashboardURL: cfg.DashboardURL}
}

// Routes returns a chi.Router with all auth routes mounted.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/refresh", h.refresh)
	r.Get("/google", h.googleLogin)
	r.Get("/google/callback", h.googleCallback)
	return r
}

// --- Request / Response types ---

type registerRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RegionCode string `json:"region_code"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// --- Handlers ---

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	if err := validator.ValidateEmail(req.Email); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}
	if err := validator.ValidatePassword(req.Password); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}
	if err := validator.ValidateRegionCode(req.RegionCode); err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("auth: hash password: %v", err)
		apierror.InternalError(w, "failed to process registration")
		return
	}

	var userID string
	err = h.pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, region_code) VALUES ($1, $2, $3) RETURNING id`,
		req.Email, hash, req.RegionCode,
	).Scan(&userID)
	if err != nil {
		if isDuplicateKeyError(err) {
			apierror.Conflict(w, "email already registered")
			return
		}
		log.Printf("auth: insert user: %v", err)
		apierror.InternalError(w, "failed to create account")
		return
	}

	h.issueTokens(w, r, userID, http.StatusCreated)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	var userID, passwordHash, status string
	err := h.pool.QueryRow(r.Context(),
		`SELECT id, password_hash, status FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &passwordHash, &status)
	if err != nil {
		if err == pgx.ErrNoRows {
			apierror.Unauthorized(w, "invalid email or password")
			return
		}
		log.Printf("auth: query user: %v", err)
		apierror.InternalError(w, "login failed")
		return
	}

	if status != "active" {
		apierror.Forbidden(w, "account is not active")
		return
	}

	match, err := VerifyPassword(req.Password, passwordHash)
	if err != nil || !match {
		apierror.Unauthorized(w, "invalid email or password")
		return
	}

	h.issueTokens(w, r, userID, http.StatusOK)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.BadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		apierror.BadRequest(w, "refresh_token is required")
		return
	}

	tokenHash := hashRefreshToken(req.RefreshToken)

	var userID string
	var expiresAt time.Time
	err := h.pool.QueryRow(r.Context(),
		`SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			apierror.Unauthorized(w, "invalid refresh token")
			return
		}
		log.Printf("auth: query refresh token: %v", err)
		apierror.InternalError(w, "token refresh failed")
		return
	}

	if time.Now().After(expiresAt) {
		apierror.Unauthorized(w, "refresh token expired")
		return
	}

	// Rotate: delete old token before issuing new one.
	_, _ = h.pool.Exec(r.Context(),
		`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)

	h.issueTokens(w, r, userID, http.StatusOK)
}

// --- Helpers ---

// issueTokens generates an access + refresh token pair, persists the refresh
// token hash, and writes the JSON response with the given HTTP status code.
func (h *Handler) issueTokens(w http.ResponseWriter, r *http.Request, userID string, statusCode int) {
	accessToken, err := h.jwtMgr.GenerateAccessToken(userID)
	if err != nil {
		log.Printf("auth: generate access token: %v", err)
		apierror.InternalError(w, "failed to generate tokens")
		return
	}

	refreshToken, err := h.jwtMgr.GenerateRefreshToken()
	if err != nil {
		log.Printf("auth: generate refresh token: %v", err)
		apierror.InternalError(w, "failed to generate tokens")
		return
	}

	tokenHash := hashRefreshToken(refreshToken)
	_, err = h.pool.Exec(r.Context(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		log.Printf("auth: store refresh token: %v", err)
		apierror.InternalError(w, "failed to generate tokens")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(authResponse{ //nolint:errcheck
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	})
}

// hashRefreshToken returns the hex-encoded SHA-256 of the raw token string.
func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// isDuplicateKeyError reports whether err is a PostgreSQL unique-violation (23505).
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
