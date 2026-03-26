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
// Enforces plan-specific limits. -1 means unlimited.
func (t *Tracker) CheckLimit(ctx context.Context, orgID, metric string) (bool, int, int, error) {
	plan, err := t.GetOrgPlan(ctx, orgID)
	if err != nil {
		return true, 0, 0, err
	}

	limit := LimitForPlanMetric(plan, metric)
	if limit < 0 {
		return true, 0, 0, nil // unlimited
	}
	if limit == 0 {
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

// LimitForPlanMetric returns the limit for a metric on a given plan. -1 = unlimited.
func LimitForPlanMetric(plan, metric string) int {
	limits := LimitsForPlan(plan)
	if v, ok := limits[metric]; ok {
		return v
	}
	return 0
}

// LimitsForPlan returns all limits for a given plan.
func LimitsForPlan(plan string) map[string]int {
	switch plan {
	case "pro":
		return map[string]int{"secrets_count": 500, "agents_count": 25, "requests_today": 10000}
	case "team":
		return map[string]int{"secrets_count": -1, "agents_count": -1, "requests_today": 50000}
	default: // free
		return map[string]int{"secrets_count": FreeTierMaxSecrets, "agents_count": FreeTierMaxAgents, "requests_today": FreeTierMaxReqDay}
	}
}
