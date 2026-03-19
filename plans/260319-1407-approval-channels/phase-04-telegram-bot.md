# Phase 04: Telegram Bot

**Priority:** P2
**Status:** COMPLETED — 2026-03-19

## Architecture

```
User sends /start to Telegram bot
  └─▶ POST /api/v1/webhooks/telegram (Telegram Update)
       └─▶ extract chat_id, store in user_notification_channels (via linking token)

CreateRequest
  └─▶ channelStore.GetPreferred(ownerUserID, "telegram") → chat_id
  └─▶ telegramAdapter.SendApprovalRequest(chat_id, requestID, ...)
       └─▶ POST api.telegram.org/bot{TOKEN}/sendMessage with InlineKeyboardMarkup

Approver taps button
  └─▶ Telegram sends callback_query Update to webhook
       └─▶ parse data="approve:{requestID}" | "reject:{requestID}"
       └─▶ call ApproveBySystem / RejectBySystem
       └─▶ answerCallbackQuery + editMessageText (outcome)
```

## Linking Flow (how user connects Telegram account)
1. Dashboard shows bot link: `https://t.me/{BOT_USERNAME}?start={link_token}`
2. Server generates a short-lived link token stored in `telegram_link_tokens` (UUID, user_id, expires 10min)
3. User clicks link → opens Telegram → `/start {link_token}`
4. Webhook receives update → validates token → upserts `user_notification_channels` with chat_id → marks verified

## Env Vars
```
TELEGRAM_BOT_TOKEN=...
TELEGRAM_BOT_USERNAME=valtbot  # for link generation
```

## Files

| File | Change |
|------|--------|
| `server/internal/database/migrations/000028_telegram_link_tokens.up.sql` | Create |
| `server/internal/notify/telegram.go` | Create — Telegram adapter |
| `server/internal/notify/telegram_webhook.go` | Create — Update handler |
| `server/internal/notify/service.go` | Modify — route to Telegram |
| `server/internal/config/config.go` | Modify — Telegram env vars |
| `server/cmd/server/main.go` | Modify — register Telegram webhook route |
| `dashboard/src/app/(dashboard)/settings/notifications/page.tsx` | Modify — add Telegram link button |

---

## Task 1: DB migration for linking tokens

- [ ] Write `000028_telegram_link_tokens.up.sql`:

```sql
CREATE TABLE telegram_link_tokens (
    token    TEXT PRIMARY KEY,
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at  TIMESTAMPTZ
);
```

- [ ] Down: `DROP TABLE IF EXISTS telegram_link_tokens;`

- [ ] Run `make migrate-up`

- [ ] Commit:
```bash
git add server/internal/database/migrations/000028_*
git commit -m "feat(db): add telegram_link_tokens table"
```

---

## Task 2: telegram.go — adapter

**Files:**
- Create: `server/internal/notify/telegram.go`

- [ ] Write adapter (plain HTTP, no SDK):

```go
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramAdapter struct {
	token  string
	client *http.Client
}

func NewTelegramAdapter(token string) *TelegramAdapter {
	return &TelegramAdapter{token: token, client: &http.Client{}}
}

func (t *TelegramAdapter) apiURL(method string) string {
	return fmt.Sprintf("https://api.telegram.org/bot%s/%s", t.token, method)
}

func (t *TelegramAdapter) SendApprovalRequest(ctx context.Context, chatID int64, requestID, secretName, requester, reason string) error {
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text": fmt.Sprintf("*Access Request*\nSecret: %s\nRequester: %s\nReason: %s",
			secretName, requester, reason),
		"parse_mode": "Markdown",
		"reply_markup": map[string]interface{}{
			"inline_keyboard": [][]map[string]string{
				{
					{"text": "✓ Approve", "callback_data": "approve:" + requestID},
					{"text": "✗ Reject",  "callback_data": "reject:" + requestID},
				},
			},
		},
	}
	return t.post(ctx, "sendMessage", payload)
}

func (t *TelegramAdapter) EditMessage(ctx context.Context, chatID int64, messageID int, text string) error {
	payload := map[string]interface{}{
		"chat_id": chatID, "message_id": messageID, "text": text,
	}
	return t.post(ctx, "editMessageText", payload)
}

func (t *TelegramAdapter) AnswerCallbackQuery(ctx context.Context, callbackID, text string) error {
	payload := map[string]interface{}{"callback_query_id": callbackID, "text": text}
	return t.post(ctx, "answerCallbackQuery", payload)
}

func (t *TelegramAdapter) post(ctx context.Context, method string, payload interface{}) error {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", t.apiURL(method), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil { return fmt.Errorf("telegram api: %w", err) }
	defer resp.Body.Close()
	var result struct{ OK bool `json:"ok"`; Description string `json:"description"` }
	json.NewDecoder(resp.Body).Decode(&result)
	if !result.OK { return fmt.Errorf("telegram error: %s", result.Description) }
	return nil
}
```

- [ ] Run: `cd server && go build ./internal/notify/...`

- [ ] Commit:
```bash
git add server/internal/notify/telegram.go
git commit -m "feat(notify): add Telegram bot adapter for approval messages"
```

---

## Task 3: telegram_webhook.go — Update handler

**Files:**
- Create: `server/internal/notify/telegram_webhook.go`

- [ ] Write webhook handler:

```go
package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type TelegramWebhookHandler struct {
	telegram     *TelegramAdapter
	channels     *ChannelStore
	pool         interface{ QueryRow(ctx context.Context, sql string, args ...interface{}) interface{ Scan(dest ...interface{}) error } }
	approver     SlackApprover // reuse same interface — ApproveBySystem/RejectBySystem
	botUsername  string
}

// TelegramUpdate is the minimal Telegram Update payload we care about.
type TelegramUpdate struct {
	Message *struct {
		Chat struct{ ID int64 `json:"id"` } `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
	CallbackQuery *struct {
		ID      string `json:"id"`
		Data    string `json:"data"`
		Message struct {
			MessageID int             `json:"message_id"`
			Chat      struct{ ID int64 `json:"id"` } `json:"chat"`
		} `json:"message"`
		From struct{ ID int64 `json:"id"` } `json:"from"`
	} `json:"callback_query"`
}

func (h *TelegramWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var update TelegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusOK) // always 200 to Telegram
		return
	}

	// /start {link_token} — link user's Telegram account
	if update.Message != nil && strings.HasPrefix(update.Message.Text, "/start ") {
		linkToken := strings.TrimPrefix(update.Message.Text, "/start ")
		chatID := update.Message.Chat.ID
		h.handleLinkToken(r.Context(), linkToken, chatID)
	}

	// Callback from Approve/Reject button
	if update.CallbackQuery != nil {
		cq := update.CallbackQuery
		parts := strings.SplitN(cq.Data, ":", 2)
		if len(parts) == 2 {
			action, requestID := parts[0], parts[1]
			var outcome string
			switch action {
			case "approve":
				if err := h.approver.ApproveBySystem(r.Context(), requestID, "telegram-action"); err != nil {
					outcome = "Failed: " + err.Error()
				} else {
					outcome = "✓ Approved"
				}
			case "reject":
				if err := h.approver.RejectBySystem(r.Context(), requestID, "telegram-action", "Rejected via Telegram"); err != nil {
					outcome = "Failed: " + err.Error()
				} else {
					outcome = "✗ Rejected"
				}
			}
			_ = h.telegram.AnswerCallbackQuery(r.Context(), cq.ID, outcome)
			_ = h.telegram.EditMessage(r.Context(), cq.Message.Chat.ID, cq.Message.MessageID, outcome)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TelegramWebhookHandler) handleLinkToken(ctx context.Context, token string, chatID int64) {
	// Validate token + get user_id from DB, then upsert notification channel
	// Implementation: query telegram_link_tokens WHERE token=$1 AND used_at IS NULL AND expires_at > NOW()
	// then channelStore.Upsert(userID, "telegram", fmt.Sprint(chatID)) + mark token used
}
```

- [ ] Register route in `main.go` (public — Telegram calls without auth):
```go
r.Post("/api/v1/webhooks/telegram", telegramWebhookHandler.Handle)
```

- [ ] Generate link token endpoint (authenticated — user calls from dashboard):
```go
// POST /me/telegram-link → generates token, returns t.me/{BOT_USERNAME}?start={token}
r.Post("/me/telegram-link", channelHandler.GenerateTelegramLink)
```

- [ ] Run: `cd server && go build ./cmd/server`

- [ ] Commit:
```bash
git add server/internal/notify/telegram_webhook.go
git commit -m "feat(notify): add Telegram webhook handler for approve/reject callbacks and account linking"
```

---

## Task 4: Wire into service + dashboard

- [ ] Add `telegram *TelegramAdapter` to `notify.Service`, route in `NotifyApprovalNeeded`:

```go
telegramHandle, _ := s.channels.GetPreferred(ctx, ownerUserID, "telegram")
if telegramHandle != "" && s.telegram != nil {
    var chatID int64
    fmt.Sscan(telegramHandle, &chatID)
    _ = s.telegram.SendApprovalRequest(ctx, chatID, requestID, secretName, requester, reason)
}
```

- [ ] Update settings notifications page: add "Connect Telegram" button that calls `POST /me/telegram-link` and opens the returned URL

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/notify/service.go dashboard/src/app/\(dashboard\)/settings/notifications/page.tsx
git commit -m "feat(notify): route approval notifications to Telegram if user has linked account"
```

---

## Success Criteria
- User can link Telegram by clicking bot link in dashboard settings
- Approver receives Telegram DM with inline Approve/Reject buttons
- Tapping button approves/rejects and updates the message
- Unlinked users fall back to email
