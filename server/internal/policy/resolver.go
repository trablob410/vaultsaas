package policy

import "fmt"

const (
	PolicySourceDefault          = "default"
	PolicySourceTemplate         = "template"
	PolicySourceTemplateOverride = "template+override"
)

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
