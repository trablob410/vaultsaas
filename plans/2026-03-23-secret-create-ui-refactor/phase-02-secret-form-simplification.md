# Phase 02 — Secret form simplification

Status: completed  
Effort: 4h  
Priority: P2

## Objective
Make create/edit secret dialog focused, short, and predictable.

## Requirements
- Remove embedded policy section from `SecretForm`.
- Keep fields: name, description, type, source, value(new only).
- Keep current validation; tighten required markers.
- Expose callback to open policy modal after create.

## UX rules
- Primary action text:
  - Create mode: `Create secret`
  - Edit mode: `Save changes`
- Cancel closes modal; prompt only when dirty.
- `value` hidden on edit unless product explicitly needs inline rotation.

## Steps
1. [x] Remove `SecretPolicyBindingSection` usage and related hook calls.
2. [x] Keep submit path only for secret create/update.
3. [x] Add optional `onCreated(secretId: string)` callback in props.
4. [x] After create success: call `onCreated`, then `onSuccess`.
5. [x] Add dirty-state check for accidental close.

## Implementation notes
- `SecretForm` now keeps only core secret fields and submit behavior.
- Close actions (cancel/backdrop) prompt only when form is dirty.
- CTA labels follow UX rules exactly: `Create secret` / `Save changes`.
- Value field remains create-only; hidden in edit mode.

## Bug fix update (2026-03-23)
- Fixed create-secret project scoping regression: create flow now forwards `project_id` from `valt_current_project`.
- `SecretForm` accepts `projectId` and includes it in `api.secrets.create(...)` payload.
- API typing updated to include optional `project_id` for create requests.
- Backend vault create path now accepts `project_id`, validates UUID format, and persists it to `secrets.project_id`.

## Testing focus
- Create secret success and failure paths unchanged.
- Edit secret path unchanged.
- No policy API call from `SecretForm`.
- New secret created from dashboard includes expected `project_id` in backend record.

## Success criteria
- Form visually shorter and single-purpose.
- Secret CRUD behavior unchanged.
