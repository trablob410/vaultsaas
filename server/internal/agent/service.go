package agent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Agent represents an agent identity.
type Agent struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	AgentType     string    `json:"agent_type"`
	AuthMethod    string    `json:"auth_method"`
	AllowedScopes []string  `json:"allowed_scopes"`
	MaxSessionTTL int       `json:"max_session_ttl"`
	Status        string    `json:"status"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Token represents an agent auth token (token_hash never exposed).
type Token struct {
	ID         string     `json:"id"`
	AgentID    string     `json:"agent_id"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	// PlainToken is only set on creation, never stored.
	PlainToken string `json:"token,omitempty"`
}

// Service handles agent identity and token operations.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a new agent Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Create inserts a new agent identity.
func (s *Service) Create(ctx context.Context, projectID, name, description, agentType, createdBy string, maxSessionTTL int, allowedScopes []string) (*Agent, error) {
	if agentType == "" {
		agentType = "custom"
	}
	if maxSessionTTL <= 0 {
		maxSessionTTL = 3600
	}
	if allowedScopes == nil {
		allowedScopes = []string{}
	}

	var a Agent
	err := s.pool.QueryRow(ctx,
		`INSERT INTO agent_identities
		    (project_id, name, description, agent_type, auth_method, allowed_scopes, max_session_ttl, created_by)
		 VALUES ($1, $2, $3, $4, 'token', $5, $6, $7)
		 RETURNING id, project_id, name, description, agent_type, auth_method,
		           allowed_scopes, max_session_ttl, status, created_by, created_at, updated_at`,
		projectID, name, description, agentType, allowedScopes, maxSessionTTL, createdBy,
	).Scan(&a.ID, &a.ProjectID, &a.Name, &a.Description, &a.AgentType, &a.AuthMethod,
		&a.AllowedScopes, &a.MaxSessionTTL, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting agent: %w", err)
	}
	return &a, nil
}

// Get returns an agent by ID.
func (s *Service) Get(ctx context.Context, agentID string) (*Agent, error) {
	var a Agent
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, name, description, agent_type, auth_method,
		        allowed_scopes, max_session_ttl, status, created_by, created_at, updated_at
		 FROM agent_identities WHERE id = $1`,
		agentID,
	).Scan(&a.ID, &a.ProjectID, &a.Name, &a.Description, &a.AgentType, &a.AuthMethod,
		&a.AllowedScopes, &a.MaxSessionTTL, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying agent: %w", err)
	}
	return &a, nil
}

// ListByProject returns all agents in a project.
func (s *Service) ListByProject(ctx context.Context, projectID string) ([]Agent, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, name, description, agent_type, auth_method,
		        allowed_scopes, max_session_ttl, status, created_by, created_at, updated_at
		 FROM agent_identities WHERE project_id = $1
		 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.Name, &a.Description, &a.AgentType, &a.AuthMethod,
			&a.AllowedScopes, &a.MaxSessionTTL, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning agent: %w", err)
		}
		agents = append(agents, a)
	}
	if agents == nil {
		agents = []Agent{}
	}
	return agents, nil
}

// IssueToken generates a new token for an agent.
func (s *Service) IssueToken(ctx context.Context, agentID string, scopes []string, expiresAt *time.Time) (*Token, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("generating token bytes: %w", err)
	}
	plainToken := base64.RawURLEncoding.EncodeToString(raw)

	h := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(h[:])

	if scopes == nil {
		scopes = []string{}
	}

	var t Token
	err := s.pool.QueryRow(ctx,
		`INSERT INTO agent_tokens (agent_id, token_hash, scopes, expires_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, agent_id, scopes, expires_at, last_used_at, created_at, revoked_at`,
		agentID, tokenHash, scopes, expiresAt,
	).Scan(&t.ID, &t.AgentID, &t.Scopes, &t.ExpiresAt, &t.LastUsedAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting token: %w", err)
	}
	t.PlainToken = plainToken
	return &t, nil
}

// RevokeToken marks a token as revoked.
func (s *Service) RevokeToken(ctx context.Context, tokenID string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE agent_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`,
		tokenID,
	)
	if err != nil {
		return fmt.Errorf("revoking token: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ValidateToken hashes plainToken and looks it up, checking expiry/revocation.
func (s *Service) ValidateToken(ctx context.Context, plainToken string) (*Token, error) {
	h := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(h[:])

	var t Token
	err := s.pool.QueryRow(ctx,
		`SELECT id, agent_id, scopes, expires_at, last_used_at, created_at, revoked_at
		 FROM agent_tokens
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		   AND (expires_at IS NULL OR expires_at > NOW())`,
		tokenHash,
	).Scan(&t.ID, &t.AgentID, &t.Scopes, &t.ExpiresAt, &t.LastUsedAt, &t.CreatedAt, &t.RevokedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("validating token: %w", err)
	}

	// Update last_used_at asynchronously — best-effort, don't fail validation on error.
	_, _ = s.pool.Exec(ctx,
		`UPDATE agent_tokens SET last_used_at = NOW() WHERE id = $1`,
		t.ID,
	)
	return &t, nil
}

// ListTokens returns non-revoked tokens for an agent.
func (s *Service) ListTokens(ctx context.Context, agentID string) ([]Token, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, agent_id, scopes, expires_at, last_used_at, created_at, revoked_at
		 FROM agent_tokens
		 WHERE agent_id = $1 AND revoked_at IS NULL
		 ORDER BY created_at DESC`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying tokens: %w", err)
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(&t.ID, &t.AgentID, &t.Scopes, &t.ExpiresAt, &t.LastUsedAt, &t.CreatedAt, &t.RevokedAt); err != nil {
			return nil, fmt.Errorf("scanning token: %w", err)
		}
		tokens = append(tokens, t)
	}
	if tokens == nil {
		tokens = []Token{}
	}
	return tokens, nil
}
