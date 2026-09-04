package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartPartitionManager launches a background goroutine that ensures
// audit_logs partitions exist for the current month and the next 2 months.
// It runs immediately on start, then every 24 hours until ctx is cancelled.
func StartPartitionManager(ctx context.Context, pool *pgxpool.Pool) {
	go func() {
		EnsurePartitions(ctx, pool)

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Partition manager stopped")
				return
			case <-ticker.C:
				EnsurePartitions(ctx, pool)
			}
		}
	}()
}

func EnsurePartitions(ctx context.Context, pool *pgxpool.Pool) {
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		t := now.AddDate(0, i, 0)
		if err := createPartition(ctx, pool, t); err != nil {
			log.Printf("Failed to create partition for %s: %v", t.Format("2006-01"), err)
		}
	}
}

func createPartition(ctx context.Context, pool *pgxpool.Pool, t time.Time) error {
	name := fmt.Sprintf("audit_logs_%d_%02d", t.Year(), t.Month())
	from := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)

	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs
		FOR VALUES FROM ('%s') TO ('%s')`,
		name,
		from.Format("2006-01-02"),
		to.Format("2006-01-02"),
	)

	_, err := pool.Exec(ctx, query)
	return err
}
