# Phase 01: DB Migrations + Policy Resolver

**Priority:** P0 — foundation for all policy customization
**Status:** COMPLETED — 2026-03-19

## Context
- `server/internal/policy/engine.go` — `Policy` struct, `MergePolicy`, `ForCredentialType`
- `server/internal/workflow/service.go:54` — `p := policy.ForCredentialType(input.CredentialType)` — this is what we replace
- `MergePolicy` already exists but only merges duration/limit caps (restrictive direction only)

## Architecture

```
CustomPolicy (stored as JSONB):
  approvers          []string   // user IDs — overrides tier defaults
  max_duration_min   int        // cap (only lower than tier max)
  auto_approve       *bool      // explicit override (nil = use tier default)
  block              bool       // never auto-approve even Tier 1
  require_reason     *bool
  min_reason_length  int
  business_hours     string     // "09:00-18:00 Mon-Fri Asia/Bangkok" or ""
  escalate_after_min int        // 0 = no escalation

Resolution order (most specific wins):
  secret.policy_config → project.policy_config → ForCredentialType(credentialType)
```

## Files

| File | Change |
|------|--------|
| `server/internal/database/migrations/000029_policy_config_columns.up.sql` | Create |
| `server/internal/database/migrations/000029_policy_config_columns.down.sql` | Create |
| `server/internal/policy/custom_policy.go` | Create |
| `server/internal/policy/resolver.go` | Create |
| `server/internal/workflow/service.go` | Modify — use Resolver |

---

## Task 1: DB migration

**Files:**
- Create: `server/internal/database/migrations/000029_policy_config_columns.up.sql`
- Create: `server/internal/database/migrations/000029_policy_config_columns.down.sql`

- [ ] Write up migration:

```sql
ALTER TABLE secrets  ADD COLUMN IF NOT EXISTS policy_config JSONB;
ALTER TABLE projects ADD COLUMN IF NOT EXISTS policy_config JSONB;
```

- [ ] Write down migration:

```sql
ALTER TABLE secrets  DROP COLUMN IF EXISTS policy_config;
ALTER TABLE projects DROP COLUMN IF EXISTS policy_config;
```

- [ ] Run: `make migrate-up` — verify no errors

- [ ] Commit:
```bash
git add server/internal/database/migrations/000029_*
git commit -m "feat(db): add policy_config JSONB columns to secrets and projects"
```

---

## Task 2: custom_policy.go — CustomPolicy struct

**Files:**
- Create: `server/internal/policy/custom_policy.go`
- Test: `server/internal/policy/custom_policy_test.go`

- [ ] Write struct + validation:

```go
package policy

import (
	"fmt"
	"regexp"
)

// CustomPolicy holds user-defined overrides. Zero values mean "use default".
type CustomPolicy struct {
	// Approvers: explicit list of user IDs required to approve. Empty = tier default behavior.
	Approvers []string `json:"approvers,omitempty"`

	// MaxDurationMinutes: cap on session duration. 0 = use tier default.
	// Policy enforces min(custom, tier_default).
	MaxDurationMinutes int `json:"max_duration_minutes,omitempty"`

	// AutoApprove: nil = use tier default. true/false = explicit override.
	AutoApprove *bool `json:"auto_approve,omitempty"`

	// Block: if true, always require approval regardless of tier (overrides AutoApprove).
	Block bool `json:"block,omitempty"`

	// RequireReason: nil = use tier default.
	RequireReason *bool `json:"require_reason,omitempty"`

	// MinReasonLength: 0 = use tier default.
	MinReasonLength int `json:"min_reason_length,omitempty"`

	// BusinessHours: "09:00-18:00 Mon-Fri Asia/Bangkok" or "". Empty = no restriction.
	BusinessHours string `json:"business_hours,omitempty"`

	// EscalateAfterMinutes: 0 = no escalation.
	EscalateAfterMinutes int `json:"escalate_after_minutes,omitempty"`

	// EscalateToUserID: who to notify if escalation triggers.
	EscalateToUserID string `json:"escalate_to_user_id,omitempty"`
}

var businessHoursPattern = regexp.MustCompile(`^\d{2}:\d{2}-\d{2}:\d{2} \w{3}-\w{3} \S+$`)

// Validate returns an error if the custom policy has invalid values.
func (c *CustomPolicy) Validate() error {
	if c.MaxDurationMinutes < 0 {
		return fmt.Errorf("max_duration_minutes cannot be negative")
	}
	if c.MinReasonLength < 0 {
		return fmt.Errorf("min_reason_length cannot be negative")
	}
	if c.EscalateAfterMinutes < 0 {
		return fmt.Errorf("escalate_after_minutes cannot be negative")
	}
	if c.BusinessHours != "" && !businessHoursPattern.MatchString(c.BusinessHours) {
		return fmt.Errorf("business_hours must be format '09:00-18:00 Mon-Fri Asia/Bangkok'")
	}
	if c.EscalateAfterMinutes > 0 && c.EscalateToUserID == "" {
		return fmt.Errorf("escalate_to_user_id required when escalate_after_minutes is set")
	}
	return nil
}

// ApplyTo merges this CustomPolicy onto a base Policy (tier defaults).
// CustomPolicy wins on any field it explicitly sets.
func (c *CustomPolicy) ApplyTo(base Policy) Policy {
	merged := base

	if c.MaxDurationMinutes > 0 && c.MaxDurationMinutes < base.MaxDurationMinutes {
		merged.MaxDurationMinutes = c.MaxDurationMinutes
	}
	if c.MinReasonLength > base.MinReasonLength {
		merged.MinReasonLength = c.MinReasonLength
	}
	if c.RequireReason != nil {
		merged.RequireReason = *c.RequireReason
	}
	if c.AutoApprove != nil {
		merged.AllowAutoApprove = *c.AutoApprove
		if !*c.AutoApprove {
			merged.RequireApproval = true
		}
	}
	if c.Block {
		merged.AllowAutoApprove = false
		merged.RequireApproval = true
	}

	return merged
}
```

- [ ] Write tests for `Validate` and `ApplyTo`:

```go
func TestCustomPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		policy  CustomPolicy
		wantErr bool
	}{
		{"valid empty", CustomPolicy{}, false},
		{"negative duration", CustomPolicy{MaxDurationMinutes: -1}, true},
		{"valid business hours", CustomPolicy{BusinessHours: "09:00-18:00 Mon-Fri Asia/Bangkok"}, false},
		{"invalid business hours", CustomPolicy{BusinessHours: "bad"}, true},
		{"escalate without user", CustomPolicy{EscalateAfterMinutes: 30}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomPolicyApplyTo(t *testing.T) {
	base := DefaultPolicyForTier(TierLow) // AllowAutoApprove=true, MaxDurationMinutes=480
	block := CustomPolicy{Block: true}
	result := block.ApplyTo(base)
	if result.AllowAutoApprove {
		t.Error("Block should disable AllowAutoApprove")
	}
	if !result.RequireApproval {
		t.Error("Block should enable RequireApproval")
	}
}
```

- [ ] Run: `cd server && go test ./internal/policy/... -v`

- [ ] Commit:
```bash
git add server/internal/policy/custom_policy.go server/internal/policy/custom_policy_test.go
git commit -m "feat(policy): add CustomPolicy struct with validation and merge logic"
```

---

## Task 3: resolver.go — DB-aware policy resolution

**Files:**
- Create: `server/internal/policy/resolver.go`
- Test: `server/internal/policy/resolver_test.go`

- [ ] Write `Resolver`:

```go
package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Resolver resolves the effective Policy for a secret, merging custom overrides.
type Resolver struct {
	pool *pgxpool.Pool
}

func NewResolver(pool *pgxpool.Pool) *Resolver {
	return &Resolver{pool: pool}
}

// Resolve returns the effective Policy for a secret.
// Resolution order: secret custom → project custom → tier defaults.
func (r *Resolver) Resolve(ctx context.Context, secretID, credentialType string) (Policy, []string, error) {
	base := ForCredentialType(credentialType)

	// Fetch secret policy_config + project_id
	var secretPolicyCfg []byte
	var projectID *string
	err := r.pool.QueryRow(ctx,
		`SELECT policy_config, project_id FROM secrets WHERE id = $1 AND deleted_at IS NULL`,
		secretID,
	).Scan(&secretPolicyCfg, &projectID)
	if err != nil {
		return base, nil, fmt.Errorf("fetching secret policy: %w", err)
	}

	// Apply project policy first (lower priority)
	var customApprovers []string
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
```

- [ ] Update `workflow/service.go` to use Resolver:

```go
// Add Resolver field to Service
type Service struct {
	pool     *pgxpool.Pool
	resolver *policy.Resolver
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, resolver: policy.NewResolver(pool)}
}

// In CreateRequest — replace:
//   p := policy.ForCredentialType(input.CredentialType)
// With:
	p, customApprovers, err := s.resolver.Resolve(ctx, input.SecretID, input.CredentialType)
	if err != nil {
		return nil, fmt.Errorf("resolving policy: %w", err)
	}
	_ = customApprovers // used in Phase 02 for approval_steps routing
```

- [ ] Run: `cd server && go build ./...`

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/policy/resolver.go server/internal/workflow/service.go
git commit -m "feat(policy): add DB-aware policy Resolver; wire into workflow service"
```

---

## Success Criteria
- `secrets.policy_config` and `projects.policy_config` columns exist
- `CustomPolicy.Validate()` rejects invalid configs
- `Resolver.Resolve()` correctly applies secret → project → tier defaults
- All existing tests still pass (resolver falls back to tier defaults when no custom config)
