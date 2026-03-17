package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteAuditLog writes an audit log entry with enhanced fields.
func WriteAuditLog(ctx context.Context, pool *pgxpool.Pool, entry AuditEntry) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO audit_logs (user_id, action, resource_type, resource_id, event_type, status, ip_address, user_agent, region_code, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7::inet, $8, $9, $10)`,
		entry.UserID, entry.Action, entry.ResourceType, entry.ResourceID,
		entry.EventType, entry.Status, entry.IPAddress, entry.UserAgent, entry.RegionCode,
		entry.Metadata,
	)
	if err != nil {
		return fmt.Errorf("writing audit log: %w", err)
	}
	return nil
}

// AuditEntry represents a structured audit log entry.
type AuditEntry struct {
	UserID       string `json:"user_id"`
	Action       string `json:"action"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	EventType    string `json:"event_type"` // action, auth, access, system
	Status       string `json:"status"`     // success, failure, denied
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	RegionCode   string `json:"region_code,omitempty"`
	Metadata     string `json:"metadata"`
}

// NewAuditEntry creates an AuditEntry with sensible defaults.
func NewAuditEntry(userID, action, resourceType, resourceID string) AuditEntry {
	return AuditEntry{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventType:    "action",
		Status:       "success",
		Metadata:     "{}",
	}
}
