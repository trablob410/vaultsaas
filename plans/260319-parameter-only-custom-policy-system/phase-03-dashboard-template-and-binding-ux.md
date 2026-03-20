# Phase 3: Dashboard UX — Templates + Secret Binding

## Context Links
- [plan.md](./plan.md)
- [phase-01](./phase-01-data-model-api-and-version-history.md)
- `dashboard/src/app/(dashboard)/`

## Overview
- **Priority:** P2
- **Status:** partial
- **Description:** Add policy template management UI and secret policy binding UX with effective policy preview and weaker warnings.

## Goals
1. Add policy templates page for project scope.
2. Add secret form controls for template selection and overrides.
3. Show weaker-warning UX non-blocking.
4. Keep forms simple and parameter-only.

## UX Scope
1. New route: `/projects/[id]/policies`
   - list templates
   - clone template
   - create custom template
   - edit custom template (creates new version)
   - version history drawer/list
2. Secret create/edit:
   - select template + version
   - optional override panel
   - effective policy preview card
   - warning chips for weaker overrides
3. Secret detail:
   - show assigned template/version and active warnings

## Permission UX
- Hide/disable template management actions unless owner/admin.
- Keep binding controls aligned with secret write permissions.

## Implementation Steps
1. Add typed API client methods and TS models for templates/bindings/permissions.
2. Build templates page with create/clone/edit/version history flows.
3. Extend secret form with template picker + override inputs + live preview.
4. Show non-blocking warning banner/chips on weaker settings.
5. Add dashboard tests for key flows and payload integrity.

## Todo List
- [x] Add policy models and API client methods
- [x] Add policies page under project scope
- [x] Add template version history UX
- [x] Add secret binding form section + preview
- [x] Add weaker override warning UX
- [x] Add dashboard tests for templates + binding flows

## Success Criteria
1. User can create/clone/edit template and see version history.
2. User can bind template to secret and set overrides.
3. Weaker overrides save successfully with visible warnings.
4. Unauthorized roles cannot manage templates.

## Risks
- UI complexity creep.
  - Mitigation: parameter-only form, no condition builder, no matrix editor.

## Dependencies
- Depends on Phase 1 APIs and Phase 2 enforcement data contracts.

## Output
- Users can manage and apply parameter-only policies from dashboard.
