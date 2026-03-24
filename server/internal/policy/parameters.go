package policy

import "fmt"

const (
	keyMaxDurationMinutes = "max_duration_minutes"
	keyRequireApproval    = "require_approval"
	keyAllowAutoApprove   = "allow_auto_approve"
	keyRequireReason      = "require_reason"
	keyMinReasonLength    = "min_reason_length"
	keyMaxRequestsPerDay  = "max_requests_per_day"
	keyCoolDownMinutes    = "cool_down_minutes"
	keySingleUse          = "single_use"
	keyNotifyOnAccess     = "notify_on_access"
	keyRequireConsent     = "require_consent"
)

var allowedParameterKeys = map[string]struct{}{
	keyMaxDurationMinutes: {},
	keyRequireApproval:    {},
	keyAllowAutoApprove:   {},
	keyRequireReason:      {},
	keyMinReasonLength:    {},
	keyMaxRequestsPerDay:  {},
	keyCoolDownMinutes:    {},
	keySingleUse:          {},
	keyNotifyOnAccess:     {},
	keyRequireConsent:     {},
}

// PolicyParameters is the canonical parameter-only policy schema for v1.
type PolicyParameters struct {
	MaxDurationMinutes int  `json:"max_duration_minutes"`
	RequireApproval    bool `json:"require_approval"`
	AllowAutoApprove   bool `json:"allow_auto_approve"`
	RequireReason      bool `json:"require_reason"`
	MinReasonLength    int  `json:"min_reason_length"`
	MaxRequestsPerDay  int  `json:"max_requests_per_day"`
	CoolDownMinutes    int  `json:"cool_down_minutes"`
	SingleUse          bool `json:"single_use"`
	NotifyOnAccess     bool `json:"notify_on_access"`
	RequireConsent     bool `json:"require_consent"`
}

func DefaultParametersForCredentialType(credentialType string) PolicyParameters {
	return ParametersFromPolicy(ForCredentialType(credentialType))
}

func ParametersFromPolicy(p Policy) PolicyParameters {
	return PolicyParameters{
		MaxDurationMinutes: p.MaxDurationMinutes,
		RequireApproval:    p.RequireApproval,
		AllowAutoApprove:   p.AllowAutoApprove,
		RequireReason:      p.RequireReason,
		MinReasonLength:    p.MinReasonLength,
		MaxRequestsPerDay:  p.MaxRequestsPerDay,
		CoolDownMinutes:    p.CoolDownMinutes,
		SingleUse:          p.SingleUse,
		NotifyOnAccess:     p.NotifyOnAccess,
		RequireConsent:     p.RequireConsent,
	}
}

func (p PolicyParameters) ToPolicy(tier RiskTier) Policy {
	return Policy{
		RiskTier:           tier,
		MaxDurationMinutes: p.MaxDurationMinutes,
		RequireApproval:    p.RequireApproval,
		AllowAutoApprove:   p.AllowAutoApprove,
		RequireReason:      p.RequireReason,
		MinReasonLength:    p.MinReasonLength,
		MaxRequestsPerDay:  p.MaxRequestsPerDay,
		CoolDownMinutes:    p.CoolDownMinutes,
		SingleUse:          p.SingleUse,
		NotifyOnAccess:     p.NotifyOnAccess,
		RequireConsent:     p.RequireConsent,
	}
}

func ensureAllowedKeys(input map[string]any) error {
	for key := range input {
		if _, ok := allowedParameterKeys[key]; !ok {
			return fmt.Errorf("unknown policy key: %s", key)
		}
	}
	return nil
}
