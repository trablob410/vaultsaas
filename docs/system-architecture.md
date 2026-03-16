# System Architecture

## Overview

```
Client Layer:  MCP Server (Rust, stdio) + Dashboard (Next.js)
    ↓ HTTPS
Backend:       Go Monolith (chi/v5 router)
               - Auth, Vault, Workflow, Audit, Notify modules
    ↓
Data Layer:    PostgreSQL 16 (metadata + audit) + MinIO (encrypted blobs)
    ↑
Proxy:         Caddy (reverse proxy + auto TLS)
```

## Components

| Component | Tech | Purpose |
|-----------|------|---------|
| server/ | Go 1.22+ | REST API monolith |
| dashboard/ | Next.js 15 | Web UI (SSR, client-side crypto) |
| mcp-server/ | Rust | Local MCP server for AI agents |
| PostgreSQL | v16 | Metadata, audit logs (partitioned) |
| MinIO | Latest | S3-compatible encrypted blob storage |
| Caddy | v2 | Reverse proxy, auto TLS |

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Go monolith over microservices | Simpler deploy/debug for MVP |
| Caddy over Kong | Auto TLS, simple config |
| PostgreSQL for audit logs | Partitioned tables sufficient for MVP |
| JWT RS256 over Keycloak | Fewer infra dependencies |
| State machine over Temporal | Approval workflow is simple enough |
| Docker Compose over K8s | Single-region MVP |

## Security Architecture
- Zero-knowledge: server never sees plaintext
- Envelope encryption: secret → DEK → user master key
- Master key derived client-side from password (Argon2id)
- JWT RS256 (15min access, 7day refresh)
- Audit hash chain (SHA-256)
