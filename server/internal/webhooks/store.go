package webhooks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Webhook represents a registered org webhook endpoint.
type Webhook struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	URL         string    `json:"url"`
	SecretKey   string    `json:"secret_key,omitempty"`
	Events      []string  `json:"events"`
	Active      bool      `json:"active"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store manages org_webhooks table operations.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a webhook Store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Create registers a new webhook. Returns the webhook with full secret_key (shown once).
func (s *Store) Create(ctx context.Context, orgID, url, description string, events []string) (*Webhook, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("webhook URL must use HTTPS")
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("at least one event required")
	}

	// Enforce max 10 per org
	var count int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM org_webhooks WHERE org_id = $1`, orgID).Scan(&count); err != nil {
		return nil, fmt.Errorf("counting webhooks: %w", err)
	}
	if count >= 10 {
		return nil, fmt.Errorf("maximum 10 webhooks per org")
	}

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("generating secret: %w", err)
	}
	secretKey := hex.EncodeToString(secretBytes)

	var w Webhook
	err := s.pool.QueryRow(ctx,
		`INSERT INTO org_webhooks (org_id, url, secret_key, events, description)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, org_id, url, secret_key, events, active, description, created_at`,
		orgID, url, secretKey, events, description,
	).Scan(&w.ID, &w.OrgID, &w.URL, &w.SecretKey, &w.Events, &w.Active, &w.Description, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting webhook: %w", err)
	}
	return &w, nil
}

// List returns all webhooks for an org with masked secret keys.
func (s *Store) List(ctx context.Context, orgID string) ([]Webhook, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, url, secret_key, events, active, description, created_at
		 FROM org_webhooks WHERE org_id = $1 ORDER BY created_at`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing webhooks: %w", err)
	}
	defer rows.Close()

	var result []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.OrgID, &w.URL, &w.SecretKey, &w.Events, &w.Active, &w.Description, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		// Mask secret key — show only last 8 chars
		if len(w.SecretKey) > 8 {
			w.SecretKey = "****" + w.SecretKey[len(w.SecretKey)-8:]
		}
		result = append(result, w)
	}
	if result == nil {
		result = []Webhook{}
	}
	return result, nil
}

// Delete removes a webhook.
func (s *Store) Delete(ctx context.Context, orgID, webhookID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM org_webhooks WHERE id = $1 AND org_id = $2`,
		webhookID, orgID,
	)
	return err
}

// ListActive returns active webhooks for an org matching a specific event (with full secret).
func (s *Store) ListActive(ctx context.Context, orgID, event string) ([]Webhook, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, url, secret_key, events, active, description, created_at
		 FROM org_webhooks
		 WHERE org_id = $1 AND active = true AND $2 = ANY(events)`,
		orgID, event,
	)
	if err != nil {
		return nil, fmt.Errorf("listing active webhooks: %w", err)
	}
	defer rows.Close()

	var result []Webhook
	for rows.Next() {
		var w Webhook
		if err := rows.Scan(&w.ID, &w.OrgID, &w.URL, &w.SecretKey, &w.Events, &w.Active, &w.Description, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning webhook: %w", err)
		}
		result = append(result, w)
	}
	return result, nil
}
