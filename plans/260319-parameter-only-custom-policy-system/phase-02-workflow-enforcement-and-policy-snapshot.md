# Phase 2: Workflow Enforcement + Applied Policy Snapshot

## Context Links
- [plan.md](./plan.md)
- [phase-01](./phase-01-data-model-api-and-version-history.md)
- `server/internal/workflow/service.go`
- `server/internal/workflow/handler.go`

## Overview
- **Priority:** P1
- **Status:** pending
- **Description:** Switch workflow to effective policy resolution from template/binding and persist request-time policy snapshots.

## Goals
1. Replace direct tier default checks with resolver in workflow.
2. Persist `applied_policy` snapshot + template/version + warnings on request create.
3. Keep deterministic enforcement order.
4. Add feature flag for safe rollout.

## Enforcement Touchpoints
1. `CreateRequest` service:
   - reason gate
   - duration cap
   - daily limit
   - cooldown
   - initial status
   - persist applied snapshot
2. `CreateRequest` handler:
   - notify decision from effective policy
3. `GetCredential` handler:
   - single-use revoke from effective policy

## Runtime Safety
- If runtime policy resolution fails:
  - before full rollout: fallback to credential-type default + high-severity log/metric
  - after stabilization: configurable fail-closed mode

## Implementation Steps
1. Integrate resolver into workflow service path.
2. Add data writes for applied snapshot columns on access request creation.
3. Update query scans/models to include snapshot metadata where needed.
4. Integrate weaker warning fields into audit/metrics.
5. Add feature flag `POLICY_ENFORCEMENT_V2_ENABLED`.
6. Add unit + integration tests for old/new behavior parity and new logic.

## Todo List
- [ ] Wire resolver into `CreateRequest`
- [ ] Persist `applied_policy` and metadata on request create
- [ ] Wire notify and single-use checks to resolved policy
- [ ] Add fallback behavior + telemetry
- [ ] Add feature flag gating
- [ ] Add tests for deterministic enforcement and parity

## Success Criteria
1. Two secrets with same credential type can enforce different policy outcomes.
2. Request stores immutable applied policy snapshot/version.
3. Existing behavior preserved when no binding exists.
4. Feature flag can disable new path quickly.

## Risks
- Extra DB reads in hot path.
  - Mitigation: minimal join strategy, optional short-lived cache later only if needed.

## Dependencies
- Depends on Phase 1 tables and APIs.

## Output
- Live enforcement uses template/binding effective policy reliably.
