package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"
)

func TestGoogleLogin_MissingOAuthConfig(t *testing.T) {
	h := &Handler{oauthConfig: &oauth2.Config{}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google", nil)
	rec := httptest.NewRecorder()

	h.googleLogin(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestGoogleCallback_MissingOAuthConfig(t *testing.T) {
	h := &Handler{oauthConfig: &oauth2.Config{}}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback?state=abc&code=xyz", nil)
	rec := httptest.NewRecorder()

	h.googleCallback(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
