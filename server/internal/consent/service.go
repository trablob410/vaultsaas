package consent

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles consent recording and checking.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a consent Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// RecordConsent records a user consent entry.
func (s *Service) RecordConsent(ctx context.Context, input ConsentInput) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO user_consent_logs (user_id, secret_id, credential_type, consent_type, granted, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6::inet, $7)`,
		input.UserID, input.SecretID, input.CredentialType, input.ConsentType,
		input.Granted, nilIfEmpty(input.IPAddress), nilIfEmpty(input.UserAgent),
	)
	if err != nil {
		return fmt.Errorf("recording consent: %w", err)
	}
	return nil
}

// HasConsent checks if a user has granted consent for a specific secret and type.
func (s *Service) HasConsent(ctx context.Context, userID, secretID, consentType string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM user_consent_logs
			WHERE user_id = $1 AND secret_id = $2 AND consent_type = $3 AND granted = true
		)`,
		userID, secretID, consentType,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking consent: %w", err)
	}
	return exists, nil
}

// ConsentInput represents a consent recording request.
type ConsentInput struct {
	UserID         string `json:"user_id"`
	SecretID       string `json:"secret_id"`
	CredentialType string `json:"credential_type"`
	ConsentType    string `json:"consent_type"`
	Granted        bool   `json:"granted"`
	IPAddress      string `json:"ip_address,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
