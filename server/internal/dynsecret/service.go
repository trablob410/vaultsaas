package dynsecret

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service manages dynamic secret providers and leases.
type Service struct {
	db *pgxpool.Pool
}

// NewService creates a new Service.
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// CreateProvider persists a new provider config.
// TODO(Phase13): encrypt config with project key before storing.
func (s *Service) CreateProvider(ctx context.Context, projectID, name, providerType string, config map[string]string, userID string) (*ProviderConfig, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	var pc ProviderConfig
	err = s.db.QueryRow(ctx, `
		INSERT INTO dynamic_providers (project_id, name, provider_type, config_enc, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, project_id, name, provider_type, status`,
		projectID, name, providerType, raw, userID,
	).Scan(&pc.ID, &pc.ProjectID, &pc.Name, &pc.ProviderType, &pc.Status)
	if err != nil {
		return nil, fmt.Errorf("insert provider: %w", err)
	}
	pc.Config = config
	return &pc, nil
}

// ListProviders returns all providers for a project.
func (s *Service) ListProviders(ctx context.Context, projectID string) ([]ProviderConfig, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, name, provider_type, config_enc, status
		FROM dynamic_providers WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProviderConfig
	for rows.Next() {
		var pc ProviderConfig
		var raw []byte
		if err := rows.Scan(&pc.ID, &pc.ProjectID, &pc.Name, &pc.ProviderType, &raw, &pc.Status); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &pc.Config) // best effort
		out = append(out, pc)
	}
	return out, rows.Err()
}

// GetProvider returns a single provider by ID.
func (s *Service) GetProvider(ctx context.Context, id string) (*ProviderConfig, error) {
	var pc ProviderConfig
	var raw []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, name, provider_type, config_enc, status
		FROM dynamic_providers WHERE id = $1`, id,
	).Scan(&pc.ID, &pc.ProjectID, &pc.Name, &pc.ProviderType, &raw, &pc.Status)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(raw, &pc.Config)
	return &pc, nil
}

// newProviderInstance instantiates the correct Provider implementation.
func newProviderInstance(pc *ProviderConfig) (Provider, error) {
	switch pc.ProviderType {
	case "postgres":
		return &PostgresProvider{config: *pc}, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", pc.ProviderType)
	}
}

// CreateLease generates credentials from a provider and persists the lease.
// TODO: encrypt secret_data_enc with project key.
func (s *Service) CreateLease(ctx context.Context, providerID, agentID, requestID string, ttlSeconds int) (*Lease, error) {
	pc, err := s.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	prov, err := newProviderInstance(pc)
	if err != nil {
		return nil, err
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	lease, err := prov.Create(ctx, LeaseRequest{TTL: ttl, AgentID: agentID, RequestID: requestID})
	if err != nil {
		return nil, fmt.Errorf("provider create: %w", err)
	}

	credBytes, err := json.Marshal(lease.Credentials)
	if err != nil {
		return nil, err
	}

	var agentIDPtr, reqIDPtr *string
	if agentID != "" {
		agentIDPtr = &agentID
	}
	if requestID != "" {
		reqIDPtr = &requestID
	}

	err = s.db.QueryRow(ctx, `
		INSERT INTO dynamic_leases (provider_id, agent_id, access_request_id, secret_data_enc, ttl_seconds, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		providerID, agentIDPtr, reqIDPtr, credBytes, ttlSeconds, lease.ExpiresAt,
	).Scan(&lease.ID)
	if err != nil {
		return nil, fmt.Errorf("insert lease: %w", err)
	}
	lease.ProviderID = providerID
	return lease, nil
}

// RevokeLease marks a lease revoked and drops the backend credential.
func (s *Service) RevokeLease(ctx context.Context, leaseID string) error {
	var providerID string
	var credRaw []byte
	err := s.db.QueryRow(ctx, `
		SELECT provider_id, secret_data_enc FROM dynamic_leases
		WHERE id = $1 AND revoked_at IS NULL`, leaseID,
	).Scan(&providerID, &credRaw)
	if err != nil {
		return fmt.Errorf("lease not found: %w", err)
	}

	var creds map[string]string
	_ = json.Unmarshal(credRaw, &creds)

	_, err = s.db.Exec(ctx, `UPDATE dynamic_leases SET revoked_at = now() WHERE id = $1`, leaseID)
	if err != nil {
		return err
	}

	// Best-effort backend revocation using username as identifier
	if username, ok := creds["username"]; ok {
		pc, pcErr := s.GetProvider(ctx, providerID)
		if pcErr == nil {
			prov, provErr := newProviderInstance(pc)
			if provErr == nil {
				_ = prov.Revoke(ctx, username)
			}
		}
	}
	return nil
}

// ListActiveLeases returns non-revoked leases for a provider.
func (s *Service) ListActiveLeases(ctx context.Context, providerID string) ([]LeaseInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, provider_id, agent_id, access_request_id, ttl_seconds, expires_at, revoked_at, created_at
		FROM dynamic_leases WHERE provider_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LeaseInfo
	for rows.Next() {
		var li LeaseInfo
		if err := rows.Scan(&li.ID, &li.ProviderID, &li.AgentID, &li.AccessRequestID,
			&li.TTLSeconds, &li.ExpiresAt, &li.RevokedAt, &li.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, li)
	}
	return out, rows.Err()
}

// StartExpiryWorker runs a background goroutine that marks expired leases every 60s.
func (s *Service) StartExpiryWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, _ = s.db.Exec(ctx,
					`UPDATE dynamic_leases SET revoked_at = now() WHERE expires_at < now() AND revoked_at IS NULL`)
			case <-ctx.Done():
				return
			}
		}
	}()
}
