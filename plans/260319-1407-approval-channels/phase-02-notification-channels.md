# Phase 02: Notification Channel Settings

**Priority:** P1 — prerequisite for Slack + Telegram phases
**Status:** COMPLETED — 2026-03-19

## Context
Users need to link their Slack/Telegram accounts so the notify service knows where to send alerts. This phase adds the DB table, CRUD API, and dashboard settings UI.

## Architecture

```
user_notification_channels table
  ├── channel_type: 'email' | 'slack' | 'telegram'
  ├── channel_handle: email addr | slack_user_id | telegram_chat_id
  └── verified: bool (Telegram/Slack require verification step)

GET  /api/v1/me/notification-channels     → list user's channels
POST /api/v1/me/notification-channels     → add channel
DELETE /api/v1/me/notification-channels/{id} → remove channel

notify.Service.RouteNotification(ctx, userID, ...) → picks preferred channel
```

## Files

| File | Change |
|------|--------|
| `server/internal/database/migrations/000027_user_notification_channels.up.sql` | Create |
| `server/internal/database/migrations/000027_user_notification_channels.down.sql` | Create |
| `server/internal/notify/channel_store.go` | Create |
| `server/internal/notify/channel_handler.go` | Create |
| `server/internal/notify/service.go` | Modify — add channel lookup + routing |
| `server/cmd/server/main.go` | Modify — register channel routes |
| `dashboard/src/lib/api/notification-channels.ts` | Create |
| `dashboard/src/app/(dashboard)/settings/notifications/page.tsx` | Create |

---

## Task 1: DB migration

- [ ] Write `000027_user_notification_channels.up.sql`:

```sql
CREATE TABLE user_notification_channels (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_type TEXT NOT NULL CHECK (channel_type IN ('slack', 'telegram', 'email')),
    handle       TEXT NOT NULL,   -- slack_user_id | telegram_chat_id | email address
    verified     BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, channel_type)   -- one channel per type per user
);
```

- [ ] Write down migration:
```sql
DROP TABLE IF EXISTS user_notification_channels;
```

- [ ] Run `make migrate-up` — verify clean

- [ ] Commit:
```bash
git add server/internal/database/migrations/000027_*
git commit -m "feat(db): add user_notification_channels table"
```

---

## Task 2: channel_store.go

**Files:**
- Create: `server/internal/notify/channel_store.go`

- [ ] Write store:

```go
package notify

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationChannel represents a user's linked notification channel.
type NotificationChannel struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ChannelType string `json:"channel_type"`
	Handle      string `json:"handle"`
	Verified    bool   `json:"verified"`
}

// ChannelStore manages user notification channel preferences.
type ChannelStore struct {
	pool *pgxpool.Pool
}

func NewChannelStore(pool *pgxpool.Pool) *ChannelStore {
	return &ChannelStore{pool: pool}
}

func (s *ChannelStore) List(ctx context.Context, userID string) ([]NotificationChannel, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, channel_type, handle, verified FROM user_notification_channels WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing channels: %w", err)
	}
	defer rows.Close()
	var out []NotificationChannel
	for rows.Next() {
		var c NotificationChannel
		if err := rows.Scan(&c.ID, &c.UserID, &c.ChannelType, &c.Handle, &c.Verified); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil { out = []NotificationChannel{} }
	return out, nil
}

func (s *ChannelStore) Upsert(ctx context.Context, userID, channelType, handle string) (*NotificationChannel, error) {
	var c NotificationChannel
	err := s.pool.QueryRow(ctx,
		`INSERT INTO user_notification_channels (user_id, channel_type, handle)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, channel_type) DO UPDATE SET handle = EXCLUDED.handle, verified = false
		 RETURNING id, user_id, channel_type, handle, verified`,
		userID, channelType, handle,
	).Scan(&c.ID, &c.UserID, &c.ChannelType, &c.Handle, &c.Verified)
	if err != nil {
		return nil, fmt.Errorf("upserting channel: %w", err)
	}
	return &c, nil
}

func (s *ChannelStore) Delete(ctx context.Context, userID, channelID string) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM user_notification_channels WHERE id = $1 AND user_id = $2`,
		channelID, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}

// GetPreferred returns the user's preferred channel handle for a given type, or "" if none.
func (s *ChannelStore) GetPreferred(ctx context.Context, userID, channelType string) (string, error) {
	var handle string
	err := s.pool.QueryRow(ctx,
		`SELECT handle FROM user_notification_channels WHERE user_id = $1 AND channel_type = $2 AND verified = true`,
		userID, channelType,
	).Scan(&handle)
	if err != nil {
		return "", nil // not found = no preference
	}
	return handle, nil
}
```

- [ ] Run: `cd server && go build ./internal/notify/...`

- [ ] Commit:
```bash
git add server/internal/notify/channel_store.go
git commit -m "feat(notify): add ChannelStore for user notification channel preferences"
```

---

## Task 3: channel_handler.go REST endpoints

**Files:**
- Create: `server/internal/notify/channel_handler.go`

- [ ] Write handler:

```go
package notify

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
	"github.com/valt-dev/valt/server/pkg/validator"
)

type ChannelHandler struct {
	store *ChannelStore
}

func NewChannelHandler(store *ChannelStore) *ChannelHandler {
	return &ChannelHandler{store: store}
}

// List handles GET /me/notification-channels
func (h *ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	channels, err := h.store.List(r.Context(), userID)
	if err != nil {
		apierror.InternalError(w, "failed to list channels")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"channels": channels})
}

type upsertChannelBody struct {
	ChannelType string `json:"channel_type"`
	Handle      string `json:"handle"`
}

// Upsert handles POST /me/notification-channels
func (h *ChannelHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	var body upsertChannelBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apierror.BadRequest(w, "invalid body")
		return
	}
	if body.ChannelType != "slack" && body.ChannelType != "telegram" && body.ChannelType != "email" {
		apierror.BadRequest(w, "channel_type must be slack, telegram, or email")
		return
	}
	if body.Handle == "" {
		apierror.BadRequest(w, "handle required")
		return
	}
	ch, err := h.store.Upsert(r.Context(), userID, body.ChannelType, body.Handle)
	if err != nil {
		apierror.InternalError(w, "failed to save channel")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ch)
}

// Delete handles DELETE /me/notification-channels/{id}
func (h *ChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	channelID := chi.URLParam(r, "id")
	if _, err := validator.ValidateUUID(channelID); err != nil {
		apierror.BadRequest(w, "invalid channel id")
		return
	}
	if err := h.store.Delete(r.Context(), userID, channelID); err != nil {
		apierror.NotFound(w, "channel not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] Register routes in `main.go` (inside authenticated group):
```go
channelStore := notify.NewChannelStore(pool)
channelHandler := notify.NewChannelHandler(channelStore)

r.Get("/me/notification-channels", channelHandler.List)
r.Post("/me/notification-channels", channelHandler.Upsert)
r.Delete("/me/notification-channels/{id}", channelHandler.Delete)
```

- [ ] Run: `cd server && go build ./cmd/server`

- [ ] Run: `make test-unit`

- [ ] Commit:
```bash
git add server/internal/notify/channel_handler.go server/cmd/server/main.go
git commit -m "feat(notify): add notification channel CRUD API endpoints"
```

---

## Task 4: Dashboard — Notifications settings page

**Files:**
- Create: `dashboard/src/lib/api/notification-channels.ts`
- Create: `dashboard/src/app/(dashboard)/settings/notifications/page.tsx`

- [ ] Write `notification-channels.ts` API client:

```typescript
import { apiFetch } from '@/lib/api-client'

export interface NotificationChannel {
  id: string
  channel_type: 'slack' | 'telegram' | 'email'
  handle: string
  verified: boolean
}

export async function listNotificationChannels(): Promise<NotificationChannel[]> {
  const res = await apiFetch('/me/notification-channels')
  return res.channels
}

export async function upsertNotificationChannel(
  channelType: string,
  handle: string
): Promise<NotificationChannel> {
  return apiFetch('/me/notification-channels', {
    method: 'POST',
    body: JSON.stringify({ channel_type: channelType, handle }),
  })
}

export async function deleteNotificationChannel(id: string): Promise<void> {
  await apiFetch(`/me/notification-channels/${id}`, { method: 'DELETE' })
}
```

- [ ] Write `notifications/page.tsx` — settings page with channel list + add/remove UI (shadcn Card + Input + Button)

- [ ] Add "Notifications" link to Settings sidebar nav

- [ ] Run: `cd dashboard && npm run lint`

- [ ] Run: `cd dashboard && npm test`

- [ ] Commit:
```bash
git add dashboard/src/lib/api/notification-channels.ts dashboard/src/app/(dashboard)/settings/notifications/
git commit -m "feat(dashboard): add notification channel settings page"
```

---

## Success Criteria
- User can add Slack user ID, Telegram chat ID, or email to their account
- API returns 409 if same channel_type already set (upsert replaces)
- Dashboard shows linked channels with remove option
