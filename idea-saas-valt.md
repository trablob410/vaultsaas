# AI VAULT SaaS - Human-in-the-Loop Secret Management

## 1. TỔNG QUAN DỰ ÁN

### 1.1 Vấn Đề

AI Agent (Claude, Cursor, Windsurf, Copilot...) ngày càng cần truy cập secrets (API keys, DB credentials, cloud tokens) để thực thi tác vụ. Hiện tại:

- Developer hardcode secrets hoặc paste trực tiếp vào chat → **rủi ro bảo mật**
- Không có cơ chế approve/reject khi AI muốn dùng secret → **thiếu kiểm soát**
- Secrets không bị giới hạn thời gian hay phạm vi → **vi phạm least privilege**
- Không có audit trail cho AI access → **không compliance được**

### 1.2 Giải Pháp

**Valt** - MCP-native secret vault với human approval workflow:

1. Developer lưu secrets vào Valt (encrypted end-to-end)
2. AI Agent request access qua MCP Protocol
3. Developer nhận notification, approve/reject
4. AI nhận temporary credential (tự hủy sau N phút)
5. Mọi action được audit log

### 1.3 Unique Value Proposition

| Differentiator | Chi tiết |
|----------------|----------|
| **MCP-native** | Tích hợp trực tiếp với Claude, Cursor, Windsurf qua MCP Protocol |
| **Human-in-the-loop** | Approval workflow bắt buộc cho mọi AI access request |
| **Zero-Knowledge** | Server không bao giờ thấy plaintext secrets |
| **Temporary credentials** | Auto-expire, scoped, revocable |
| **Compliance-ready** | Audit trail, data localization, consent management |

### 1.4 Target Customer

- **Primary**: Dev teams (5-50 người) đang dùng AI coding assistants
- **Secondary**: Enterprise security teams cần kiểm soát AI access
- **Geography**: Vietnam & Southeast Asia trước, mở rộng EU/US sau

---

## 2. BUSINESS MODEL

### 2.1 Pricing

| Tier | Giá | Bao gồm |
|------|-----|----------|
| **Free** | $0 | 5 secrets, 1 user, email approval, 7-day audit log |
| **Pro** | $19/user/tháng | Unlimited secrets, team (up to 10), Zalo/Slack notify, 90-day audit |
| **Team** | $39/user/tháng | Unlimited team, RBAC, SSO, 1-year audit, priority support |
| **Enterprise** | Custom | Self-hosted, multi-region, compliance reports, SLA, 2+ year audit |

### 2.2 Revenue Streams

- SaaS subscription (primary)
- Self-hosted license fee (enterprise)
- Compliance consulting & setup service
- Premium support SLA

### 2.3 Go-to-Market Strategy

1. Open-source Local MCP Server → build trust & adoption
2. Free tier → developer-led growth
3. Content marketing: blog posts về AI security, MCP tutorials
4. Community: Discord cho developers
5. Direct sales cho enterprise (VN market)

---

## 3. KIẾN TRÚC HỆ THỐNG

### 3.1 MVP Architecture (Phase 1)

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌──────────────┐            ┌──────────────┐           │
│  │ Local MCP    │            │ Web Dashboard│           │
│  │ Server       │            │ (Next.js)    │           │
│  │ (Rust)       │            │              │           │
│  └──────┬───────┘            └──────┬───────┘           │
│         └────────────┬──────────────┘                    │
│                      │ HTTPS (TLS 1.3)                   │
└──────────────────────┼──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    BACKEND (Go Monolith)                  │
│                                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Auth     │ │ Vault    │ │ Workflow │ │ Notify   │   │
│  │ Module   │ │ Module   │ │ Module   │ │ Module   │   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│  ┌──────────┐ ┌──────────┐                               │
│  │ Audit    │ │ API      │                               │
│  │ Module   │ │ Router   │                               │
│  └──────────┘ └──────────┘                               │
│                                                          │
│  Caddy (reverse proxy + auto TLS)                        │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    DATA LAYER                             │
│  ┌──────────────┐         ┌──────────────┐               │
│  │ PostgreSQL   │         │ MinIO        │               │
│  │ (metadata +  │         │ (ciphertext  │               │
│  │  audit logs) │         │  storage)    │               │
│  └──────────────┘         └──────────────┘               │
└──────────────────────────────────────────────────────────┘
```

### 3.2 Quyết Định Kiến Trúc

| Quyết định | Lý do |
|------------|-------|
| **Go monolith** thay vì microservices | Đơn giản, dễ debug, dễ deploy. Tách microservices khi cần scale |
| **Caddy** thay vì Kong | Auto TLS, config đơn giản, đủ cho MVP |
| **PostgreSQL** cho cả audit log | Không cần Elasticsearch ở MVP. Partitioned table đủ hiệu năng |
| **JWT + bcrypt** thay vì Keycloak | Giảm infra complexity. Thêm SSO provider sau |
| **State machine** thay vì Temporal | Approval workflow đơn giản, không cần durable execution engine |
| **Docker Compose** thay vì K8s | Đủ cho single-region MVP, lên K8s khi multi-region |
| **Email notification** trước | Channel đơn giản nhất, thêm Zalo/Slack ở Phase 2 |

### 3.3 Technology Stack

| Layer | Công nghệ | Lý do |
|-------|-----------|-------|
| **Backend** | Go 1.22+ | Performance, crypto stdlib, dễ recruit |
| **Local MCP Server** | Rust | Security, low footprint trên máy user |
| **Web Dashboard** | Next.js 15 + TypeScript | SSR, App Router, shadcn/ui |
| **Database** | PostgreSQL 16 | pgcrypto, partitioning, ACID |
| **Object Storage** | MinIO (single instance) | S3-compatible, self-hostable |
| **Reverse Proxy** | Caddy | Auto TLS, simple config |
| **Containerization** | Docker + Docker Compose | Dev parity, easy deploy |

---

## 4. CẤU TRÚC THƯ MỤC

```
valt/
├── README.md
├── LICENSE (Apache 2.0)
├── SECURITY.md
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.example
├── Makefile
│
├── docs/
│   ├── architecture.md
│   ├── api-spec.md
│   ├── mcp-spec.md
│   ├── security-model.md
│   ├── compliance.md
│   └── deployment.md
│
├── server/                          # Go monolith backend
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go            # App configuration
│   │   ├── auth/
│   │   │   ├── handler.go           # Auth HTTP handlers
│   │   │   ├── jwt.go               # JWT token management
│   │   │   ├── middleware.go         # Auth middleware
│   │   │   └── password.go          # Password hashing (Argon2id)
│   │   ├── vault/
│   │   │   ├── handler.go           # Vault HTTP handlers
│   │   │   ├── service.go           # Vault business logic
│   │   │   ├── encryption.go        # Envelope encryption (AES-256-GCM)
│   │   │   └── storage.go           # MinIO integration
│   │   ├── workflow/
│   │   │   ├── handler.go           # Approval HTTP handlers
│   │   │   ├── service.go           # Approval state machine
│   │   │   └── credential.go        # Temporary credential management
│   │   ├── notify/
│   │   │   ├── service.go           # Notification router
│   │   │   └── email.go             # Email channel (MVP)
│   │   ├── audit/
│   │   │   ├── handler.go           # Audit log query handlers
│   │   │   ├── logger.go            # Audit event logging
│   │   │   └── hash-chain.go        # Hash chain integrity
│   │   ├── middleware/
│   │   │   ├── rate-limit.go        # Rate limiting
│   │   │   ├── cors.go              # CORS config
│   │   │   └── security-headers.go  # Security headers
│   │   └── database/
│   │       ├── postgres.go          # DB connection
│   │       └── migrations/          # SQL migrations
│   └── pkg/
│       ├── crypto/                  # Shared crypto utilities
│       └── validator/               # Input validation
│
├── mcp-server/                      # Rust Local MCP Server
│   ├── Cargo.toml
│   ├── src/
│   │   ├── main.rs
│   │   ├── mcp/
│   │   │   ├── mod.rs
│   │   │   ├── protocol.rs          # MCP protocol implementation
│   │   │   ├── tools.rs             # MCP tools (request_access, etc.)
│   │   │   └── resources.rs         # MCP resources (vault://secrets)
│   │   ├── client/
│   │   │   ├── mod.rs
│   │   │   └── api.rs               # Valt API client
│   │   ├── keychain/
│   │   │   ├── mod.rs
│   │   │   └── secure-storage.rs    # OS keychain integration
│   │   └── config.rs
│   └── Dockerfile
│
├── dashboard/                       # Next.js Web Dashboard
│   ├── package.json
│   ├── next.config.ts
│   ├── Dockerfile
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx             # Landing/login
│   │   │   ├── (auth)/
│   │   │   │   ├── login/
│   │   │   │   └── register/
│   │   │   ├── (dashboard)/
│   │   │   │   ├── secrets/         # Secret management
│   │   │   │   ├── approvals/       # Pending approvals
│   │   │   │   ├── audit/           # Audit log viewer
│   │   │   │   └── settings/        # Account settings
│   │   │   └── api/                 # API routes (BFF)
│   │   ├── components/
│   │   │   ├── ui/                  # shadcn/ui components
│   │   │   ├── secrets/
│   │   │   ├── approvals/
│   │   │   └── audit/
│   │   └── lib/
│   │       ├── api-client.ts        # Backend API client
│   │       ├── crypto.ts            # Client-side encryption
│   │       └── auth.ts              # Auth utilities
│   └── tailwind.config.ts
│
├── tests/
│   ├── unit/
│   ├── integration/
│   ├── e2e/
│   └── security/
│
└── scripts/
    ├── setup-dev.sh
    ├── migrate.sh
    ├── seed.sh
    └── backup.sh
```

---

## 5. API SPECIFICATION

### 5.1 REST API

```yaml
openapi: 3.0.3
info:
  title: Valt API
  version: 1.0.0
  description: Human-in-the-Loop Secret Management for AI Agents

servers:
  - url: https://api.valt.dev

security:
  - BearerAuth: []

paths:
  # ── Auth ──
  /api/v1/auth/register:
    post:
      summary: Register new user
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [email, password, region_code]
              properties:
                email: { type: string, format: email }
                password: { type: string, minLength: 12 }
                region_code: { type: string, enum: [vn, sg, eu, us] }
      responses:
        201: { description: User created }

  /api/v1/auth/login:
    post:
      summary: Login
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [email, password]
              properties:
                email: { type: string }
                password: { type: string }
      responses:
        200:
          description: JWT token pair
          content:
            application/json:
              schema:
                type: object
                properties:
                  access_token: { type: string }
                  refresh_token: { type: string }
                  expires_in: { type: integer }

  /api/v1/auth/refresh:
    post:
      summary: Refresh access token
      responses:
        200: { description: New token pair }

  # ── Secrets ──
  /api/v1/secrets:
    post:
      summary: Create encrypted secret
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [name, encrypted_blob, encrypted_dek]
              properties:
                name: { type: string, maxLength: 255 }
                encrypted_blob: { type: string, description: Client-side encrypted value }
                encrypted_dek: { type: string, description: Encrypted Data Encryption Key }
                metadata: { type: object }
                policy:
                  type: object
                  properties:
                    max_duration_minutes: { type: integer, default: 60 }
                    require_reason: { type: boolean, default: true }
                    allowed_agent_ids: { type: array, items: { type: string } }
      responses:
        201:
          content:
            application/json:
              schema:
                type: object
                properties:
                  id: { type: string, format: uuid }
                  name: { type: string }
                  created_at: { type: string, format: date-time }

    get:
      summary: List secrets (metadata only, no values)
      parameters:
        - name: page
          in: query
          schema: { type: integer, default: 1 }
        - name: limit
          in: query
          schema: { type: integer, default: 20 }
      responses:
        200: { description: Paginated list of secret metadata }

  /api/v1/secrets/{secret_id}:
    get:
      summary: Get secret metadata
      responses:
        200: { description: Secret metadata }
    put:
      summary: Update secret
      responses:
        200: { description: Secret updated }
    delete:
      summary: Soft delete secret
      responses:
        204: { description: Secret deleted }

  # ── Access Requests ──
  /api/v1/secrets/{secret_id}/access-requests:
    post:
      summary: Request access to a secret (triggers approval workflow)
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [reason, requester_type]
              properties:
                reason: { type: string, minLength: 20 }
                requester_type: { type: string, enum: [ai_agent, human] }
                ai_agent_id: { type: string }
                requested_duration_minutes: { type: integer, default: 30, maximum: 480 }
      responses:
        202:
          content:
            application/json:
              schema:
                type: object
                properties:
                  request_id: { type: string, format: uuid }
                  status: { type: string, enum: [pending] }

  /api/v1/access-requests:
    get:
      summary: List access requests (for approval dashboard)
      parameters:
        - name: status
          in: query
          schema: { type: string, enum: [pending, approved, rejected, expired] }
      responses:
        200: { description: List of access requests }

  /api/v1/access-requests/{request_id}/approve:
    post:
      summary: Approve access request
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                duration_minutes: { type: integer }
      responses:
        200: { description: Request approved, credential generated }

  /api/v1/access-requests/{request_id}/reject:
    post:
      summary: Reject access request
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                reason: { type: string }
      responses:
        200: { description: Request rejected }

  # ── Credentials ──
  /api/v1/credentials/{request_id}:
    get:
      summary: Retrieve temporary credential (only after approval)
      responses:
        200:
          content:
            application/json:
              schema:
                type: object
                properties:
                  credential_data: { type: string, description: Encrypted for recipient }
                  expires_at: { type: string, format: date-time }

  /api/v1/credentials/{request_id}/revoke:
    post:
      summary: Revoke credential early
      responses:
        200: { description: Credential revoked }

  # ── Audit ──
  /api/v1/audit/logs:
    get:
      summary: Query audit logs
      parameters:
        - name: start_date
          in: query
          schema: { type: string, format: date }
        - name: end_date
          in: query
          schema: { type: string, format: date }
        - name: event_type
          in: query
          schema: { type: string }
        - name: page
          in: query
          schema: { type: integer, default: 1 }
      responses:
        200: { description: Paginated audit log entries }

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
```

### 5.2 MCP Protocol Specification

```
Server Info:
  Name: valt-mcp-server
  Version: 1.0.0
  Protocol: Model Context Protocol 1.0

Resources:
  vault://secrets          → List available secrets (metadata only)
  vault://requests/{id}    → Status of an access request
  vault://audit/today      → Today's audit log summary

Tools:
  request_secret_access:
    Description: Request access to a secret with reason
    Input:
      secret_id: uuid (required)
      reason: string, min 20 chars (required)
      duration_minutes: integer, default 30 (optional)
    Output:
      request_id: uuid
      status: "pending"
      message: "Approval request sent to owner"

  check_approval_status:
    Description: Check if an access request has been approved
    Input:
      request_id: uuid (required)
    Output:
      status: "pending" | "approved" | "rejected" | "expired"
      credential_available: boolean

  get_credential:
    Description: Retrieve temporary credential (only after approval)
    Input:
      request_id: uuid (required)
    Output:
      credential_type: string
      credential_data: object (decrypted locally)
      expires_at: datetime

  revoke_credential:
    Description: Revoke a previously granted credential
    Input:
      request_id: uuid (required)
    Output:
      revoked: boolean
      revoked_at: datetime

  list_my_secrets:
    Description: List secrets you have access to
    Input: (none)
    Output:
      secrets: array of {id, name, created_at}

Sampling:
  Disabled — no secret data is ever sent to LLM providers
```

---

## 6. DATABASE SCHEMA

```sql
-- ============================================
-- PostgreSQL 16 Schema for Valt
-- ============================================

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    region_code VARCHAR(10) NOT NULL DEFAULT 'vn',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_region ON users(region_code);

-- Secrets metadata (NO plaintext values stored)
CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    storage_key VARCHAR(512) NOT NULL,
    encrypted_dek BYTEA NOT NULL,
    policy JSONB NOT NULL DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_secrets_user ON secrets(user_id);
CREATE INDEX idx_secrets_deleted ON secrets(deleted_at) WHERE deleted_at IS NULL;

-- Access requests
CREATE TABLE access_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id),
    requester_user_id UUID NOT NULL REFERENCES users(id),
    requester_type VARCHAR(20) NOT NULL CHECK (requester_type IN ('ai_agent', 'human')),
    ai_agent_id VARCHAR(255),
    reason TEXT NOT NULL,
    requested_duration_minutes INTEGER NOT NULL DEFAULT 30,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'revoked')),
    decided_by UUID REFERENCES users(id),
    decided_at TIMESTAMPTZ,
    rejection_reason TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_requests_secret ON access_requests(secret_id);
CREATE INDEX idx_requests_status ON access_requests(status);
CREATE INDEX idx_requests_requester ON access_requests(requester_user_id);
CREATE INDEX idx_requests_pending ON access_requests(status, created_at)
    WHERE status = 'pending';

-- Credential sessions
CREATE TABLE credential_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_request_id UUID NOT NULL REFERENCES access_requests(id),
    credential_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'revoked', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    usage_count INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_sessions_request ON credential_sessions(access_request_id);
CREATE INDEX idx_sessions_active ON credential_sessions(status, expires_at)
    WHERE status = 'active';

-- Audit logs (partitioned by month)
CREATE TABLE audit_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    event_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_type VARCHAR(50) NOT NULL,
    user_id UUID,
    resource_type VARCHAR(50),
    resource_id UUID,
    action VARCHAR(50) NOT NULL,
    status VARCHAR(20),
    ip_address INET,
    user_agent TEXT,
    region_code VARCHAR(10) NOT NULL DEFAULT 'vn',
    metadata JSONB DEFAULT '{}',
    hash_prev VARCHAR(64),
    PRIMARY KEY (id, event_time)
) PARTITION BY RANGE (event_time);

-- Create monthly partitions (auto-create via pg_partman or cron)
CREATE TABLE audit_logs_2026_03 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE audit_logs_2026_04 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE INDEX idx_audit_time ON audit_logs(event_time);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_type ON audit_logs(event_type);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);
CREATE INDEX idx_refresh_user ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_hash ON refresh_tokens(token_hash);
```

---

## 7. SECURITY MODEL

### 7.1 Encryption

```
Data At Rest:
  Algorithm:       AES-256-GCM
  Key Derivation:  Argon2id (password → Master Key)
  Key Management:  Envelope Encryption
    - Secret encrypted with DEK (Data Encryption Key)
    - DEK encrypted with User Master Key
    - User Master Key derived on client, never sent to server
    - Server stores: encrypted_blob (MinIO) + encrypted_dek (PostgreSQL)

Data In Transit:
  TLS:             1.3 minimum (enforced by Caddy)
  API Auth:        JWT (access token 15min + refresh token 7 days)

Key Storage:
  Client (MCP Server): OS Keychain (macOS Keychain, Windows DPAPI, Linux Secret Service)
  Server:              Environment variables (MVP), HSM (Phase 3)
  Key Rotation:        Manual (MVP), automatic yearly (Phase 3)

Key Recovery:
  Method:          Recovery key generated at registration (Shamir's Secret Sharing, Phase 2)
  MVP:             User downloads recovery key file, stores securely
```

### 7.2 Authentication & Authorization

```
Authentication:
  Method:          Email + password (Argon2id hash)
  JWT:             RS256, 15-minute access token
  Refresh:         7-day refresh token, stored hashed in DB
  2FA:             TOTP (Phase 2)

Authorization:
  Model:           Owner-based (MVP) → RBAC (Phase 2)
  Rules:
    - User can only access own secrets
    - Approval required for every AI access
    - Credential auto-expires after approved duration

Rate Limiting:
  API:             100 requests/minute per user
  Login:           5 attempts/minute per IP
  Access Request:  10/hour per user
```

### 7.3 Threat Model

| Threat | Mitigation |
|--------|------------|
| Compromised AI Agent sends malicious request | Human approval required, reason must be >20 chars, agent ID logged |
| Compromised MCP Server on user machine | MCP Server only stores auth token (not secrets), secrets decrypted in-memory only |
| Server database breach | Zero-knowledge: server has only ciphertext + encrypted DEK, no master key |
| Insider threat (admin) | Admin cannot decrypt secrets, all actions audit logged with hash chain |
| Replay attack on credential | Credentials have expiry + usage count limit, single-use option available |
| Man-in-the-middle | TLS 1.3 enforced, certificate pinning on MCP Server (Phase 2) |
| Brute force password | Argon2id (slow hash), rate limiting, account lockout after 10 failures |

### 7.4 Security Headers

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 0
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss:; frame-ancestors 'none';
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

---

## 8. COMPLIANCE

### 8.1 MVP Scope (Vietnam - NĐ-13/2023)

| Requirement | Implementation |
|-------------|----------------|
| Data localization | All data stored in Vietnam DC |
| Explicit consent | Consent checkbox at registration, stored in DB |
| Audit log retention | 24 months minimum (partitioned PostgreSQL) |
| Breach notification | Incident response plan + 72hr notification template |
| Privacy policy | Vietnamese language, accessible from dashboard |

### 8.2 Future Compliance (Phase 3+)

| Region | Standard | Key Requirements |
|--------|----------|------------------|
| EU | GDPR | Right to erasure (crypto shredding), DPO, SCCs for cross-border |
| US | CCPA/CPRA | Opt-out mechanism, deletion requests |
| SG | PDPA | DPO, 3-day breach notification |

### 8.3 Audit Log Events

```
auth.login, auth.logout, auth.register, auth.password_change
secret.create, secret.update, secret.delete, secret.list
access.request, access.approve, access.reject, access.expire
credential.create, credential.use, credential.revoke, credential.expire
consent.given, consent.withdrawn
```

---

## 9. CORE WORKFLOWS

### 9.1 AI Secret Access (Primary Flow)

```
AI Agent                MCP Server            Valt API              Owner
   │                       │                     │                    │
   │ request_secret_access │                     │                    │
   │──────────────────────>│                     │                    │
   │                       │ POST /access-request│                    │
   │                       │────────────────────>│                    │
   │                       │                     │ Send notification  │
   │                       │                     │───────────────────>│
   │                       │   202 Accepted      │                    │
   │                       │<────────────────────│                    │
   │   {request_id, pending}                     │                    │
   │<──────────────────────│                     │                    │
   │                       │                     │                    │
   │ check_approval_status │                     │    (reviews)       │
   │──────────────────────>│ GET /access-requests│                    │
   │                       │────────────────────>│                    │
   │  {status: "pending"}  │                     │                    │
   │<──────────────────────│                     │                    │
   │                       │                     │                    │
   │                       │                     │ POST /approve      │
   │                       │                     │<───────────────────│
   │                       │                     │   200 OK           │
   │                       │                     │───────────────────>│
   │                       │                     │                    │
   │ check_approval_status │                     │                    │
   │──────────────────────>│ GET /access-requests│                    │
   │                       │────────────────────>│                    │
   │  {status: "approved"} │                     │                    │
   │<──────────────────────│                     │                    │
   │                       │                     │                    │
   │ get_credential        │                     │                    │
   │──────────────────────>│ GET /credentials    │                    │
   │                       │────────────────────>│                    │
   │                       │ {encrypted_data}    │                    │
   │                       │<────────────────────│                    │
   │  (decrypted locally)  │                     │                    │
   │<──────────────────────│                     │                    │
   │                       │                     │                    │
   │  Uses credential...   │                     │                    │
   │                       │                     │                    │
   │  (auto-expires after N minutes)             │                    │
```

### 9.2 Approval State Machine

```
                ┌──────────┐
                │ PENDING  │
                └────┬─────┘
                     │
              ┌──────┴──────┐
              │             │
         ┌────▼────┐  ┌────▼────┐
         │APPROVED │  │REJECTED │
         └────┬────┘  └─────────┘
              │
         ┌────▼────┐
         │ ACTIVE  │ (credential issued)
         └────┬────┘
              │
       ┌──────┴──────┐
       │             │
  ┌────▼────┐  ┌────▼────┐
  │ EXPIRED │  │ REVOKED │
  └─────────┘  └─────────┘
```

---

## 10. TESTING REQUIREMENTS

```
Unit Tests:
  Coverage: 80% minimum
  Framework: Go testing (server), cargo test (MCP), vitest (dashboard)
  Focus: encryption, validation, state machine, auth

Integration Tests:
  Scope: API endpoints with real PostgreSQL + MinIO
  Tools: testcontainers-go
  Run: every PR

E2E Tests:
  Scope: full approval workflow
  Tools: Playwright (dashboard)
  Scenarios:
    - Register → create secret → AI request → approve → get credential → expire
    - Register → create secret → AI request → reject
    - Credential auto-expire after duration

Security Tests:
  SAST: golangci-lint, cargo clippy
  Dependency scan: trivy
  Input fuzzing: go-fuzz for encryption/parsing
```

---

## 11. DEPLOYMENT (MVP)

### 11.1 Prerequisites

```
Infrastructure:
  - 1x VPS (8 vCPU, 16GB RAM, 200GB SSD) - Vietnam DC
  - Domain: valt.dev (or similar)
  - SMTP credentials (email notification)

Software:
  - Docker 24+
  - Docker Compose 2.20+
  - Git
```

### 11.2 Quick Start (Development)

```bash
git clone https://github.com/your-org/valt.git
cd valt
cp .env.example .env  # edit with your values

# Start all services
docker compose up -d

# Run migrations
docker compose exec server go run cmd/migrate/main.go up

# Seed dev data
docker compose exec server go run cmd/seed/main.go

# Access:
# - API:       https://localhost:8443
# - Dashboard: https://localhost:3000
# - MinIO:     https://localhost:9001

# Run tests
make test
```

### 11.3 Production (Single VPS)

```bash
# On VPS
git clone https://github.com/your-org/valt.git
cd valt
cp .env.example .env.prod  # configure production values

# Deploy with production compose
docker compose -f docker-compose.prod.yml up -d

# Caddy handles TLS automatically via Let's Encrypt
# Configure DNS: api.valt.dev → VPS IP
#                app.valt.dev → VPS IP

# Verify
curl https://api.valt.dev/health
```

---

## 12. ROADMAP

### Phase 0: Validation (2 tuần)
- [ ] Interview 10-20 dev teams đang dùng AI coding tools
- [ ] Confirm pain point: secret management cho AI Agent
- [ ] Validate willingness to pay
- [ ] Finalize feature prioritization

### Phase 1: Lean MVP (6-8 tuần)
- [ ] Go backend monolith (auth, vault, workflow, audit, email notify)
- [ ] Rust Local MCP Server
- [ ] Next.js dashboard (secrets, approvals, audit viewer)
- [ ] Client-side encryption (AES-256-GCM envelope encryption)
- [ ] Email approval notification
- [ ] Deploy single VPS Vietnam
- [ ] Beta launch with 5-10 early adopters

### Phase 2: Product-Market Fit (4-6 tuần)
- [ ] Zalo OA + Slack notification channels
- [ ] VSCode extension
- [ ] Team management (invite, roles)
- [ ] TOTP 2FA
- [ ] Recovery key (Shamir's Secret Sharing)
- [ ] Improved dashboard UX
- [ ] Public launch

### Phase 3: Scale (4-6 tuần)
- [ ] Kubernetes migration
- [ ] Multi-region: Singapore
- [ ] SSO (OIDC provider integration)
- [ ] RBAC policies
- [ ] API key management (beyond secrets)
- [ ] SOC 2 Type 1 preparation

### Phase 4: Enterprise (ongoing)
- [ ] Self-hosted deployment option
- [ ] EU region (Frankfurt) + GDPR compliance
- [ ] HSM integration
- [ ] Compliance reporting dashboard
- [ ] Mobile app (Flutter)
- [ ] ISO 27001 / SOC 2 Type 2

---

## PHỤ LỤC

### A. Glossary

| Term | Definition |
|------|------------|
| DEK | Data Encryption Key - key dùng để encrypt secret data |
| MEK | Master Encryption Key - key dùng để encrypt DEK |
| MCP | Model Context Protocol - protocol cho AI Agent giao tiếp với tools |
| Zero-Knowledge | Kiến trúc server không thể decrypt user data |
| Envelope Encryption | Mã hóa nhiều lớp: data → DEK → MEK |
| WORM | Write Once Read Many - storage cho audit logs |
| Human-in-the-loop | Con người phải approve trước khi AI được truy cập |

### B. References

- [Model Context Protocol Spec](https://modelcontextprotocol.io)
- Nghị định 13/2023/NĐ-CP (Vietnam Data Protection)
- OWASP Top 10 2023
- NIST Cybersecurity Framework
