package policy

import "testing"

func validInput() map[string]any {
	return map[string]any{
		keyMaxDurationMinutes: 60,
		keyRequireApproval:    true,
		keyAllowAutoApprove:   false,
		keyRequireReason:      true,
		keyMinReasonLength:    20,
		keyMaxRequestsPerDay:  20,
		keyCoolDownMinutes:    5,
		keySingleUse:          false,
		keyNotifyOnAccess:     true,
		keyRequireConsent:     false,
	}
}

func TestValidateParametersSuccess(t *testing.T) {
	params, err := ValidateParameters(validInput())
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if params.MaxDurationMinutes != 60 {
		t.Fatalf("max_duration_minutes mismatch")
	}
}

func TestValidateParametersRejectUnknownKey(t *testing.T) {
	input := validInput()
	input["unknown"] = 1
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected unknown key error")
	}
}

func TestValidateParametersRejectTypeMismatch(t *testing.T) {
	input := validInput()
	input[keyMaxDurationMinutes] = "60"
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected type mismatch error")
	}
}

func TestValidateParametersAcceptFloat64IntegerFromJSON(t *testing.T) {
	input := validInput()
	input[keyMaxDurationMinutes] = float64(60)
	input[keyMinReasonLength] = float64(20)
	input[keyMaxRequestsPerDay] = float64(20)
	input[keyCoolDownMinutes] = float64(5)
	params, err := ValidateParameters(input)
	if err != nil {
		t.Fatalf("expected success for integer-like float64: %v", err)
	}
	if params.MaxDurationMinutes != 60 {
		t.Fatalf("unexpected value: %d", params.MaxDurationMinutes)
	}
}

func TestValidateParametersRejectFloat64Fractional(t *testing.T) {
	input := validInput()
	input[keyMaxDurationMinutes] = float64(60.5)
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected fractional number reject")
	}
}

func TestValidateParametersBounds(t *testing.T) {
	input := validInput()
	input[keyMaxDurationMinutes] = 1441
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected bounds error")
	}
}

func TestValidateParametersCrossFieldAutoApproveConflict(t *testing.T) {
	input := validInput()
	input[keyAllowAutoApprove] = true
	input[keyRequireApproval] = true
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected auto-approve conflict error")
	}
}

func TestValidateParametersCrossFieldSingleUseDuration(t *testing.T) {
	input := validInput()
	input[keySingleUse] = true
	input[keyMaxDurationMinutes] = 300
	_, err := ValidateParameters(input)
	if err == nil {
		t.Fatalf("expected single_use max duration error")
	}
}

func TestValidateParametersRequireReasonFalseForcesMinReasonZero(t *testing.T) {
	input := validInput()
	input[keyRequireReason] = false
	input[keyMinReasonLength] = 120
	params, err := ValidateParameters(input)
	if err != nil {
		t.Fatalf("expected success when forcing min_reason_length: %v", err)
	}
	if params.MinReasonLength != 0 {
		t.Fatalf("expected min_reason_length=0, got %d", params.MinReasonLength)
	}
}
