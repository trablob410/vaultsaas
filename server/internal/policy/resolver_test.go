package policy

import "testing"

func TestResolveEffectivePolicyMergeAndWarnings(t *testing.T) {
	template := PolicyParameters{
		MaxDurationMinutes: 60,
		RequireApproval:    true,
		AllowAutoApprove:   false,
		RequireReason:      true,
		MinReasonLength:    20,
		MaxRequestsPerDay:  20,
		CoolDownMinutes:    5,
		SingleUse:          true,
		NotifyOnAccess:     true,
		RequireConsent:     true,
	}
	override := map[string]any{
		keyMaxDurationMinutes: 120,
		keyRequireApproval:    false,
		keyRequireReason:      false,
		keyMinReasonLength:    0,
		keyMaxRequestsPerDay:  40,
		keyCoolDownMinutes:    1,
		keySingleUse:          false,
		keyNotifyOnAccess:     false,
		keyRequireConsent:     false,
	}

	effective, warnings, err := ResolveEffectivePolicy(template, override)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if effective.MaxDurationMinutes != 120 {
		t.Fatalf("expected override max_duration_minutes")
	}
	if len(warnings) == 0 {
		t.Fatalf("expected weakening warnings")
	}
}

func TestResolveEffectivePolicyRejectUnknownKey(t *testing.T) {
	template := DefaultParametersForCredentialType("db_credential")
	_, _, err := ResolveEffectivePolicy(template, map[string]any{"bad_key": true})
	if err == nil {
		t.Fatalf("expected unknown key error")
	}
}

func TestResolveEffectivePolicyNoOverrideUsesTemplateSource(t *testing.T) {
	template := DefaultParametersForCredentialType("db_credential")
	effective, warnings, err := ResolveEffectivePolicy(template, nil)
	if err != nil {
		t.Fatalf("expected success with nil override: %v", err)
	}
	if effective != template {
		t.Fatalf("expected effective to equal template")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
	obs := LastResolutionObservation()
	if obs.Source != PolicySourceTemplate {
		t.Fatalf("expected source=%s got %s", PolicySourceTemplate, obs.Source)
	}
}

func TestCompareWeakeningDeterministic(t *testing.T) {
	template := PolicyParameters{
		MaxDurationMinutes: 60,
		RequireApproval:    true,
		AllowAutoApprove:   false,
		RequireReason:      true,
		MinReasonLength:    10,
		MaxRequestsPerDay:  20,
		CoolDownMinutes:    5,
		SingleUse:          true,
		NotifyOnAccess:     true,
		RequireConsent:     true,
	}
	effective := PolicyParameters{
		MaxDurationMinutes: 120,
		RequireApproval:    false,
		AllowAutoApprove:   true,
		RequireReason:      false,
		MinReasonLength:    0,
		MaxRequestsPerDay:  40,
		CoolDownMinutes:    0,
		SingleUse:          false,
		NotifyOnAccess:     false,
		RequireConsent:     false,
	}
	warnings := CompareWeakening(template, effective)
	if len(warnings) != 10 {
		t.Fatalf("expected 10 warnings, got %d", len(warnings))
	}
}
