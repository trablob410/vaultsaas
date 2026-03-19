---
title: "Per-secret policy architecture options (MVP)"
description: "Practical options to enforce policy per secret with clear trade-offs and low-risk rollout"
status: pending
priority: P2
effort: 6h
branch: fix/seed-schema-migration
tags: [policy, workflow, security, mvp, architecture]
created: 2026-03-19
---

# Per-secret policy system for Valt (MVP)

## Goal

Add **real per-secret policy enforcement** across Go API + Dashboard + MCP, keeping MVP velocity high.

## Current state (from codebase)

- `secrets.policy` already exists as `JSONB NOT NULL DEFAULT '{}'`.
- Secret create API already accepts/stores `policy` string.
- Runtime enforcement in `workflow` uses `policy.ForCredentialType(secret.CredentialType)` only.
- So today: policy is effectively **credential-type default**, not truly per-secret.
- Enforcement points live in one place (`workflow/service.go`, plus single-use in `workflow/handler.go`) which is good DRY leverage.

## MVP constraints / design principles

- YAGNI: no generic policy DSL, no external policy engine.
- KISS: keep policy decisions inside Go monolith, no extra infra.
- DRY: one policy resolution path used by dashboard + MCP requests.
- Keep backward-compatible behavior for existing secrets (`{}` policy).

## Decision criteria

1. Simplicity of implementation
2. Maintainability (typed validation, testability)
3. Runtime performance
4. Rollout risk (migrations, behavior regressions)

---

## Option 1 — JSONB override on `secrets.policy` (recommended for MVP)

### Architecture

- Keep `secrets.policy` as source of per-secret override.
- Add a small policy resolver in Go:
  - `default = policy.ForCredentialType(secret.CredentialType)`
  - parse `secret.policy` JSON override
  - sanitize/validate fields (only allow stricter or bounded values)
  - `effective = Merge(default, override)`
- Replace direct `ForCredentialType(...)` calls in workflow paths with `ResolvePolicy(secret)`.
- Return `effective_policy` (or normalized `policy`) in secret/detail and optional request/detail responses for UX transparency.

### Data model

- No new table.
- Optional DB check constraints later (not required MVP).

### Performance

- Per request: one small JSON parse + merge.
- Negligible vs DB I/O and crypto operations.
- Can cache in-memory later if needed, likely unnecessary now.

### Pros

- Lowest change surface, fastest to ship.
- No migration complexity beyond optional cleanup/backfill.
- Leverages existing schema + API payload.
- Easy incremental rollout with feature flag.

### Cons

- Weak DB-level typing/validation.
- Policy schema evolution handled in app code.
- Slight risk of malformed stored JSON if validation missed on one write path.

### Rollout risk

- **Low** if fallback defaults preserved when parse/validate fails.

---

## Option 2 — New typed `secret_policies` table (1:1 with secret)

### Architecture

- Add `secret_policies(secret_id PK/FK, max_duration_minutes, require_reason, ... )` typed columns.
- On create/update secret, write typed policy row.
- Workflow joins `secrets` + `secret_policies`, computes effective policy.
- Keep credential-type defaults as fallback when no row exists.

### Data model

- New table + migration + backfill job.
- Can add strict DB constraints (`CHECK` ranges).

### Performance

- Extra join in hot path (small but real).
- Could mitigate by embedding policy columns directly in `secrets` instead.

### Pros

- Strong typing, easier data quality guarantees.
- Cleaner analytics/querying on policy fields.
- Better long-term maintainability if policy complexity grows.

### Cons

- More schema + service + handler + test churn.
- Higher rollout and migration risk for MVP.
- DRY risk if dual-write/dual-read transition gets messy.

### Rollout risk

- **Medium** due to migration/backfill and extra query coupling.

---

## Option 3 — Policy profiles + per-secret reference (plus inline override)

### Architecture

- Add `policy_profiles` (project-level reusable templates).
- Secret stores `policy_profile_id` + optional override JSON.
- Effective policy = profile default -> credential default fallback -> secret override.

### Pros

- Best consistency for many secrets.
- Good for enterprise governance model.

### Cons

- Overbuild for MVP (new domain + UI + lifecycle + permission semantics).
- More cognitive load for users.
- Highest rollout risk and scope creep.

### Rollout risk

- **High** for current stage.

---

## Trade-off summary matrix

| Option | Simplicity | Maintainability | Performance | Rollout risk |
|---|---|---|---|---|
| 1. JSONB override on `secrets.policy` | **High** | Medium | **High** | **Low** |
| 2. Typed `secret_policies` table | Medium | **High** | Medium | Medium |
| 3. Profiles + overrides | Low | Medium/High (later) | Medium | **High** |

## Recommendation

Pick **Option 1** for MVP.

Reason:
- ships fastest with smallest blast radius
- fits existing schema and request flow
- preserves KISS/YAGNI while enabling actual per-secret behavior now
- keeps clean upgrade path to Option 2 later if policy schema hardens

## Recommended target MVP policy scope (80/20)

Support only fields already in engine and directly enforced today:

- `max_duration_minutes`
- `require_reason`
- `min_reason_length`
- `max_requests_per_day`
- `cool_down_minutes`
- `require_approval`
- `allow_auto_approve`
- `single_use`
- `notify_on_access`

Guardrails:
- allow only bounded values
- do not allow override to weaken critical defaults unless explicitly intended by product decision
- invalid override -> reject on write (or soft-fallback + audit log during early rollout)

---

## Phased rollout strategy

### Phase 0 — Contract + observability (no behavior change)

- Define canonical `PolicyOverride` schema in Go.
- Add resolver path with default-only behavior toggle OFF.
- Add metrics/logs:
  - policy resolve success/failure
  - fallback-to-default count
  - per-field override usage

Exit criteria:
- zero resolver errors in staging with default-only mode.

### Phase 1 — Write path enablement (dashboard + API)

- Validate/sanitize policy on secret create/update.
- Store normalized JSON into `secrets.policy`.
- Dashboard: add minimal advanced policy section behind feature flag.
- MCP: no protocol change required (policy applied server-side).

Exit criteria:
- policy saved/retrieved correctly for new secrets.

### Phase 2 — Read-time enforcement switch

- Switch workflow enforcement to `ResolvePolicy(secret)`.
- Keep fallback to credential defaults on parse/validation fail.
- Audit event includes effective policy snapshot hash or key fields (optional MVP+).

Exit criteria:
- staged e2e confirms duration/reason/limit/cooldown/single-use follow per-secret config.

### Phase 3 — Progressive rollout

- Enable by project/org cohort.
- Monitor rejection spikes, approval latency, error rates.
- If stable, default ON for all new secrets; then all secrets.

Exit criteria:
- no significant increase in workflow error rate; support tickets stable.

### Phase 4 — Hardening (post-MVP)

- Add DB constraints or migrate to Option 2 if needed.
- Add policy diff/audit views in dashboard.
- Add import/export presets only if demand justifies.

---

## Risks and mitigations

- Risk: malformed legacy policy JSON
  - Mitigation: strict parser with fallback + telemetry + repair script.
- Risk: accidental policy weakening
  - Mitigation: enforce clamp rules (cannot exceed safe bounds by tier).
- Risk: behavior surprise for existing secrets
  - Mitigation: default empty override keeps old behavior; feature-flag rollout.
- Risk: duplicated enforcement logic
  - Mitigation: single resolver used by all workflow entry points.

## Minimal acceptance criteria for MVP

1. Same secret type, different secrets can enforce different duration/reason/limit behavior.
2. Dashboard can set/view policy override for a secret.
3. MCP requests automatically follow same server-side policy without client update.
4. Existing secrets with `{}` keep current behavior.
5. Audit/metrics can show effective policy was applied.

## Unresolved questions

1. Should secret owners be allowed to make policy *less strict* than tier defaults, or only stricter?
2. Should policy be editable by `project admin` only, or also `member` with write access?
3. For rollout, should invalid stored policy hard-fail requests or soft-fallback + warn?
