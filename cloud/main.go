package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	cloudapi "github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/api"
	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	githubapp "github.com/hoanggvu410/NeverMile_Hackathon2026/github-app"
)

func main() {
	databaseURL := mustEnv("DATABASE_URL")
	port := getEnv("PORT", "8080")
	jwtSecret := mustEnv("JWT_SECRET")
	githubClientID := getEnv("GITHUB_CLIENT_ID", "")
	githubClientSecret := getEnv("GITHUB_CLIENT_SECRET", "")

	database, err := dbpkg.Connect(databaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	migrationsDir := getEnv("MIGRATIONS_DIR", "db/migrations")
	if err := dbpkg.Migrate(database, migrationsDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied")

	// GitHub App is optional — PR comment endpoint returns 503 when not configured.
	var app *githubapp.App
	githubAppIDStr := getEnv("GITHUB_APP_ID", "")
	githubPrivateKeyPath := getEnv("GITHUB_APP_PRIVATE_KEY_PATH", "")
	githubWebhookSecret := getEnv("GITHUB_WEBHOOK_SECRET", "")
	if githubAppIDStr != "" && githubPrivateKeyPath != "" {
		appID, parseErr := strconv.ParseInt(githubAppIDStr, 10, 64)
		if parseErr != nil {
			log.Fatalf("invalid GITHUB_APP_ID: %v", parseErr)
		}
		app, err = githubapp.NewApp(appID, githubPrivateKeyPath, githubWebhookSecret)
		if err != nil {
			log.Printf("warning: github app init failed (%v) — POST /v1/pr/comment will be unavailable", err)
			app = nil
		}
	} else {
		log.Println("GITHUB_APP_ID/GITHUB_APP_PRIVATE_KEY_PATH not set — PR comment feature disabled")
	}

	router := cloudapi.NewRouter(database, app, jwtSecret, githubClientID, githubClientSecret)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
