# Phase 05 — Telemetry, QA, rollout guardrails

Status: pending  
Effort: 2h  
Priority: P2

## Objective
Validate UX improvement safely and detect regressions early.

## Instrumentation events
- `secret_create_opened`
- `secret_create_submitted`
- `secret_create_succeeded`
- `secret_policy_modal_opened` (source: post-create | row-action)
- `secret_policy_saved`
- `secret_create_abandoned`

## Metrics
- Median create completion time.
- Create abandonment rate.
- % secrets still unbound after 1h.
- Policy modal open-after-create conversion.

## QA plan
- Manual QA matrix:
  - create secret then apply policy
  - create secret skip policy then apply from row action
  - edit secret + edit policy independent
  - no project selected state
  - template version switch + override warnings
- Regression tests for API client integration paths.

## Rollout
- Feature flag: `NEXT_PUBLIC_SPLIT_SECRET_POLICY_FLOW`.
- Stage rollout internal first.
- Add short in-app hint for changed flow.

## Success criteria
- No critical regressions.
- Baseline metrics improve after rollout window.
