package api

import (
	"database/sql"
	"net/http"
	"time"

	githubapp "github.com/hoanggvu410/NeverMile_Hackathon2026/github-app"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/api/handlers"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

func chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

func NewRouter(database *sql.DB, app *githubapp.App, jwtSecret, clientID, clientSecret string) http.Handler {
	mux := http.NewServeMux()

	requireKey := middleware.RequireAPIKey(database)
	requireJWT := middleware.RequireJWT(database, jwtSecret)

	// ── Auth ──────────────────────────────────────────────────────────────────
	mux.Handle("GET /v1/auth/github",
		handlers.HandleGitHubRedirect(clientID))

	mux.Handle("GET /v1/auth/github/callback",
		handlers.HandleGitHubCallback(database, jwtSecret, clientID, clientSecret))

	mux.Handle("GET /v1/auth/me",
		chain(handlers.HandleMe(), requireJWT))

	mux.Handle("POST /v1/auth/api-key",
		chain(handlers.HandleCreateAPIKey(database),
			requireJWT,
			middleware.RateLimit(10, time.Hour)))

	mux.Handle("GET /v1/auth/api-keys",
		chain(handlers.HandleListAPIKeys(database), requireJWT))

	mux.Handle("DELETE /v1/auth/api-keys/{id}",
		chain(handlers.HandleRevokeAPIKey(database), requireJWT))

	// ── Contexts ──────────────────────────────────────────────────────────────
	mux.Handle("POST /v1/contexts/sync",
		chain(handlers.HandleSync(database),
			requireKey,
			middleware.RateLimit(100, time.Minute)))

	mux.Handle("GET /v1/contexts/search",
		chain(handlers.HandleSearch(database),
			requireKey,
			middleware.RateLimit(200, time.Minute)))

	mux.Handle("POST /v1/contexts/publish",
		chain(handlers.HandlePublish(database), requireKey))

	mux.Handle("GET /v1/contexts/{id}",
		chain(handlers.HandleGetContext(database), requireKey))

	// ── PR bot ────────────────────────────────────────────────────────────────
	mux.Handle("POST /v1/pr/comment",
		chain(handlers.HandlePostPR(database, app),
			requireKey,
			middleware.RateLimit(60, time.Minute)))

	// ── Health ────────────────────────────────────────────────────────────────
	mux.Handle("GET /health", handlers.HandleHealth(database))

	return mux
}
