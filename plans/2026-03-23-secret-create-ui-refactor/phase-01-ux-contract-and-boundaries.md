# Phase 01 — UX contract + component boundaries

Status: pending  
Effort: 3h  
Priority: P2

## Objective
Define new UX contract and strict component responsibilities before refactor.

## Requirements
- Separate creation concerns from policy concerns.
- Reuse existing API methods.
- Keep behavior parity for edit flows.

## Component boundary decisions
- `SecretForm`:
  - Owns only secret fields and submit for create/update secret record.
  - Emits `onCreated(secretId)` when create succeeds.
- `SecretPolicyBindingDialog` (new):
  - Owns template/version/override state and save binding.
  - Accepts `secretId`, `open`, `onClose`, `onSaved`.
- `useSecretPolicyBinding`:
  - Decouple from secret-form open state.
  - Consume `secretId` + `open` as explicit inputs.

## Files to modify/create
- Create: `dashboard/src/components/secrets/secret-policy-binding-dialog.tsx`
- Update: `dashboard/src/components/secrets/use-secret-policy-binding.ts`
- Update: `dashboard/src/components/secrets/secret-form.tsx`
- Update: `dashboard/src/components/secrets/secret-list.tsx`

## Steps
1. Define prop contracts (typescript interfaces) for new dialog and callbacks.
2. Map old in-form policy interactions to new explicit entry points.
3. Define transition states (create->policy) and cancellation states.

## Success criteria
- Contracts agreed and stable; no circular state ownership.
- No duplicate policy state between components.

## Risks
- Hidden reliance on `secret` object in existing hook.

## Mitigation
- Pass `secretId` as source of truth, fetch binding within policy dialog.
