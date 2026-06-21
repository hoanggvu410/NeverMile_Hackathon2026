package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

// agentsMDDest is the filename written into the target repo by mcp install.
const agentsMDDest = "AGENTS.md"

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage the GitWhy2 MCP server configuration",
	}
	cmd.AddCommand(newMCPInstallCmd())
	cmd.AddCommand(newMCPUninstallCmd())
	return cmd
}

func newMCPInstallCmd() *cobra.Command {
	var binaryPath string
	var global bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Register the GitWhy2 MCP server in Claude Code and Cursor for this repo",
		Long: `Writes a gitwhy2 MCP server entry into .claude/settings.json and .cursor/mcp.json
with cwd set to the current git repository root. Run this once inside any project
that should use GitWhy2. The MCP binary is auto-detected alongside git-why.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := contextpkg.FindGitRoot()
			if err != nil {
				return err
			}

			mcpBin, err := resolveMCPBinary(binaryPath)
			if err != nil {
				return err
			}

			entry := mcpServerEntry{
				Command: mcpBin,
				Args:    []string{},
				CWD:     root,
			}

			claudeTarget, cursorTarget := mcpConfigPaths(root, global)

			if err := writeMCPConfig(claudeTarget, "gitwhy2", entry); err != nil {
				return fmt.Errorf("claude code config: %w", err)
			}
			fmt.Printf("claude code: %s\n", claudeTarget)

			if err := writeMCPConfig(cursorTarget, "gitwhy2", entry); err != nil {
				return fmt.Errorf("cursor config: %w", err)
			}
			fmt.Printf("cursor:      %s\n", cursorTarget)

			agentsDest := filepath.Join(root, agentsMDDest)
			if err := os.WriteFile(agentsDest, agentsMD, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not write AGENTS.md: %v\n", err)
			} else {
				fmt.Printf("agents.md:   %s\n", agentsDest)
			}

			fmt.Printf("\ngitwhy2 MCP registered — repo: %s\n", root)
			fmt.Println("Restart your IDE or reload the MCP server list to pick up the change.")
			return nil
		},
	}

	cmd.Flags().StringVar(&binaryPath, "binary", "", "path to gitwhy2-mcp binary (auto-detected if omitted)")
	cmd.Flags().BoolVar(&global, "global", false, "write to user-global config (~/.claude, ~/.cursor) instead of project config")
	return cmd
}

func newMCPUninstallCmd() *cobra.Command {
	var global bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the GitWhy2 MCP server entry from Claude Code and Cursor configs",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := contextpkg.FindGitRoot()
			if err != nil {
				return err
			}

			claudeTarget, cursorTarget := mcpConfigPaths(root, global)

			removedClaude := removeMCPEntry(claudeTarget, "gitwhy2")
			removedCursor := removeMCPEntry(cursorTarget, "gitwhy2")

			if removedClaude {
				fmt.Printf("removed from: %s\n", claudeTarget)
			}
			if removedCursor {
				fmt.Printf("removed from: %s\n", cursorTarget)
			}
			if !removedClaude && !removedCursor {
				fmt.Println("gitwhy2 entry not found in any config file")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "remove from user-global config instead of project config")
	return cmd
}

// mcpConfigPaths returns the Claude Code and Cursor config file paths.
func mcpConfigPaths(gitRoot string, global bool) (claudePath, cursorPath string) {
	if global {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude", "settings.json"),
			filepath.Join(home, ".cursor", "mcp.json")
	}
	return filepath.Join(gitRoot, ".claude", "settings.json"),
		filepath.Join(gitRoot, ".cursor", "mcp.json")
}

// resolveMCPBinary finds the gitwhy2-mcp binary.
// If explicit is set, that path is used directly. Otherwise the binary is
// looked up alongside the running git-why executable.
func resolveMCPBinary(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("binary not found at %s: %w", explicit, err)
		}
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		return abs, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate own executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)

	candidates := []string{
		"gitwhy2-mcp.exe", "gitwhy2-mcp",
		"git-why-mcp.exe", "git-why-mcp",
	}
	for _, name := range candidates {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf(
		"gitwhy2-mcp binary not found alongside %s\n"+
			"Build it with: go build -o %s ./mcp/\n"+
			"Or pass --binary <path>",
		exe, filepath.Join(dir, "gitwhy2-mcp"),
	)
}

// mcpServerEntry is the JSON shape written into mcpServers.
type mcpServerEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	CWD     string   `json:"cwd"`
}

// writeMCPConfig merges a named server entry into an existing or new config JSON file.
// Existing keys outside mcpServers are preserved. The gitwhy2 entry is always overwritten.
func writeMCPConfig(path, serverName string, entry mcpServerEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	var config map[string]any
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &config)
	}
	if config == nil {
		config = map[string]any{}
	}

	servers, _ := config["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers[serverName] = entry
	config["mcpServers"] = servers

	out, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0644)
}

// removeMCPEntry deletes the named server from the config and returns true if it existed.
func removeMCPEntry(path, serverName string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var config map[string]any
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}
	servers, ok := config["mcpServers"].(map[string]any)
	if !ok {
		return false
	}
	if _, exists := servers[serverName]; !exists {
		return false
	}
	delete(servers, serverName)
	config["mcpServers"] = servers
	out, _ := json.MarshalIndent(config, "", "  ")
	_ = os.WriteFile(path, append(out, '\n'), 0644)
	return true
}
