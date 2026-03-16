# Security Model

## Encryption

### Data at Rest
- Algorithm: AES-256-GCM
- Key Derivation: Argon2id (password → master key)
- Envelope Encryption:
  - Secret encrypted with DEK (Data Encryption Key)
  - DEK encrypted with user master key
  - Master key derived client-side, never sent to server
  - Server stores: encrypted_blob (MinIO) + encrypted_dek (PostgreSQL)

### Data in Transit
- TLS 1.3 minimum (Caddy)
- JWT RS256 (15min access + 7day refresh)

### Key Storage
- Client (MCP Server): OS Keychain
- Server: Environment variables (MVP), HSM (future)

## Authentication
- Email + password (Argon2id hash)
- JWT RS256 access tokens (15 min)
- Refresh tokens (7 days, stored hashed in DB)

## Authorization
- Owner-based (MVP): users access only own secrets
- Every AI access requires human approval
- Credentials auto-expire after approved duration

## Rate Limiting
- API: 100 req/min per user
- Login: 5 attempts/min per IP
- Access requests: 10/hour per user

## Audit
- All events logged with hash chain (SHA-256)
- Partitioned by month in PostgreSQL
- Events: auth, secret, access, credential, consent

## Threat Mitigations
| Threat | Mitigation |
|--------|-----------|
| Compromised AI agent | Human approval required |
| Database breach | Zero-knowledge architecture |
| Insider threat | Admin can't decrypt, audit hash chain |
| Replay attack | Credential expiry + usage limits |
| Brute force | Argon2id + rate limiting + lockout |
