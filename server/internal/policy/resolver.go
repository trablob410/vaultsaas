package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PolicySourceDefault          = "default"
	PolicySourceTemplate         = "template"
	PolicySourceTemplateOverride = "template+override"
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
// Resolution order (most specific wins): secret.policy_config -> project.policy_config -> tier defaults.
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

// ResolveEffectivePolicy resolves policy parameters from a template with optional overrides.
func ResolveEffectivePolicy(template PolicyParameters, override map[string]any) (PolicyParameters, []string, error) {
	source := PolicySourceTemplate
	if override == nil {
		override = map[string]any{}
	}
	if len(override) > 0 {
		source = PolicySourceTemplateOverride
	}
	if err := ensureAllowedKeys(override); err != nil {
		ObserveResolution(source, "validation_failed", nil)
		return PolicyParameters{}, nil, err
	}

	merged, err := mergeOverride(template, override)
	if err != nil {
		ObserveResolution(source, "validation_failed", nil)
		return PolicyParameters{}, nil, err
	}
	warnings := CompareWeakening(template, merged)
	ObserveResolution(source, "ok", warnings)
	return merged, warnings, nil
}

// CompareWeakening detects policy parameters that are weaker than the template.
func CompareWeakening(template, effective PolicyParameters) []string {
	warnings := make([]string, 0, 10)
	if effective.MaxDurationMinutes > template.MaxDurationMinutes {
		warnings = append(warnings, "weaker:max_duration_minutes")
	}
	if template.RequireApproval && !effective.RequireApproval {
		warnings = append(warnings, "weaker:require_approval")
	}
	if !template.AllowAutoApprove && effective.AllowAutoApprove {
		warnings = append(warnings, "weaker:allow_auto_approve")
	}
	if template.RequireReason && !effective.RequireReason {
		warnings = append(warnings, "weaker:require_reason")
	}
	if effective.MinReasonLength < template.MinReasonLength {
		warnings = append(warnings, "weaker:min_reason_length")
	}
	if effective.MaxRequestsPerDay > template.MaxRequestsPerDay {
		warnings = append(warnings, "weaker:max_requests_per_day")
	}
	if effective.CoolDownMinutes < template.CoolDownMinutes {
		warnings = append(warnings, "weaker:cool_down_minutes")
	}
	if template.SingleUse && !effective.SingleUse {
		warnings = append(warnings, "weaker:single_use")
	}
	if template.NotifyOnAccess && !effective.NotifyOnAccess {
		warnings = append(warnings, "weaker:notify_on_access")
	}
	if template.RequireConsent && !effective.RequireConsent {
		warnings = append(warnings, "weaker:require_consent")
	}
	return warnings
}

func mergeOverride(template PolicyParameters, override map[string]any) (PolicyParameters, error) {
	m := map[string]any{
		keyMaxDurationMinutes: template.MaxDurationMinutes,
		keyRequireApproval:    template.RequireApproval,
		keyAllowAutoApprove:   template.AllowAutoApprove,
		keyRequireReason:      template.RequireReason,
		keyMinReasonLength:    template.MinReasonLength,
		keyMaxRequestsPerDay:  template.MaxRequestsPerDay,
		keyCoolDownMinutes:    template.CoolDownMinutes,
		keySingleUse:          template.SingleUse,
		keyNotifyOnAccess:     template.NotifyOnAccess,
		keyRequireConsent:     template.RequireConsent,
	}
	for key, value := range override {
		if _, ok := m[key]; !ok {
			return PolicyParameters{}, fmt.Errorf("unknown policy key: %s", key)
		}
		m[key] = value
	}
	return ValidateParameters(m)
}
