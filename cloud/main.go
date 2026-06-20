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
	githubClientID := mustEnv("GITHUB_CLIENT_ID")
	githubClientSecret := mustEnv("GITHUB_CLIENT_SECRET")
	githubAppIDStr := mustEnv("GITHUB_APP_ID")
	githubPrivateKeyPath := mustEnv("GITHUB_APP_PRIVATE_KEY_PATH")
	githubWebhookSecret := mustEnv("GITHUB_WEBHOOK_SECRET")

	appID, err := strconv.ParseInt(githubAppIDStr, 10, 64)
	if err != nil {
		log.Fatalf("invalid GITHUB_APP_ID: %v", err)
	}

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

	app, err := githubapp.NewApp(appID, githubPrivateKeyPath, githubWebhookSecret)
	if err != nil {
		log.Fatalf("init github app: %v", err)
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
