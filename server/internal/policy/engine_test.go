package policy

import "testing"

func TestDeriveRiskTier(t *testing.T) {
	cases := []struct {
		ctype    string
		expected RiskTier
	}{
		{"api_key", TierLow},
		{"db_credential", TierMedium},
		{"ssh_key", TierMedium},
		{"oauth_token", TierMedium},
		{"cloud_credential", TierHigh},
		{"personal_session", TierHigh},
		{"unknown_type", TierMedium}, // default
	}
	for _, c := range cases {
		got := DeriveRiskTier(c.ctype)
		if got != c.expected {
			t.Errorf("DeriveRiskTier(%q)=%d, want %d", c.ctype, got, c.expected)
		}
	}
}

func TestDefaultPolicyForTier(t *testing.T) {
	p1 := DefaultPolicyForTier(TierLow)
	if p1.RequireApproval {
		t.Error("TierLow should not require approval")
	}
	if !p1.AllowAutoApprove {
		t.Error("TierLow should allow auto-approve")
	}
	if p1.MaxDurationMinutes != 480 {
		t.Errorf("TierLow MaxDuration=%d, want 480", p1.MaxDurationMinutes)
	}
	if p1.RiskTier != TierLow {
		t.Errorf("TierLow RiskTier=%d", p1.RiskTier)
	}

	p2 := DefaultPolicyForTier(TierMedium)
	if !p2.RequireApproval {
		t.Error("TierMedium should require approval")
	}
	if !p2.RequireReason {
		t.Error("TierMedium should require reason")
	}
	if p2.AllowAutoApprove {
		t.Error("TierMedium should not allow auto-approve")
	}
	if p2.MaxDurationMinutes != 60 {
		t.Errorf("TierMedium MaxDuration=%d, want 60", p2.MaxDurationMinutes)
	}

	p3 := DefaultPolicyForTier(TierHigh)
	if !p3.SingleUse {
		t.Error("TierHigh should be single-use")
	}
	if !p3.NotifyOnAccess {
		t.Error("TierHigh should notify on access")
	}
	if !p3.RequireConsent {
		t.Error("TierHigh should require consent")
	}
	if p3.MaxDurationMinutes != 15 {
		t.Errorf("TierHigh MaxDuration=%d, want 15", p3.MaxDurationMinutes)
	}

	p4 := DefaultPolicyForTier(TierCritical)
	if p4.RiskTier != TierCritical {
		t.Errorf("TierCritical RiskTier=%d", p4.RiskTier)
	}
	if p4.MaxDurationMinutes != 5 {
		t.Errorf("TierCritical MaxDuration=%d, want 5", p4.MaxDurationMinutes)
	}
}

func TestForCredentialType(t *testing.T) {
	p := ForCredentialType("api_key")
	if p.RiskTier != TierLow {
		t.Errorf("api_key tier=%d, want TierLow", p.RiskTier)
	}

	p = ForCredentialType("db_credential")
	if p.RiskTier != TierMedium {
		t.Errorf("db_credential tier=%d, want TierMedium", p.RiskTier)
	}

	p = ForCredentialType("cloud_credential")
	if p.RiskTier != TierHigh {
		t.Errorf("cloud_credential tier=%d, want TierHigh", p.RiskTier)
	}

	// unknown falls back to medium
	p = ForCredentialType("unknown")
	if p.RiskTier != TierMedium {
		t.Errorf("unknown tier=%d, want TierMedium", p.RiskTier)
	}
}

func TestMergePolicy(t *testing.T) {
	defaults := DefaultPolicyForTier(TierMedium) // MaxDuration=60, MaxReqPerDay=20, CoolDown=5

	// Tighten duration
	user := Policy{MaxDurationMinutes: 30}
	merged := MergePolicy(user, defaults)
	if merged.MaxDurationMinutes != 30 {
		t.Errorf("merged MaxDuration=%d, want 30", merged.MaxDurationMinutes)
	}

	// Cannot loosen duration (user sets higher than default)
	user2 := Policy{MaxDurationMinutes: 120}
	merged2 := MergePolicy(user2, defaults)
	if merged2.MaxDurationMinutes != 60 {
		t.Errorf("loosen should not work: MaxDuration=%d, want 60", merged2.MaxDurationMinutes)
	}

	// Tighten MaxRequestsPerDay
	user3 := Policy{MaxRequestsPerDay: 5}
	merged3 := MergePolicy(user3, defaults)
	if merged3.MaxRequestsPerDay != 5 {
		t.Errorf("merged MaxReqPerDay=%d, want 5", merged3.MaxRequestsPerDay)
	}

	// Zero values fall back to defaults
	userZero := Policy{}
	mergedZero := MergePolicy(userZero, defaults)
	if mergedZero.MaxDurationMinutes != defaults.MaxDurationMinutes {
		t.Errorf("zero MaxDuration=%d, want %d", mergedZero.MaxDurationMinutes, defaults.MaxDurationMinutes)
	}
	if mergedZero.MaxRequestsPerDay != defaults.MaxRequestsPerDay {
		t.Errorf("zero MaxRequestsPerDay=%d, want %d", mergedZero.MaxRequestsPerDay, defaults.MaxRequestsPerDay)
	}

	// Tighten CoolDown (larger is stricter for cool-down)
	user4 := Policy{CoolDownMinutes: 10}
	merged4 := MergePolicy(user4, defaults)
	if merged4.CoolDownMinutes != 10 {
		t.Errorf("merged CoolDown=%d, want 10", merged4.CoolDownMinutes)
	}
}
