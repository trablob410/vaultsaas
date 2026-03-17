package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CredentialSession represents an active credential session.
type CredentialSession struct {
	ID              string     `json:"id"`
	AccessRequestID string     `json:"access_request_id"`
	CredentialType  *string    `json:"credential_type,omitempty"`
	Status          string     `json:"status"`
	ExpiresAt       time.Time  `json:"expires_at"`
	UsageCount      int        `json:"usage_count"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CredentialManager handles temporary credential lifecycle.
type CredentialManager struct {
	pool *pgxpool.Pool
}

// NewCredentialManager creates a CredentialManager.
func NewCredentialManager(pool *pgxpool.Pool) *CredentialManager {
	return &CredentialManager{pool: pool}
}

// IssueCredential creates a credential session for an approved request.
func (m *CredentialManager) IssueCredential(ctx context.Context, requestID, credentialType string, durationMinutes int) (*CredentialSession, error) {
	expiresAt := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	var session CredentialSession
	err := m.pool.QueryRow(ctx,
		`INSERT INTO credential_sessions (access_request_id, credential_type, status, expires_at)
		 VALUES ($1, $2, 'active', $3)
		 RETURNING id, access_request_id, credential_type, status, expires_at, usage_count, revoked_at, created_at`,
		requestID, credentialType, expiresAt,
	).Scan(&session.ID, &session.AccessRequestID, &session.CredentialType,
		&session.Status, &session.ExpiresAt, &session.UsageCount, &session.RevokedAt, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("issuing credential: %w", err)
	}
	return &session, nil
}

// GetCredential retrieves an active credential session and increments usage.
func (m *CredentialManager) GetCredential(ctx context.Context, requestID string) (*CredentialSession, error) {
	var session CredentialSession
	err := m.pool.QueryRow(ctx,
		`UPDATE credential_sessions
		 SET usage_count = usage_count + 1
		 WHERE access_request_id = $1 AND status = 'active' AND expires_at > NOW()
		 RETURNING id, access_request_id, credential_type, status, expires_at, usage_count, revoked_at, created_at`,
		requestID,
	).Scan(&session.ID, &session.AccessRequestID, &session.CredentialType,
		&session.Status, &session.ExpiresAt, &session.UsageCount, &session.RevokedAt, &session.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no active credential for this request")
		}
		return nil, fmt.Errorf("getting credential: %w", err)
	}
	return &session, nil
}

// RevokeCredential revokes an active credential session.
func (m *CredentialManager) RevokeCredential(ctx context.Context, requestID string) error {
	tag, err := m.pool.Exec(ctx,
		`UPDATE credential_sessions
		 SET status = 'revoked', revoked_at = NOW()
		 WHERE access_request_id = $1 AND status = 'active'`,
		requestID,
	)
	if err != nil {
		return fmt.Errorf("revoking credential: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no active credential to revoke")
	}

	// Also update access request status
	_, _ = m.pool.Exec(ctx,
		`UPDATE access_requests SET status = 'revoked' WHERE id = $1 AND status IN ('approved', 'active')`,
		requestID,
	)
	return nil
}

// AutoRevokeIfSingleUse revokes the credential if usage reaches 1 (single-use).
func (m *CredentialManager) AutoRevokeIfSingleUse(ctx context.Context, requestID string, singleUse bool) error {
	if !singleUse {
		return nil
	}
	return m.RevokeCredential(ctx, requestID)
}
