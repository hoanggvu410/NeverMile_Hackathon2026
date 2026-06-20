package handlers

import (
	"database/sql"
	"net/http"
	"time"

	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

func HandleSync(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		var req SyncRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body")
			return
		}
		if len(req.Contexts) == 0 {
			quotaUsed, _ := dbpkg.GetMonthlyUsage(database, user.ID)
			quotaRemaining := -1
			if user.Plan == "free" {
				quotaRemaining = 20 - quotaUsed
				if quotaRemaining < 0 {
					quotaRemaining = 0
				}
			}
			writeJSON(w, http.StatusOK, SyncResponse{
				Synced:         []string{},
				Failed:         []string{},
				QuotaUsed:      quotaUsed,
				QuotaRemaining: quotaRemaining,
			})
			return
		}

		// Collect all local IDs from the request.
		allIDs := make([]string, 0, len(req.Contexts))
		for _, c := range req.Contexts {
			if c.LocalID != "" {
				allIDs = append(allIDs, c.LocalID)
			}
		}

		// Determine which are already in the DB (quota-free re-syncs).
		existing, err := dbpkg.GetExistingLocalIDs(database, user.ID, allIDs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check existing contexts")
			return
		}
		existingSet := make(map[string]bool, len(existing))
		for _, id := range existing {
			existingSet[id] = true
		}
		newCount := 0
		for _, id := range allIDs {
			if !existingSet[id] {
				newCount++
			}
		}

		// Enforce quota for free-tier users.
		if user.Plan == "free" && newCount > 0 {
			if err := dbpkg.CheckAndIncrementQuota(database, user.ID, newCount); err != nil {
				if err == dbpkg.ErrQuotaExceeded {
					writeError(w, http.StatusTooManyRequests, "CONTEXT_QUOTA_EXCEEDED",
						"Free tier sync quota (20/month) exceeded")
					return
				}
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check quota")
				return
			}
		}

		// Convert request items to ContextRow.
		rows := make([]dbpkg.ContextRow, 0, len(req.Contexts))
		var failed []string
		for _, item := range req.Contexts {
			if item.LocalID == "" || item.Prompt == "" {
				failed = append(failed, item.LocalID)
				continue
			}
			var contextTS time.Time
			if item.ContextTS != "" {
				contextTS, _ = time.Parse(time.RFC3339, item.ContextTS)
			}
			if contextTS.IsZero() {
				contextTS = time.Now().UTC()
			}
			files := item.Files
			if files == nil {
				files = []string{}
			}
			commits := item.Commits
			if commits == nil {
				commits = []string{}
			}
			rows = append(rows, dbpkg.ContextRow{
				LocalID:              item.LocalID,
				Prompt:               item.Prompt,
				Reasoning:            item.Reasoning,
				Decisions:            item.Decisions,
				RejectedAlternatives: item.RejectedAlternatives,
				TradeOffs:            item.TradeOffs,
				Files:                files,
				Commits:              commits,
				Domain:               item.Domain,
				Topic:                item.Topic,
				Agent:                item.Agent,
				Model:                item.Model,
				ContextTS:            contextTS,
			})
		}

		synced, err := dbpkg.SyncContexts(database, user.ID, rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sync contexts")
			return
		}

		// Collect IDs that were attempted but not in synced list.
		syncedSet := make(map[string]bool, len(synced))
		for _, id := range synced {
			syncedSet[id] = true
		}
		for _, row := range rows {
			if !syncedSet[row.LocalID] {
				failed = append(failed, row.LocalID)
			}
		}
		if failed == nil {
			failed = []string{}
		}

		// Build quota numbers for response.
		quotaUsed, _ := dbpkg.GetMonthlyUsage(database, user.ID)
		quotaRemaining := -1
		if user.Plan == "free" {
			quotaRemaining = 20 - quotaUsed
			if quotaRemaining < 0 {
				quotaRemaining = 0
			}
		}

		writeJSON(w, http.StatusOK, SyncResponse{
			Synced:         synced,
			Failed:         failed,
			QuotaUsed:      quotaUsed,
			QuotaRemaining: quotaRemaining,
		})
	}
}
