package context

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FindGitRoot walks up from the current working directory until it finds .git/.
func FindGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not inside a git repository")
		}
		dir = parent
	}
}

// HeadCommitHash returns the output of `git rev-parse HEAD` (trimmed).
func HeadCommitHash() (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentBranch returns the current branch name.
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RepoName returns "owner/repo" parsed from the origin remote URL,
// or falls back to the current directory's base name.
func RepoName() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err == nil {
		url := strings.TrimSpace(string(out))
		// Strip trailing .git
		url = strings.TrimSuffix(url, ".git")
		// Handle SSH: git@github.com:owner/repo
		if idx := strings.Index(url, ":"); idx != -1 && !strings.Contains(url[:idx], "/") {
			url = url[idx+1:]
		} else if idx := strings.Index(url, "//"); idx != -1 {
			// Handle HTTPS: https://github.com/owner/repo
			url = url[idx+2:]
			if slashIdx := strings.Index(url, "/"); slashIdx != -1 {
				url = url[slashIdx+1:]
			}
		}
		if url != "" {
			return url, nil
		}
	}
	// Fallback: use cwd basename
	dir, err2 := os.Getwd()
	if err2 != nil {
		return "unknown/repo", nil
	}
	return filepath.Base(dir), nil
}

// GitUserName returns git config user.name.
func GitUserName() (string, error) {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return "", fmt.Errorf("git config user.name: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GitWhyDir returns the absolute path to .git/gitwhy given a git root.
func GitWhyDir(gitRoot string) string {
	return filepath.Join(gitRoot, ".git", "gitwhy")
}

// ContextsDir returns the absolute path to .git/gitwhy/contexts.
func ContextsDir(gitRoot string) string {
	return filepath.Join(GitWhyDir(gitRoot), "contexts")
}

// EnsureGitWhyDirs creates .git/gitwhy/contexts/ (and parents) if absent.
func EnsureGitWhyDirs(gitRoot string) error {
	return os.MkdirAll(ContextsDir(gitRoot), 0755)
}
