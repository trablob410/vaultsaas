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

// NewChannelStore creates a ChannelStore.
func NewChannelStore(pool *pgxpool.Pool) *ChannelStore {
	return &ChannelStore{pool: pool}
}

// List returns all notification channels for a user.
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
	if out == nil {
		out = []NotificationChannel{}
	}
	return out, nil
}

// Upsert adds or updates a channel (one per type per user).
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

// Delete removes a channel by ID, scoped to the user.
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

// GetPreferred returns the verified handle for a channel type, or "" if none.
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
