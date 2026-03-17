package rbac

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valt-dev/valt/server/internal/auth"
	"github.com/valt-dev/valt/server/pkg/apierror"
)

// Middleware returns an HTTP middleware that checks if the authenticated user
// has the required permission for the project in the URL parameter.
// projectParam is the chi URL param name (e.g. "project_id").
func Middleware(db *pgxpool.Pool, projectParam string, resource Resource, action Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := auth.UserIDFromContext(r.Context())
			if userID == "" {
				apierror.Unauthorized(w, "authentication required")
				return
			}

			projectID := chi.URLParam(r, projectParam)
			if projectID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Look up user's role in this project
			var role string
			err := db.QueryRow(r.Context(),
				`SELECT role FROM project_memberships WHERE project_id = $1 AND user_id = $2`,
				projectID, userID,
			).Scan(&role)
			if err != nil {
				// Not a member — deny
				apierror.Forbidden(w, "access denied")
				return
			}

			if !Can(RoleFromProjectMembership(role), resource, action) {
				apierror.Forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
