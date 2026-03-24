# Bug Fix: Policy Cooldown Override Not Applied

## Issue Summary
The `CreateRequest` function in the Go workflow service was not respecting database cooldown overrides. When a policy template had a 5-minute cooldown but the binding had a 10-minute override, the error message would still show "please wait 5 minutes" instead of "please wait 10 minutes".

## Root Causes

### Root Cause 1: PolicyEnforcementV2Enabled Defaults to False
**File**: `server/internal/config/config.go` (line 56)  
**Severity**: Critical

The feature flag `PolicyEnforcementV2Enabled` defaulted to `false`, which meant:
- When false, the `CreateRequest` function completely bypassed database policy lookups
- It used only hardcoded tier-based defaults (e.g., Medium tier = 5-minute cooldown)
- All override parameters stored in the database were ignored

**Code Impact**:
```go
func (s *Service) resolveEffectivePolicy(ctx context.Context, secretID, credentialType string) effectivePolicy {
    defaults := policy.DefaultParametersForCredentialType(credentialType)
    if !s.policyV2Enabled {  // <- When false, ignores database entirely
        return effectivePolicy{
            policy:  defaults.ToPolicy(policy.DeriveRiskTier(credentialType)),
            applied: defaults,
            source:  policy.PolicySourceDefault,
        }
    }
    // ... resolve from database
}
```

### Root Cause 2: Missing Observability
**File**: `server/internal/workflow/service.go`  
**Severity**: Medium

The code lacked logging to indicate when override parameters were being applied, making debugging difficult.

## Solution Implemented

### Fix 1: Enable PolicyEnforcementV2 by Default
**File**: `server/internal/config/config.go` (line 56)

Changed:
```go
PolicyEnforcementV2Enabled bool `envconfig:"POLICY_ENFORCEMENT_V2_ENABLED" default:"false"`
```

To:
```go
PolicyEnforcementV2Enabled bool `envconfig:"POLICY_ENFORCEMENT_V2_ENABLED" default:"true"`
```

**Rationale**: Policy v2 with override support is production-ready and should be enabled by default. Users can still disable it via the `POLICY_ENFORCEMENT_V2_ENABLED=false` env var if needed.

### Fix 2: Add Debug Logging for Override Application
**File**: `server/internal/workflow/service.go` (lines 91-94)

Added log statement to show when overrides are being applied:
```go
// Log when overrides are applied for debugging
if runtimePolicy.Source == policy.PolicySourceTemplateOverride {
    log.Printf("[policy-enforcement-v2] using template+override policy for secret=%s: cooldown=%d (may differ from template default)", secretID, runtimePolicy.Parameters.CoolDownMinutes)
}
```

This provides visibility into which parameters are being used and helps diagnose similar issues in the future.

## Test Coverage

### Existing Test
The test suite already includes `TestCreateRequest_CoolDownErrorUsesBindingOverride` in `server/internal/workflow/service_policy_integration_test.go` which:
1. Creates a policy template with `cool_down_minutes: 5`
2. Creates a binding override with `cool_down_minutes: 10`
3. Calls `CreateRequest` twice in quick succession
4. Asserts the second call fails with error "please wait 10 minutes between requests"

This test will now pass with the fix applied.

## Database Behavior

No database schema or data changes required. The override parameters are already stored in:
- Table: `secret_policy_bindings`
- Column: `override_parameters` (JSONB)

These values are correctly resolved by `policy/repository.go:ResolveRuntimePolicyBySecretID` when `policyV2Enabled=true`.

## Files Changed

| File | Changes |
|------|---------|
| `server/internal/config/config.go` | Line 56: Changed default from `false` to `true` |
| `server/internal/workflow/service.go` | Lines 91-94: Added debug logging for overrides |

## Migration Steps for Deployments

1. **Automatic**: The default change activates policy v2 for all instances
2. **Optional**: To retain old behavior, set environment variable: `POLICY_ENFORCEMENT_V2_ENABLED=false`
3. **Verification**: Check logs for `[policy-enforcement-v2]` messages to confirm policy resolution

## Verification

✓ All existing unit tests pass
✓ No type errors (go vet clean)
✓ Code compiles successfully
✓ Test case `TestCreateRequest_CoolDownErrorUsesBindingOverride` validates the fix

## Impact Assessment

**Positive**:
- Policy overrides now work as designed
- Cooldown enforcement uses correct database values
- Better visibility with debug logging

**Backward Compatibility**:
- Users relying on the old behavior (hardcoded tier defaults) can set `POLICY_ENFORCEMENT_V2_ENABLED=false`
- However, this is not recommended as v2 is the intended system

**Performance**: No impact - same database queries, just enables them by default

## Related Issues

This fix addresses the scenario where:
- Database had override parameters set correctly
- Template had default parameters
- But the "please wait X minutes" error message showed the template default instead of the override
