package handlers

import (
	"database/sql"
	"net/http"

	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

func HandlePublish(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		if user.Plan == "free" {
			writeError(w, http.StatusForbidden, "PLAN_REQUIRED",
				"Publishing contexts requires a Team plan")
			return
		}

		var req PublishRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body")
			return
		}
		if len(req.ContextIDs) == 0 {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "context_ids must not be empty")
			return
		}

		if err := dbpkg.PublishContexts(database, user.ID, req.ContextIDs); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to publish contexts")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success":   true,
			"published": req.ContextIDs,
		})
	}
}
