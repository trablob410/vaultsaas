---
title: "Parameter-only custom policy system for secrets"
description: "Design and rollout plan for create/edit/apply policy using templates, enforced in workflow without DSL or arbitrary code."
status: pending
priority: P1
effort: 14h
branch: fix/seed-schema-migration
tags: [policy, workflow, security, dashboard, api, mvp]
created: 2026-03-19
---

# Objective

Implement policy system where users can:
- create policy
- edit policy
- apply policy to secrets

Constraints:
- parameter-only (no arbitrary code, no DSL)
- align with current Go monolith + Next.js dashboard + existing workflow paths
- keep YAGNI/KISS/DRY

Decision lock from product:
- secret overrides can be weaker than template, but must produce warnings
- template management permission = owner/admin only
- policy version history required in v1
- agent visibility starts enforcement-only; policy detail readable only when user explicitly permitted that agent

# Explicit non-goals

1. No policy scripting language.
2. No user-defined expressions/conditions (`if`, `eval`, regex rules, CEL/Rego).
3. No external policy engine/service.
4. No cross-org/global policy inheritance tree.
5. No retrofitting dynamic secret provider policy in v1.
6. No AB testing engine for policy decisions.
7. No breaking migration for existing `secrets.policy` data.

# Current-state findings (from code)

1. `secrets.policy` already exists as `JSONB` and is persisted.
2. Runtime enforcement today ignores `secrets.policy`; it uses `policy.ForCredentialType(secret.CredentialType)` in workflow.
3. Enforcement touchpoints are concentrated (good):
   - `workflow.Service.CreateRequest` (reason, duration cap, daily limit, cooldown, auto-approve)
   - `workflow.Handler.CreateRequest` (notification based on policy)
   - `workflow.Handler.GetCredential` (single-use auto-revoke)
4. Dashboard currently edits secret metadata/value only; no policy UI.

Implication: fastest safe path = keep current enforcement architecture, insert one policy resolution layer reused by all workflow entry points.

# Recommended design (MVP)

## 1) Product model

Two-level model, minimal:

1. **Policy Template** (project-scoped reusable object)
   - system templates (built-in defaults users can clone)
   - user custom templates (editable)
2. **Secret Policy Binding**
   - each secret references one template
   - optional per-secret parameter overrides (strictly bounded)
   - weaker-than-template values allowed, but returned/logged as warnings

Effective policy = `template.parameters` + `secret.override_parameters` (override wins if valid).

Reasoning:
- satisfies “templates users can clone/customize”
- supports “apply policy to secrets” cleanly
- avoids DSL/engine complexity

## 2) Parameter schema (v1)

Keep to already-enforced knobs only (DRY, no new policy semantics):

```json
{
  "max_duration_minutes": 60,
  "require_approval": true,
  "allow_auto_approve": false,
  "require_reason": true,
  "min_reason_length": 20,
  "max_requests_per_day": 20,
  "cool_down_minutes": 5,
  "single_use": false,
  "notify_on_access": true,
  "require_consent": false
}
```

No other parameters in v1.

## 3) Enforcement design in existing Go workflow paths

Add new resolver in `server/internal/policy`:

- `ResolveEffectivePolicy(ctx, secretID)` or `ResolveEffectivePolicy(secret, template, override)`
- outputs typed struct used by workflow

Use resolver at 3 existing points:

1. `workflow/service.go::CreateRequest`
   - replace direct `ForCredentialType(...)`
   - enforce reason, duration cap, daily limit, cooldown, initial status
2. `workflow/handler.go::CreateRequest`
   - use resolved policy for notify decision
3. `workflow/handler.go::GetCredential`
   - use resolved policy for single-use revoke

Rule: no direct policy checks outside resolver + these workflow gates.

# Data model plan

## New tables

### `policy_templates`

- `id UUID PK`
- `project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE`
- `name VARCHAR(120) NOT NULL`
- `description TEXT NOT NULL DEFAULT ''`
- `is_system BOOLEAN NOT NULL DEFAULT false`
- `base_credential_type VARCHAR(50) NULL` (optional hint; null = generic)
- `parameters JSONB NOT NULL`
- `created_by UUID NULL REFERENCES users(id)`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- unique `(project_id, lower(name))`

### `policy_template_versions`

- `id UUID PK`
- `template_id UUID NOT NULL REFERENCES policy_templates(id) ON DELETE CASCADE`
- `version INT NOT NULL` (starts at 1, increments per edit)
- `parameters JSONB NOT NULL`
- `change_note TEXT NOT NULL DEFAULT ''`
- `created_by UUID NULL REFERENCES users(id)`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- unique `(template_id, version)`

### `secret_policy_bindings`

- `secret_id UUID PK REFERENCES secrets(id) ON DELETE CASCADE`
- `template_id UUID NOT NULL REFERENCES policy_templates(id)`
- `template_version INT NOT NULL`
- `override_parameters JSONB NOT NULL DEFAULT '{}'`
- `override_warnings JSONB NOT NULL DEFAULT '[]'`
- `updated_by UUID NULL REFERENCES users(id)`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

### `secret_policy_agent_permissions`

- `id UUID PK`
- `secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE`
- `agent_id UUID NOT NULL REFERENCES agent_identities(id) ON DELETE CASCADE`
- `can_read_effective_policy BOOLEAN NOT NULL DEFAULT true`
- `granted_by_user_id UUID NOT NULL REFERENCES users(id)`
- `granted_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `expires_at TIMESTAMPTZ NULL`
- `revoked_at TIMESTAMPTZ NULL`
- unique active permission by `(secret_id, agent_id)` where `revoked_at IS NULL`

## Existing table use

- keep `secrets.policy` as backward-compatible mirror for phase migration only.
- v1 write both places during migration window (short-term dual write), then deprecate reads from `secrets.policy`.

## Access request snapshot (v1 policy history)

Add to `access_requests`:
- `applied_policy JSONB NOT NULL DEFAULT '{}'`
- `applied_template_id UUID NULL`
- `applied_template_version INT NULL`
- `applied_policy_source VARCHAR(32) NOT NULL DEFAULT 'default'`
- `applied_policy_warnings JSONB NOT NULL DEFAULT '[]'`

Reason: preserve explainability even if template later changes.

## Seeded system templates

Per project, on first policy page visit or project creation:

1. `Default API Key`
2. `Default DB Credential`
3. `Default SSH Key`
4. `Default OAuth Token`
5. `Default Cloud Credential`
6. `Default Personal Session`

Parameter values from current `policy.DefaultPolicyForTier` mapping.

# API design

Base: `/api/v1`

## Policy template endpoints

1. `GET /projects/{project_id}/policy-templates`
   - list templates
2. `POST /projects/{project_id}/policy-templates`
   - create custom template
3. `GET /policy-templates/{template_id}`
   - template detail
4. `PUT /policy-templates/{template_id}`
   - create new template version (system templates not editable)
5. `POST /policy-templates/{template_id}/clone`
   - clone system/custom template into same project
6. `DELETE /policy-templates/{template_id}`
   - only if not bound to secrets
7. `GET /policy-templates/{template_id}/versions`
   - list version history

Permission: owner/admin only for create/edit/delete/clone template actions.

## Secret policy binding endpoints

1. `GET /secrets/{secret_id}/policy-binding`
    - returns template + overrides + effective policy
2. `PUT /secrets/{secret_id}/policy-binding`
    - assign template version, optional overrides, returns warnings array
3. optional bulk endpoint later (non-MVP):
    - `POST /projects/{project_id}/secrets/policy-binding:bulk`

## Agent visibility endpoints (v1)

1. `POST /secrets/{secret_id}/policy-agent-permissions`
   - grant one agent read access to effective policy details
2. `DELETE /secrets/{secret_id}/policy-agent-permissions/{agent_id}`
   - revoke grant

Default for agents: no policy detail access, enforcement-only behavior.

## Response shape (binding)

```json
{
  "secret_id": "...",
  "template": {"id": "...", "name": "Default DB Credential", "is_system": true},
  "template_version": 3,
  "override_parameters": {"max_duration_minutes": 30},
  "override_warnings": ["weaker:max_duration_minutes"],
  "effective_policy": {"max_duration_minutes": 30, "require_approval": true, "...": "..."}
}
```

## Auth/RBAC

- policy template CRUD: owner/admin only.
- binding update: users with secret write permission.
- workflow enforcement remains dual-auth compatible for user/agent requests.
- agent policy detail read requires explicit grant by user on that secret.

# Validation rules

Central validator in `internal/policy` used by template create/edit and binding update.

Hard bounds (v1):

- `max_duration_minutes`: 1..1440
- `min_reason_length`: 0..500
- `max_requests_per_day`: 1..1000
- `cool_down_minutes`: 0..1440

Cross-field checks:

1. `require_reason=false` => force `min_reason_length=0`
2. `allow_auto_approve=true` => `require_approval=false`
3. `single_use=true` with `max_duration_minutes > 240` => reject (risk guard)
4. if `require_approval=true` and `allow_auto_approve=true` => reject

Safety policy:

- Reject unknown keys (prevents silent config drift).
- Reject type mismatch.
- Reject invalid template reference / cross-project binding.
- Allow weaker overrides, but compute warning codes and persist them.

Weaker override warning examples:
- `weaker:max_duration_minutes`
- `weaker:require_approval`
- `weaker:allow_auto_approve`
- `weaker:min_reason_length`
- `weaker:max_requests_per_day`
- `weaker:cool_down_minutes`
- `weaker:single_use`
- `weaker:notify_on_access`
- `weaker:require_consent`

# Enforcement algorithm (deterministic)

For each access request:

1. load secret
2. load secret binding; if none, fallback default by credential type
3. load template params (if binding exists)
4. merge + validate => effective policy
5. compute weaker-warning codes by comparing effective vs template
5. enforce in existing order:
   - reason gate
   - duration cap
   - daily limit
   - cooldown
   - initial status (auto-approve or pending)
6. on credential retrieval, apply single-use revoke rule
7. persist applied snapshot + template version + warnings on `access_requests`
8. emit audit + metrics with policy source (`default` | `template` | `template+override`)

Fallback behavior:
- if policy row corrupt/unreadable: fail-closed for create/update APIs, fail-open-to-default only for runtime reads initially behind feature flag + high-severity log.

# UX flow (dashboard)

## A. Policy Templates page (new)

Path: `/projects/[id]/policies`

Flows:
1. list templates (system + custom)
2. clone system template
3. create custom template from scratch
4. edit custom template
5. delete unused custom template

## B. Secret create/edit integration

In secret form:
1. select template (required)
2. optional “Customize for this secret” accordion (override fields)
3. inline validation + preview card showing effective policy

## C. Secret detail

Show:
- assigned template
- overridden fields badge
- effective policy summary
- CTA: “Edit policy binding”

## D. Approvals UX impact

No new page required. Existing approvals page can show computed constraints hint (duration cap/reason requirement) when rejecting invalid requests.

## E. Weaker override UX behavior

- Show non-blocking warning banner when user saves weaker-than-template override.
- Show warning chips near weaker fields.
- Save still allowed.

# Rollout phases

Phase docs:
- [phase-00-foundations-validator-and-observability.md](./phase-00-foundations-validator-and-observability.md)
- [phase-01-data-model-api-and-version-history.md](./phase-01-data-model-api-and-version-history.md)
- [phase-02-workflow-enforcement-and-policy-snapshot.md](./phase-02-workflow-enforcement-and-policy-snapshot.md)
- [phase-03-dashboard-template-and-binding-ux.md](./phase-03-dashboard-template-and-binding-ux.md)
- [phase-04-launch-hardening-agent-policy-visibility.md](./phase-04-launch-hardening-agent-policy-visibility.md)

## Phase 0 — Foundations (2h)

- add policy schema/constants/validator/resolver
- add metrics + structured logs
- no behavior change yet

## Phase 1 — Data + API (4h)

- migrations for `policy_templates`, `secret_policy_bindings`
- migrations for `policy_template_versions`, `secret_policy_agent_permissions`, access-request snapshot columns
- template CRUD + clone APIs
- binding read/write API
- seed system templates

### Phase 1 current status update (2026-03-20)

- **Implementation progress:** core schema + API scaffolding implemented in `server/internal/policy` and migration `000026` added.
- **Review state:** completed code review; phase remains open pending hardening/test closure.
- **Key blockers to close Phase 1:**
  1. Resolve `/secrets` route overlap risk for policy-binding and policy-agent-permission endpoints.
  2. ~~Normalize policy validation errors to return `400 bad_request` instead of internal error paths.~~ ✅ done (2026-03-20)
  3. Harden concurrency behavior for template versioning and permission grant/revoke.
  4. Add integration tests for all Phase 1 policy APIs.

### Phase 1 maintainability update (2026-03-20)

- `server/internal/policy` was consolidated from many micro-files into concern-based files:
  - `handler.go`, `service.go`, `repository.go`, `models.go`, `errors.go`.
- No behavior expansion; structural simplification only for scanability and maintainability.


## Phase 2 — Workflow enforcement switch (3h)

- wire resolver into workflow service/handler touchpoints
- add fallback mode flag `POLICY_ENFORCEMENT_V2_ENABLED`

## Phase 3 — Dashboard UX (3h)

- templates page
- secret form binding controls
- effective policy preview

## Phase 4 — Harden + launch (2h)

- migration/backfill from `secrets.policy` when possible
- staged enable by internal project then all projects
- monitor metrics; rollback via feature flag
- gate agent policy-detail read behind explicit grant check

# Metrics and observability

## Product metrics

1. `% secrets with template binding`
2. `% secrets with override`
3. template clone count / week
4. policy edit frequency

## Enforcement metrics

1. `policy_resolution_total{source,status}`
2. `policy_validation_fail_total{stage}`
3. `access_request_rejected_total{reason=policy_*}`
4. `access_request_auto_approved_total`
5. `single_use_auto_revoke_total`

## Reliability metrics

1. p95 latency on create request endpoint
2. DB query count delta in workflow path
3. runtime fallback-to-default count (should trend to zero)
4. `agent_policy_read_denied_total`
5. `agent_policy_read_granted_total`

# Risks and mitigations

1. **Policy sprawl** (too many templates)
   - mitigate: naming rules + archive/delete unused + show usage count
2. **Broken bindings after template delete**
   - mitigate: block delete if bindings exist
3. **Permission gaps in project scope**
   - mitigate: enforce project ownership checks on all template/binding endpoints
4. **Behavior regressions in workflow**
   - mitigate: feature flag + golden tests against existing default behavior
5. **Dual-write drift (`secrets.policy` vs new tables)**
   - mitigate: one-way migration plan, short dual-write window, audit job
6. **Policy version confusion (binding to wrong version)**
   - mitigate: explicit `template_version` in binding + UI labels + snapshot on request

# Test strategy

## Unit (Go)

1. validator tests: bounds, unknown keys, cross-field rules
2. resolver tests: default/template/override precedence
3. enforcement tests in `workflow/service` for each gate
4. backward compatibility tests (no binding => old default behavior)

## Integration (Go + DB)

1. template CRUD and clone semantics
2. binding updates + project isolation
3. workflow request creation with effective policy
4. delete template with active bindings returns 409
5. template edit creates new version, older version remains queryable
6. agent permission required for policy-detail read route

## Dashboard tests

1. template list/create/clone/edit form behavior
2. secret form policy selection and preview
3. API client type-safe payload contract tests

## E2E smoke

1. create custom template -> bind to secret -> request access -> enforcement observed
2. same credential type, two secrets with different limits => different outcomes
3. weaker override saved -> warning visible + enforcement uses weaker effective policy
4. agent without grant cannot read effective policy detail; with grant can read

# Implementation notes (YAGNI/KISS/DRY guardrails)

1. Reuse existing `internal/policy` package; do not create new service process.
2. Keep policy parameter set fixed in v1; expanding keys requires explicit schema version bump.
3. Use one resolver and one validator shared by API + workflow.
4. Do not add bulk operations until clear demand.
5. Keep API error format unchanged (`{"error": {"code","message"}}`).

# Open migration strategy

1. New secrets: must create binding at secret creation (or bind default template automatically).
2. Existing secrets:
   - if `secrets.policy='{}'`: bind credential-type default template
   - if non-empty JSON policy parseable: map to nearest template + put residual keys in override
    - if unparsable: bind default and flag for manual review report

# Unresolved questions

1. Should agent policy-read grants be per-agent permanent until revoke, or default-expiring (e.g., 24h)?
2. Should binding default to latest template version, or pin explicit selected version always? (recommended: pin explicit)
