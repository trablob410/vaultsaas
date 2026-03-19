package notify

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ActionToken represents a single-use approve/reject token.
type ActionToken struct {
	ID        string
	RequestID string
	Action    string
	ExpiresAt time.Time
}

// ActionTokenStore manages single-use action tokens backed by Postgres.
type ActionTokenStore struct {
	pool *pgxpool.Pool
}

// NewActionTokenStore creates an ActionTokenStore.
func NewActionTokenStore(pool *pgxpool.Pool) *ActionTokenStore {
	return &ActionTokenStore{pool: pool}
}

// Create generates a random token, stores its SHA-256 hash, returns the raw token.
func (s *ActionTokenStore) Create(ctx context.Context, requestID, action string, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	token := hex.EncodeToString(raw)
	hash := sha256Token(token)
	expiresAt := time.Now().Add(ttl)

	_, err := s.pool.Exec(ctx,
		`INSERT INTO request_action_tokens (request_id, action, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		requestID, action, hash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("storing token: %w", err)
	}
	return token, nil
}

// Consume validates and atomically marks the token as used. Returns the token metadata.
// Returns an error if the token is invalid, already used, or expired.
func (s *ActionTokenStore) Consume(ctx context.Context, rawToken string) (*ActionToken, error) {
	hash := sha256Token(rawToken)
	var t ActionToken
	err := s.pool.QueryRow(ctx,
		`UPDATE request_action_tokens
		 SET used_at = NOW()
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		 RETURNING id, request_id, action, expires_at`,
		hash,
	).Scan(&t.ID, &t.RequestID, &t.Action, &t.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("token invalid or expired")
	}
	return &t, nil
}

func sha256Token(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
