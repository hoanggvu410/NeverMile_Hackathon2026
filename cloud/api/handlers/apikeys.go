package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"net/http"

	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

func HandleCreateAPIKey(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		// Free tier: max 3 active keys
		if user.Plan == "free" {
			count, err := dbpkg.CountActiveAPIKeys(database, user.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to count keys")
				return
			}
			if count >= 3 {
				writeError(w, http.StatusForbidden, "PLAN_REQUIRED", "Free tier allows at most 3 active API keys")
				return
			}
		}

		var req CreateAPIKeyRequest
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body")
			return
		}

		// Generate: "gw_live_" + 32 random bytes base64url (no padding)
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate key")
			return
		}
		plaintext := "gw_live_" + base64.RawURLEncoding.EncodeToString(raw)

		// Store SHA-256 hash only
		sum := sha256.Sum256([]byte(plaintext))
		keyHash := hex.EncodeToString(sum[:])

		apiKey, err := dbpkg.CreateAPIKey(database, user.ID, req.Name, keyHash)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create key")
			return
		}

		writeJSON(w, http.StatusCreated, CreateAPIKeyResponse{
			APIKey: plaintext,
			KeyID:  apiKey.ID,
			Name:   apiKey.Name,
		})
	}
}

func HandleListAPIKeys(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		keys, err := dbpkg.ListAPIKeysByUser(database, user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list keys")
			return
		}

		infos := make([]APIKeyInfo, 0, len(keys))
		for _, k := range keys {
			if k.RevokedAt != nil {
				continue // omit revoked keys from listing
			}
			info := APIKeyInfo{
				ID:        k.ID,
				Name:      k.Name,
				CreatedAt: k.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			}
			if k.LastUsedAt != nil {
				s := k.LastUsedAt.UTC().Format("2006-01-02T15:04:05Z")
				info.LastUsedAt = &s
			}
			infos = append(infos, info)
		}

		writeJSON(w, http.StatusOK, map[string]any{"api_keys": infos})
	}
}

func HandleRevokeAPIKey(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}

		keyID := r.PathValue("id")
		if keyID == "" {
			writeError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Missing key ID")
			return
		}

		if err := dbpkg.RevokeAPIKey(database, keyID, user.ID); err != nil {
			writeError(w, http.StatusNotFound, "CONTEXT_NOT_FOUND", "Key not found or already revoked")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"success": true})
	}
}
