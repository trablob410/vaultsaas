# Phase 0: Foundations — Validator + Resolver + Observability

## Context Links
- [plan.md](./plan.md)
- `server/internal/policy/engine.go`
- `server/internal/workflow/service.go`

## Overview
- **Priority:** P1
- **Status:** completed
- **Description:** Build shared policy schema, validation, and resolution primitives with no behavior change yet.

## Goals
1. Add canonical parameter-only policy schema.
2. Add reusable validator for template/binding write paths.
3. Add resolver returning effective policy + weaker warnings.
4. Add metrics/log fields to measure parse/validation/resolution quality.

## Scope
- In scope: internal policy package only, unit tests, observability hooks.
- Out of scope: DB migrations, API endpoints, dashboard UI, enforcement switch.

## Design Notes
- Keep one source of truth in `internal/policy`.
- Reject unknown keys and wrong types.
- Weaker-than-template is allowed; return warning codes, do not block save.

## Implementation Steps
1. Define `PolicyParameters` typed struct for v1 keys only.
2. Implement `ValidateParameters(input)` with bounds + cross-field checks.
3. Implement `ResolveEffectivePolicy(template, override)`:
   - merge override over template
   - validate merged result
   - compute weaker warning list vs template
4. Add helper `CompareWeakening(template, effective) []string`.
5. Add structured log fields: source, warning_count, warning_codes.
6. Add metrics stubs/counters in policy package callsites.

## Todo List
- [x] Define v1 parameter schema type
- [x] Implement central validator
- [x] Implement resolver + weaker warning detector
- [x] Add unit tests for validator/resolver
- [x] Add observability fields/counters (no behavior switch)

## Success Criteria
1. Validator rejects unknown keys/type mismatch.
2. Resolver returns deterministic effective policy.
3. Weaker overrides generate warning codes and still resolve.
4. No workflow behavior change in this phase.

## Risks
- Over-coupling schema to transport payloads.
  - Mitigation: keep internal type and map from API DTOs later.

## Dependencies
- None. Base phase.

## Output
- Internal policy primitives ready for API + workflow phases.
