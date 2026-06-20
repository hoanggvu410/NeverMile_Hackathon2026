package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
)

type contextKey string

const UserContextKey   contextKey = "user"
const APIKeyContextKey contextKey = "api_key"

// Claims is the JWT payload used for dashboard (GitHub OAuth) authentication.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func RequireAPIKey(database *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Missing or invalid Authorization header")
				return
			}
			rawKey := strings.TrimPrefix(auth, "Bearer ")

			sum := sha256.Sum256([]byte(rawKey))
			keyHash := hex.EncodeToString(sum[:])

			apiKey, err := dbpkg.GetAPIKeyByHash(database, keyHash)
			if err != nil || apiKey == nil || apiKey.RevokedAt != nil {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid or revoked API key")
				return
			}

			user, err := dbpkg.GetUserByID(database, apiKey.UserID)
			if err != nil || user == nil {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "User not found")
				return
			}
			if !user.IsActive {
				writeError(w, http.StatusForbidden, "ACCOUNT_DISABLED", "Account is disabled")
				return
			}

			go func() {
				if err := dbpkg.UpdateAPIKeyLastUsed(database, apiKey.ID); err != nil {
					log.Printf("UpdateAPIKeyLastUsed %s: %v", apiKey.ID, err)
				}
			}()

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, APIKeyContextKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireJWT(database *sql.DB, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Missing Authorization header")
				return
			}
			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid or expired token")
				return
			}

			user, err := dbpkg.GetUserByID(database, claims.UserID)
			if err != nil || user == nil {
				writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "User not found")
				return
			}
			if !user.IsActive {
				writeError(w, http.StatusForbidden, "ACCOUNT_DISABLED", "Account is disabled")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *dbpkg.User {
	v := ctx.Value(UserContextKey)
	if v == nil {
		return nil
	}
	u, _ := v.(*dbpkg.User)
	return u
}

func APIKeyFromContext(ctx context.Context) *dbpkg.APIKey {
	v := ctx.Value(APIKeyContextKey)
	if v == nil {
		return nil
	}
	k, _ := v.(*dbpkg.APIKey)
	return k
}

// writeError is a local helper so middleware does not import handlers.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"success":false,"error":{"code":%q,"message":%q}}`, code, message)
}
