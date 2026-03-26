package auth

import (
	"sync"
	"time"
)

const (
	maxLoginAttempts = 5
	lockoutWindow    = 15 * time.Minute
)

// loginAttempt tracks failed login attempts for an email.
type loginAttempt struct {
	failures []time.Time
}

// LoginLockout is an in-memory rate limiter for login attempts.
// Locks accounts after maxLoginAttempts failures within lockoutWindow.
type LoginLockout struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

// NewLoginLockout creates a new login lockout tracker.
func NewLoginLockout() *LoginLockout {
	l := &LoginLockout{attempts: make(map[string]*loginAttempt)}
	// Background cleanup of stale entries every 5 minutes
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			l.cleanup()
		}
	}()
	return l
}

// IsLocked returns true if the email has exceeded the max login attempts
// within the lockout window.
func (l *LoginLockout) IsLocked(email string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		return false
	}

	// Count recent failures within the window
	cutoff := time.Now().Add(-lockoutWindow)
	recent := 0
	for _, t := range a.failures {
		if t.After(cutoff) {
			recent++
		}
	}
	return recent >= maxLoginAttempts
}

// RecordFailure records a failed login attempt for an email.
func (l *LoginLockout) RecordFailure(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		a = &loginAttempt{}
		l.attempts[email] = a
	}
	a.failures = append(a.failures, time.Now())

	// Keep only recent failures to bound memory
	cutoff := time.Now().Add(-lockoutWindow)
	filtered := a.failures[:0]
	for _, t := range a.failures {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	a.failures = filtered
}

// ClearFailures resets the failure count for an email (on successful login).
func (l *LoginLockout) ClearFailures(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, email)
}

// cleanup removes stale entries older than the lockout window.
func (l *LoginLockout) cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-lockoutWindow)
	for email, a := range l.attempts {
		recent := 0
		for _, t := range a.failures {
			if t.After(cutoff) {
				recent++
			}
		}
		if recent == 0 {
			delete(l.attempts, email)
		}
	}
}
