package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage the post-commit hook",
	}
	cmd.AddCommand(newHookInstallCmd())
	cmd.AddCommand(newHookUninstallCmd())
	return cmd
}

func newHookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the post-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := contextpkg.FindGitRoot()
			if err != nil {
				return err
			}
			if err := installHook(root); err != nil {
				return err
			}
			fmt.Println("hook installed: .git/hooks/post-commit")
			fmt.Println("commit hashes will now be appended to .git/gitwhy/pending_commits after each commit")
			return nil
		},
	}
}

func newHookUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the post-commit hook",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := contextpkg.FindGitRoot()
			if err != nil {
				return err
			}
			if err := uninstallHook(root); err != nil {
				return err
			}
			fmt.Println("hook uninstalled")
			return nil
		},
	}
}

// hookStanza is the shell block written into post-commit.
func hookStanza() string {
	return `# gitwhy: append HEAD hash for pending_commits
git rev-parse HEAD >> "$(git rev-parse --git-dir)/gitwhy/pending_commits"
`
}

// installHook writes (or appends) the gitwhy stanza to .git/hooks/post-commit
// and ensures the file is executable.
func installHook(gitRoot string) error {
	hookPath := filepath.Join(gitRoot, ".git", "hooks", "post-commit")
	stanza := hookStanza()

	// Read existing content
	existing, _ := os.ReadFile(hookPath)
	if strings.Contains(string(existing), "gitwhy") {
		fmt.Println("hook already installed")
		return nil
	}

	var content string
	if len(existing) == 0 {
		content = "#!/bin/sh\n" + stanza
	} else {
		src := string(existing)
		if !strings.HasSuffix(src, "\n") {
			src += "\n"
		}
		content = src + stanza
	}

	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("writing post-commit hook: %w", err)
	}
	// Ensure executable bit
	return os.Chmod(hookPath, 0755)
}

// uninstallHook removes the gitwhy stanza from .git/hooks/post-commit.
func uninstallHook(gitRoot string) error {
	hookPath := filepath.Join(gitRoot, ".git", "hooks", "post-commit")
	data, err := os.ReadFile(hookPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	stanza := hookStanza()
	updated := strings.ReplaceAll(string(data), "# gitwhy: append HEAD hash for pending_commits\n", "")
	updated = strings.ReplaceAll(updated, stanza, "")

	// If only the shebang remains, remove the file
	trimmed := strings.TrimSpace(updated)
	if trimmed == "" || trimmed == "#!/bin/sh" {
		return os.Remove(hookPath)
	}

	return os.WriteFile(hookPath, []byte(updated), 0755)
}
