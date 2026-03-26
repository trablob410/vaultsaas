# Phase Implementation Report

## Executed Phase
- Phase: fix-3-production-bugs (ad-hoc, no plan dir)
- Plan: none
- Status: completed

## Files Modified
- `server/internal/workflow/handler.go` — Bug 1 fix (~15 lines changed in CreateRequest)
- `server/internal/vault/service.go` — Bug 2 fix (~20 lines changed in UpdateSecret + new nullableBytes helper)

## Tasks Completed
- [x] Bug 1: Fixed empty UUID insert for user access requests
- [x] Bug 2: Fixed UpdateSecret 500 caused by unconditional DEK nullification
- [x] Bug 3: Verified auth middleware — no fix required (false positive)

## Bug Details

### Bug 1 — workflow/handler.go: CreateRequest empty UUID
**Root cause:** `requestAgentID := body.AIAgentID` allowed a user to submit a non-empty `ai_agent_id` string in the JSON body even when `RequesterType` was `human`. The service's guard `if input.AIAgentID != ""` would then set a non-nil `*string`, and Postgres would attempt to insert it as a UUID, failing with "invalid input syntax for type uuid" if the value wasn't a valid UUID.

**Fix:** Replaced the body-passthrough logic with explicit rules:
- If `agentID` is in context (authenticated agent): use it.
- If `requester_type == "ai_agent"` without a context agent: require and use `body.AIAgentID`, return 400 if empty.
- Otherwise (`human`): `requestAgentID` stays `""` → service stores `nil`.

### Bug 2 — vault/service.go: UpdateSecret nulls encrypted_dek
**Root cause:** The UPDATE query used `encrypted_dek = $3` unconditionally. When a caller sends only `name`/`description` (no new value), both `EncryptedBlob` and `EncryptedDEK` are nil/empty. Passing `nil` as `$3` sets the column to NULL. Any subsequent credential delivery attempt then finds `len(secret.EncryptedDEK) == 0` and returns an empty value — or the DB throws a NOT NULL constraint violation, producing a 500.

**Fix:** Changed query to `encrypted_dek = COALESCE($3, encrypted_dek)` and introduced `nullableBytes()` helper that converts empty `[]byte` to `nil` interface, so Postgres sees SQL NULL and COALESCE preserves the existing column value. Blob in MinIO is only overwritten when `len(input.EncryptedBlob) > 0` (already correct, unchanged).

### Bug 3 — auth/middleware.go: false positive
**Verdict:** No issue. `AuthMiddleware` correctly blocks unauthenticated requests with 401. All write routes are in the `auth.AuthMiddleware` group; dual-auth routes use `dualAuthMiddleware` which also requires a valid Bearer token. Public endpoints (`/action-tokens/*/redeem`, `/webhooks/*`, `/auth/*`) are intentionally unauthenticated. Tester confusion likely stemmed from these public routes returning 200/400 without auth.

## Tests Status
- Type check: pass (`go build ./...` — zero errors, zero output)
- Unit tests: not run (no DB available in this environment)
- Integration tests: not run

## Issues Encountered
None.

## Next Steps
- Run integration tests against a live DB to confirm the COALESCE fix doesn't regress existing update flows.
- Consider adding a unit test for `UpdateSecret` with nil DEK input to guard against regression.
