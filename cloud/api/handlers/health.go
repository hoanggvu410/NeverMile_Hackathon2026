package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

func HandleHealth(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil {
			dbStatus = "error"
		}

		status := "ok"
		if dbStatus != "ok" {
			status = "degraded"
		}

		writeJSON(w, http.StatusOK, HealthResponse{
			Status:    status,
			Version:   "1.0.0",
			DB:        dbStatus,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
