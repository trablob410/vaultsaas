# Phase 04 — List actions + unbound indicators

Status: completed  
Effort: 2h  
Priority: P2

## Objective
Expose explicit policy-management entry and prevent missed policy setup.

## Requirements
- Add `Manage Policies` row action in `secret-list.tsx`.
- Add unbound indicator badge/icon for secrets with no binding.
- Post-create handoff: prompt user to open policy modal.

## Steps
1. [x] Extend row actions UI (menu or inline button) with Manage Policies.
2. [x] Add local state for policy dialog open + selected secret.
3. [x] Hook create-success path to set selected secret and open policy modal or CTA.
4. [x] Add lightweight binding status field retrieval strategy:
   - short term: fetch binding per selected secret on demand (KISS)
   - optional improvement: backend list includes binding summary (out of current scope)

## Implementation notes
- `secret-list.tsx` now includes explicit `Manage Policies` row action wired to `SecretPolicyBindingDialog`.
- Post-create handoff CTA shown after successful create: `Apply policy now` opens the policy modal with selected secret.
- Unbound indicator implemented as `No policy` outline badge + warning icon and tooltip text: “Secret has no access policy bound.”
- Binding state is fetched via `api.policies.getBinding(secretId)` per listed secret and interpreted with conservative tri-state (`true|false|null`) to avoid false bound state on transient errors.

## UX details
- Indicator text: `No policy` (warning/outline badge).
- Tooltip explains impact: “Secret has no access policy bound.”

## Success criteria
- User can manage policy directly from list.
- Reduced chance of orphaned/unbound secret.
