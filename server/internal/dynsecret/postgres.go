package dynsecret

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PostgresProvider creates temporary DB roles with TTL.
// Config keys: host, port, database, admin_user, admin_password, ssl_mode
type PostgresProvider struct {
	config ProviderConfig
}

func (p *PostgresProvider) Type() string { return "postgres" }

func (p *PostgresProvider) connStr() string {
	host := p.config.Config["host"]
	port := p.config.Config["port"]
	if port == "" {
		port = "5432"
	}
	db := p.config.Config["database"]
	user := p.config.Config["admin_user"]
	pass := p.config.Config["admin_password"]
	sslMode := p.config.Config["ssl_mode"]
	if sslMode == "" {
		sslMode = "require"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pass, host, port, db, sslMode)
}

func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Create issues a new temporary role via CREATE ROLE ... VALID UNTIL.
func (p *PostgresProvider) Create(ctx context.Context, req LeaseRequest) (*Lease, error) {
	conn, err := pgx.Connect(ctx, p.connStr())
	if err != nil {
		return nil, fmt.Errorf("postgres provider connect: %w", err)
	}
	defer conn.Close(ctx)

	suffix, err := randHex(4)
	if err != nil {
		return nil, err
	}
	password, err := randHex(16)
	if err != nil {
		return nil, err
	}

	username := "valt_" + suffix
	expiresAt := time.Now().Add(req.TTL)
	validUntil := expiresAt.UTC().Format(time.RFC3339)
	db := p.config.Config["database"]

	_, err = conn.Exec(ctx,
		fmt.Sprintf(`CREATE ROLE "%s" LOGIN PASSWORD '%s' VALID UNTIL '%s' CONNECTION LIMIT 5`,
			username, password, validUntil))
	if err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	_, err = conn.Exec(ctx, fmt.Sprintf(`GRANT CONNECT ON DATABASE "%s" TO "%s"`, db, username))
	if err != nil {
		// Best-effort cleanup
		_, _ = conn.Exec(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS "%s"`, username))
		return nil, fmt.Errorf("grant connect: %w", err)
	}

	return &Lease{
		Credentials: map[string]string{
			"username": username,
			"password": password,
			"host":     p.config.Config["host"],
			"port":     p.config.Config["port"],
			"database": db,
			"ssl_mode": p.config.Config["ssl_mode"],
		},
		ExpiresAt:  expiresAt,
		ProviderID: p.config.ID,
	}, nil
}

// Revoke drops the temporary role.
func (p *PostgresProvider) Revoke(ctx context.Context, leaseID string) error {
	// leaseID here is the username stored in credentials
	conn, err := pgx.Connect(ctx, p.connStr())
	if err != nil {
		return fmt.Errorf("postgres provider connect: %w", err)
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS "%s"`, leaseID))
	return err
}

// Renew extends the VALID UNTIL timestamp for an existing role.
func (p *PostgresProvider) Renew(ctx context.Context, leaseID string, ttl time.Duration) (*Lease, error) {
	conn, err := pgx.Connect(ctx, p.connStr())
	if err != nil {
		return nil, fmt.Errorf("postgres provider connect: %w", err)
	}
	defer conn.Close(ctx)
	newExpiry := time.Now().Add(ttl)
	_, err = conn.Exec(ctx,
		fmt.Sprintf(`ALTER ROLE "%s" VALID UNTIL '%s'`, leaseID, newExpiry.UTC().Format(time.RFC3339)))
	if err != nil {
		return nil, fmt.Errorf("renew role: %w", err)
	}
	return &Lease{
		ID:         leaseID,
		ExpiresAt:  newExpiry,
		ProviderID: p.config.ID,
	}, nil
}
