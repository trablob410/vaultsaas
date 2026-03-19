# valt CLI Implementation Plan

> **For agentic workers:** REQUIRED: Use `/ck:plan` in execute mode (subagent-driven or sequential) to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Single Go binary that handles both new-user onboarding (setup, MCP install) and daily developer workflows (get secret, inject env, request access).

**Architecture:** Cobra CLI at `valt-cli/` root. Config at `~/.valt/config.toml`. Auth token in OS keychain (`zalando/go-keyring`). HTTP client reuses same REST API as dashboard. `valt run` is the killer feature — injects secrets as env vars before running a command.

**Tech Stack:** Go 1.22+, `spf13/cobra`, `zalando/go-keyring`, `BurntSushi/toml`, `goreleaser` for release pipeline

**Context:** Brainstorm → `plans/reports/brainstorm-260319-1407-phase2-roadmap.md`

---

## Status

| Phase | Description | Status |
|-------|-------------|--------|
| 01 | Project scaffold + config + keychain + API client | COMPLETED — 2026-03-19 |
| 02 | `valt setup` + `valt mcp install` | COMPLETED — 2026-03-19 |
| 03 | `valt get` + `valt list` + `valt run` | COMPLETED — 2026-03-19 |
| 04 | `valt request` + `valt status` | COMPLETED — 2026-03-19 |
| 05 | Build pipeline + install script | COMPLETED — 2026-03-19 |

## Directory Structure

```
valt-cli/
├── cmd/
│   ├── root.go          — cobra root, persistent flags, version
│   ├── setup.go         — interactive setup wizard
│   ├── mcp.go           — mcp install subcommand
│   ├── get.go           — get secret value
│   ├── list.go          — list accessible secrets
│   ├── run.go           — inject secrets + exec command
│   ├── request.go       — create access request
│   └── status.go        — check request approval status
├── internal/
│   ├── config/
│   │   └── config.go    — read/write ~/.valt/config.toml
│   ├── keychain/
│   │   └── keychain.go  — OS keychain token storage
│   └── api/
│       └── client.go    — HTTP client to Valt REST API
├── go.mod
├── go.sum
└── .goreleaser.yml
```

## Key Dependencies

```go
require (
    github.com/spf13/cobra      v1.8.0
    github.com/zalando/go-keyring v0.2.4
    github.com/BurntSushi/toml  v1.3.2
)
```

## Phases

- [Phase 01](phase-01-scaffold-config-keychain-client.md) — Scaffold + config + keychain + API client
- [Phase 02](phase-02-setup-mcp-install.md) — `valt setup` + `valt mcp install`
- [Phase 03](phase-03-get-list-run.md) — `valt get` + `valt list` + `valt run`
- [Phase 04](phase-04-request-status.md) — `valt request` + `valt status`
- [Phase 05](phase-05-build-pipeline.md) — Build pipeline + install script
