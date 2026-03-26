package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Worker polls notification_jobs and delivers pending notifications.
type Worker struct {
	jobStore *JobStore
	email    Notifier
	slack    *SlackAdapter
	telegram *TelegramAdapter
	zalo     *ZaloAdapter
	interval time.Duration
}

// NewWorker creates a notification delivery Worker.
func NewWorker(jobStore *JobStore, email Notifier, slack *SlackAdapter, telegram *TelegramAdapter, zalo *ZaloAdapter) *Worker {
	return &Worker{
		jobStore: jobStore,
		email:    email,
		slack:    slack,
		telegram: telegram,
		zalo:     zalo,
		interval: 30 * time.Second,
	}
}

// Start runs the worker loop. Blocks until ctx is cancelled.
func (w *Worker) Start(ctx context.Context) {
	log.Println("[notify-worker] started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	cleanTicker := time.NewTicker(1 * time.Hour)
	defer cleanTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[notify-worker] shutting down")
			return
		case <-ticker.C:
			w.processBatch(ctx)
		case <-cleanTicker.C:
			if n, err := w.jobStore.CleanOld(ctx, 7); err == nil && n > 0 {
				log.Printf("[notify-worker] cleaned %d old jobs", n)
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	jobs, err := w.jobStore.FetchReady(ctx, 10)
	if err != nil {
		log.Printf("[notify-worker] fetch error: %v", err)
		return
	}

	for _, job := range jobs {
		deliverErr := w.deliver(ctx, job)
		if deliverErr == nil {
			if err := w.jobStore.MarkSent(ctx, job.ID); err != nil {
				log.Printf("[notify-worker] mark sent error: %v", err)
			}
			continue
		}

		newCount := job.RetryCount + 1
		if newCount >= job.MaxRetries {
			if err := w.jobStore.MarkExhausted(ctx, job.ID, deliverErr.Error()); err != nil {
				log.Printf("[notify-worker] mark exhausted error: %v", err)
			}
			log.Printf("[notify-worker] job %s exhausted after %d retries: %v", job.ID, newCount, deliverErr)
		} else {
			if err := w.jobStore.MarkRetry(ctx, job.ID, deliverErr.Error(), newCount); err != nil {
				log.Printf("[notify-worker] mark retry error: %v", err)
			}
		}
	}
}

func (w *Worker) deliver(ctx context.Context, job NotificationJob) error {
	switch job.ChannelType {
	case "email":
		if w.email == nil {
			return fmt.Errorf("email notifier not configured")
		}
		var payload struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(job.PayloadJSON, &payload); err != nil {
			return fmt.Errorf("unmarshal email payload: %w", err)
		}
		return w.email.Send(ctx, job.Recipient, job.Subject, payload.Body)

	case "slack":
		if w.slack == nil {
			return fmt.Errorf("slack adapter not configured")
		}
		var payload struct {
			RequestID  string `json:"request_id"`
			SecretName string `json:"secret_name"`
			Requester  string `json:"requester"`
			Reason     string `json:"reason"`
		}
		if err := json.Unmarshal(job.PayloadJSON, &payload); err != nil {
			return fmt.Errorf("unmarshal slack payload: %w", err)
		}
		return w.slack.SendApprovalRequest(ctx, job.Recipient, payload.RequestID, payload.SecretName, payload.Requester, payload.Reason)

	case "telegram":
		if w.telegram == nil {
			return fmt.Errorf("telegram adapter not configured")
		}
		var payload struct {
			ChatID     int64  `json:"chat_id"`
			RequestID  string `json:"request_id"`
			SecretName string `json:"secret_name"`
			Requester  string `json:"requester"`
			Reason     string `json:"reason"`
		}
		if err := json.Unmarshal(job.PayloadJSON, &payload); err != nil {
			return fmt.Errorf("unmarshal telegram payload: %w", err)
		}
		return w.telegram.SendApprovalRequest(ctx, payload.ChatID, payload.RequestID, payload.SecretName, payload.Requester, payload.Reason)

	case "zalo":
		if w.zalo == nil {
			return fmt.Errorf("zalo adapter not configured")
		}
		var payload struct {
			ZaloUserID string `json:"zalo_user_id"`
			Message    string `json:"message"`
		}
		if err := json.Unmarshal(job.PayloadJSON, &payload); err != nil {
			return fmt.Errorf("unmarshal zalo payload: %w", err)
		}
		return w.zalo.SendNotification(ctx, payload.ZaloUserID, payload.Message)

	default:
		return fmt.Errorf("unknown channel type: %s", job.ChannelType)
	}
}
