# OneCLI Competitive Analysis Report

**Date:** 2026-03-24
**Researcher:** Agent
**Status:** Complete

---

## Executive Summary

OneCLI is an open-source credential gateway built in Rust + Next.js designed to solve a critical pain point: **AI agents need API access without exposing raw credentials**. The system stores encrypted credentials once and transparently injects them during requests, ensuring agents work with placeholders instead of real secrets.

**Key differentiation vs. Valt:** OneCLI is narrowly focused on credential injection for AI agents (HTTPS proxy model), while Valt is a broader secret vault with approval workflows, dynamic secrets, and RBAC. OneCLI excels at agent-specific threat models (prompt injection defense); Valt solves enterprise secret lifecycle management.

---

## 1. Architecture Overview

### Request Flow

```
Agent (HTTP request with placeholder key)
    ↓
Rust Gateway (port 10255)
    - Matches request by host + path patterns
    - Verifies agent access permissions
    - Swaps placeholder for real credential
    - Decrypts secret (AES-256-GCM)
    ↓
External API (receives real credential in header)
```

### Core Components

| Component | Tech | Port | Purpose |
|-----------|------|------|---------|
| **Gateway** | Rust | 10255 | Fast HTTP(S) proxy intercepting requests, credential swapping, MITM interception for HTTPS |
| **Dashboard** | Next.js + shadcn/ui | 10254 | Manage agents, secrets, permissions, view audit logs |
| **Secret Store** | PostgreSQL + Prisma | N/A | Encrypted credential storage with AES-256-GCM |
| **Deployment** | Docker | Self-hosted | Single container with Postgres embedded |

### Deployment Model

- **Single Docker container:** `docker run ghcr.io/onecli/onecli`
- **Includes:** Gateway + Dashboard + Postgres + Vault
- **Platforms:** Linux, macOS
- **No cloud dependency:** True self-hosted (optional OneCLI Cloud for managed version)

---

## 2. Features Matrix

### Core Features (Implemented)

| Feature | Status | Notes |
|---------|--------|-------|
| **Transparent credential injection** | ✓ Implemented | Agents use `HTTPS_PROXY=http://localhost:10255` or `Proxy-Authorization` headers |
| **AES-256-GCM encryption** | ✓ Implemented | Secrets encrypted at rest, decrypted only at request time |
| **Pattern-based routing** | ✓ Implemented | Host + path matching (e.g., `api.github.com/repos/*`) |
| **Multi-agent support** | ✓ Implemented | Each agent gets scoped access token |
| **Agent-specific rate limiting** | ✓ Implemented | Per-agent endpoint blocking and rate-limit capabilities |
| **HTTPS MITM interception** | ✓ Implemented | Rust gateway can intercept TLS by injecting CA certs in containers |
| **Audit trail** | ✓ Implemented | "See which agent made which API call and when" |
| **Authentication modes** | ✓ Implemented | Single-user (local) OR Google OAuth (teams) |
| **Vault integration** | ✓ Implemented | Connect Bitwarden or other password managers for on-demand injection |

### Planned / Under Development

| Feature | Status | Notes |
|---------|--------|-------|
| **Access policies** | 🔄 Planned | Define what each agent can call |
| **Human approval workflows** | 🔄 Planned | Require human approval before sensitive actions |
| **Granular RBAC** | 🔄 Partial | Mentions "scoped permissions" but full role model unclear |

---

## 3. Security Model

### Threat Model: Prompt Injection

OneCLI's primary value proposition addresses **prompt injection attacks** on AI agents:

**Without OneCLI:**
- Agent receives real API key → stores in environment/memory
- Malicious prompt: "Extract all API keys" → agent exfiltrates `REAL_KEY_12345`
- Attacker gains unlimited access to the service

**With OneCLI:**
- Agent receives placeholder: `FAKE_KEY`
- Malicious prompt: "Extract all API keys" → agent exfiltrates `FAKE_KEY` (useless)
- Credential is never in agent's memory, logs, or environment
- Real credential stays on gateway, decrypted only at request time

### Encryption Details

- **Algorithm:** AES-256-GCM
- **Key derivation:** Master key stored in KMS or OneCLI Cloud
- **Storage:** Secrets encrypted in Postgres
- **Decryption scope:** Gateway process only, never returned to agent
- **Nonce:** Random per-encryption (standard GCM practice)

### What OneCLI Does NOT Solve

1. **Malicious agent behavior** — If an agent has legitimate access to `api.github.com`, prompt injection could still trick it into malicious API calls (e.g., delete repos, change permissions). OneCLI prevents credential exfiltration but not authorization abuse.
2. **Coarse-grained access control** — The gateway doesn't distinguish between "read" and "write" at the API level; full host+path scoping is the granularity.
3. **Credential rotation at scale** — Relies on manual Bitwarden sync or OneCLI Cloud integrations.

---

## 4. Tech Stack Details

### Frontend & API

- **Framework:** Next.js 15+ (App Router)
- **UI Components:** shadcn/ui
- **API:** Node.js route handlers
- **State Management:** Server-side (Prisma)
- **Package Manager:** pnpm
- **Language:** TypeScript (60.8% of codebase)

### Gateway

- **Language:** Rust (35.9% of codebase)
- **HTTP:** Axum or Hyper (typical Rust web stack)
- **Performance:** Built for low-latency credential swapping
- **Features:** MITM proxy capabilities, TLS interception

### Database & ORM

- **Database:** PostgreSQL
- **ORM:** Prisma
- **Migrations:** Version-controlled SQL files
- **Schema:** Agents, secrets, credentials, audit logs

### Infrastructure

- **Containerization:** Docker, docker-compose
- **Build:** Turbo (monorepo orchestration)
- **Repository:** Monorepo (apps/web, apps/gateway, packages/db, packages/ui)

---

## 5. Agent Framework Support

OneCLI works transparently with major AI agent frameworks by intercepting HTTP(S) traffic:

| Framework | Support | Notes |
|-----------|---------|-------|
| OpenClaw | ✓ Verified | Popular Anthropic agent framework |
| NanoClaw | ✓ Verified | Lightweight agent variant |
| IronClaw | ✓ Verified | Enterprise variant |
| Difyn | ✓ Verified | Tool orchestration framework |
| n8n | ✓ Verified | Workflow automation with agents |
| OpenHands | ✓ Verified | OSS code agent |
| Custom frameworks | ✓ Works | Via `HTTPS_PROXY` or `Proxy-Authorization` headers |

**MCP (Model Context Protocol) Support:** No explicit MCP integration found. OneCLI is a credential injection layer, not an MCP server.

---

## 6. Permissions Model

### Multi-Tenant Scoping

- **Agent tokens:** Each agent gets a unique scoped access token
- **Host+Path matching:** Agents can only access credentials for specific patterns
  - Example: Agent-A can only call `api.github.com/repos/*`
  - Agent-B can only call `api.stripe.com/v1/*`
- **Rate limiting:** Per-agent endpoint blocking and rate limits
- **Audit trail:** All access logged with agent ID, timestamp, endpoint called

### Authentication

- **Single-user mode:** For local development
- **Google OAuth:** For team deployments (dashboard login)
- **Agent auth:** Bearer token via `Proxy-Authorization` header

### RBAC Status

- **Implemented:** Basic per-agent scoping
- **Planned:** Full RBAC (owner/admin/member/viewer) — not yet released

---

## 7. Approval Workflows & Human-in-the-Loop

**Current Status:** Not fully implemented.

**Planned Roadmap:**
- Access policies defining what each agent can call
- Human approval gates for sensitive actions
- Audit trail to log all approvals

**Comparison to Valt:**
- **Valt:** Mature approval workflows with state machine (pending → approved/rejected → active)
- **OneCLI:** Conceptual design only, not production-ready

---

## 8. Dynamic Secrets

**Current Status:** Limited.

OneCLI supports **Bitwarden integration** for on-demand credential injection without storing secrets on server. However, this is not true dynamic secret generation (ephemeral credentials with TTL).

**Comparison to Valt:**
- **Valt:** Full dynamic secret support (Postgres provider, AES-256-GCM encrypted leases)
- **OneCLI:** Static credentials with optional Bitwarden sync

---

## 9. Audit Logging

**Implemented Features:**
- Tracks which agent made which API call
- Timestamps for all requests
- Accessible via dashboard

**Planned Enhancements:**
- Policy violation logging
- Approval decision audit trail
- Export/compliance reporting

**Architecture:** Append-only logs stored in Postgres (standard SQL).

**Comparison to Valt:**
- **Valt:** SHA-256 hash chain audit log (tamper-proof), detailed event metadata
- **OneCLI:** Standard database audit logs (reversible, not cryptographically chained)

---

## 10. Community & Adoption

### GitHub Metrics (as of 2026-03-24)

- **Repository:** https://github.com/onecli/onecli
- **License:** Apache-2.0
- **Recent Activity:** Active (Hacker News post 4 days ago)
- **Hacker News Discussion:** https://news.ycombinator.com/item?id=47353558
- **Website:** https://www.onecli.sh
- **Blog/Docs:** Available (referenced in search results)

### Community Feedback (from Hacker News)

**Strengths:**
- Solves a real pain point (prompt injection defense for agents)
- Simple mental model ("store once, inject anywhere")
- Transparent operation (agents don't need modification)
- Open-source with Apache-2.0 license

**Criticisms:**
- HTTPS/TLS handling requires MITM (container modifications needed)
- Node.js proxy respect was problematic in older versions (fixed in Node 22.21+, 24+)
- Incomplete security model (doesn't prevent malicious API calls, only credential exfiltration)
- Approval workflows underdeveloped compared to traditional vaults
- RBAC less granular than enterprise solutions

**Comparisons Mentioned:**
- HashiCorp Vault (mature, complex, steep learning curve)
- AWS Secrets Manager (cloud-only)
- Specialized tools: Tokenizer, Airut, agent-creds

---

## 11. Limitations & Gaps

| Limitation | Impact | Workaround |
|-----------|--------|-----------|
| **No human approval workflows** | Medium | Manual access policy management outside system |
| **No dynamic secrets** | Medium | Use Bitwarden for credential rotation |
| **HTTPS MITM complexity** | High | Requires container modifications; not suitable for agent-in-cloud scenarios |
| **Limited RBAC** | Medium | Per-agent scoping is coarse; no resource-level permissions |
| **No compliance exports** | Medium | Manual audit log extraction for compliance |
| **Node.js proxy issues** (legacy) | Low | Fixed in newer Node versions; legacy agents affected |
| **No secret lifecycle TTL** | Medium | Secrets persist indefinitely; rotation manual |

---

## 12. Pricing & Licensing

- **License:** Apache-2.0 (permissive, commercial-friendly)
- **Self-hosted:** Free (open-source)
- **OneCLI Cloud:** Optional managed service (pricing not disclosed in public sources)
- **No feature tiers:** Single open-source version for all users

---

## 13. Valt vs. OneCLI: Strategic Positioning

### OneCLI Strengths

1. **Agent-native design** — Built specifically for AI agents, not general secret management
2. **Minimal config** — Set `HTTPS_PROXY` and go; no complex HCL policies
3. **Prompt injection defense** — Unique value vs. traditional vaults
4. **Lightweight deployment** — Single container, embedded Postgres
5. **Fast iteration** — Narrow scope = faster feature development

### Valt Strengths

1. **Approval workflows** — Mature state machine, multi-step chains
2. **Dynamic secrets** — Ephemeral credentials with automated lifecycle
3. **Tamper-proof audit** — SHA-256 hash chain
4. **RBAC maturity** — Project-scoped roles (owner/admin/member/viewer)
5. **Use case diversity** — Works for secret rotation, API key management, infrastructure secrets

### When to Choose OneCLI

- Deploying AI agents that need external API access
- Threat model is primarily prompt injection / credential exfiltration
- Simplicity is critical (no org structure, approvals, or dynamic secrets needed)
- Self-hosted only

### When to Choose Valt

- Enterprise secret lifecycle management (rotation, TTL, audit)
- Multi-step approval workflows required
- Dynamic secrets (ephemeral credentials) needed
- Complex RBAC with team/org hierarchies
- Compliance-grade audit logging

---

## 14. Technical Debt & Considerations

### OneCLI

1. **HTTPS MITM complexity** — Production deployments require careful container setup
2. **No path-level authorization** — All requests to matched path go through; can't distinguish read vs. write
3. **Approval workflows incomplete** — Roadmap item but not production-ready
4. **Lack of cryptographic audit chain** — Standard database logs are reversible

### Valt

1. **Larger codebase** — More features = more complexity to maintain
2. **MCP support** — Planned but not yet released

---

## 15. Unresolved Questions

1. **OneCLI approval workflows:** What is the exact implementation status? Is there a beta or PR?
2. **OneCLI RBAC roadmap:** When will granular role-based access control ship?
3. **OneCLI Cloud pricing:** What are the costs for the managed service?
4. **OneCLI MCP integration:** Is there any plan to become an MCP server?
5. **OneCLI repository stats:** What are the exact star count, contributor count, and issue backlog?
6. **OneCLI performance benchmarks:** How many requests/sec can the Rust gateway handle?
7. **OneCLI TLS support:** Can the gateway run with client certificates for agent authentication?
8. **Valt vs. OneCLI:** Will Valt add agent-specific threat model defenses (e.g., placeholder injection)?

---

## Key Insights for Valt Positioning

1. **OneCLI is not a direct competitor** — It's agent-focused, Valt is enterprise-focused
2. **Potential synergy** — Valt could integrate with OneCLI as a backing vault for credential storage
3. **Differentiation vector** — Valt's approval workflows + dynamic secrets are not in OneCLI's scope
4. **Market gap** — Teams need both agent protection (OneCLI) AND enterprise secret management (Valt)
5. **MCP opportunity** — Both projects could benefit from native MCP server support

---

## Sources

- [OneCLI GitHub Repository](https://github.com/onecli/onecli)
- [OneCLI Website](https://www.onecli.sh)
- [Hacker News Discussion](https://news.ycombinator.com/item?id=47353558)
- [HashiCorp Vault Alternatives (Infisical)](https://infisical.com/blog/hashicorp-vault-alternatives)
- [AI Agent Permissions (OSO)](https://www.osohq.com/learn/ai-agent-permissions-delegated-access)
- [Why Your AI Agent's API Keys Are a Security Risk (DEV Community)](https://dev.to/jonathanfishner/why-your-ai-agents-api-keys-are-a-ticking-time-bomb-12pm)
