package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	}))
	defer srv.Close()
	c := New(srv.URL, "test-token")
	var out map[string]string
	if err := c.Get(context.Background(), "/health", &out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "ok" {
		t.Error("unexpected response")
	}
}

func TestClientGet_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"}) //nolint:errcheck
	}))
	defer srv.Close()
	c := New(srv.URL, "bad-token")
	var out map[string]string
	err := c.Get(context.Background(), "/secrets", &out)
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}
