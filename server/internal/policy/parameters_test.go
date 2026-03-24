package policy

import "testing"

func TestParametersRoundTrip(t *testing.T) {
	p := DefaultPolicyForTier(TierHigh)
	params := ParametersFromPolicy(p)
	policyRoundTrip := params.ToPolicy(TierHigh)

	if policyRoundTrip.MaxDurationMinutes != p.MaxDurationMinutes {
		t.Fatalf("max_duration mismatch: got %d want %d", policyRoundTrip.MaxDurationMinutes, p.MaxDurationMinutes)
	}
	if policyRoundTrip.RequireApproval != p.RequireApproval {
		t.Fatalf("require_approval mismatch")
	}
	if policyRoundTrip.AllowAutoApprove != p.AllowAutoApprove {
		t.Fatalf("allow_auto_approve mismatch")
	}
	if policyRoundTrip.RequireReason != p.RequireReason {
		t.Fatalf("require_reason mismatch")
	}
	if policyRoundTrip.MinReasonLength != p.MinReasonLength {
		t.Fatalf("min_reason_length mismatch")
	}
	if policyRoundTrip.MaxRequestsPerDay != p.MaxRequestsPerDay {
		t.Fatalf("max_requests_per_day mismatch")
	}
	if policyRoundTrip.CoolDownMinutes != p.CoolDownMinutes {
		t.Fatalf("cool_down_minutes mismatch")
	}
	if policyRoundTrip.SingleUse != p.SingleUse {
		t.Fatalf("single_use mismatch")
	}
	if policyRoundTrip.NotifyOnAccess != p.NotifyOnAccess {
		t.Fatalf("notify_on_access mismatch")
	}
	if policyRoundTrip.RequireConsent != p.RequireConsent {
		t.Fatalf("require_consent mismatch")
	}
}

func TestDefaultParametersForCredentialType(t *testing.T) {
	params := DefaultParametersForCredentialType("db_credential")
	if !params.RequireApproval {
		t.Fatalf("db_credential should require approval")
	}
	if params.MaxDurationMinutes != 60 {
		t.Fatalf("max_duration_minutes mismatch: got %d want 60", params.MaxDurationMinutes)
	}
}
