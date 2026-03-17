package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApprovalStep represents one step in an approval chain.
type ApprovalStep struct {
	ID             string     `json:"id"`
	RequestID      string     `json:"request_id"`
	StepOrder      int        `json:"step_order"`
	ApproverUserID string     `json:"approver_user_id"`
	Status         string     `json:"status"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateApprovalChain creates a multi-step approval chain for a request.
func CreateApprovalChain(ctx context.Context, db *pgxpool.Pool, requestID string, approverUserIDs []string) ([]ApprovalStep, error) {
	if len(approverUserIDs) == 0 {
		return nil, fmt.Errorf("at least one approver is required")
	}

	steps := make([]ApprovalStep, 0, len(approverUserIDs))
	for i, approverID := range approverUserIDs {
		var step ApprovalStep
		err := db.QueryRow(ctx,
			`INSERT INTO approval_steps (request_id, step_order, approver_user_id, status)
			 VALUES ($1, $2, $3, 'pending')
			 RETURNING id, request_id, step_order, approver_user_id, status, decided_at, created_at`,
			requestID, i+1, approverID,
		).Scan(&step.ID, &step.RequestID, &step.StepOrder, &step.ApproverUserID,
			&step.Status, &step.DecidedAt, &step.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("inserting approval step %d: %w", i+1, err)
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// GetCurrentStep returns the current pending step for a request.
// Returns nil if all steps are done.
func GetCurrentStep(ctx context.Context, db *pgxpool.Pool, requestID string) (*ApprovalStep, error) {
	var step ApprovalStep
	err := db.QueryRow(ctx,
		`SELECT id, request_id, step_order, approver_user_id, status, decided_at, created_at
		 FROM approval_steps
		 WHERE request_id = $1 AND status = 'pending'
		 ORDER BY step_order ASC
		 LIMIT 1`,
		requestID,
	).Scan(&step.ID, &step.RequestID, &step.StepOrder, &step.ApproverUserID,
		&step.Status, &step.DecidedAt, &step.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getting current step: %w", err)
	}
	return &step, nil
}

// AdvanceChain approves or rejects the current step for the given request.
// Returns true if the entire chain is now complete (all steps approved).
// Any rejection immediately marks the chain as rejected (returns false).
func AdvanceChain(ctx context.Context, db *pgxpool.Pool, requestID, approverID string, approve bool, _ string) (bool, error) {
	step, err := GetCurrentStep(ctx, db, requestID)
	if err != nil {
		return false, err
	}
	if step == nil {
		return false, fmt.Errorf("no pending step found for request")
	}
	if step.ApproverUserID != approverID {
		return false, fmt.Errorf("approver mismatch: expected %s", step.ApproverUserID)
	}

	newStatus := "approved"
	if !approve {
		newStatus = "rejected"
	}

	_, err = db.Exec(ctx,
		`UPDATE approval_steps
		 SET status = $1, decided_at = NOW()
		 WHERE id = $2`,
		newStatus, step.ID,
	)
	if err != nil {
		return false, fmt.Errorf("updating step status: %w", err)
	}

	if !approve {
		return false, nil
	}

	// Check if all steps are now approved
	var pendingCount int
	err = db.QueryRow(ctx,
		`SELECT COUNT(*) FROM approval_steps WHERE request_id = $1 AND status = 'pending'`,
		requestID,
	).Scan(&pendingCount)
	if err != nil {
		return false, fmt.Errorf("checking remaining steps: %w", err)
	}

	return pendingCount == 0, nil
}
