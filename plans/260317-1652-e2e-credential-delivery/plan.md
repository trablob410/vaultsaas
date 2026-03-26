---
title: "E2E Credential Delivery"
description: "Wire server-side encryption and credential value delivery so secrets are usable end-to-end"
status: completed
priority: P0
effort: 4h
branch: master
tags: [vault, encryption, mcp, workflow, dashboard]
created: 2026-03-17
completed: 2026-03-24
---

# E2E Credential Delivery

## Problem

Dashboard sends plaintext `value` but handler expects base64 `encrypted_blob`/`encrypted_dek` -- value silently dropped. `GetCredential` returns only session metadata, no secret value. MCP `get_credential` never delivers actual credentials to AI agents. Approvals table shows truncated UUID instead of secret name, raw minutes for duration.

## Architecture Decision

Server-side symmetric encryption with master KEK (AES-256-GCM). Envelope encryption: random DEK encrypts value, master key wraps DEK. Same pattern as AWS Secrets Manager / HashiCorp Vault. MCP server does NOT decrypt -- server delivers plaintext over authenticated TLS channel.

## Phases

| # | Phase | Effort | Status | File |
|---|-------|--------|--------|------|
| 1 | Server Crypto Layer | 30min | ✅ completed | [phase-01](./phase-01-server-crypto-layer.md) |
| 2 | Vault Create: Encrypt + Store | 45min | ✅ completed | [phase-02](./phase-02-vault-create-encrypt.md) |
| 3 | Credential Delivery: Decrypt + Return | 45min | ✅ completed | [phase-03](./phase-03-credential-delivery.md) |
| 4 | MCP get_credential | 30min | ✅ completed | [phase-04](./phase-04-mcp-get-credential.md) |
| 5 | Dashboard UI Fixes | 30min | ✅ completed | [phase-05](./phase-05-dashboard-ui-fixes.md) |

## Dependencies

- Phase 2 depends on Phase 1 (crypto functions)
- Phase 3 depends on Phase 1 (decrypt functions) + Phase 2 (stored blobs)
- Phase 4 depends on Phase 3 (API returns `value`)
- Phase 5 independent, can run in parallel

## Key Files

- `server/pkg/crypto/storage-key.go` -- existing, add `aes.go` sibling
- `server/internal/config/config.go` -- add `VaultMasterKey`
- `server/internal/vault/handler.go` -- accept `value`, encrypt on create
- `server/internal/vault/service.go` -- add `GetSecretByID` (no owner check)
- `server/internal/workflow/handler.go` -- `GetCredential` fetches+decrypts blob
- `server/internal/workflow/credential.go` -- add `Value` to response struct
- `server/cmd/server/main.go` -- wire master key through
- `mcp-server/src/tools.rs` -- return `value` from `get_credential`
- `mcp-server/src/client.rs` -- no struct changes needed (uses serde_json::Value)
- `dashboard/src/components/approvals/approval-list.tsx` -- duration + secret name
- `dashboard/src/types/api.ts` -- add `secret_name` to `AccessRequest`
- `server/internal/workflow/service.go` -- JOIN secret name in `ListPending`
