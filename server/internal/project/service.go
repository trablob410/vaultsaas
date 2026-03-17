package project

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectMember struct {
	ProjectID   string    `json:"project_id"`
	UserID      string    `json:"user_id"`
	Role        string    `json:"role"`
	Permissions string    `json:"permissions"` // raw JSON string
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Create inserts a new project under the given workspace.
func (s *Service) Create(ctx context.Context, workspaceID, name, slug string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (workspace_id, name, slug)
		 VALUES ($1, $2, $3)
		 RETURNING id, workspace_id, name, slug, created_at, updated_at`,
		workspaceID, name, slug,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting project: %w", err)
	}
	return &p, nil
}

// Get returns a project by ID.
func (s *Service) Get(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`SELECT id, workspace_id, name, slug, created_at, updated_at
		 FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying project: %w", err)
	}
	return &p, nil
}

// ListByWorkspace returns all projects in a workspace.
func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, workspace_id, name, slug, created_at, updated_at
		 FROM projects WHERE workspace_id = $1
		 ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []Project{}
	}
	return projects, nil
}

// AddMember adds a user to a project with the given role.
func (s *Service) AddMember(ctx context.Context, projectID, userID, role string) (*ProjectMember, error) {
	var m ProjectMember
	err := s.pool.QueryRow(ctx,
		`INSERT INTO project_memberships (project_id, user_id, role, permissions)
		 VALUES ($1, $2, $3, '{}')
		 ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role
		 RETURNING project_id, user_id, role, permissions::text, created_at`,
		projectID, userID, role,
	).Scan(&m.ProjectID, &m.UserID, &m.Role, &m.Permissions, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding project member: %w", err)
	}
	return &m, nil
}

// ListMembers returns all members of a project.
func (s *Service) ListMembers(ctx context.Context, projectID string) ([]ProjectMember, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT project_id, user_id, role, permissions::text, created_at
		 FROM project_memberships WHERE project_id = $1
		 ORDER BY created_at ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying project members: %w", err)
	}
	defer rows.Close()

	var members []ProjectMember
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Role, &m.Permissions, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning project member: %w", err)
		}
		members = append(members, m)
	}
	if members == nil {
		members = []ProjectMember{}
	}
	return members, nil
}
