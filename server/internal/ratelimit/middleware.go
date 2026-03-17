package ratelimit

import (
	"net/http"
)

// Middleware returns an HTTP middleware that rate-limits by X-Agent-ID header.
// rpm is the requests-per-minute limit per agent.
// Requests without X-Agent-ID header pass through (human users).
func (l *RedisLimiter) Middleware(rpm int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			agentID := r.Header.Get("X-Agent-ID")
			if agentID == "" {
				next.ServeHTTP(w, r)
				return
			}
			allowed, err := l.Allow(r.Context(), agentID, rpm)
			if err != nil || allowed {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`)) //nolint:errcheck
		})
	}
}
