package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	githubapp "github.com/hoanggvu410/NeverMile_Hackathon2026/github-app"
	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
)

func HandlePostPR(database *sql.DB, app *githubapp.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PRCommentRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body")
			return
		}
		if req.ContextLocalID == "" || req.Repo == "" || req.PRNumber <= 0 {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
				"context_local_id, repo, and pr_number are required")
			return
		}

		if app == nil {
			writeError(w, http.StatusServiceUnavailable, "GITHUB_APP_NOT_CONFIGURED",
				"GitHub App is not configured on this server")
			return
		}

		row, err := dbpkg.GetContextByLocalID(database, req.ContextLocalID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch context")
			return
		}
		if row == nil {
			writeError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Context not found")
			return
		}

		body := githubapp.FormatPRComment(row)

		commentURL, err := app.PostPRComment(r.Context(), req.Repo, req.PRNumber, body)
		if err != nil {
			if errors.Is(err, githubapp.ErrAppNotInstalled) {
				writeError(w, http.StatusForbidden, "PR_GITHUB_APP_NOT_INSTALLED",
					"gitwhy-bot is not installed on this repository")
				return
			}
			writeError(w, http.StatusBadGateway, "PR_COMMENT_FAILED",
				"Failed to post GitHub PR comment")
			return
		}

		if err := dbpkg.RecordPRComment(database, row.ID, req.Repo, req.PRNumber, commentURL); err != nil {
			// Non-fatal: comment was posted, just logging to stderr is sufficient.
			_ = err
		}

		writeJSON(w, http.StatusCreated, PRCommentResponse{CommentURL: commentURL})
	}
}
