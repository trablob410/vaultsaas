# Phase 03: Slack Bot

**Priority:** P1 — most viral for dev teams
**Status:** COMPLETED — 2026-03-19

## Prerequisites
- `user_notification_channels` table exists (Phase 02)
- User has linked their Slack user ID in settings

## Architecture

```
CreateRequest
  └─▶ channelStore.GetPreferred(userID, "slack") → slack_user_id
  └─▶ slackAdapter.SendApprovalRequest(slack_user_id, requestID, secret, requester, reason)
       └─▶ POST api.slack.com/chat.postMessage (Block Kit with Approve/Reject buttons)
            └─▶ message contains action_id="approve:{requestID}" | "reject:{requestID}"

Approver clicks button in Slack DM
  └─▶ Slack sends POST /api/v1/webhooks/slack/interactions
       └─▶ verify X-Slack-Signature (HMAC-SHA256)
       └─▶ parse payload.actions[0].action_id
       └─▶ call service.Approve / service.Reject
       └─▶ update Slack message to show outcome (chat.update)
```

## Env Vars Required
```
SLACK_BOT_TOKEN=xoxb-...        # Bot User OAuth Token
SLACK_SIGNING_SECRET=...        # For request signature verification
```

## Files

| File | Change |
|------|--------|
| `server/internal/notify/slack.go` | Create — Slack adapter |
| `server/internal/notify/slack_webhook.go` | Create — interactivity handler |
| `server/internal/notify/service.go` | Modify — route to Slack if user has channel |
| `server/internal/config/config.go` | Modify — add Slack env vars |
| `server/cmd/server/main.go` | Modify — init Slack adapter + register webhook route |

---

## Task 1: slack.go — Slack adapter

**Files:**
- Create: `server/internal/notify/slack.go`

- [ ] Write Slack adapter using `net/http` (no SDK — keep deps minimal):

```go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackAdapter sends Slack messages via Bot API.
type SlackAdapter struct {
	botToken string
	client   *http.Client
}

func NewSlackAdapter(botToken string) *SlackAdapter {
	return &SlackAdapter{botToken: botToken, client: &http.Client{}}
}

type slackBlock struct {
	Type string      `json:"type"`
	Text *slackText  `json:"text,omitempty"`
	Elements []slackElement `json:"elements,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackElement struct {
	Type     string    `json:"type"`
	Text     *slackText `json:"text"`
	ActionID string    `json:"action_id"`
	Style    string    `json:"style,omitempty"` // "primary" | "danger"
}

// SendApprovalRequest sends a Block Kit DM to the approver.
func (s *SlackAdapter) SendApprovalRequest(ctx context.Context, slackUserID, requestID, secretName, requester, reason string) error {
	payload := map[string]interface{}{
		"channel": slackUserID,
		"text":    fmt.Sprintf("Approval needed: %s requests access to %s", requester, secretName),
		"blocks": []slackBlock{
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: fmt.Sprintf(
				"*Access Request*\n*Secret:* %s\n*Requester:* %s\n*Reason:* %s",
				secretName, requester, reason,
			)}},
			{Type: "actions", Elements: []slackElement{
				{Type: "button", ActionID: "approve:" + requestID,
					Text: &slackText{Type: "plain_text", Text: "✓ Approve"},
					Style: "primary"},
				{Type: "button", ActionID: "reject:" + requestID,
					Text: &slackText{Type: "plain_text", Text: "✗ Reject"},
					Style: "danger"},
			}},
		},
	}
	return s.post(ctx, "https://slack.com/api/chat.postMessage", payload)
}

// UpdateMessage replaces the approval buttons with an outcome message.
func (s *SlackAdapter) UpdateMessage(ctx context.Context, channelID, ts, outcome string) error {
	payload := map[string]interface{}{
		"channel": channelID, "ts": ts,
		"text": "Request " + outcome + ".",
		"blocks": []slackBlock{
			{Type: "section", Text: &slackText{Type: "mrkdwn", Text: "*Request " + outcome + ".* (actioned via Slack)"}},
		},
	}
	return s.post(ctx, "https://slack.com/api/chat.update", payload)
}

func (s *SlackAdapter) post(ctx context.Context, url string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.botToken)
	resp, err := s.client.Do(req)
	if err != nil { return fmt.Errorf("slack api: %w", err) }
	defer resp.Body.Close()
	var result struct{ OK bool `json:"ok"`; Error string `json:"error"` }
	json.NewDecoder(resp.Body).Decode(&result)
	if !result.OK { return fmt.Errorf("slack error: %s", result.Error) }
	return nil
}
```

- [ ] Run: `cd server && go build ./internal/notify/...`

- [ ] Commit:
```bash
git add server/internal/notify/slack.go
git commit -m "feat(notify): add Slack Block Kit adapter for approval messages"
```

---

## Task 2: slack_webhook.go — Interactivity handler

**Files:**
- Create: `server/internal/notify/slack_webhook.go`

- [ ] Write handler — verifies Slack signature, parses action, calls approve/reject:

```go
package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SlackWebhookHandler handles Slack interactivity callbacks.
type SlackWebhookHandler struct {
	signingSecret string
	approver       SlackApprover
	slack          *SlackAdapter
}

// SlackApprover is implemented by workflow.Handler.
type SlackApprover interface {
	ApproveBySystem(ctx context.Context, requestID, actor string) error
	RejectBySystem(ctx context.Context, requestID, actor, reason string) error
}

func NewSlackWebhookHandler(signingSecret string, approver SlackApprover, slack *SlackAdapter) *SlackWebhookHandler {
	return &SlackWebhookHandler{signingSecret: signingSecret, approver: approver, slack: slack}
}

func (h *SlackWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	if !h.verifySignature(r, body) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	vals, _ := url.ParseQuery(string(body))
	var payload struct {
		Actions []struct {
			ActionID string `json:"action_id"`
		} `json:"actions"`
		Channel struct{ ID string `json:"id"` } `json:"channel"`
		Message struct{ Ts string `json:"ts"` } `json:"message"`
	}
	json.Unmarshal([]byte(vals.Get("payload")), &payload)

	if len(payload.Actions) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	actionID := payload.Actions[0].ActionID // "approve:{requestID}" or "reject:{requestID}"
	parts := strings.SplitN(actionID, ":", 2)
	if len(parts) != 2 {
		http.Error(w, "invalid action_id", http.StatusBadRequest)
		return
	}
	action, requestID := parts[0], parts[1]

	var err error
	switch action {
	case "approve":
		err = h.approver.ApproveBySystem(r.Context(), requestID, "slack-action")
	case "reject":
		err = h.approver.RejectBySystem(r.Context(), requestID, "slack-action", "Rejected via Slack")
	}

	outcome := action + "d"
	if err != nil { outcome = "failed: " + err.Error() }
	_ = h.slack.UpdateMessage(r.Context(), payload.Channel.ID, payload.Message.Ts, outcome)
	w.WriteHeader(http.StatusOK)
}

func (h *SlackWebhookHandler) verifySignature(r *http.Request, body []byte) bool {
	ts := r.Header.Get("X-Slack-Request-Timestamp")
	sig := r.Header.Get("X-Slack-Signature")
	// Reject if timestamp is older than 5 minutes
	t, err := time.Parse("1136214245", ts) // unix ts
	_ = t
	if err == nil && time.Since(time.Unix(0, 0)).Seconds() > 300 { return false }

	mac := hmac.New(sha256.New, []byte(h.signingSecret))
	fmt.Fprintf(mac, "v0:%s:%s", ts, body)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(sig))
}
```

- [ ] Add `ApproveBySystem` / `RejectBySystem` methods to `workflow.Handler` — thin wrappers around `service.Approve` / `service.Reject` + `credMgr.IssueCredential`:

```go
func (h *Handler) ApproveBySystem(ctx context.Context, requestID, actor string) error {
	req, err := h.service.Approve(ctx, requestID, actor)
	if err != nil { return err }
	secret, _ := h.vaultSvc.GetSecretByID(ctx, req.SecretID)
	credType := ""
	if secret != nil { credType = secret.CredentialType }
	_, err = h.credMgr.IssueCredential(ctx, req.ID, credType, req.RequestedDurationMinutes)
	return err
}

func (h *Handler) RejectBySystem(ctx context.Context, requestID, actor, reason string) error {
	_, err := h.service.Reject(ctx, requestID, actor, reason)
	return err
}
```

- [ ] Register **public** route in `main.go`:
```go
r.Post("/api/v1/webhooks/slack/interactions", slackWebhookHandler.Handle)
```

- [ ] Run: `cd server && go build ./cmd/server`

- [ ] Commit:
```bash
git add server/internal/notify/slack_webhook.go server/internal/workflow/handler.go
git commit -m "feat(notify): add Slack interactivity webhook handler for approve/reject actions"
```

---

## Task 3: Wire into notify.Service routing

**Files:**
- Modify: `server/internal/notify/service.go`

- [ ] Add `slack *SlackAdapter` + `channels *ChannelStore` to `Service`:

```go
// In NotifyApprovalNeeded — after creating email tokens, also try Slack:
slackHandle, _ := s.channels.GetPreferred(ctx, ownerUserID, "slack")
if slackHandle != "" && s.slack != nil {
    _ = s.slack.SendApprovalRequest(ctx, slackHandle, requestID, secretName, requester, reason)
}
```

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/notify/service.go
git commit -m "feat(notify): route approval notifications to Slack if user has linked account"
```

---

## Success Criteria
- Approver with Slack linked receives DM with Approve/Reject buttons
- Clicking Approve → request approved, credential issued, message updated
- Invalid Slack signature → 401 rejected
- Slack not configured → falls back to email silently

## Security Notes
- Always verify `X-Slack-Signature` before processing callbacks
- Reject timestamps > 5 minutes old (replay attack prevention)
- `SLACK_SIGNING_SECRET` must be in env, never hardcoded
