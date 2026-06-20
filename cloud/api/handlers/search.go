package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

func HandleSearch(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		q := r.URL.Query().Get("q")
		if q == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Missing required query param: q")
			return
		}

		limit := 5
		if ls := r.URL.Query().Get("limit"); ls != "" {
			if n, err := strconv.Atoi(ls); err == nil && n > 0 {
				limit = n
			}
		}

		// Fetch teams the user belongs to so published team contexts are included.
		teamIDs, err := dbpkg.GetTeamIDsByUser(database, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch team membership")
			return
		}

		results, err := dbpkg.SearchContexts(database, user.ID, teamIDs, q, limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Search failed")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"results": results,
			"total":   len(results),
		})
	}
}
