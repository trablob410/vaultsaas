package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status=%d, want 200", rec.Code)
	}

	headers := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"X-XSS-Protection":       "0",
	}
	for header, expected := range headers {
		got := rec.Header().Get(header)
		if got != expected {
			t.Errorf("%s: got %q, want %q", header, got, expected)
		}
	}
	// HSTS should be set
	if hsts := rec.Header().Get("Strict-Transport-Security"); hsts == "" {
		t.Error("HSTS header should be set")
	}
	// CSP should be set
	if csp := rec.Header().Get("Content-Security-Policy"); csp == "" {
		t.Error("CSP header should be set")
	}
}

func TestSecurityHeaders_PassThrough(t *testing.T) {
	// Verify that the handler still calls next
	var nextCalled bool
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusCreated)
	}))
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !nextCalled {
		t.Error("SecurityHeaders should call next handler")
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("status=%d, want 201", rec.Code)
	}
}
