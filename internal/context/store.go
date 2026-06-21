package context

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ErrNotFound is returned by Get when no context matches the given ID.
var ErrNotFound = errors.New("context not found")

// Store is the file-system store rooted at .git/gitwhy/contexts/.
type Store struct {
	gitRoot     string
	gitWhyDir   string
	contextsDir string
}

// NewStore auto-detects the git root and ensures required directories exist.
func NewStore() (*Store, error) {
	root, err := FindGitRoot()
	if err != nil {
		return nil, err
	}
	if err := EnsureGitWhyDirs(root); err != nil {
		return nil, err
	}
	return &Store{
		gitRoot:     root,
		gitWhyDir:   GitWhyDir(root),
		contextsDir: ContextsDir(root),
	}, nil
}

// Save writes ctx to .git/gitwhy/contexts/<domain>/<topic>/<id>.md.
// Missing fields are auto-filled from git. After writing, pending_commits is
// consumed: its hashes are appended to ctx.Commits, the file is re-rendered
// and overwritten, then pending_commits is truncated.
func (s *Store) Save(ctx Context) (string, error) {
	if ctx.ID == "" {
		ctx.ID = GenerateID()
	}
	if ctx.Date.IsZero() {
		ctx.Date = time.Now().UTC()
	}
	if ctx.SavedBy == "" {
		if name, err := GitUserName(); err == nil {
			ctx.SavedBy = name
		}
	}
	if ctx.Repository == "" {
		if repo, err := RepoName(); err == nil {
			ctx.Repository = repo
		}
	}
	if ctx.Branch == "" {
		if branch, err := CurrentBranch(); err == nil {
			ctx.Branch = branch
		}
	}

	domain := ctx.Domain
	if domain == "" {
		domain = "_"
	}
	topic := ctx.Topic
	if topic == "" {
		topic = "_"
	}

	dir := filepath.Join(s.contextsDir, domain, topic)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	path := s.contextPath(domain, topic, ctx.ID)

	// Write initial file
	if err := os.WriteFile(path, []byte(Render(&ctx)), 0644); err != nil {
		return "", err
	}

	// Consume pending_commits
	pending, err := s.consumePendingCommits()
	if err == nil && len(pending) > 0 {
		ctx.Commits = append(ctx.Commits, pending...)
		// Rewrite with updated commits
		_ = os.WriteFile(path, []byte(Render(&ctx)), 0644)
	}

	return ctx.ID, nil
}

// Get reads and parses the context file matching id via a recursive walk.
func (s *Store) Get(id string) (*Context, error) {
	var found string
	err := filepath.WalkDir(s.contextsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasSuffix(d.Name(), ".md") && strings.Contains(d.Name(), id) {
			found = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == "" {
		return nil, ErrNotFound
	}
	data, err := os.ReadFile(found)
	if err != nil {
		return nil, err
	}
	ctx, err := Parse(string(data))
	if err != nil {
		return nil, err
	}
	// Recover domain/topic from path
	rel, _ := filepath.Rel(s.contextsDir, found)
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 3 {
		ctx.Domain = parts[0]
		ctx.Topic = parts[1]
	}
	return ctx, nil
}

// Search does a case-insensitive substring scan across all text fields of every context.
func (s *Store) Search(query string) ([]ContextSummary, error) {
	all, err := s.walkContexts()
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var results []ContextSummary
	for _, summary := range all {
		// Load full context to search all fields
		ctx, err := s.Get(summary.ID)
		if err != nil {
			continue
		}
		haystack := strings.ToLower(strings.Join([]string{
			ctx.Title,
			ctx.Prompt,
			ctx.WhatWasDone,
			ctx.Reasoning,
			ctx.KeyDecisions,
			ctx.RejectedAlternatives,
			ctx.RisksAndOpenQuestions,
			ctx.Topic,
		}, " "))
		if strings.Contains(haystack, q) {
			results = append(results, summary)
		}
	}
	return results, nil
}

// List returns all ContextSummary values, optionally filtered by domain and/or topic.
// Empty string means "all". Sorted newest-first.
func (s *Store) List(domain, topic string) ([]ContextSummary, error) {
	all, err := s.walkContexts()
	if err != nil {
		return nil, err
	}
	if domain == "" && topic == "" {
		return all, nil
	}
	var out []ContextSummary
	for _, c := range all {
		if domain != "" && !strings.HasPrefix(c.Domain, domain) {
			continue
		}
		if topic != "" && !strings.HasPrefix(c.Topic, topic) {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}

// Status collects StatusInfo without modifying anything.
func (s *Store) Status() StatusInfo {
	info := StatusInfo{
		IsGitRepo: true,
		GitRoot:   s.gitRoot,
		GitWhyDir: s.gitWhyDir,
	}
	summaries, err := s.walkContexts()
	if err == nil {
		info.LocalContextCount = len(summaries)
	}
	pendingPath := filepath.Join(s.gitWhyDir, "pending_commits")
	if data, err := os.ReadFile(pendingPath); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				info.PendingCommits = append(info.PendingCommits, line)
			}
		}
	}
	return info
}

// GitWhyDir returns the path to the .git/gitwhy directory for this repo.
func (s *Store) GitWhyDir() string { return s.gitWhyDir }

// contextPath builds the full file path for a given domain/topic/id triple.
func (s *Store) contextPath(domain, topic, id string) string {
	return filepath.Join(s.contextsDir, domain, topic, id+".md")
}

// walkContexts returns all ContextSummary values by walking contextsDir.
func (s *Store) walkContexts() ([]ContextSummary, error) {
	var summaries []ContextSummary
	err := filepath.WalkDir(s.contextsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		ctx, err := Parse(string(data))
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(s.contextsDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		domain, topic := "_", "_"
		if len(parts) == 3 {
			domain = parts[0]
			topic = parts[1]
		}
		// Fall back to fields parsed from the file itself (handles multi-level domain paths).
		if (domain == "_" || topic == "_") && (ctx.Domain != "" || ctx.Topic != "") {
			if ctx.Domain != "" {
				domain = ctx.Domain
			}
			if ctx.Topic != "" {
				topic = ctx.Topic
			}
		}
		summaries = append(summaries, ContextSummary{
			ID:     ctx.ID,
			Title:  ctx.Title,
			Domain: domain,
			Topic:  topic,
			Date:   ctx.Date,
			Prompt: ctx.Prompt,
			Path:   path,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Sort newest-first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Date.After(summaries[j].Date)
	})
	return summaries, nil
}

// consumePendingCommits reads .git/gitwhy/pending_commits, returns lines, then truncates.
func (s *Store) consumePendingCommits() ([]string, error) {
	path := filepath.Join(s.gitWhyDir, "pending_commits")
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var hashes []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			hashes = append(hashes, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return hashes, err
	}
	// Truncate
	if err := f.Truncate(0); err != nil {
		return hashes, err
	}
	return hashes, nil
}
