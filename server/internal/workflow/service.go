package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valt-dev/valt/server/internal/policy"
)

// AccessRequest represents an access request row.
type AccessRequest struct {
	ID                       string     `json:"id"`
	SecretID                 string     `json:"secret_id"`
	SecretName               string     `json:"secret_name,omitempty"`
	RequesterUserID          string     `json:"requester_user_id"`
	RequesterType            string     `json:"requester_type"`
	AIAgentID                *string    `json:"ai_agent_id,omitempty"`
	Status                   string     `json:"status"`
	Reason                   *string    `json:"reason,omitempty"`
	RejectionReason          *string    `json:"rejection_reason,omitempty"`
	RequestedDurationMinutes int        `json:"requested_duration_minutes"`
	DecidedBy                *string    `json:"decided_by,omitempty"`
	DecidedAt                *time.Time `json:"decided_at,omitempty"`
	ExpiresAt                *time.Time `json:"expires_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
}

// CreateRequestInput holds the data for creating an access request.
type CreateRequestInput struct {
	SecretID        string
	RequesterUserID string
	RequesterType   string
	AIAgentID       string
	Reason          string
	DurationMinutes int
	CredentialType  string // from the secret
}

// Service handles the approval workflow.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a workflow Service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// CreateRequest creates a new access request, enforcing policy.
func (s *Service) CreateRequest(ctx context.Context, input CreateRequestInput) (*AccessRequest, error) {
	p := policy.ForCredentialType(input.CredentialType)

	// Enforce reason requirements
	if p.RequireReason && len(input.Reason) < p.MinReasonLength {
		return nil, fmt.Errorf("reason must be at least %d characters for %s credentials", p.MinReasonLength, input.CredentialType)
	}

	// Cap duration
	dur := input.DurationMinutes
	if dur <= 0 || dur > p.MaxDurationMinutes {
		dur = p.MaxDurationMinutes
	}

	// Check daily request limit
	var dailyCount int
	var dailyErr error
	if input.RequesterUserID != "" {
		dailyErr = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM access_requests
			 WHERE requester_user_id = $1 AND secret_id = $2
			   AND created_at >= NOW() - INTERVAL '24 hours'`,
			input.RequesterUserID, input.SecretID,
		).Scan(&dailyCount)
	} else {
		dailyErr = s.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM access_requests
			 WHERE ai_agent_id = $1 AND secret_id = $2
			   AND created_at >= NOW() - INTERVAL '24 hours'`,
			input.AIAgentID, input.SecretID,
		).Scan(&dailyCount)
	}
	if dailyErr != nil {
		return nil, fmt.Errorf("checking daily limit: %w", dailyErr)
	}
	if dailyCount >= p.MaxRequestsPerDay {
		return nil, fmt.Errorf("daily request limit (%d) exceeded", p.MaxRequestsPerDay)
	}

	// Check cool-down
	if p.CoolDownMinutes > 0 {
		var recentCount int
		var coolErr error
		if input.RequesterUserID != "" {
			coolErr = s.pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM access_requests
				 WHERE requester_user_id = $1 AND secret_id = $2
				   AND created_at >= NOW() - INTERVAL '1 minute' * $3`,
				input.RequesterUserID, input.SecretID, p.CoolDownMinutes,
			).Scan(&recentCount)
		} else {
			coolErr = s.pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM access_requests
				 WHERE ai_agent_id = $1 AND secret_id = $2
				   AND created_at >= NOW() - INTERVAL '1 minute' * $3`,
				input.AIAgentID, input.SecretID, p.CoolDownMinutes,
			).Scan(&recentCount)
		}
		if coolErr != nil {
			return nil, fmt.Errorf("checking cool-down: %w", coolErr)
		}
		if recentCount > 0 {
			return nil, fmt.Errorf("please wait %d minutes between requests", p.CoolDownMinutes)
		}
	}

	// Determine initial status
	initialStatus := "pending"
	if p.AllowAutoApprove && !p.RequireApproval {
		initialStatus = "approved"
	}

	var req AccessRequest
	var aiAgentID *string
	if input.AIAgentID != "" {
		aiAgentID = &input.AIAgentID
	}
	var requesterUserID *string
	if input.RequesterUserID != "" {
		requesterUserID = &input.RequesterUserID
	}

	insertErr := s.pool.QueryRow(ctx,
		`INSERT INTO access_requests (secret_id, requester_user_id, requester_type, ai_agent_id, status, reason, requested_duration_minutes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, secret_id, COALESCE(requester_user_id, '') AS requester_user_id, requester_type, ai_agent_id, status, reason, requested_duration_minutes, decided_by, decided_at, expires_at, created_at`,
		input.SecretID, requesterUserID, input.RequesterType, aiAgentID,
		initialStatus, input.Reason, dur,
	).Scan(&req.ID, &req.SecretID, &req.RequesterUserID, &req.RequesterType,
		&req.AIAgentID, &req.Status, &req.Reason, &req.RequestedDurationMinutes,
		&req.DecidedBy, &req.DecidedAt, &req.ExpiresAt, &req.CreatedAt)
	if insertErr != nil {
		return nil, fmt.Errorf("inserting access request: %w", insertErr)
	}

	return &req, nil
}

// ListPending returns pending access requests for the secret owner.
func (s *Service) ListPending(ctx context.Context, ownerUserID, status string, limit, offset int) ([]AccessRequest, int, error) {
	filterStatus := "pending"
	if status != "" {
		filterStatus = status
	}

	var total int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM access_requests ar
		 JOIN secrets s ON s.id = ar.secret_id
		 WHERE s.user_id = $1 AND ar.status = $2 AND s.deleted_at IS NULL`,
		ownerUserID, filterStatus,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting requests: %w", err)
	}

	rows, err := s.pool.Query(ctx,
		`SELECT ar.id, ar.secret_id, COALESCE(s.name, '') AS secret_name,
		        COALESCE(ar.requester_user_id, '') AS requester_user_id, ar.requester_type, ar.ai_agent_id,
		        ar.status, ar.reason, ar.requested_duration_minutes, ar.decided_by, ar.decided_at, ar.expires_at, ar.created_at
		 FROM access_requests ar
		 JOIN secrets s ON s.id = ar.secret_id AND s.user_id = $1 AND s.deleted_at IS NULL
		 WHERE ar.status = $2
		 ORDER BY ar.created_at DESC LIMIT $3 OFFSET $4`,
		ownerUserID, filterStatus, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("querying requests: %w", err)
	}
	defer rows.Close()

	var requests []AccessRequest
	for rows.Next() {
		var r AccessRequest
		if err := rows.Scan(&r.ID, &r.SecretID, &r.SecretName, &r.RequesterUserID, &r.RequesterType,
			&r.AIAgentID, &r.Status, &r.Reason, &r.RequestedDurationMinutes,
			&r.DecidedBy, &r.DecidedAt, &r.ExpiresAt, &r.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scanning request: %w", err)
		}
		requests = append(requests, r)
	}
	if requests == nil {
		requests = []AccessRequest{}
	}

	return requests, total, nil
}

// Approve approves a pending access request.
func (s *Service) Approve(ctx context.Context, requestID, approverUserID string) (*AccessRequest, error) {
	var req AccessRequest
	err := s.pool.QueryRow(ctx,
		`UPDATE access_requests
		 SET status = 'approved', decided_by = $1, decided_at = NOW(),
		     expires_at = NOW() + (requested_duration_minutes || ' minutes')::INTERVAL
		 WHERE id = $2 AND status = 'pending'
		 RETURNING id, secret_id, COALESCE(requester_user_id, '') AS requester_user_id, requester_type, ai_agent_id, status, reason,
		           requested_duration_minutes, decided_by, decided_at, expires_at, created_at`,
		approverUserID, requestID,
	).Scan(&req.ID, &req.SecretID, &req.RequesterUserID, &req.RequesterType,
		&req.AIAgentID, &req.Status, &req.Reason, &req.RequestedDurationMinutes,
		&req.DecidedBy, &req.DecidedAt, &req.ExpiresAt, &req.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("request not found or not pending")
		}
		return nil, fmt.Errorf("approving request: %w", err)
	}
	return &req, nil
}

// Reject rejects a pending access request.
func (s *Service) Reject(ctx context.Context, requestID, approverUserID, rejectionReason string) (*AccessRequest, error) {
	var req AccessRequest
	err := s.pool.QueryRow(ctx,
		`UPDATE access_requests
		 SET status = 'rejected', decided_by = $1, decided_at = NOW(), rejection_reason = $3
		 WHERE id = $2 AND status = 'pending'
		 RETURNING id, secret_id, COALESCE(requester_user_id, '') AS requester_user_id, requester_type, ai_agent_id, status, reason,
		           rejection_reason, requested_duration_minutes, decided_by, decided_at, expires_at, created_at`,
		approverUserID, requestID, rejectionReason,
	).Scan(&req.ID, &req.SecretID, &req.RequesterUserID, &req.RequesterType,
		&req.AIAgentID, &req.Status, &req.Reason, &req.RejectionReason,
		&req.RequestedDurationMinutes, &req.DecidedBy, &req.DecidedAt, &req.ExpiresAt, &req.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("request not found or not pending")
		}
		return nil, fmt.Errorf("rejecting request: %w", err)
	}
	return &req, nil
}

// IsAssignedApprover returns true if userID has an entry in approval_steps for the given requestID.
func (s *Service) IsAssignedApprover(ctx context.Context, requestID, userID string) (bool, error) {
	var cnt int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM approval_steps WHERE request_id = $1 AND approver_user_id = $2`,
		requestID, userID).Scan(&cnt)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// GetRequestByID fetches a single access request.
func (s *Service) GetRequestByID(ctx context.Context, requestID string) (*AccessRequest, error) {
	var req AccessRequest
	err := s.pool.QueryRow(ctx,
		`SELECT id, secret_id, COALESCE(requester_user_id, '') AS requester_user_id, requester_type, ai_agent_id, status, reason,
		        rejection_reason, requested_duration_minutes, decided_by, decided_at, expires_at, created_at
		 FROM access_requests WHERE id = $1`,
		requestID,
	).Scan(&req.ID, &req.SecretID, &req.RequesterUserID, &req.RequesterType,
		&req.AIAgentID, &req.Status, &req.Reason, &req.RejectionReason,
		&req.RequestedDurationMinutes, &req.DecidedBy, &req.DecidedAt, &req.ExpiresAt, &req.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting request: %w", err)
	}
	return &req, nil
}
