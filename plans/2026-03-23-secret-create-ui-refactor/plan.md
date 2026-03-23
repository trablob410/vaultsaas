---
title: "Refactor Secret Create UI Flow"
description: "Split secret creation and policy application into clearer two-step UX with dedicated policy modal."
status: pending
priority: P2
effort: 15h
branch: feat/custom-policy
tags: [dashboard, ux, secrets, policy, shadcn]
created: 2026-03-23
---

# Refactor Secret Create UI Flow

## Goal
Reduce form clutter in secret create/edit UX. Separate policy controls into dedicated modal. Keep security posture. Keep existing API contract.

## Scope (MVP)
- `SecretForm` handles core secret fields only.
- New `SecretPolicyBindingDialog` handles template/version/override config.
- Add `Manage Policies` action from secrets list.
- Add unbound-policy indicator in secrets list.
- Add tracking events for UX outcome.

## Non-Goals
- No backend contract changes.
- No redesign of policy template management pages.
- No workflow engine changes.

## Current State Summary
- `secret-form.tsx` renders secret + policy blocks in one dialog.
- `secret-policy-binding-section.tsx` contains heavy controls/preview/warnings.
- `use-secret-policy-binding.ts` state tightly coupled to secret form open state.
- UX overloaded; long modal; weak progressive disclosure.

## Target State
- Two dialogs, two mental models:
  1) create/edit secret data
  2) manage policy binding
- Explicit flow handoff after create.
- Easier scan/read/complete.

## Phases
| # | Phase | Effort | Status | File |
|---|---|---:|---|---|
| 1 | UX contract + component boundaries | 3h | pending | [phase-01-ux-contract-and-boundaries](./phase-01-ux-contract-and-boundaries.md) |
| 2 | Refactor create/edit secret dialog | 4h | pending | [phase-02-secret-form-simplification](./phase-02-secret-form-simplification.md) |
| 3 | Build dedicated policy dialog | 4h | pending | [phase-03-policy-dialog-extraction](./phase-03-policy-dialog-extraction.md) |
| 4 | Integrate list actions + indicators | 2h | pending | [phase-04-list-actions-and-indicators](./phase-04-list-actions-and-indicators.md) |
| 5 | Telemetry, QA, rollout guardrails | 2h | pending | [phase-05-telemetry-qa-rollout](./phase-05-telemetry-qa-rollout.md) |

## Dependencies
- Existing API client methods in `src/lib/api-client.ts` for secrets + policy binding.
- shadcn/ui primitives already present.
- Local project context from `valt_current_project` behavior unchanged.

## Risks
- Users may skip policy step after create.
- State regressions when switching create/edit/policy quickly.
- Hidden coupling in current hook state.

## Bug fix updates
- 2026-03-23: Fixed secret-create project scoping regression.
  - Dashboard now forwards `project_id` from `valt_current_project` during create.
  - Vault create API now accepts/validates `project_id` and persists to `secrets.project_id`.

## Risk Mitigations
- Post-create CTA: “Apply policy now”.
- Unbound badge in list.
- Feature-flag rollout + instrumentation.

## Design Inputs
- [UI/UX report](./reports/ui-ux-designer-report.md)

## Unresolved questions
- Should policy modal auto-open immediately after create, or use toast CTA only?
- Do we block “Done” if secret has no binding, or allow with warning?
- Should `source` stay in basic section or move to advanced?
