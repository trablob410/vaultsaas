# Phase 1: Data Model + API + Version History

## Context Links
- [plan.md](./plan.md)
- [phase-00](./phase-00-foundations-validator-and-observability.md)
- `server/internal/database/migrations/`

## Overview
- **Priority:** P1
- **Status:** pending
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
- [ ] Add migrations for new policy tables
- [ ] Add migration for `access_requests` policy snapshot columns
- [ ] Implement template APIs with version-on-edit behavior
- [ ] Implement binding APIs with warning return payload
- [ ] Implement agent permission grant/revoke APIs
- [ ] Seed system templates
- [ ] Integration tests for all policy APIs

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
