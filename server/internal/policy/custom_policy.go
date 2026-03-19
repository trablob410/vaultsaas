package policy

import (
	"fmt"
	"regexp"
)

// CustomPolicy holds user-defined overrides on top of tier defaults.
// Zero values mean "use tier default".
type CustomPolicy struct {
	// Approvers: explicit user IDs required to approve. Empty = tier default behavior.
	Approvers []string `json:"approvers,omitempty"`

	// MaxDurationMinutes: cap on session duration. 0 = use tier default.
	// Enforces min(custom, tier_default).
	MaxDurationMinutes int `json:"max_duration_minutes,omitempty"`

	// AutoApprove: nil = use tier default. Explicit true/false overrides tier.
	AutoApprove *bool `json:"auto_approve,omitempty"`

	// Block: always require approval regardless of tier (overrides AutoApprove).
	Block bool `json:"block,omitempty"`

	// RequireReason: nil = use tier default.
	RequireReason *bool `json:"require_reason,omitempty"`

	// MinReasonLength: 0 = use tier default.
	MinReasonLength int `json:"min_reason_length,omitempty"`

	// BusinessHours: "09:00-18:00 Mon-Fri Asia/Bangkok" or "". Empty = no restriction.
	BusinessHours string `json:"business_hours,omitempty"`

	// EscalateAfterMinutes: 0 = no escalation.
	EscalateAfterMinutes int `json:"escalate_after_minutes,omitempty"`

	// EscalateToUserID: who to notify on escalation. Required when EscalateAfterMinutes > 0.
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
		return fmt.Errorf("business_hours must match format '09:00-18:00 Mon-Fri Asia/Bangkok'")
	}
	if c.EscalateAfterMinutes > 0 && c.EscalateToUserID == "" {
		return fmt.Errorf("escalate_to_user_id required when escalate_after_minutes is set")
	}
	return nil
}

// ApplyTo merges this CustomPolicy onto a base Policy, returning the effective policy.
// Custom values win over base values when explicitly set.
func (c *CustomPolicy) ApplyTo(base Policy) Policy {
	merged := base

	// Cap duration (only lower, never raise)
	if c.MaxDurationMinutes > 0 && c.MaxDurationMinutes < base.MaxDurationMinutes {
		merged.MaxDurationMinutes = c.MaxDurationMinutes
	}
	// Raise min reason length
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
	// Block overrides AutoApprove
	if c.Block {
		merged.AllowAutoApprove = false
		merged.RequireApproval = true
	}

	return merged
}
