package policy

import "testing"

func TestCustomPolicyValidate(t *testing.T) {
	tests := []struct {
		name    string
		policy  CustomPolicy
		wantErr bool
	}{
		{"valid empty", CustomPolicy{}, false},
		{"negative duration", CustomPolicy{MaxDurationMinutes: -1}, true},
		{"negative reason length", CustomPolicy{MinReasonLength: -1}, true},
		{"valid business hours", CustomPolicy{BusinessHours: "09:00-18:00 Mon-Fri Asia/Bangkok"}, false},
		{"invalid business hours", CustomPolicy{BusinessHours: "bad"}, true},
		{"escalate without user", CustomPolicy{EscalateAfterMinutes: 30}, true},
		{"escalate with user", CustomPolicy{EscalateAfterMinutes: 30, EscalateToUserID: "user-1"}, false},
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

	t.Run("block disables auto-approve", func(t *testing.T) {
		cp := CustomPolicy{Block: true}
		result := cp.ApplyTo(base)
		if result.AllowAutoApprove {
			t.Error("Block should disable AllowAutoApprove")
		}
		if !result.RequireApproval {
			t.Error("Block should enable RequireApproval")
		}
	})

	t.Run("max duration capped lower", func(t *testing.T) {
		cp := CustomPolicy{MaxDurationMinutes: 60}
		result := cp.ApplyTo(base)
		if result.MaxDurationMinutes != 60 {
			t.Errorf("expected 60, got %d", result.MaxDurationMinutes)
		}
	})

	t.Run("max duration not raised", func(t *testing.T) {
		cp := CustomPolicy{MaxDurationMinutes: 9999}
		result := cp.ApplyTo(base)
		if result.MaxDurationMinutes != base.MaxDurationMinutes {
			t.Errorf("expected %d, got %d", base.MaxDurationMinutes, result.MaxDurationMinutes)
		}
	})

	t.Run("auto_approve false forces require_approval", func(t *testing.T) {
		f := false
		cp := CustomPolicy{AutoApprove: &f}
		result := cp.ApplyTo(base)
		if result.AllowAutoApprove {
			t.Error("AutoApprove=false should disable AllowAutoApprove")
		}
		if !result.RequireApproval {
			t.Error("AutoApprove=false should enable RequireApproval")
		}
	})
}
