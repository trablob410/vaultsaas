---
phase: "3.4"
title: "Notification Reliability"
priority: P1
status: pending
effort: 5h
---

# Phase 3.4: Notification Reliability

## Context Links
- [Notify service](../../server/internal/notify/service.go) — current fire-and-forget dispatch
- [Email sender](../../server/internal/notify/email.go) — SMTP delivery
- [Slack adapter](../../server/internal/notify/slack.go) — Slack API calls
- [Telegram adapter](../../server/internal/notify/telegram.go) — Telegram API calls

## Overview

Current notification dispatch is fire-and-forget — failures are logged but not retried. Add a `notification_jobs` queue table with background worker that retries failed deliveries with exponential backoff (max 3 retries).

## Key Insights
- Existing `NotifyApprovalNeeded` calls email/Slack/Telegram inline, ignoring errors
- Simple DB-backed queue is sufficient — no need for Redis/RabbitMQ (KISS)
- 30s poll interval balances latency vs DB load
- Exponential backoff: 1min, 5min, 25min

## Requirements

### Functional
- `notification_jobs` table: queues pending notifications
- Insert job on notification dispatch instead of direct send
- Background goroutine polls every 30s for pending/retryable jobs
- Retry up to 3 times with exponential backoff
- Mark jobs as `sent`, `failed`, or `exhausted`
- Job cleanup: delete sent jobs older than 7 days (background)

### Non-Functional
- Worker must be safe for single-instance (SELECT FOR UPDATE SKIP LOCKED)
- Graceful shutdown: stop worker on context cancellation
- No new dependencies

## Architecture

```
NotifyApprovalNeeded()
  → INSERT INTO notification_jobs (channel_type, recipient, payload_json, status='pending')
  → return immediately

Background worker (every 30s):
  → SELECT ... WHERE status IN ('pending','retrying') AND next_retry_at <= NOW() FOR UPDATE SKIP LOCKED LIMIT 10
  → For each job: attempt delivery
    → Success: UPDATE status='sent'
    → Failure: IF retry_count < 3: UPDATE status='retrying', retry_count++, next_retry_at = NOW() + backoff
               ELSE: UPDATE status='exhausted'
```

## Related Code Files

### Create
- `server/internal/notify/job_store.go` — notification_jobs CRUD
- `server/internal/notify/worker.go` — background poll + deliver
- `server/internal/database/migrations/000033_notification_jobs.up.sql`
- `server/internal/database/migrations/000033_notification_jobs.down.sql`

### Modify
- `server/internal/notify/service.go` — enqueue instead of direct send
- `server/cmd/server/main.go` — start worker

## Implementation Steps

### 1. DB Migration (000033)

```sql
-- 000033_notification_jobs.up.sql
CREATE TABLE notification_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_type  VARCHAR(20) NOT NULL, -- 'email', 'slack', 'telegram'
    recipient     TEXT NOT NULL,         -- email addr, slack user ID, telegram chat ID
    subject       TEXT DEFAULT '',
    payload_json  JSONB NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, retrying, sent, exhausted
    retry_count   INTEGER NOT NULL DEFAULT 0,
    max_retries   INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error    TEXT DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notification_jobs_pending ON notification_jobs(status, next_retry_at)
  WHERE status IN ('pending', 'retrying');
```

### 2. notify/job_store.go (~70 lines)

```go
type NotificationJob struct {
    ID           string
    ChannelType  string
    Recipient    string
    Subject      string
    PayloadJSON  json.RawMessage
    Status       string
    RetryCount   int
    MaxRetries   int
    NextRetryAt  time.Time
    LastError    string
}

type JobStore struct { pool *pgxpool.Pool }

func NewJobStore(pool) *JobStore
func (s *JobStore) Enqueue(ctx, channelType, recipient, subject string, payload interface{}) error
func (s *JobStore) FetchReady(ctx, limit int) ([]NotificationJob, error)
  // SELECT ... WHERE status IN ('pending','retrying') AND next_retry_at <= NOW()
  // FOR UPDATE SKIP LOCKED LIMIT $1
func (s *JobStore) MarkSent(ctx, jobID string) error
func (s *JobStore) MarkRetry(ctx, jobID, lastError string, retryCount int) error
  // Backoff: next_retry_at = NOW() + (5^retryCount) minutes → 1min, 5min, 25min
func (s *JobStore) MarkExhausted(ctx, jobID, lastError string) error
func (s *JobStore) CleanOld(ctx, olderThanDays int) (int, error)
  // DELETE WHERE status='sent' AND created_at < NOW() - interval
```

### 3. notify/worker.go (~80 lines)

```go
type Worker struct {
    jobStore  *JobStore
    email     Notifier
    slack     *SlackAdapter
    telegram  *TelegramAdapter
    interval  time.Duration
}

func NewWorker(jobStore, email, slack, telegram) *Worker
func (w *Worker) Start(ctx context.Context)
```

`Start` logic:
```go
ticker := time.NewTicker(w.interval)
defer ticker.Stop()
for {
    select {
    case <-ctx.Done(): return
    case <-ticker.C: w.processBatch(ctx)
    }
}
```

`processBatch`:
1. `FetchReady(ctx, 10)`
2. For each job, attempt delivery based on `channel_type`:
   - `email`: `w.email.Send(ctx, recipient, subject, payloadBody)`
   - `slack`: decode payload JSON, call `w.slack.SendApprovalRequest(ctx, token, ...)`
   - `telegram`: decode payload JSON, call `w.telegram.SendApprovalRequest(ctx, chatID, ...)`
3. On success: `MarkSent`
4. On failure: if `retryCount < maxRetries` → `MarkRetry`, else `MarkExhausted`

Run cleanup once per hour (separate ticker or piggyback on batch): `CleanOld(ctx, 7)`.

### 4. Update notify/service.go

Replace direct send calls with job enqueue:

```go
// Before:
if s.email != nil && to != "" {
    s.email.Send(ctx, to, subject, body)
}

// After:
if s.jobStore != nil && to != "" {
    s.jobStore.Enqueue(ctx, "email", to, subject, map[string]string{
        "body": body,
    })
}
```

Same pattern for Slack and Telegram — enqueue with channel-specific payload.

If `jobStore` is nil (not configured), fall back to direct send (backward compatible).

### 5. Wire in main.go

```go
jobStore := notify.NewJobStore(pool)
notifyWorker := notify.NewWorker(jobStore, emailNotifier, slackAdapter, telegramAdapter)
go notifyWorker.Start(ctx) // respects ctx cancellation for graceful shutdown

// Pass jobStore to notify.NewService(...)
```

## Todo

- [ ] Create migration 000033
- [ ] Implement notify/job_store.go
- [ ] Implement notify/worker.go
- [ ] Update notify/service.go to enqueue jobs
- [ ] Wire worker in main.go
- [ ] Unit tests: enqueue, fetch, retry backoff, mark transitions
- [ ] Integration test: email delivery → mark sent

## Success Criteria
- Notifications are queued instead of sent inline
- Failed deliveries retry up to 3 times with exponential backoff
- Worker polls every 30s, processes batch of 10
- Graceful shutdown stops worker
- Old sent jobs cleaned up after 7 days
- Backward compatible: direct send if jobStore not available

## Security Considerations
- Job payload may contain secret names and requester info — stored in DB (same security as access_requests)
- No sensitive credentials in payload (tokens resolved at send time, not stored in job)
- SKIP LOCKED prevents double-processing in future multi-instance scenarios

## Risk Assessment
- **30s latency**: acceptable for notifications; users won't notice vs instant
- **DB load**: partial index + SKIP LOCKED keeps queries fast; 10 jobs/batch is light
- **Worker crash**: jobs remain in DB, picked up on restart (durable by design)
