# Phase 03 — Dedicated policy dialog extraction

Status: completed  
Effort: 4h  
Priority: P2

## Objective
Move all policy template binding and override UX into dedicated modal.

## Requirements
- New dialog includes:
  - Template selector
  - Version selector
  - Override toggle + parameter fields
  - Weaker-setting warnings
  - Effective policy preview
- Save action updates binding via existing API.

## Architecture
- UI composition:
  - `SecretPolicyBindingDialog`
    - wraps existing content from `secret-policy-binding-section.tsx`
  - keep `secret-policy-binding-section.tsx` as internal presentational block or fold into dialog if simpler.
- Hook:
  - adapt `useSecretPolicyBinding(open, secret?)` to `useSecretPolicyBinding({open, secretId})`
  - load binding lazily when dialog opens.

## Steps
1. [x] Create dialog shell with header/footer and save/cancel handlers.
2. [x] Port section content into dialog body.
3. [x] Add save handler: `api.policies.updateBinding(secretId, payload)`.
4. [x] Add loading/error states scoped to dialog.
5. [x] Ensure warning badges and preview remain accurate.

## Implementation notes
- UI was refined in `secret-policy-binding-dialog.tsx` and `secret-policy-binding-section.tsx` for dedicated modal UX, edge-case messaging, and loading/empty/error presentation.
- Backend wiring kept in existing path via `api.policies.updateBinding(secretId, payload)`.
- Hook updated to support lazy loading states and guarded version handling (`template_version >= 1`).

## Edge cases
- No project selected.
- Secret exists but no binding yet.
- Template deleted while modal open.

## Success criteria
- Policy operations possible without opening secret form.
- Existing policy behavior parity maintained.
