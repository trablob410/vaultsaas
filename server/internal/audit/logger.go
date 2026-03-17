package audit

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry represents a structured audit log entry.
type Entry struct {
	ID           string    `json:"id"`
	EventTime    time.Time `json:"event_time"`
	UserID       string    `json:"user_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	EventType    string    `json:"event_type"`
	Status       string    `json:"status"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	RegionCode   string    `json:"region_code,omitempty"`
	Metadata     string    `json:"metadata"`
	HashPrev     string    `json:"hash_prev,omitempty"`
}

// Logger writes audit log entries to the database with hash chaining.
type Logger struct {
	pool     *pgxpool.Pool
	lastHash string
}

// NewLogger creates a new audit Logger.
func NewLogger(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

// Log writes a single audit entry.
func (l *Logger) Log(ctx context.Context, e Entry) error {
	if e.EventType == "" {
		e.EventType = "action"
	}
	if e.Status == "" {
		e.Status = "success"
	}
	if e.Metadata == "" {
		e.Metadata = "{}"
	}

	hash := ComputeHash(e, l.lastHash)

	var id string
	err := l.pool.QueryRow(ctx,
		`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, event_type, status, ip_address, user_agent, region_code, metadata, hash_prev)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8, $9, $10, $11)
		 RETURNING id`,
		nilIfEmpty(e.UserID), e.Action, e.ResourceType, nilIfEmpty(e.ResourceID),
		e.EventType, e.Status, nilIfEmpty(e.IPAddress), nilIfEmpty(e.UserAgent),
		nilIfEmpty(e.RegionCode), e.Metadata, hash,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("writing audit log: %w", err)
	}

	l.lastHash = hash
	return nil
}

// LogFromRequest creates and writes an audit entry extracting IP/UA from request.
func (l *Logger) LogFromRequest(r *http.Request, userID, action, resourceType, resourceID string) {
	e := Entry{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventType:    "action",
		Status:       "success",
		IPAddress:    extractIP(r),
		UserAgent:    r.UserAgent(),
	}
	if err := l.Log(r.Context(), e); err != nil {
		log.Printf("audit log error: %v", err)
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr
	host := r.RemoteAddr
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
	}
	return host
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
