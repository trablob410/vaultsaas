# Phase 1: Data Model + API + Version History

## Context Links
- [plan.md](./plan.md)
- [phase-00](./phase-00-foundations-validator-and-observability.md)
- `server/internal/database/migrations/`

## Overview
- **Priority:** P1
- **Status:** in-review
- **Description:** Add policy templates, template versions, secret bindings, agent policy-read permissions, and API endpoints.

## Goals
1. Introduce tables for template model + v1 version history.
2. Add template CRUD/clone/version endpoints.
3. Add secret policy binding read/write endpoints.
4. Add policy-agent permission grant/revoke endpoints.

## Data Changes
1. Create `policy_templates`.
2. Create `policy_template_versions`.
3. Create `secret_policy_bindings` (with `template_version`, `override_warnings`).
4. Create `secret_policy_agent_permissions`.
5. Alter `access_requests` with applied policy snapshot columns.

## API Scope
1. Template endpoints (owner/admin only): list/create/get/update-version/clone/delete/list-versions.
2. Binding endpoints: get/update binding + effective policy.
3. Permission endpoints: grant/revoke agent policy-detail read on secret.

## Authorization Rules
1. Template management: owner/admin only.
2. Secret binding update: secret write permission.
3. Agent policy-detail read grant/revoke: owner/admin only.

## Implementation Steps
1. Write migrations with indexes + FK + uniqueness constraints.
2. Add policy repository/service methods for templates/versions/bindings/permissions.
3. Add handlers/routes for template, binding, permission endpoints.
4. Enforce project-scope checks on every endpoint.
5. Seed system templates mapped from existing tier defaults.
6. Add integration tests for CRUD/version/binding/permission behavior.

## Todo List
- [x] Add migrations for new policy tables
- [x] Add migration for `access_requests` policy snapshot columns
- [x] Implement template APIs with version-on-edit behavior
- [x] Implement binding APIs with warning return payload
- [x] Implement agent permission grant/revoke APIs
- [x] Seed system templates
- [x] Integration tests for all policy APIs

## Latest Code Review (2026-03-20)

### Review Status
- Review completed for Phase 01 implementation in `server/internal/policy`, migration `000026`, and route wiring in `server/cmd/server/main.go`.
- Outcome: **needs hardening before phase closure**.

### Findings
1. **Critical:** Route overlap risk for `/secrets/{secret_id}/policy-*` endpoints due to `/secrets` mount ordering.
2. **High:** Invalid override payloads may surface as 500 instead of 400 in some resolver/repository paths.
3. **High:** Concurrency risk in template version creation (`MAX(version)+1`) and permission grant (revoke-then-insert).
4. **High:** Integration tests for policy APIs are still missing.
5. **Medium:** Seed idempotency edge case when template name conflicts with non-system template.
6. **Medium:** Binding does not enforce `base_credential_type` compatibility with secret credential type.

### Follow-up Tasks (from review)
- [ ] Add route-level integration test to verify policy endpoints are reachable under current router structure
- [x] Normalize validation failures to `ErrPolicyInvalid` (return 400 for invalid override/type/key cases)
- [ ] Make template version creation concurrency-safe (transaction/lock or retry on conflict)
- [ ] Make permission grant flow atomic and race-safe
- [ ] Add integration tests for template CRUD/version/binding/permission endpoints
- [ ] Decide and implement `base_credential_type` compatibility enforcement at binding time

## Phase 01 Module Structure Update (2026-03-20)

- Refactored `server/internal/policy` to reduce over-splitting and keep separation by concern.
- Consolidated implementation into:
  - `service.go` (all service/use-case logic)
  - `repository.go` (all repository/query/scan logic)
  - `handler.go` (all API handlers + request DTOs + response helpers)
  - plus `models.go` and `errors.go`
- Removed micro-split files for template/binding/permission service/repository/handler fragments.
- Handler input validation paths were unified to consistently map malformed payloads/route params to `ErrPolicyInvalid`.

## Success Criteria
1. Editing template creates new version row, old versions preserved.
2. Binding pins template version explicitly.
3. Overrides can be weaker and return warning codes.
4. Owner/admin restriction enforced for template management.
5. Agent permission grant/revoke works and is project-scoped safe.

## Risks
- Dual-write ambiguity with legacy `secrets.policy`.
  - Mitigation: short transition, read priority from new tables only once phase 2 starts.

## Dependencies
- Depends on Phase 0 validator/resolver types.

## Output
- Persistent model + APIs ready for enforcement and dashboard phases.
