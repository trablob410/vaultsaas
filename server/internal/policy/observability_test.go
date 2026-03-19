package policy

import "testing"

func TestObserveResolutionAndCounters(t *testing.T) {
	ObserveResolution(PolicySourceTemplateOverride, "ok", []string{"weaker:max_duration_minutes"})
	obs := LastResolutionObservation()
	if obs.Source != PolicySourceTemplateOverride {
		t.Fatalf("source mismatch")
	}
	if obs.WarningCount != 1 {
		t.Fatalf("warning_count mismatch")
	}

	ObserveValidationFailure("validate_parameters")
	c := SnapshotCounters()
	if c["policy_validation_fail_total|stage=validate_parameters"] < 1 {
		t.Fatalf("expected validation fail counter")
	}
}
