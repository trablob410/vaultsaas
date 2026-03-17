package audit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
	"github.com/valt-dev/valt/server/pkg/validator"
)

// Handler serves audit log query endpoints.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new audit Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// Routes returns the audit router.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/logs", h.queryLogs)
	return r
}

type auditLogResponse struct {
	Logs  []Entry `json:"logs"`
	Total int     `json:"total"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
}

func (h *Handler) queryLogs(w http.ResponseWriter, r *http.Request) {
	_ = auth.UserIDFromContext(r.Context())

	pg, err := validator.ValidatePagination(
		r.URL.Query().Get("page"),
		r.URL.Query().Get("limit"),
	)
	if err != nil {
		apierror.BadRequest(w, err.Error())
		return
	}

	eventType := r.URL.Query().Get("event_type")
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	query := `SELECT id, event_time, user_id, action, resource_type, resource_id,
	                 event_type, status, ip_address::TEXT, user_agent, region_code, metadata, hash_prev
	          FROM audit_logs WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, eventType)
		argIdx++
	}
	if startStr != "" {
		start, parseErr := time.Parse(time.RFC3339, startStr)
		if parseErr != nil {
			apierror.BadRequest(w, "invalid start date format, use RFC3339")
			return
		}
		query += fmt.Sprintf(" AND event_time >= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND event_time >= $%d", argIdx)
		args = append(args, start)
		argIdx++
	}
	if endStr != "" {
		end, parseErr := time.Parse(time.RFC3339, endStr)
		if parseErr != nil {
			apierror.BadRequest(w, "invalid end date format, use RFC3339")
			return
		}
		query += fmt.Sprintf(" AND event_time <= $%d", argIdx)
		countQuery += fmt.Sprintf(" AND event_time <= $%d", argIdx)
		args = append(args, end)
		argIdx++
	}

	var total int
	if err := h.pool.QueryRow(r.Context(), countQuery, args...).Scan(&total); err != nil {
		log.Printf("Failed to count audit logs: %v", err)
		apierror.InternalError(w, "failed to query audit logs")
		return
	}

	query += fmt.Sprintf(" ORDER BY event_time DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pg.Limit, pg.Offset)

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		log.Printf("Failed to query audit logs: %v", err)
		apierror.InternalError(w, "failed to query audit logs")
		return
	}
	defer rows.Close()

	var logs []Entry
	for rows.Next() {
		var e Entry
		var userID, resourceID, ipAddr, ua, regionCode, hashPrev *string
		if err := rows.Scan(&e.ID, &e.EventTime, &userID, &e.Action, &e.ResourceType, &resourceID,
			&e.EventType, &e.Status, &ipAddr, &ua, &regionCode, &e.Metadata, &hashPrev); err != nil {
			log.Printf("Failed to scan audit log: %v", err)
			apierror.InternalError(w, "failed to read audit logs")
			return
		}
		if userID != nil {
			e.UserID = *userID
		}
		if resourceID != nil {
			e.ResourceID = *resourceID
		}
		if ipAddr != nil {
			e.IPAddress = *ipAddr
		}
		if ua != nil {
			e.UserAgent = *ua
		}
		if regionCode != nil {
			e.RegionCode = *regionCode
		}
		if hashPrev != nil {
			e.HashPrev = *hashPrev
		}
		logs = append(logs, e)
	}
	if logs == nil {
		logs = []Entry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(auditLogResponse{
		Logs:  logs,
		Total: total,
		Page:  pg.Page,
		Limit: pg.Limit,
	})
}
