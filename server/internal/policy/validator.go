package policy

import (
	"fmt"
	"math"
)

const (
	maxInt = int(^uint(0) >> 1)
	minInt = -maxInt - 1
)

func ValidateParameters(input map[string]any) (PolicyParameters, error) {
	if input == nil {
		return validationFail("validate_parameters", "policy parameters are required")
	}
	if err := ensureAllowedKeys(input); err != nil {
		return validationErr("validate_parameters", err)
	}

	maxDuration, err := intField(input, keyMaxDurationMinutes)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	requireApproval, err := boolField(input, keyRequireApproval)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	allowAutoApprove, err := boolField(input, keyAllowAutoApprove)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	requireReason, err := boolField(input, keyRequireReason)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	minReasonLength, err := intField(input, keyMinReasonLength)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	maxRequestsPerDay, err := intField(input, keyMaxRequestsPerDay)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	coolDownMinutes, err := intField(input, keyCoolDownMinutes)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	singleUse, err := boolField(input, keySingleUse)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	notifyOnAccess, err := boolField(input, keyNotifyOnAccess)
	if err != nil {
		return validationErr("validate_parameters", err)
	}
	requireConsent, err := boolField(input, keyRequireConsent)
	if err != nil {
		return validationErr("validate_parameters", err)
	}

	if !requireReason {
		minReasonLength = 0
	}

	params := PolicyParameters{
		MaxDurationMinutes: maxDuration,
		RequireApproval:    requireApproval,
		AllowAutoApprove:   allowAutoApprove,
		RequireReason:      requireReason,
		MinReasonLength:    minReasonLength,
		MaxRequestsPerDay:  maxRequestsPerDay,
		CoolDownMinutes:    coolDownMinutes,
		SingleUse:          singleUse,
		NotifyOnAccess:     notifyOnAccess,
		RequireConsent:     requireConsent,
	}
	if err := validateParameterBounds(params); err != nil {
		return validationErr("validate_parameters", err)
	}
	if err := validateCrossFields(params); err != nil {
		return validationErr("validate_parameters", err)
	}
	return params, nil
}

func validationFail(stage, message string) (PolicyParameters, error) {
	ObserveValidationFailure(stage)
	return PolicyParameters{}, fmt.Errorf("%s", message)
}

func validationErr(stage string, err error) (PolicyParameters, error) {
	ObserveValidationFailure(stage)
	return PolicyParameters{}, err
}

func intField(input map[string]any, key string) (int, error) {
	v, ok := input[key]
	if !ok {
		return 0, fmt.Errorf("missing policy key: %s", key)
	}
	switch iv := v.(type) {
	case int:
		return iv, nil
	case int32:
		return int(iv), nil
	case int64:
		if iv < int64(minInt) || iv > int64(maxInt) {
			return 0, fmt.Errorf("invalid type for %s: int64 out of range", key)
		}
		return int(iv), nil
	case float64:
		if math.Trunc(iv) != iv {
			return 0, fmt.Errorf("invalid type for %s: expected integer number", key)
		}
		if iv < float64(minInt) || iv > float64(maxInt) {
			return 0, fmt.Errorf("invalid type for %s: number out of range", key)
		}
		return int(iv), nil
	default:
		return 0, fmt.Errorf("invalid type for %s: expected int", key)
	}
}

func boolField(input map[string]any, key string) (bool, error) {
	v, ok := input[key]
	if !ok {
		return false, fmt.Errorf("missing policy key: %s", key)
	}
	bv, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("invalid type for %s: expected bool", key)
	}
	return bv, nil
}

func validateParameterBounds(p PolicyParameters) error {
	if p.MaxDurationMinutes < 1 || p.MaxDurationMinutes > 1440 {
		return fmt.Errorf("max_duration_minutes out of range [1,1440]")
	}
	if p.MinReasonLength < 0 || p.MinReasonLength > 500 {
		return fmt.Errorf("min_reason_length out of range [0,500]")
	}
	if p.MaxRequestsPerDay < 1 || p.MaxRequestsPerDay > 1000 {
		return fmt.Errorf("max_requests_per_day out of range [1,1000]")
	}
	if p.CoolDownMinutes < 0 || p.CoolDownMinutes > 1440 {
		return fmt.Errorf("cool_down_minutes out of range [0,1440]")
	}
	return nil
}

func validateCrossFields(p PolicyParameters) error {
	if p.AllowAutoApprove && p.RequireApproval {
		return fmt.Errorf("allow_auto_approve cannot be true when require_approval is true")
	}
	if p.SingleUse && p.MaxDurationMinutes > 240 {
		return fmt.Errorf("single_use=true requires max_duration_minutes <= 240")
	}
	return nil
}
