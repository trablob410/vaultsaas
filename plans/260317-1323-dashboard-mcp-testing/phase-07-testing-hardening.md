# Phase 7: Testing & Hardening

## Context Links
- Go backend packages: `server/internal/` (auth, vault, workflow, audit, consent, policy, notify, middleware, config, database)
- Existing interfaces: `vault.Storage` (mockable), `notify.Notifier` (mockable)
- DI pattern: manual constructor injection in `server/cmd/server/main.go`
- Pure-logic packages: `pkg/apierror`, `pkg/validator`, `pkg/crypto`, `internal/policy`, `internal/audit/hash-chain`
- Dashboard: `dashboard/` (Phase 5 output)
- MCP server: `mcp-server/` (Phase 6 output)
- Zero test files exist today

## Overview
- **Priority:** P0
- **Status:** pending
- **Effort:** 8h
- Unit tests (80%+ coverage): Go, Rust, TypeScript
- Integration tests: Go API endpoints with real Postgres + MinIO via testcontainers-go
- E2E: full approval workflow
- Security linting: golangci-lint, cargo clippy, trivy

## Key Insights
- Go services use `*pgxpool.Pool` directly -- no repo interfaces. Handler tests use `httptest` + real chi router
- Pure-logic packages testable without DB: policy engine, hash chain, validators, apierror, crypto
- `vault.Storage` interface enables mock for unit tests (avoid MinIO dependency)
- `notify.Notifier` interface enables mock for notification tests
- Auth middleware testable with generated RSA test keys + JWTManager
- testcontainers-go spins real Postgres+MinIO per test -- slow but reliable
- Dashboard crypto.ts is pure functions -- easy to unit test with vitest
- Rust MCP: protocol parsing and tool dispatch are pure functions, HTTP client mockable

## Requirements

### Functional
- Go unit tests: all packages in `internal/` and `pkg/`
- Go integration tests: full API request/response cycle with real DB
- Rust unit tests: protocol, tools, resources, crypto modules
- TypeScript unit tests: crypto.ts, api-client.ts, component rendering
- E2E test: register -> create secret -> request access -> approve -> get credential -> revoke

### Non-Functional
- 80% minimum line coverage (Go, Rust)
- Tests run in CI without external dependencies (testcontainers provides them)
- No flaky tests (no timing-dependent assertions)
- Security scan: no critical/high vulnerabilities

## Architecture

### Test Organization
```
server/
  internal/
    auth/
      handler_test.go         # httptest + chi, mock JWT keys
      password_test.go        # Argon2id hash/verify roundtrip
      jwt_test.go             # Token generate/validate with test keys
      middleware_test.go       # Auth header parsing, context injection
    vault/
      service_test.go         # Unit: mock Storage interface
      handler_test.go         # httptest + chi
      storage_test.go         # Integration only (needs MinIO)
    workflow/
      service_test.go         # Unit: state transitions
      handler_test.go         # httptest + chi
      credential_test.go      # Unit: issue/expire/revoke logic
    policy/
      engine_test.go          # Pure logic: tier mapping, policy merge
    audit/
      hash-chain_test.go      # Pure logic: chain verification
      logger_test.go          # Integration (needs DB)
    middleware/
      rate-limit_test.go      # Unit: sliding window logic
      security-headers_test.go # Unit: header assertions
    notify/
      service_test.go         # Unit: mock Notifier
    consent/
      service_test.go         # Integration (needs DB)
  pkg/
    apierror/
      apierror_test.go        # Unit: JSON response format
    validator/
      validator_test.go       # Unit: email, password, UUID, pagination
    crypto/
      crypto_test.go          # Unit: storage key generation
  tests/
    integration/
      api_test.go             # testcontainers: full API lifecycle
      e2e_test.go             # testcontainers: full approval workflow

dashboard/
  src/
    lib/
      __tests__/
        crypto.test.ts        # AES-256-GCM roundtrip
        api-client.test.ts    # Fetch mock, error handling
    components/
      __tests__/
        secret-list.test.tsx  # Render with mock data
        approval-list.test.tsx

mcp-server/src/
  protocol_test.rs            # (in protocol.rs as #[cfg(test)])
  tools_test.rs               # (in tools.rs as #[cfg(test)])
  resources_test.rs
  crypto_test.rs
```

### Testcontainers Setup (Go Integration)
```go
// tests/integration/testutil.go
func SetupTestEnv(t *testing.T) (*pgxpool.Pool, vault.Storage, func()) {
    // 1. Start Postgres container (16-alpine)
    // 2. Run migrations via golang-migrate
    // 3. Start MinIO container
    // 4. Return pool, storage, cleanup func
}
```

## Related Code Files

### Files to Create (Go)
- `server/internal/auth/handler_test.go`
- `server/internal/auth/password_test.go`
- `server/internal/auth/jwt_test.go`
- `server/internal/auth/middleware_test.go`
- `server/internal/vault/service_test.go`
- `server/internal/vault/handler_test.go`
- `server/internal/workflow/service_test.go`
- `server/internal/workflow/handler_test.go`
- `server/internal/workflow/credential_test.go`
- `server/internal/policy/engine_test.go`
- `server/internal/audit/hash-chain_test.go`
- `server/internal/middleware/rate-limit_test.go`
- `server/internal/middleware/security-headers_test.go`
- `server/internal/notify/service_test.go`
- `server/pkg/apierror/apierror_test.go`
- `server/pkg/validator/validator_test.go`
- `server/pkg/crypto/crypto_test.go`
- `server/tests/integration/testutil.go`
- `server/tests/integration/api_test.go`
- `server/tests/integration/e2e_test.go`

### Files to Create (Dashboard)
- `dashboard/vitest.config.ts`
- `dashboard/src/lib/__tests__/crypto.test.ts`
- `dashboard/src/lib/__tests__/api-client.test.ts`
- `dashboard/src/components/__tests__/secret-list.test.tsx`
- `dashboard/src/components/__tests__/approval-list.test.tsx`

### Files to Modify
- `server/go.mod` -- add testcontainers-go, testify
- `dashboard/package.json` -- add vitest, @testing-library/react, jsdom, happy-dom
- `Makefile` -- add test targets (test-unit, test-integration, test-e2e)

### New Dependencies
```
# Go
github.com/testcontainers/testcontainers-go v0.34
github.com/testcontainers/testcontainers-go/modules/postgres
github.com/testcontainers/testcontainers-go/modules/minio
github.com/stretchr/testify v1.9

# Dashboard
vitest ^3
@testing-library/react ^16
@testing-library/jest-dom ^6
happy-dom ^15

# Rust (dev-dependencies)
# No new deps -- use built-in #[cfg(test)] + assert macros
```

## Implementation Steps

### Step 1: Go Pure-Logic Unit Tests (1.5h)
Start here -- no DB needed, fast feedback.

1. `pkg/validator/validator_test.go`:
   - Valid/invalid emails, passwords (length, complexity), UUIDs, pagination
   - Edge cases: empty strings, SQL injection strings, unicode

2. `pkg/apierror/apierror_test.go`:
   - Each error function (BadRequest, Unauthorized, etc.) returns correct status + JSON format

3. `pkg/crypto/crypto_test.go`:
   - StorageKey format: `users/{uid}/secrets/{sid}`

4. `internal/policy/engine_test.go`:
   - `DeriveRiskTier`: all 6 known types + unknown default
   - `DefaultPolicyForTier`: verify all fields for each tier
   - `ForCredentialType`: integration of tier+policy
   - `MergePolicy`: user overrides respected, can only tighten (not loosen)

5. `internal/audit/hash-chain_test.go`:
   - Chain of 3 entries: verify each hash links to previous
   - Tamper detection: modify middle entry, verify chain breaks

### Step 2: Go Auth Unit Tests (1h)
1. `internal/auth/password_test.go`:
   - Hash + Verify roundtrip
   - Verify fails with wrong password
   - Hash output contains Argon2id identifier

2. `internal/auth/jwt_test.go`:
   - Generate test RSA keys in TestMain
   - Generate access token -> validate -> get userID
   - Expired token -> validation fails
   - Tampered token -> validation fails
   - Refresh token generation (random, sufficient length)

3. `internal/auth/middleware_test.go`:
   - Missing Authorization header -> 401
   - Invalid format (no Bearer) -> 401
   - Invalid token -> 401
   - Valid token -> context has userID, next handler called

### Step 3: Go Service Unit Tests (1.5h)
1. `internal/vault/service_test.go`:
   - Create mock Storage implementation (in-memory map)
   - Test through service would need DB -- mark as integration
   - Focus on testing handler HTTP layer with httptest

2. `internal/vault/handler_test.go`:
   - Use httptest.NewRecorder + chi.NewRouter
   - Inject real handler with mock pool (or skip, do in integration)
   - Test request validation: missing fields, invalid UUIDs

3. `internal/workflow/service_test.go`:
   - State transition validation (unit testable if we extract validation logic)
   - Policy enforcement: reason length, duration caps

4. `internal/middleware/rate-limit_test.go`:
   - Burst requests -> first N pass, N+1 gets 429
   - Window expires -> requests pass again

5. `internal/middleware/security-headers_test.go`:
   - Apply middleware -> response has X-Content-Type-Options, X-Frame-Options, etc.

6. `internal/notify/service_test.go`:
   - Mock Notifier: verify Send called with correct args
   - Nil notifier (no-op mode): no error, no panic

### Step 4: Go Integration Tests (2h)
1. `server/tests/integration/testutil.go`:
   - `SetupTestEnv(t)` using testcontainers-go:
     - Postgres 16 container with port mapping
     - Run migrations using golang-migrate (embed from `internal/database/`)
     - MinIO container
     - Return pool, MinIOStorage, cleanup func
   - Helper: `CreateTestUser(t, pool)` -- insert user, return ID
   - Helper: `GetTestJWT(t, jwtMgr, userID)` -- generate valid access token

2. `server/tests/integration/api_test.go`:
   - Build full chi router (same as main.go but with test containers)
   - Test: Register -> Login -> Get tokens
   - Test: Create secret -> List -> Get -> Update -> Delete
   - Test: Auth required (no token -> 401)
   - Test: Rate limiting (burst requests -> 429)

3. `server/tests/integration/e2e_test.go`:
   - Full workflow:
     1. Register user
     2. Login, get tokens
     3. Create secret (with encrypted blob)
     4. Create access request
     5. Verify policy enforcement (Tier 1 auto-approves)
     6. Get credential
     7. Revoke credential
     8. Verify audit log entries exist
   - Test Tier 2 (requires manual approval):
     1. Create db_credential secret
     2. Request access -> status=pending
     3. Approve -> status=approved
     4. Get credential -> success
     5. Verify notification attempted (mock SMTP)

### Step 5: Dashboard Tests (1h)
1. Setup:
   - `vitest.config.ts` with happy-dom environment
   - Add test script to `package.json`: `"test": "vitest run"`

2. `src/lib/__tests__/crypto.test.ts`:
   - `generateDEK` returns 256-bit key
   - Encrypt -> decrypt roundtrip with random data
   - `wrapDEK` -> `unwrapDEK` roundtrip
   - Decrypt with wrong key fails
   - Note: uses Web Crypto API -- needs happy-dom or polyfill

3. `src/lib/__tests__/api-client.test.ts`:
   - Mock fetch globally
   - `listSecrets` calls correct URL, returns typed data
   - 401 response triggers redirect/error
   - Network error handled gracefully

4. Component tests (smoke tests):
   - `secret-list.test.tsx`: renders table headers, renders rows from mock data
   - `approval-list.test.tsx`: renders pending items, approve button present

### Step 6: Rust Tests (0.5h)
1. `protocol.rs` `#[cfg(test)]` module:
   - Parse valid JSON-RPC request
   - Parse request with missing method -> error
   - Serialize response matches expected JSON

2. `tools.rs` `#[cfg(test)]`:
   - `list_tools()` returns 5 tools with valid schemas
   - Tool dispatch: unknown tool name -> error
   - Argument validation: missing required field -> error

3. `resources.rs` `#[cfg(test)]`:
   - `list_resources()` returns 3 resources
   - URI parsing: `vault://secrets` -> correct handler
   - Invalid URI -> error

4. `crypto.rs` `#[cfg(test)]`:
   - AES-256-GCM decrypt with known test vector
   - Decrypt with wrong key -> error

### Step 7: Security Scanning + CI (0.5h)
1. Add to Makefile:
   ```makefile
   test-unit:
       cd server && go test ./internal/... ./pkg/... -v -count=1
   test-integration:
       cd server && go test ./tests/integration/... -v -count=1 -timeout 5m
   test-dashboard:
       cd dashboard && npm test
   test-mcp:
       cd mcp-server && cargo test
   test: test-unit test-integration test-dashboard test-mcp
   lint:
       cd server && golangci-lint run ./...
       cd dashboard && npm run lint
       cd mcp-server && cargo clippy -- -D warnings
   security:
       trivy fs --severity HIGH,CRITICAL .
   ```

2. golangci-lint config (`.golangci.yml`):
   - Enable: gosec, govet, errcheck, staticcheck, unused
   - Disable: funlen (we have 200-line rule manually)

3. Verify all pass: `make lint && make test && make security`

## Todo List
- [ ] Go pure-logic tests (validator, apierror, crypto, policy, hash-chain)
- [ ] Go auth tests (password, JWT, middleware)
- [ ] Go service/handler tests (vault, workflow, notify, rate-limit, security-headers)
- [ ] Go integration setup (testcontainers: Postgres + MinIO + migrations)
- [ ] Go integration tests (API lifecycle)
- [ ] Go E2E test (full approval workflow)
- [ ] Dashboard vitest setup
- [ ] Dashboard crypto.ts tests
- [ ] Dashboard api-client tests
- [ ] Dashboard component smoke tests
- [ ] Rust unit tests (protocol, tools, resources, crypto)
- [ ] Makefile test targets
- [ ] golangci-lint config + run
- [ ] trivy security scan
- [ ] 80%+ coverage verified

## Success Criteria
- `make test` passes all suites (Go, Dashboard, Rust)
- `make lint` passes all linters
- `make security` reports no HIGH/CRITICAL vulnerabilities
- Go coverage >= 80% (`go test -coverprofile`)
- E2E test completes full approval workflow with real containers
- No flaky tests (3 consecutive runs pass)

## Risk Assessment
| Risk | Impact | Mitigation |
|------|--------|------------|
| testcontainers slow in CI | Medium | Parallel test packages; reuse containers where possible |
| Docker not available in CI | High | Ensure CI runner has Docker; separate unit vs integration jobs |
| Web Crypto API in vitest | Medium | Use happy-dom which polyfills SubtleCrypto; fallback to @peculiar/webcrypto |
| Go handler tests need real DB | Medium | Focus unit tests on pure logic; integration tests cover DB paths |
| Coverage target hard for handlers | Low | 80% is for pure-logic; handlers covered by integration tests |

## Security Considerations
- Test keys generated at test time, never committed
- Test containers use random ports, no port conflicts
- No production credentials in test code
- trivy scans all Dockerfiles and dependencies
- golangci-lint with gosec catches common Go security issues

## Next Steps
- CI/CD pipeline: GitHub Actions workflow running `make lint test security`
- Performance tests: benchmark crypto operations, API latency under load
- Penetration testing: manual review of auth flow, CSRF, injection vectors
