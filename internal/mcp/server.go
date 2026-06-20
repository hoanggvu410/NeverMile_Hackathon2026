package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the mark3labs MCP server with a context store.
type MCPServer struct {
	store  *contextpkg.Store
	server *server.MCPServer
}

// NewMCPServer constructs and registers all 8 tools on the MCP server.
func NewMCPServer(store *contextpkg.Store) (*MCPServer, error) {
	s := &MCPServer{
		store:  store,
		server: server.NewMCPServer("gitwhy", "0.1.0"),
	}
	s.registerTools()
	return s, nil
}

// Serve starts the stdio transport loop. Blocks until stdin closes.
func (s *MCPServer) Serve() error {
	return server.ServeStdio(s.server)
}

func (s *MCPServer) registerTools() {
	s.server.AddTool(mcp.NewTool("gitwhy_status",
		mcp.WithDescription("Check setup state: git repo detection, pending commit count, local context count."),
	), s.handleStatus)

	s.server.AddTool(mcp.NewTool("gitwhy_save",
		mcp.WithDescription("Save development context (reasoning, decisions, trade-offs) for the current session."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The original user prompt given to the AI agent")),
		mcp.WithString("reasoning", mcp.Required(), mcp.Description("Agent's explanation of its approach and methodology")),
		mcp.WithString("decisions", mcp.Required(), mcp.Description("Key choices made with rationale")),
		mcp.WithString("rejected_alternatives", mcp.Description("Options considered but discarded, and why")),
		mcp.WithString("what_was_done", mcp.Description("Summary of what was actually implemented")),
		mcp.WithString("risks", mcp.Description("Risks and open questions")),
		mcp.WithString("verification", mcp.Description("How to verify the changes")),
		mcp.WithArray("files", mcp.Description("Source files affected (strings)")),
		mcp.WithArray("commits", mcp.Description("Git commit hashes to link (auto-detected from HEAD if omitted)")),
		mcp.WithString("domain", mcp.Description("Hierarchical domain label, e.g. 'backend/auth'")),
		mcp.WithString("topic", mcp.Description("Topic slug, e.g. 'jwt-migration'")),
		mcp.WithString("agent", mcp.Description("Agent name, e.g. 'claude-code'")),
		mcp.WithString("title", mcp.Description("Short title for this context (defaults to first 72 chars of prompt)")),
	), s.handleSave)

	s.server.AddTool(mcp.NewTool("gitwhy_get",
		mcp.WithDescription("Retrieve a saved context by its ID."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Context ID, e.g. 'ctx_a1B2c3D4'")),
	), s.handleGet)

	s.server.AddTool(mcp.NewTool("gitwhy_search",
		mcp.WithDescription("Search saved contexts by keyword. Searches all text fields locally."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("domain", mcp.Description("Optional: filter by domain prefix")),
		mcp.WithString("topic", mcp.Description("Optional: filter by topic prefix")),
	), s.handleSearch)

	s.server.AddTool(mcp.NewTool("gitwhy_list",
		mcp.WithDescription("Browse saved contexts by domain/topic hierarchy."),
		mcp.WithString("domain", mcp.Description("Filter by domain (optional)")),
		mcp.WithString("topic", mcp.Description("Filter by topic (optional)")),
	), s.handleList)

	s.server.AddTool(mcp.NewTool("gitwhy_sync",
		mcp.WithDescription("Upload local contexts to the cloud (requires API key). Cloud sync coming soon."),
		mcp.WithString("context_id", mcp.Description("Sync a specific context. If omitted, sync all pending.")),
	), s.handleSync)

	s.server.AddTool(mcp.NewTool("gitwhy_publish",
		mcp.WithDescription("Share synced contexts with your team (requires Team plan). Coming soon."),
		mcp.WithString("context_id", mcp.Description("Publish a specific context. If omitted, publish all synced.")),
	), s.handlePublish)

	s.server.AddTool(mcp.NewTool("gitwhy_post_pr",
		mcp.WithDescription("Post a context summary as a GitHub PR comment via gitwhy-bot. Coming soon."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Context ID to post")),
		mcp.WithString("repo", mcp.Required(), mcp.Description("GitHub repo in owner/repo format")),
		mcp.WithString("pr_number", mcp.Required(), mcp.Description("Pull request number")),
	), s.handlePostPR)
}

func (s *MCPServer) handleStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info := s.store.Status()
	out, _ := json.Marshal(map[string]any{
		"is_git_repo":         info.IsGitRepo,
		"git_root":            info.GitRoot,
		"gitwhy_dir":          info.GitWhyDir,
		"local_context_count": info.LocalContextCount,
		"pending_commits":     info.PendingCommits,
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handleSave(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	prompt, err := req.RequireString("prompt")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reasoning, err := req.RequireString("reasoning")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	decisions, err := req.RequireString("decisions")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ctx := contextpkg.Context{
		Prompt:               prompt,
		Reasoning:            reasoning,
		KeyDecisions:         decisions,
		RejectedAlternatives: req.GetString("rejected_alternatives", ""),
		WhatWasDone:          req.GetString("what_was_done", ""),
		RisksAndOpenQuestions: req.GetString("risks", ""),
		Verification:         req.GetString("verification", ""),
		Domain:               req.GetString("domain", ""),
		Topic:                req.GetString("topic", ""),
		Agent:                req.GetString("agent", ""),
		Title:                req.GetString("title", ""),
		Date:                 time.Now().UTC(),
	}

	// Files
	for _, f := range req.GetStringSlice("files", nil) {
		ctx.Files = append(ctx.Files, contextpkg.FileEntry{File: f, Status: "modified"})
	}
	// Commits
	ctx.Commits = req.GetStringSlice("commits", nil)

	// If no commits provided, try to auto-detect HEAD
	if len(ctx.Commits) == 0 {
		if hash, err := contextpkg.HeadCommitHash(); err == nil {
			ctx.Commits = []string{hash}
		}
	}

	id, err := s.store.Save(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("save failed: %v", err)), nil
	}

	out, _ := json.Marshal(map[string]any{
		"success":   true,
		"id":        id,
		"timestamp": ctx.Date.Format(time.RFC3339),
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handleGet(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	ctx, err := s.store.Get(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("get failed: %v", err)), nil
	}
	return mcp.NewToolResultText(contextpkg.Render(ctx)), nil
}

func (s *MCPServer) handleSearch(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	results, err := s.store.Search(query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Apply optional domain/topic filter
	domain := req.GetString("domain", "")
	topic := req.GetString("topic", "")
	if domain != "" || topic != "" {
		filtered := results[:0]
		for _, r := range results {
			if domain != "" && !strings.HasPrefix(r.Domain, domain) {
				continue
			}
			if topic != "" && !strings.HasPrefix(r.Topic, topic) {
				continue
			}
			filtered = append(filtered, r)
		}
		results = filtered
	}

	items := make([]map[string]any, 0, len(results))
	for _, r := range results {
		items = append(items, map[string]any{
			"id":     r.ID,
			"title":  r.Title,
			"domain": r.Domain,
			"topic":  r.Topic,
			"date":   r.Date.Format(time.RFC3339),
			"prompt": r.Prompt,
		})
	}
	out, _ := json.Marshal(map[string]any{
		"results": items,
		"total":   len(items),
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handleList(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := req.GetString("domain", "")
	topic := req.GetString("topic", "")
	summaries, err := s.store.List(domain, topic)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("list failed: %v", err)), nil
	}
	items := make([]map[string]any, 0, len(summaries))
	for _, s := range summaries {
		items = append(items, map[string]any{
			"id":     s.ID,
			"title":  s.Title,
			"domain": s.Domain,
			"topic":  s.Topic,
			"date":   s.Date.Format(time.RFC3339),
		})
	}
	out, _ := json.Marshal(map[string]any{"contexts": items, "total": len(items)})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handleSync(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, _ := json.Marshal(map[string]any{
		"message": "Cloud sync coming soon. Visit gitwhy.dev to join the waitlist.",
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handlePublish(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, _ := json.Marshal(map[string]any{
		"message": "Team publish coming soon. Visit gitwhy.dev to join the waitlist.",
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handlePostPR(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, _ := json.Marshal(map[string]any{
		"message": "PR comment bot coming soon. Visit gitwhy.dev to join the waitlist.",
	})
	return mcp.NewToolResultText(string(out)), nil
}
