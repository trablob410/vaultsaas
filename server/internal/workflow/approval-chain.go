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
	ID              string     `json:"id"`
	RequestID       string     `json:"request_id"`
	StepOrder       int        `json:"step_order"`
	ApproverUserID  string     `json:"approver_user_id"`
	Status          string     `json:"status"`
	DecidedAt       *time.Time `json:"decided_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreateApprovalChain creates a multi-step approval chain for a request.
// All steps are inserted in a single transaction — partial failures roll back.
func CreateApprovalChain(ctx context.Context, db *pgxpool.Pool, requestID string, approverUserIDs []string) ([]ApprovalStep, error) {
	if len(approverUserIDs) == 0 {
		return nil, fmt.Errorf("at least one approver is required")
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	steps := make([]ApprovalStep, 0, len(approverUserIDs))
	for i, approverID := range approverUserIDs {
		var step ApprovalStep
		err := tx.QueryRow(ctx,
			`INSERT INTO approval_steps (request_id, step_order, approver_user_id, status)
			 VALUES ($1, $2, $3, 'pending')
			 RETURNING id, request_id, step_order, approver_user_id, status, decided_at, rejection_reason, created_at`,
			requestID, i+1, approverID,
		).Scan(&step.ID, &step.RequestID, &step.StepOrder, &step.ApproverUserID,
			&step.Status, &step.DecidedAt, &step.RejectionReason, &step.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("inserting approval step %d: %w", i+1, err)
		}
		steps = append(steps, step)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing approval chain: %w", err)
	}
	return steps, nil
}

// GetCurrentStep returns the current pending step for a request.
// Returns nil if all steps are done.
func GetCurrentStep(ctx context.Context, db *pgxpool.Pool, requestID string) (*ApprovalStep, error) {
	var step ApprovalStep
	err := db.QueryRow(ctx,
		`SELECT id, request_id, step_order, approver_user_id, status, decided_at, rejection_reason, created_at
		 FROM approval_steps
		 WHERE request_id = $1 AND status = 'pending'
		 ORDER BY step_order ASC
		 LIMIT 1`,
		requestID,
	).Scan(&step.ID, &step.RequestID, &step.StepOrder, &step.ApproverUserID,
		&step.Status, &step.DecidedAt, &step.RejectionReason, &step.CreatedAt)
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
// Runs in a serializable transaction to prevent TOCTOU races.
func AdvanceChain(ctx context.Context, db *pgxpool.Pool, requestID, approverID string, approve bool, reason string) (bool, error) {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return false, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Lock the current pending step to prevent concurrent advances
	var step ApprovalStep
	err = tx.QueryRow(ctx,
		`SELECT id, request_id, step_order, approver_user_id, status, decided_at, rejection_reason, created_at
		 FROM approval_steps
		 WHERE request_id = $1 AND status = 'pending'
		 ORDER BY step_order ASC
		 LIMIT 1
		 FOR UPDATE`,
		requestID,
	).Scan(&step.ID, &step.RequestID, &step.StepOrder, &step.ApproverUserID,
		&step.Status, &step.DecidedAt, &step.RejectionReason, &step.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, fmt.Errorf("no pending step found for request")
		}
		return false, fmt.Errorf("getting current step: %w", err)
	}

	if step.ApproverUserID != approverID {
		return false, fmt.Errorf("approver mismatch: expected %s", step.ApproverUserID)
	}

	newStatus := "approved"
	if !approve {
		newStatus = "rejected"
	}

	if _, err = tx.Exec(ctx,
		`UPDATE approval_steps SET status = $1, decided_at = NOW(), rejection_reason = $2 WHERE id = $3`,
		newStatus, reason, step.ID,
	); err != nil {
		return false, fmt.Errorf("updating step status: %w", err)
	}

	if !approve {
		_ = tx.Commit(ctx)
		return false, nil
	}

	// Check remaining pending steps within the same transaction
	var pendingCount int
	if err = tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM approval_steps WHERE request_id = $1 AND status = 'pending'`,
		requestID,
	).Scan(&pendingCount); err != nil {
		return false, fmt.Errorf("checking remaining steps: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("committing chain advance: %w", err)
	}
	return pendingCount == 0, nil
}
