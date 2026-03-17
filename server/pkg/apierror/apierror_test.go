package apierror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) errorResponse {
	t.Helper()
	var r errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return r
}

func TestBadRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	BadRequest(rec, "test message")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "bad_request" {
		t.Errorf("code=%q", r.Error.Code)
	}
	if r.Error.Message != "test message" {
		t.Errorf("msg=%q", r.Error.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	rec := httptest.NewRecorder()
	Unauthorized(rec, "unauthorized")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "unauthorized" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestForbidden(t *testing.T) {
	rec := httptest.NewRecorder()
	Forbidden(rec, "forbidden")
	if rec.Code != http.StatusForbidden {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "forbidden" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	NotFound(rec, "not found")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "not_found" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestConflict(t *testing.T) {
	rec := httptest.NewRecorder()
	Conflict(rec, "conflict")
	if rec.Code != http.StatusConflict {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "conflict" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestTooManyRequests(t *testing.T) {
	rec := httptest.NewRecorder()
	TooManyRequests(rec, "slow down")
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "rate_limit_exceeded" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestInternalError(t *testing.T) {
	rec := httptest.NewRecorder()
	InternalError(rec, "oops")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status=%d", rec.Code)
	}
	r := decode(t, rec)
	if r.Error.Code != "internal_error" {
		t.Errorf("code=%q", r.Error.Code)
	}
}

func TestContentTypeHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	BadRequest(rec, "x")
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type=%q", ct)
	}
}
