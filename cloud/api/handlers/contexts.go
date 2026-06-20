package handlers

import (
	"database/sql"
	"net/http"

	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
)

func HandleGetContext(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		localID := r.PathValue("id")
		if localID == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Missing context ID")
			return
		}

		ctx, err := dbpkg.GetContextByLocalID(database, localID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch context")
			return
		}
		if ctx == nil {
			writeError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Context not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"context": ctx,
		})
	}
}
