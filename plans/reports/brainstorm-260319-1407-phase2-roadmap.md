# Brainstorm: Valt Phase 2 Roadmap

**Date:** 2026-03-19
**Status:** Agreed — ready for planning

---

## Context

MVP is complete and solid: vault + workflow RBAC + audit + org hierarchy + agent identity + E2E credential delivery + dynamic secrets. Phase 2 is about making the product *feel alive* for real teams.

**Primary target:** Dev teams 5-20 ppl (Slack-first, daily AI agent usage)
**Secondary:** Enterprise/security teams (compliance, scale)

---

## The 3 Phase 2 Pillars

### Pillar 1: Approval Channels (Build first)

**Problem:** Approvers must open the dashboard to act. Kills adoption.
**Goal:** Approve/reject from wherever approvers already live.

#### 1A — Email action links (~2-3 days, zero new infra)
- Add signed one-time tokens (`HMAC-SHA256`, short TTL) to existing SMTP notification emails
- New endpoint: `POST /api/v1/requests/{id}/action?token=<signed>&action=approve|reject`
- "Approve ✓" / "Reject ✗" clickable links inline in email body
- Uses existing SMTP infrastructure — no new services

#### 1B — Slack bot (~1 week)
- Slack app with Block Kit interactive messages
- Bot DMs assigned approver on request creation
- Approve/Reject buttons → Slack sends callback → Go webhook handler
- User setting: link Slack account (store `slack_user_id`)
- New DB table: `user_notification_channels (user_id, channel_type, channel_handle, verified)`
- New route: `POST /api/v1/webhooks/slack`

#### 1C — Telegram bot (~1 week)
- Telegram bot with inline keyboard (Approve / Reject buttons)
- User sends `/start` to bot → links chat_id to Valt account
- New route: `POST /api/v1/webhooks/telegram`

#### 1D — Zalo OA (~1-2 weeks, Vietnamese market)
- Zalo Official Account with action messages
- Higher complexity due to Zalo API constraints
- Do after Slack + Telegram

**Shared architecture (1B/1C/1D):**
- `notify.Service` → pluggable channel adapters (email, slack, telegram, zalo)
- Fallback chain: user's preferred channel → email → dashboard only
- Settings UI: "Connect your notification channels" per user

---

### Pillar 2: Custom Policies (Build second)

**Problem:** Policy engine is hardcoded Tier 1-4 by credential type. No team can customize without code changes.

**Design: Per-secret rules + project-level defaults**

```
Project policy (defaults inherited by all secrets):
  default_approvers: [user_id, ...]
  default_max_duration: 4h
  default_notify_channel: slack | email | telegram
  require_reason: true, min_chars: 20

Secret policy (overrides project policy):
  approvers: [specific_user_id, ...]
  max_duration: <override>
  auto_approve: true | false (overrides tier)
  block: true (never approve, even Tier 1)
  business_hours_only: "09:00-18:00 Mon-Fri" (TZ-aware)
  escalation_after_minutes: 30 → backup_approver_id
```

**DB changes:**
- Add `policy_config JSONB` to `projects` table
- Add `policy_config JSONB` to `secrets` table
- Policy resolver: merge at request time — secret → project → global tier defaults

**UI:** Simple form drawer on project settings page + secret detail page. No YAML. Friendly toggles/selects.

---

### Pillar 3: valt CLI — All-in-one Go binary (Build third)

**Two roles in one binary:** setup/onboarding tool AND daily secret access tool.

**Why critical:** Current MCP server setup is broken for new users:
- No pre-built binaries (requires Rust toolchain + `cargo build`)
- Manual `~/.valt/config.toml` creation
- No IDE config templates (Claude Desktop, Cursor, VS Code)
- Chicken-and-egg auth: need MCP connected to use `authenticate_agent`, need auth to connect MCP

**Commands:**
```bash
# Setup & onboarding
valt setup                          # interactive wizard: API URL → browser OAuth → pick project → create agent token → keychain → print MCP config
valt mcp install --ide claude       # writes to ~/.claude/claude_desktop_config.json
valt mcp install --ide cursor       # writes to Cursor MCP settings path
valt auth status                    # show current auth state + project

# Daily use
valt list [--project X]             # list accessible secrets
valt get DB_PASSWORD                # print value to stdout
valt run -- node server.js          # inject all accessible secrets as env vars → run command
valt request DB_PASSWORD --reason "fixing prod" --duration 2h
valt status <request-id>            # poll approval status
```

**Tech:**
- Go binary (cross-compile Mac/Linux/Windows)
- Auth: `zalando/go-keyring` (OS keychain, mirrors Rust keyring pattern in MCP server)
- Distribution: GitHub Releases + `curl -sSL https://get.valt.dev | sh` install script
- `valt run` is the killer feature: secrets injected as env vars, devs never see them

---

## Sprint Plan

| Sprint | Deliverable | Why |
|--------|-------------|-----|
| S1 | Email action links | Zero infra, immediate value for existing users |
| S2 | Slack bot + notification channel settings UI | Viral — colls see the bot, team adopts |
| S3 | Custom policies (project + secret level) | Unlocks proper team workflows |
| S4 | valt CLI: `setup` + `mcp install` + `get` + `run` | Onboarding + daily DX |
| S5 | Telegram + Zalo bots | Expand reach, Vietnamese market |
| S6 | Secret tags + versioning + .env import | Secret hygiene |

---

## Lower Priority (Phase 3)

**Agent features:**
- Agent activity timeline dashboard (all agent requests in one view)
- Per-agent secret allowlist (restrict which secrets an agent can request)
- Agent trust scoring (reputation based on approval history)

**Enterprise/compliance:**
- TOTP 2FA on dashboard login
- Break-glass emergency access (bypass with extra audit trail)
- SSO / SAML
- Compliance audit export (SOC 2 format)
- Anomaly detection alerts (unusual agent activity)
- IP allowlist

**Secret lifecycle:**
- Secret expiration (the secret itself, not just sessions)
- Rotation scheduling with auto-notify
- Versioning + rollback

---

## Unresolved Questions

1. **Zalo OA** — does the team have access to Zalo OA account / dev credentials? API approval process can take weeks.
2. **Slack app distribution** — internal use only, or distributed via Slack App Directory? (Directory = review process)
3. **Email action links security** — what's the token TTL? Should tokens be single-use-strictly? (Concurrent clicks = double approve?)
4. **Business hours TZ** — whose timezone for `business_hours_only`? Secret owner's? Approver's? Org setting?
5. **CLI `valt run`** — how to handle secrets that require approval? Block until approved, skip, or error?
6. **Binary signing** — do we need code signing for macOS (Gatekeeper) and Windows (SmartScreen) from day one?
