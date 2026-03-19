package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Resolver resolves the effective Policy for a secret by merging custom overrides.
type Resolver struct {
	pool *pgxpool.Pool
}

// NewResolver creates a Resolver.
func NewResolver(pool *pgxpool.Pool) *Resolver {
	return &Resolver{pool: pool}
}

// Resolve returns the effective Policy and custom approver list for a secret.
// Resolution order (most specific wins): secret.policy_config → project.policy_config → tier defaults.
// customApprovers is empty when no custom approvers are configured.
func (r *Resolver) Resolve(ctx context.Context, secretID, credentialType string) (Policy, []string, error) {
	base := ForCredentialType(credentialType)

	// Fetch secret policy_config + project_id in one query
	var secretPolicyCfg []byte
	var projectID *string
	err := r.pool.QueryRow(ctx,
		`SELECT policy_config, project_id FROM secrets WHERE id = $1 AND deleted_at IS NULL`,
		secretID,
	).Scan(&secretPolicyCfg, &projectID)
	if err != nil {
		return base, nil, fmt.Errorf("fetching secret policy: %w", err)
	}

	var customApprovers []string

	// Apply project policy first (lower priority)
	if projectID != nil {
		var projectPolicyCfg []byte
		_ = r.pool.QueryRow(ctx,
			`SELECT policy_config FROM projects WHERE id = $1`,
			*projectID,
		).Scan(&projectPolicyCfg)
		if len(projectPolicyCfg) > 0 {
			var pc CustomPolicy
			if json.Unmarshal(projectPolicyCfg, &pc) == nil {
				base = pc.ApplyTo(base)
				customApprovers = pc.Approvers
			}
		}
	}

	// Apply secret policy (higher priority, overrides project)
	if len(secretPolicyCfg) > 0 {
		var sc CustomPolicy
		if json.Unmarshal(secretPolicyCfg, &sc) == nil {
			base = sc.ApplyTo(base)
			if len(sc.Approvers) > 0 {
				customApprovers = sc.Approvers // secret approvers win over project
			}
		}
	}

	return base, customApprovers, nil
}
