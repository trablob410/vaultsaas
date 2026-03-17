package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type googleUserInfo struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Picture     string `json:"picture"`
	VerifiedEmail bool  `json:"verified_email"`
}

func (h *Handler) googleLogin(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("oauth: generate state: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	state := hex.EncodeToString(b)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	redirectURL := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *Handler) googleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state cookie
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	// Clear state cookie
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("oauth: exchange code: %v", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Printf("oauth: fetch userinfo: %v", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("oauth: decode userinfo: %v", err)
		http.Error(w, "failed to parse user info", http.StatusInternalServerError)
		return
	}

	userID, err := h.upsertGoogleUser(r, info)
	if err != nil {
		log.Printf("oauth: upsert user: %v", err)
		http.Error(w, "account setup failed", http.StatusInternalServerError)
		return
	}

	if err := h.issueTokensWithCookies(w, r, userID); err != nil {
		log.Printf("oauth: issue tokens: %v", err)
		http.Error(w, "token issuance failed", http.StatusInternalServerError)
		return
	}

	redirectTarget, _ := url.JoinPath(h.dashboardURL, "/secrets")
	http.Redirect(w, r, redirectTarget, http.StatusTemporaryRedirect)
}

func (h *Handler) upsertGoogleUser(r *http.Request, info googleUserInfo) (string, error) {
	var userID string
	// Try to find by google_id first, then by email
	err := h.pool.QueryRow(r.Context(),
		`SELECT id FROM users WHERE google_id = $1`, info.ID,
	).Scan(&userID)
	if err == nil {
		// Update profile fields
		_, _ = h.pool.Exec(r.Context(),
			`UPDATE users SET display_name=$1, avatar_url=$2 WHERE id=$3`,
			info.Name, info.Picture, userID,
		)
		return userID, nil
	}

	// Try match by email (link existing local account)
	err = h.pool.QueryRow(r.Context(),
		`SELECT id FROM users WHERE email = $1`, info.Email,
	).Scan(&userID)
	if err == nil {
		_, updateErr := h.pool.Exec(r.Context(),
			`UPDATE users SET google_id=$1, display_name=$2, avatar_url=$3, auth_provider='google' WHERE id=$4`,
			info.ID, info.Name, info.Picture, userID,
		)
		return userID, updateErr
	}

	// Create new user
	err = h.pool.QueryRow(r.Context(),
		`INSERT INTO users (email, google_id, display_name, avatar_url, auth_provider, region_code)
		 VALUES ($1, $2, $3, $4, 'google', 'us')
		 RETURNING id`,
		info.Email, info.ID, info.Name, info.Picture,
	).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("insert google user: %w", err)
	}
	return userID, nil
}

func (h *Handler) issueTokensWithCookies(w http.ResponseWriter, r *http.Request, userID string) error {
	accessToken, err := h.jwtMgr.GenerateAccessToken(userID)
	if err != nil {
		return fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := h.jwtMgr.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("generate refresh token: %w", err)
	}

	tokenHash := hashRefreshToken(refreshToken)
	_, err = h.pool.Exec(r.Context(),
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, time.Now().Add(7*24*time.Hour),
	)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "valt_access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "valt_refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}
