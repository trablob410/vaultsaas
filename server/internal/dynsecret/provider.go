package dynsecret

import (
	"context"
	"time"
)

// LeaseRequest is the input to Provider.Create.
type LeaseRequest struct {
	TTL       time.Duration
	AgentID   string
	RequestID string
	Metadata  map[string]string
}

// Lease is the output of Provider.Create.
type Lease struct {
	ID          string
	Credentials map[string]string // e.g. {"username": "...", "password": "...", "host": "..."}
	ExpiresAt   time.Time
	ProviderID  string
}

// Provider defines the interface for dynamic secret backends.
type Provider interface {
	Type() string
	Create(ctx context.Context, req LeaseRequest) (*Lease, error)
	Revoke(ctx context.Context, leaseID string) error
	Renew(ctx context.Context, leaseID string, ttl time.Duration) (*Lease, error)
}

// ProviderConfig holds decrypted provider configuration.
type ProviderConfig struct {
	ID           string
	ProjectID    string
	Name         string
	ProviderType string
	Config       map[string]string // decrypted config
	Status       string
}

// LeaseInfo is a read-only view of a stored lease.
type LeaseInfo struct {
	ID               string     `json:"id"`
	ProviderID       string     `json:"provider_id"`
	AgentID          *string    `json:"agent_id,omitempty"`
	AccessRequestID  *string    `json:"access_request_id,omitempty"`
	TTLSeconds       int        `json:"ttl_seconds"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
