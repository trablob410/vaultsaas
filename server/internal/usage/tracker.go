package usage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Free tier limits
const (
	FreeTierMaxSecrets = 50
	FreeTierMaxAgents  = 3
	FreeTierMaxReqDay  = 1000
)

// Tracker tracks usage metrics and enforces free tier limits.
type Tracker struct {
	db *pgxpool.Pool
}

// NewTracker creates a new Tracker.
func NewTracker(db *pgxpool.Pool) *Tracker {
	return &Tracker{db: db}
}

// GetOrgPlan returns the plan for an org ("free", "pro", "enterprise").
func (t *Tracker) GetOrgPlan(ctx context.Context, orgID string) (string, error) {
	var plan string
	err := t.db.QueryRow(ctx, `SELECT plan FROM organizations WHERE id = $1`, orgID).Scan(&plan)
	return plan, err
}

// CheckLimit checks if the org is within limits for the given metric.
// Returns (allowed bool, current int, limit int, err error).
// For "free" plan: enforce limits. For other plans: always allow.
func (t *Tracker) CheckLimit(ctx context.Context, orgID, metric string) (bool, int, int, error) {
	plan, err := t.GetOrgPlan(ctx, orgID)
	if err != nil || plan != "free" {
		return true, 0, 0, err
	}

	limit := limitForMetric(metric)
	if limit <= 0 {
		return true, 0, 0, nil
	}

	current, err := t.GetCurrent(ctx, orgID, metric)
	if err != nil {
		return true, 0, 0, err // fail open on error
	}

	return current < limit, current, limit, nil
}

// GetCurrent returns the current count for a metric.
// secrets_count / agents_count: counts scoped to the org via project/workspace chain.
// requests_today: sum from usage_metrics for today.
func (t *Tracker) GetCurrent(ctx context.Context, orgID, metric string) (int, error) {
	switch metric {
	case "secrets_count":
		var count int
		err := t.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM secrets s
			JOIN projects p ON p.id = s.project_id
			JOIN workspaces w ON w.id = p.workspace_id
			WHERE w.org_id = $1
		`, orgID).Scan(&count)
		return count, err

	case "agents_count":
		var count int
		err := t.db.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM agent_identities a
			JOIN projects p ON p.id = a.project_id
			JOIN workspaces w ON w.id = p.workspace_id
			WHERE w.org_id = $1
		`, orgID).Scan(&count)
		return count, err

	case "requests_today":
		today := time.Now().UTC().Format("2006-01-02")
		var total int
		err := t.db.QueryRow(ctx, `
			SELECT COALESCE(SUM(value), 0)
			FROM usage_metrics
			WHERE org_id = $1 AND metric = 'requests_today' AND metric_date = $2
		`, orgID, today).Scan(&total)
		return total, err

	default:
		return 0, nil
	}
}

// IncrementRequests increments the daily request counter for an org.
func (t *Tracker) IncrementRequests(ctx context.Context, orgID string) error {
	today := time.Now().UTC().Format("2006-01-02")
	_, err := t.db.Exec(ctx, `
		INSERT INTO usage_metrics (org_id, metric, value, metric_date)
		VALUES ($1, 'requests_today', 1, $2)
		ON CONFLICT (org_id, metric, metric_date)
		DO UPDATE SET value = usage_metrics.value + 1, updated_at = now()
	`, orgID, today)
	return err
}

func limitForMetric(metric string) int {
	switch metric {
	case "secrets_count":
		return FreeTierMaxSecrets
	case "agents_count":
		return FreeTierMaxAgents
	case "requests_today":
		return FreeTierMaxReqDay
	default:
		return 0
	}
}
