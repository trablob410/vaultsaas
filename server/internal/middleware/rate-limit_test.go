package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsUpToMax(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.allow("key1") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverMax(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Minute)
	for i := 0; i < 3; i++ {
		rl.allow("key2")
	}
	if rl.allow("key2") {
		t.Error("4th request should be blocked")
	}
}

func TestRateLimiter_IndependentKeys(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)
	rl.allow("a")
	rl.allow("a")
	if rl.allow("a") {
		t.Error("key a: 3rd should be blocked")
	}
	if !rl.allow("b") {
		t.Error("key b: 1st should be allowed")
	}
}

func TestRateLimiter_AllowsAfterWindow(t *testing.T) {
	// Use a very short window to test expiry
	rl := NewRateLimiter(1, 50*time.Millisecond)
	if !rl.allow("key3") {
		t.Error("1st request should be allowed")
	}
	if rl.allow("key3") {
		t.Error("2nd request should be blocked")
	}
	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)
	if !rl.allow("key3") {
		t.Error("request after window should be allowed")
	}
}

func TestIPMiddleware_Returns429(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)
	handler := rl.IPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: status=%d, want 200", i+1, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("3rd request: status=%d, want 429", rec.Code)
	}
}

func TestIPMiddleware_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1*time.Minute)
	handler := rl.IPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1111"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("ip1 first: status=%d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.2:2222"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("ip2 first: status=%d (should be independent)", rec2.Code)
	}
}
