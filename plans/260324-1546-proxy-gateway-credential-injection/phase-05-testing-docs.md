---
phase: 5
title: "Testing + Documentation"
status: pending
priority: P2
effort: 4h
---

# Phase 5: Testing + Documentation

## Overview

Unit tests for gateway package, integration test for full proxy flow, and documentation updates.

## Unit Tests

### `server/internal/gateway/`

| Test File | Tests |
|-----------|-------|
| `injector_test.go` | `TestMatchPath` — glob matching edge cases |
| | `TestInjectCredential_Bearer` — header replacement |
| | `TestInjectCredential_Header` — custom header injection |
| | `TestInjectCredential_Query` — query param injection |
| `store_test.go` | `TestFindMatchingRoute` — host+path lookup |
| | `TestFindByPlaceholder` — placeholder key lookup |
| | `TestCreateRoute_AutoPlaceholder` — auto-generates placeholder |
| `server_test.go` | `TestHandleRequest_ValidAgent` — full proxy flow |
| | `TestHandleRequest_InvalidToken` — 407 response |
| | `TestHandleRequest_NoMatch` — transparent forwarding |
| | `TestHandleRequest_PlaceholderSwap` — placeholder → real key |
| `endpoint_limiter_test.go` | `TestEndpointLimit_Blocked` — 403 response |
| | `TestEndpointLimit_RateExceeded` — 429 response |

### Test Helpers

```go
// Use httptest.NewServer as fake "target API" that echoes headers
func newEchoServer() *httptest.Server {
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{
            "auth_header": r.Header.Get("Authorization"),
            "x_api_key":   r.Header.Get("X-Api-Key"),
        })
    }))
}
```

## Integration Test

`server/tests/integration/gateway_proxy_test.go`:

Full flow:
1. Create agent + token
2. Create secret with known value
3. Create proxy route (agent → host → secret)
4. Send request through gateway proxy with placeholder
5. Assert target API received real credential
6. Assert audit log entry created
7. Assert placeholder key not in forwarded request

## Documentation Updates

| File | Change |
|------|--------|
| `docs/system-architecture.md` | Add gateway component to architecture diagram |
| `docs/code-standards.md` | No changes needed |
| `README.md` | Add gateway to Quick Start + Architecture table |
| `CLAUDE.md` | Add `gateway` to internal packages list |

### README Addition

```markdown
## Proxy Gateway (Optional)

For non-MCP agents, Valt provides an HTTP proxy gateway:

1. Create agent + proxy routes in dashboard
2. Configure agent: `HTTPS_PROXY=http://valt:10256`
3. Use placeholder keys instead of real API keys
4. Valt injects real credentials at request time

Agent never sees real keys — prompt injection safe.
```

## Success Criteria
- [ ] Unit tests pass: `go test ./internal/gateway/...`
- [ ] Integration test passes with real DB
- [ ] All existing tests still pass
- [ ] Architecture docs updated with gateway component
- [ ] README includes gateway quick start
