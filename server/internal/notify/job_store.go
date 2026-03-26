package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationJob represents a queued notification delivery.
type NotificationJob struct {
	ID          string          `json:"id"`
	ChannelType string          `json:"channel_type"`
	Recipient   string          `json:"recipient"`
	Subject     string          `json:"subject"`
	PayloadJSON json.RawMessage `json:"payload_json"`
	Status      string          `json:"status"`
	RetryCount  int             `json:"retry_count"`
	MaxRetries  int             `json:"max_retries"`
	NextRetryAt time.Time       `json:"next_retry_at"`
	LastError   string          `json:"last_error"`
}

// JobStore manages notification_jobs table operations.
type JobStore struct {
	pool *pgxpool.Pool
}

// NewJobStore creates a new JobStore.
func NewJobStore(pool *pgxpool.Pool) *JobStore {
	return &JobStore{pool: pool}
}

// Enqueue inserts a new notification job.
func (s *JobStore) Enqueue(ctx context.Context, channelType, recipient, subject string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO notification_jobs (channel_type, recipient, subject, payload_json)
		 VALUES ($1, $2, $3, $4)`,
		channelType, recipient, subject, payloadJSON,
	)
	return err
}

// FetchReady returns up to `limit` jobs ready for processing, locking them.
func (s *JobStore) FetchReady(ctx context.Context, limit int) ([]NotificationJob, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, channel_type, recipient, subject, payload_json, status, retry_count, max_retries, next_retry_at, last_error
		 FROM notification_jobs
		 WHERE status IN ('pending', 'retrying') AND next_retry_at <= NOW()
		 ORDER BY next_retry_at ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching ready jobs: %w", err)
	}
	defer rows.Close()

	var jobs []NotificationJob
	for rows.Next() {
		var j NotificationJob
		if err := rows.Scan(&j.ID, &j.ChannelType, &j.Recipient, &j.Subject,
			&j.PayloadJSON, &j.Status, &j.RetryCount, &j.MaxRetries, &j.NextRetryAt, &j.LastError); err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// MarkSent marks a job as successfully delivered.
func (s *JobStore) MarkSent(ctx context.Context, jobID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE notification_jobs SET status = 'sent', updated_at = NOW() WHERE id = $1`,
		jobID,
	)
	return err
}

// MarkRetry schedules a retry with exponential backoff (5^retryCount minutes).
func (s *JobStore) MarkRetry(ctx context.Context, jobID, lastError string, retryCount int) error {
	backoffMin := math.Pow(5, float64(retryCount))
	_, err := s.pool.Exec(ctx,
		`UPDATE notification_jobs
		 SET status = 'retrying', retry_count = $1, last_error = $2,
		     next_retry_at = NOW() + ($3 || ' minutes')::INTERVAL, updated_at = NOW()
		 WHERE id = $4`,
		retryCount, lastError, fmt.Sprintf("%.0f", backoffMin), jobID,
	)
	return err
}

// MarkExhausted marks a job as permanently failed.
func (s *JobStore) MarkExhausted(ctx context.Context, jobID, lastError string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE notification_jobs SET status = 'exhausted', last_error = $1, updated_at = NOW() WHERE id = $2`,
		lastError, jobID,
	)
	return err
}

// CleanOld removes sent jobs older than the given number of days. Returns count deleted.
func (s *JobStore) CleanOld(ctx context.Context, olderThanDays int) (int, error) {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM notification_jobs WHERE status = 'sent' AND created_at < NOW() - ($1 || ' days')::INTERVAL`,
		fmt.Sprintf("%d", olderThanDays),
	)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
