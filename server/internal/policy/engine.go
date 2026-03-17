package policy

// RiskTier represents the credential risk level (1-4).
type RiskTier int

const (
	TierLow      RiskTier = 1 // api_key
	TierMedium   RiskTier = 2 // db_credential, ssh_key, oauth_token
	TierHigh     RiskTier = 3 // cloud_credential, personal_session
	TierCritical RiskTier = 4 // future: multi-approve + 2FA
)

// Policy defines access constraints for a credential type.
type Policy struct {
	RiskTier            RiskTier `json:"risk_tier"`
	MaxDurationMinutes  int      `json:"max_duration_minutes"`
	RequireApproval     bool     `json:"require_approval"`
	RequireReason       bool     `json:"require_reason"`
	MinReasonLength     int      `json:"min_reason_length"`
	SingleUse           bool     `json:"single_use"`
	NotifyOnAccess      bool     `json:"notify_on_access"`
	MaxRequestsPerDay   int      `json:"max_requests_per_day"`
	CoolDownMinutes     int      `json:"cool_down_minutes"`
	RequireConsent      bool     `json:"require_consent"`
	AllowAutoApprove    bool     `json:"allow_auto_approve"`
}

// tierMap maps credential_type to RiskTier.
var tierMap = map[string]RiskTier{
	"api_key":          TierLow,
	"db_credential":    TierMedium,
	"ssh_key":          TierMedium,
	"oauth_token":      TierMedium,
	"cloud_credential": TierHigh,
	"personal_session": TierHigh,
}

// DeriveRiskTier returns the risk tier for a credential type.
func DeriveRiskTier(credentialType string) RiskTier {
	if tier, ok := tierMap[credentialType]; ok {
		return tier
	}
	return TierMedium // default
}

// DefaultPolicyForTier returns the default policy for a given tier.
func DefaultPolicyForTier(tier RiskTier) Policy {
	switch tier {
	case TierLow:
		return Policy{
			RiskTier:           TierLow,
			MaxDurationMinutes: 480, // 8h
			RequireApproval:    false,
			RequireReason:      false,
			SingleUse:          false,
			NotifyOnAccess:     false,
			MaxRequestsPerDay:  50,
			CoolDownMinutes:    0,
			RequireConsent:     false,
			AllowAutoApprove:   true,
		}
	case TierMedium:
		return Policy{
			RiskTier:           TierMedium,
			MaxDurationMinutes: 60, // 1h
			RequireApproval:    true,
			RequireReason:      true,
			MinReasonLength:    10,
			SingleUse:          false,
			NotifyOnAccess:     false,
			MaxRequestsPerDay:  20,
			CoolDownMinutes:    5,
			RequireConsent:     false,
			AllowAutoApprove:   false,
		}
	case TierHigh:
		return Policy{
			RiskTier:           TierHigh,
			MaxDurationMinutes: 15,
			RequireApproval:    true,
			RequireReason:      true,
			MinReasonLength:    20,
			SingleUse:          true,
			NotifyOnAccess:     true,
			MaxRequestsPerDay:  10,
			CoolDownMinutes:    15,
			RequireConsent:     true,
			AllowAutoApprove:   false,
		}
	case TierCritical:
		// Stub: behaves like High for MVP
		p := DefaultPolicyForTier(TierHigh)
		p.RiskTier = TierCritical
		p.MaxDurationMinutes = 5
		p.MaxRequestsPerDay = 5
		p.CoolDownMinutes = 30
		return p
	default:
		return DefaultPolicyForTier(TierMedium)
	}
}

// ForCredentialType returns the default policy for a credential type.
func ForCredentialType(credentialType string) Policy {
	return DefaultPolicyForTier(DeriveRiskTier(credentialType))
}

// MergePolicy merges a user-provided policy override with defaults.
// User values take precedence where set; zero values fall back to defaults.
func MergePolicy(userPolicy, defaults Policy) Policy {
	merged := defaults
	if userPolicy.MaxDurationMinutes > 0 && userPolicy.MaxDurationMinutes < defaults.MaxDurationMinutes {
		merged.MaxDurationMinutes = userPolicy.MaxDurationMinutes
	}
	if userPolicy.MaxRequestsPerDay > 0 && userPolicy.MaxRequestsPerDay < defaults.MaxRequestsPerDay {
		merged.MaxRequestsPerDay = userPolicy.MaxRequestsPerDay
	}
	if userPolicy.MinReasonLength > defaults.MinReasonLength {
		merged.MinReasonLength = userPolicy.MinReasonLength
	}
	if userPolicy.CoolDownMinutes > defaults.CoolDownMinutes {
		merged.CoolDownMinutes = userPolicy.CoolDownMinutes
	}
	return merged
}
