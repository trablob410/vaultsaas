# E2E Test Failure Analysis
**Date:** 2026-03-26
**Production URL:** https://valt.turbo.ai.vn

---

## Summary
3 out of 19 executed tests failed. 5 tests skipped due to cascading failure from test #7.

---

## Failure #1: Project Creation (Test #7)

**Endpoint:** `POST /api/v1/workspaces/{ws_id}/projects`
**HTTP Status:** 400 (Bad Request)
**Expected:** 201 (Created)

### Test Data Used
```json
{
  "name": "Test Project",
  "description": "E2E test"
}
```

### Environment Context
- Workspace ID extracted from: `GET /api/v1/orgs/{id}/workspaces` response
- Auth: Valid Bearer token from successful login
- Request header: `Content-Type: application/json`

### Cascading Impact
- Blocks test #8: POST /api/v1/secrets (no project_id)
- Blocks test #12: POST /api/v1/projects/{id}/agents (no project)
- Blocks test #13: POST /api/v1/agents/{id}/tokens (no agent)
- Total cascading failures: 3

### Root Cause Hypotheses
1. **Schema mismatch** — API requires additional fields (workspace_id, owner_id, etc.)
2. **Workspace validation** — Workspace ID may be invalid or user not a member
3. **Permission issue** — User lacks "member" or "owner" role in workspace
4. **API version** — Recent changes not reflected in production docs
5. **Rate limiting** — Too many requests to same endpoint (unlikely, different params)

### Debug Actions Needed
```bash
# 1. Capture full error response
curl -v -X POST "${BASE_URL}/api/v1/workspaces/${WS_ID}/projects" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","description":"Debug"}'

# 2. Check API handler logs for validation errors
# 3. Verify workspace ownership/membership
# 4. Try with minimal payload vs full payload
```

### Related Code to Review
- File: `server/internal/project/handler.go` (ProjectCreateHandler)
- File: `server/internal/rbac/middleware.go` (permissions check)
- File: `dashboard/src/app/api/proxy/[...path]/route.ts` (request forwarding)

---

## Failure #2: Notification Channel Creation (Test #14)

**Endpoint:** `POST /api/v1/me/notification-channels`
**HTTP Status:** 400 (Bad Request)
**Expected:** 201 (Created)

### Test Data Used
```json
{
  "type": "email",
  "config": {
    "email": "test@example.com"
  }
}
```

### Environment Context
- Auth: Valid Bearer token from authenticated user
- Request header: `Content-Type: application/json`
- This endpoint should not require additional params (user from auth context)

### Impact
- Blocks users from configuring notification preferences
- Affects email notifications for approvals, secrets, etc.
- May cascade to webhook testing if user preferences required

### Root Cause Hypotheses
1. **Config schema** — May require additional fields (channel_name, enabled, etc.)
2. **Email validation** — test@example.com may fail DNS/format validation
3. **Enum mismatch** — "email" may not be valid type (should be "email_channel", "SMTP", etc.)
4. **Missing field** — May require organization_id or project_id context
5. **Validation logic** — Config object may need different structure

### Debug Actions Needed
```bash
# 1. Try alternate schemas
curl -X POST "${BASE_URL}/api/v1/me/notification-channels" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"type":"email","email":"test@example.com"}'

# 2. Try with real email
curl -X POST "${BASE_URL}/api/v1/me/notification-channels" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"type":"email","config":{"email":"testuser@gmail.com"}}'

# 3. Check if POST works with minimal data
curl -X POST "${BASE_URL}/api/v1/me/notification-channels" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"type":"email"}'
```

### Related Code to Review
- File: `server/internal/notify/handler.go` (NotificationChannelCreateHandler)
- File: `server/internal/notify/service.go` (validation logic)
- File: `docs/api-spec.md` or CLAUDE.md (schema documentation)
- Migrations: Check if notification_channels table exists and schema

### Expected vs Actual Schema
| Field | Expected | Actual | Issue |
|-------|----------|--------|-------|
| type | "email" | ? | Unknown valid values |
| config | {email: string} | ? | Unknown structure |
| enabled | - | ? | May be required |
| name | - | ? | May be required |

---

## Failure #3: Resend Verification Email (Test #21)

**Endpoint:** `POST /api/v1/auth/resend-verification`
**HTTP Status:** 401 (Unauthorized)
**Expected:** 200 or 204

### Test Data Used
```json
{
  "email": "testrun-{timestamp}@valt.dev"
}
```

### Environment Context
- Auth: **NO Bearer token** (public endpoint attempt)
- Just registered user 5 requests prior
- User not verified yet (registration doesn't auto-verify)
- Request header: `Content-Type: application/json`

### Impact
- Users cannot resend verification emails if initial email lost
- May block account recovery flows
- Email notification system appears broken for verification

### Root Cause Hypotheses
1. **Auth required** — Endpoint may require authenticated user (contradicts public flow)
2. **Rate limiting** — User hitting rate limit after multiple requests
3. **Email validation** — Server validates email doesn't exist (but we just created it)
4. **Verification state** — User might already be auto-verified in some flows
5. **Missing dependency** — SMTP not configured, causing handler to reject

### Debug Actions Needed
```bash
# 1. With auth token (authenticated user)
curl -X POST "${BASE_URL}/api/v1/auth/resend-verification" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\"}"

# 2. Without auth (public endpoint)
curl -X POST "${BASE_URL}/api/v1/auth/resend-verification" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\"}"

# 3. Different email format
curl -X POST "${BASE_URL}/api/v1/auth/resend-verification" \
  -H "Content-Type: application/json" \
  -d '{"email":"anothertest@example.com"}'

# 4. Check SMTP config
# curl ${BASE_URL}/health | grep -i smtp
```

### Related Code to Review
- File: `server/internal/auth/email_verification.go` (resend logic)
- File: `server/internal/auth/handler.go` (ResendVerificationHandler)
- File: `server/internal/notify/service.go` (email sending)
- Env vars: SMTP_HOST, SMTP_PORT, SMTP_USER
- Check: Is email notification system working at all?

### Why 401 Specifically?
- 401 Unauthorized typically means auth token required
- Not 400 (bad request) — data format likely correct
- Not 403 (forbidden) — not permission issue
- Not 404 (not found) — endpoint exists
- **Suggests:** Endpoint requires authentication

---

## Test Dependency Chain

```
Test 1: Register ✓
  ↓
Test 2: Login ✓ → Get AUTH_TOKEN
  ↓
Test 3: List Orgs ✓ → Get ORG_ID
  ↓
Test 4-6: Org operations ✓ → Get WS_ID
  ↓
Test 7: Create Project ✗ FAIL (400)
  ├─ Test 8: Create Secret - SKIP (no PROJ_ID)
  ├─ Test 12: Create Agent - SKIP (no PROJ_ID)
  └─ Test 13: Agent Token - SKIP (no AGENT_ID)

Test 14: Notification Channel ✗ FAIL (400)

Test 21: Resend Verification ✗ FAIL (401)
```

---

## Recommendations by Severity

### CRITICAL (Fix within 24 hours)
1. **Project creation (Test #7):** Blocks entire secret management workflow
   - Add error logging to capture rejection reason
   - Review schema validation
   - Verify test data matches production expectations

2. **Notification channels (Test #14):** Blocks user preferences
   - Document expected schema
   - Add input validation with clear error messages
   - Test with multiple email formats

### HIGH (Fix within sprint)
3. **Resend verification (Test #21):** Email recovery broken
   - Clarify auth requirements (public vs authenticated)
   - Fix handler to match API contract
   - Add SMTP healthcheck
   - Test email sending pipeline end-to-end

### MEDIUM (Next sprint)
4. Add request/response validation tests in unit suite
5. Add integration tests for these three endpoints
6. Implement error response standardization (all failures should include error code + message)

---

## How to Verify Fixes

After fixing each issue, re-run tests:

```bash
# Individual endpoint test
curl -i -X POST "${BASE_URL}/api/v1/workspaces/${WS_ID}/projects" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"name":"Verify Fix","description":"test"}'

# Expected: 201 with project object in response

# Full E2E suite
bash D:/vaultsaas/run_e2e_tests.sh

# Expected: 19/19 tests pass (or minimal skips)
```

---

## Additional Test Suggestions

1. **Boundary testing:**
   - Max project name length
   - Min/max description length
   - Special characters in names

2. **Error case testing:**
   - Duplicate project names
   - Invalid workspace IDs
   - Cross-org project creation attempts

3. **Notification channel variations:**
   - Slack channels (if supported)
   - Telegram (if supported)
   - Webhook targets
   - Invalid email formats

4. **Rate limit testing:**
   - Rapid requests to same endpoint
   - Multiple concurrent requests
   - Verify rate limit headers

---

**Report Generated:** 2026-03-26 16:53
**Status:** Analysis complete, awaiting debug actions
