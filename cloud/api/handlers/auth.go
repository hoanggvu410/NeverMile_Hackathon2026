package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/cloud/middleware"
)

// stateEntry holds a CSRF state token and its creation time.
type stateEntry struct {
	token     string
	createdAt time.Time
}

var oauthStates sync.Map

func init() {
	go func() {
		ticker := time.NewTicker(15 * time.Minute)
		for range ticker.C {
			now := time.Now()
			oauthStates.Range(func(k, v any) bool {
				if entry, ok := v.(stateEntry); ok {
					if now.Sub(entry.createdAt) > 10*time.Minute {
						oauthStates.Delete(k)
					}
				}
				return true
			})
		}
	}()
}

func HandleGitHubRedirect(clientID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate state")
			return
		}
		state := hex.EncodeToString(b)
		oauthStates.Store(state, stateEntry{token: state, createdAt: time.Now()})

		redirectURL := "https://github.com/login/oauth/authorize" +
			"?client_id=" + url.QueryEscape(clientID) +
			"&state=" + url.QueryEscape(state) +
			"&scope=user%3Aemail"
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

func HandleGitHubCallback(database *sql.DB, jwtSecret, clientID, clientSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		if code == "" || state == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing code or state")
			return
		}

		v, ok := oauthStates.LoadAndDelete(state)
		if !ok {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid or expired state")
			return
		}
		entry := v.(stateEntry)
		if time.Since(entry.createdAt) > 10*time.Minute {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "State expired")
			return
		}

		accessToken, err := exchangeGitHubCode(r.Context(), clientID, clientSecret, code)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to exchange GitHub code")
			return
		}

		ghUser, err := fetchGitHubUser(r.Context(), accessToken)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch GitHub user")
			return
		}

		user, err := dbpkg.UpsertUserFromGitHub(database, ghUser.id, ghUser.login, ghUser.email)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user")
			return
		}

		claims := middleware.Claims{
			UserID: user.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   user.ID,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := jwtToken.SignedString([]byte(jwtSecret))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sign token")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"token": tokenStr,
			"user": MeResponse{
				ID:          user.ID,
				Email:       user.Email,
				GitHubLogin: user.GitHubLogin,
				Plan:        user.Plan,
			},
		})
	}
}

func HandleMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := middleware.UserFromContext(r.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "INVALID_API_KEY", "Not authenticated")
			return
		}
		writeJSON(w, http.StatusOK, MeResponse{
			ID:          user.ID,
			Email:       user.Email,
			GitHubLogin: user.GitHubLogin,
			Plan:        user.Plan,
		})
	}
}

// ── unexported helpers ─────────────────────────────────────────────────────────

type githubUserInfo struct {
	id    string
	login string
	email string
}

func exchangeGitHubCode(ctx context.Context, clientID, clientSecret, code string) (string, error) {
	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth: %s: %s", result.Error, result.ErrorDesc)
	}
	return result.AccessToken, nil
}

func fetchGitHubUser(ctx context.Context, accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	email := result.Email
	if email == "" {
		email, _ = fetchPrimaryEmail(ctx, accessToken)
	}
	return &githubUserInfo{
		id:    fmt.Sprintf("%d", result.ID),
		login: result.Login,
		email: email,
	}, nil
}

func fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}
	if len(emails) > 0 {
		return emails[0].Email, nil
	}
	return "", nil
}
