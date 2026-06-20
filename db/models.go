package db

import (
	"errors"
	"time"
)

type User struct {
	ID          string
	Email       string
	GitHubID    string
	GitHubLogin string
	Plan        string // "free" | "team"
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type APIKey struct {
	ID         string
	UserID     string
	KeyHash    string
	Name       string
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

type ContextRow struct {
	ID                   string
	LocalID              string
	UserID               string
	TeamID               *string
	Prompt               string
	Reasoning            string
	Decisions            string
	RejectedAlternatives string
	TradeOffs            string
	Files                []string // deserialized from JSONB
	Commits              []string // deserialized from JSONB
	Domain               string
	Topic                string
	Agent                string
	Model                string
	IsPublished          bool
	RepoName             string
	ContextTS            time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Team struct {
	ID        string
	Name      string
	OwnerID   string
	Plan      string
	CreatedAt time.Time
}

type TeamMember struct {
	TeamID   string
	UserID   string
	Role     string // "owner" | "member"
	JoinedAt time.Time
}

type PRComment struct {
	ID         string
	ContextID  string
	Repo       string
	PRNumber   int
	CommentURL string
	PostedAt   time.Time
}

var ErrQuotaExceeded = errors.New("monthly sync quota exceeded")
