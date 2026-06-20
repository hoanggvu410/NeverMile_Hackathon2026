package handlers

import (
	"encoding/json"
	"net/http"
)

// ── Request types ──────────────────────────────────────────────────────────────

type SyncRequestItem struct {
	LocalID              string   `json:"local_id"`
	Prompt               string   `json:"prompt"`
	Reasoning            string   `json:"reasoning"`
	Decisions            string   `json:"decisions"`
	RejectedAlternatives string   `json:"rejected_alternatives"`
	TradeOffs            string   `json:"trade_offs,omitempty"`
	Files                []string `json:"files"`
	Commits              []string `json:"commits"`
	Domain               string   `json:"domain"`
	Topic                string   `json:"topic"`
	Agent                string   `json:"agent"`
	Model                string   `json:"model"`
	ContextTS            string   `json:"context_ts"` // RFC3339
}

type SyncRequest struct {
	Contexts []SyncRequestItem `json:"contexts"`
}

type PublishRequest struct {
	ContextIDs []string `json:"context_ids"` // local_ids
}

type PRCommentRequest struct {
	ContextLocalID string `json:"context_local_id"`
	Repo           string `json:"repo"`
	PRNumber       int    `json:"pr_number"`
}

type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

// ── Response types ─────────────────────────────────────────────────────────────

type SyncResponse struct {
	Synced         []string `json:"synced"`
	Failed         []string `json:"failed"`
	QuotaUsed      int      `json:"quota_used"`
	QuotaRemaining int      `json:"quota_remaining"` // -1 = unlimited (team plan)
}

type PRCommentResponse struct {
	CommentURL string `json:"comment_url"`
}

type CreateAPIKeyResponse struct {
	APIKey string `json:"api_key"` // plaintext, shown once
	KeyID  string `json:"key_id"`
	Name   string `json:"name"`
}

type APIKeyInfo struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}

type MeResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	GitHubLogin string `json:"github_login"`
	Plan        string `json:"plan"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	DB        string `json:"db"`
	Timestamp string `json:"timestamp"`
}

// ── Shared helpers ─────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
