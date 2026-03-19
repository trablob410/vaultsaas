# Phase 4: Launch Hardening + Agent Policy Visibility Controls

## Context Links
- [plan.md](./plan.md)
- [phase-02](./phase-02-workflow-enforcement-and-policy-snapshot.md)
- [phase-03](./phase-03-dashboard-template-and-binding-ux.md)

## Overview
- **Priority:** P1
- **Status:** pending
- **Description:** Stabilize rollout, migrate legacy bindings, and enforce agent visibility policy (enforcement-only by default, detail-read only by explicit user grant).

## Goals
1. Progressive rollout with safe rollback.
2. Backfill legacy secrets to template bindings.
3. Enforce agent policy detail access gate.
4. Validate metrics and SLO impact.

## Launch Rules
1. Default agent behavior: enforcement-only, no policy detail API response.
2. Agent can read effective policy details only if user granted permission for that secret.
3. Permission checks include project scope + secret ownership/admin context.

## Migration Scope
1. Auto-bind default template for secrets with empty legacy policy.
2. For parseable legacy non-empty policy, map to template + override.
3. For unparseable policy, bind default and emit review report.

## Observability Gates
1. Monitor validation fail rate.
2. Monitor fallback-to-default rate.
3. Monitor `agent_policy_read_denied_total` and grant usage trend.
4. Monitor create-request latency and error rate.

## Implementation Steps
1. Add migration/backfill job and dry-run mode.
2. Add policy-detail read guard in relevant API handlers.
3. Add audit events for grant/revoke and denied/allowed reads.
4. Enable feature flag for internal project cohort.
5. Expand rollout by cohort when metrics stable.

## Todo List
- [ ] Implement backfill + review report path
- [ ] Enforce agent policy detail grant checks
- [ ] Add audit logs for permission events
- [ ] Add rollout playbook + rollback procedure
- [ ] Run staged rollout and monitor telemetry

## Success Criteria
1. Legacy secrets successfully bound at high coverage.
2. Agents denied policy details unless explicit grant exists.
3. Granted agents can read effective policy details for permitted secrets.
4. No material regression in workflow latency/error SLO.

## Risks
- Incorrect grants causing overexposure.
  - Mitigation: deny-by-default, explicit grant, audit trail, revoke endpoint.

## Dependencies
- Depends on Phases 1–3 complete.

## Output
- Production-ready parameter-only policy system with controlled agent visibility.
