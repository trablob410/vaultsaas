# Phase 8: Error Message Audit

**Priority:** P2 | **Effort:** 3h | **Status:** pending

Review and improve all user-facing error messages for clarity, consistency, and helpfulness.

## Error Message Categories

### Authentication Errors

- [ ] "Invalid email or password" (don't reveal which is wrong)
- [ ] "Email already registered"
- [ ] "Email verification required" (with link to resend)
- [ ] "Password reset link expired" (with link to request new one)
- [ ] "Session expired, please login again"
- [ ] "Google OAuth failed, please try again"

### Vault Errors

- [ ] "Secret not found" (404)
- [ ] "You do not have permission to view this secret" (403)
- [ ] "Secret name already exists in project"
- [ ] "Invalid secret policy configuration"
- [ ] "Encryption failed, please try again"

### Workflow Errors

- [ ] "Access request already exists for this secret"
- [ ] "Access request expired" (with option to create new)
- [ ] "You do not have permission to approve"
- [ ] "Approver not found or offline"
- [ ] "Request approval chain incomplete"

### Billing Errors

- [ ] "Upgrade required" (with link to upgrade)
- [ ] "Plan limit exceeded" (specify: secrets, projects, agents)
- [ ] "Stripe connection failed" (retry button)
- [ ] "Subscription cancelled" (with option to reactivate)

### CLI Errors

- [ ] "Not authenticated. Run `valt setup` first"
- [ ] "Secret not found in project"
- [ ] "MCP server not responding"
- [ ] "Rate limit exceeded (60 requests/min)"

### General Errors

- [ ] "Something went wrong, please try again" (with error code)
- [ ] "Network timeout, please check your connection"
- [ ] "Server error, our team has been notified"

## Audit Checklist

For each error message:

- [ ] User-friendly (no technical jargon)
- [ ] Actionable (suggests next step or provides link)
- [ ] Consistent tone (professional, not sarcastic)
- [ ] Consistent formatting (title case, punctuation)
- [ ] No sensitive info exposed (no SQL errors, file paths, keys)
- [ ] Internationalization-ready (if supporting non-English)

## Standards

All error messages MUST:

1. Be in title case: "Email Already Registered"
2. Avoid pronouns: "Your account" → "Account"
3. Include action: "[x] try again", "click here to reset"
4. Never expose: database errors, file paths, API keys, user IDs
5. Keep under 2 sentences

## Testing

1. Trigger each error condition systematically
2. Screenshot the error message
3. Note if message is clear and actionable
4. Log any issues with context

## Common Issues to Fix

- Generic "Error" without explanation
- Cryptic error codes without meaning
- No guidance on what to do next
- Inconsistent capitalization/punctuation
- HTML tags visible in messages
- Timeout errors with no retry option

## Sign-Off

All user-facing errors reviewed and improved. Errors follow standard format.
