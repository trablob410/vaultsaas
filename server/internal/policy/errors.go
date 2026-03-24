package policy

import "errors"

var (
	ErrPolicyInvalid   = errors.New("policy_invalid")
	ErrPolicyForbidden = errors.New("policy_forbidden")
	ErrPolicyNotFound  = errors.New("policy_not_found")
	ErrPolicyConflict  = errors.New("policy_conflict")
)
