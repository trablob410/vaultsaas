package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Workspace struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Create inserts a new workspace under the given org.
func (s *Service) Create(ctx context.Context, orgID, name, slug string) (*Workspace, error) {
	var ws Workspace
	err := s.pool.QueryRow(ctx,
		`INSERT INTO workspaces (org_id, name, slug)
		 VALUES ($1, $2, $3)
		 RETURNING id, org_id, name, slug, created_at, updated_at`,
		orgID, name, slug,
	).Scan(&ws.ID, &ws.OrgID, &ws.Name, &ws.Slug, &ws.CreatedAt, &ws.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting workspace: %w", err)
	}
	return &ws, nil
}

// Get returns a workspace by ID.
func (s *Service) Get(ctx context.Context, id string) (*Workspace, error) {
	var ws Workspace
	err := s.pool.QueryRow(ctx,
		`SELECT id, org_id, name, slug, created_at, updated_at
		 FROM workspaces WHERE id = $1`,
		id,
	).Scan(&ws.ID, &ws.OrgID, &ws.Name, &ws.Slug, &ws.CreatedAt, &ws.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying workspace: %w", err)
	}
	return &ws, nil
}

// ListByOrg returns all workspaces belonging to an org.
func (s *Service) ListByOrg(ctx context.Context, orgID string) ([]Workspace, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, name, slug, created_at, updated_at
		 FROM workspaces WHERE org_id = $1
		 ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.OrgID, &ws.Name, &ws.Slug, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning workspace: %w", err)
		}
		workspaces = append(workspaces, ws)
	}
	if workspaces == nil {
		workspaces = []Workspace{}
	}
	return workspaces, nil
}
