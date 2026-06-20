package githubapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	dbpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/db"
)

var ErrAppNotInstalled = errors.New("gitwhy-bot not installed on repo")

type cachedToken struct {
	token     string
	expiresAt time.Time
}

type App struct {
	appID         int64
	webhookSecret string
	privateKey    []byte // raw PEM bytes
	tokenCache    sync.Map
}

func NewApp(appID int64, privateKeyPath, webhookSecret string) (*App, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	// Validate the PEM parses by creating a test transport; discard it.
	if _, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, key); err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return &App{
		appID:         appID,
		webhookSecret: webhookSecret,
		privateKey:    key,
	}, nil
}

func (a *App) GetInstallationToken(ctx context.Context, repo string) (string, error) {
	if v, ok := a.tokenCache.Load(repo); ok {
		ct := v.(cachedToken)
		if time.Now().Before(ct.expiresAt.Add(-5 * time.Minute)) {
			return ct.token, nil
		}
	}

	installationID, err := a.getInstallationID(ctx, repo)
	if err != nil {
		return "", err
	}

	itr, err := ghinstallation.New(http.DefaultTransport, a.appID, installationID, a.privateKey)
	if err != nil {
		return "", fmt.Errorf("create installation transport: %w", err)
	}
	token, err := itr.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("get installation token: %w", err)
	}

	a.tokenCache.Store(repo, cachedToken{
		token:     token,
		expiresAt: time.Now().Add(1 * time.Hour),
	})
	return token, nil
}

func (a *App) getInstallationID(ctx context.Context, repo string) (int64, error) {
	appsTransport, err := ghinstallation.NewAppsTransport(http.DefaultTransport, a.appID, a.privateKey)
	if err != nil {
		return 0, fmt.Errorf("create apps transport: %w", err)
	}
	client := &http.Client{Transport: appsTransport}

	url := fmt.Sprintf("https://api.github.com/repos/%s/installation", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("lookup installation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return 0, ErrAppNotInstalled
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GitHub installation lookup status %d", resp.StatusCode)
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode installation response: %w", err)
	}
	return result.ID, nil
}

func (a *App) PostPRComment(ctx context.Context, repo string, prNumber int, body string) (string, error) {
	token, err := a.GetInstallationToken(ctx, repo)
	if err != nil {
		return "", err // includes ErrAppNotInstalled
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments", repo, prNumber)
	payload := fmt.Sprintf(`{"body":%s}`, mustMarshalString(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post PR comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("PR or repo not found")
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("GitHub comment status %d", resp.StatusCode)
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode comment response: %w", err)
	}
	return result.HTMLURL, nil
}

// FormatPRComment renders a db.ContextRow into the standard PR comment markdown.
// github-app imports db; db does not import github-app — no cycle.
// github-app does NOT import internal/context.
func FormatPRComment(row *dbpkg.ContextRow) string {
	var sb strings.Builder

	sb.WriteString("## 🧠 GitWhy Context\n\n")
	sb.WriteString(fmt.Sprintf("**[%s](https://app.gitwhy.dev/contexts/%s)**\n", row.LocalID, row.LocalID))

	// Domain/topic subtitle
	if row.Domain != "" || row.Topic != "" {
		subtitle := row.Domain
		if row.Topic != "" {
			if subtitle != "" {
				subtitle += "/" + row.Topic
			} else {
				subtitle = row.Topic
			}
		}
		sb.WriteString(fmt.Sprintf("*%s*\n", subtitle))
	}
	sb.WriteString("\n")

	if row.Prompt != "" {
		sb.WriteString("### Original Prompt\n\n")
		for _, line := range strings.Split(row.Prompt, "\n") {
			sb.WriteString("> " + line + "\n")
		}
		sb.WriteString("\n")
	}

	if row.Decisions != "" {
		sb.WriteString("### Key Decisions\n\n")
		sb.WriteString(row.Decisions)
		if !strings.HasSuffix(row.Decisions, "\n") {
			sb.WriteByte('\n')
		}
		sb.WriteString("\n")
	}

	if row.RejectedAlternatives != "" {
		sb.WriteString("### Rejected Alternatives\n\n")
		sb.WriteString(row.RejectedAlternatives)
		if !strings.HasSuffix(row.RejectedAlternatives, "\n") {
			sb.WriteByte('\n')
		}
		sb.WriteString("\n")
	}

	if len(row.Commits) > 0 {
		sb.WriteString("### Linked Commits\n\n")
		for _, c := range row.Commits {
			short := c
			if len(c) > 7 {
				short = c[:7]
			}
			sb.WriteString(fmt.Sprintf("- `%s`\n", short))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	sb.WriteString("*Posted by [gitwhy-bot](https://gitwhy.dev)*\n")
	return sb.String()
}

func mustMarshalString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
