package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	graphpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/graph"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the mark3labs MCP server with a context store and graph.
type MCPServer struct {
	store  *contextpkg.Store
	graph  *graphpkg.Graph
	server *server.MCPServer
}

// NewMCPServer constructs and registers all tools on the MCP server.
func NewMCPServer(store *contextpkg.Store, graph *graphpkg.Graph) (*MCPServer, error) {
	s := &MCPServer{
		store:  store,
		graph:  graph,
		server: server.NewMCPServer("gitwhy2", "0.1.0"),
	}
	s.registerTools()
	return s, nil
}

// Serve starts the stdio transport loop. Blocks until stdin closes.
func (s *MCPServer) Serve() error {
	return server.ServeStdio(s.server)
}

func (s *MCPServer) registerTools() {
	s.server.AddTool(mcp.NewTool("gitwhy2_status",
		mcp.WithDescription("Check setup state: git repo detection, pending commit count, local context count."),
	), s.handleStatus)

	s.server.AddTool(mcp.NewTool("gitwhy2_save",
		mcp.WithDescription("Save development context (reasoning, decisions, trade-offs) for the current session. "+
			"GitWhy indexes what you write — vague saves produce poor search results. "+
			"Write decisions as durable constraint sentences (Use X / Do not Y / Never Z). "+
			"Include rejected alternatives with the reason they were rejected. "+
			"Frame risks as trigger→consequence: 'If X happens, Y will break'. "+
			"See AGENTS.md in the repo root for the full save contract and examples."),
		mcp.WithString("prompt", mcp.Required(), mcp.Description("The original user prompt given to the AI agent")),
		mcp.WithString("reasoning", mcp.Required(), mcp.Description(
			"Why this approach was chosen over others. Focus on the trade-off, not what was done. "+
				"BAD: 'I implemented the spacing system'. "+
				"GOOD: 'Chose a fixed spacing scale to prevent layout drift; freeform spacing caused inconsistency within two PRs.'",
		)),
		mcp.WithString("decisions", mcp.Required(), mcp.Description(
			"Durable decisions as standalone constraint sentences. Each must be understood without surrounding context. "+
				"Use constraint language: Use / Do not / Never / Prefer / Always / Avoid. Bullet list preferred. "+
				"BAD: 'fixed spacing'. "+
				"GOOD: 'Use 4/8/16/24 spacing scale. Do not introduce ad-hoc spacing values.'",
		)),
		mcp.WithString("rejected_alternatives", mcp.Description(
			"Alternatives considered but discarded. Must include the option AND the reason it was rejected. "+
				"BAD: 'tried memoization'. "+
				"GOOD: 'Rejected memoization — cache invalidation on domain changes caused stale reads in tests.'",
		)),
		mcp.WithString("what_was_done", mcp.Description("Summary of what was actually implemented (not reasoning)")),
		mcp.WithString("risks", mcp.Description(
			"Future trigger situations that would break these decisions. "+
				"Frame as: 'If X happens, Y will break.' "+
				"BAD: 'might break'. "+
				"GOOD: 'If a new planner control bypasses the spacing scale, layout will diverge — check spacing before any planner UI edit.'",
		)),
		mcp.WithString("verification", mcp.Description("How to verify the changes")),
		mcp.WithArray("files", mcp.Description("Source files affected (strings)")),
		mcp.WithArray("commits", mcp.Description("Git commit hashes to link (auto-detected from HEAD if omitted)")),
		mcp.WithString("domain", mcp.Description("Hierarchical domain label, e.g. 'frontend/planner' or 'backend/auth'")),
		mcp.WithString("topic", mcp.Description("Topic slug, e.g. 'spacing-scale' or 'jwt-migration'")),
		mcp.WithString("agent", mcp.Description("Agent name, e.g. 'claude-code'")),
		mcp.WithString("title", mcp.Description("Short title for this context (defaults to first 72 chars of prompt)")),
		mcp.WithArray("edge_hints", mcp.Description("Optional edges to write: array of {context_id, edge_type} objects")),
	), s.handleSave)

	s.server.AddTool(mcp.NewTool("gitwhy2_get",
		mcp.WithDescription("Retrieve a saved context by its ID."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Context ID, e.g. 'ctx_a1B2c3D4'")),
	), s.handleGet)

	s.server.AddTool(mcp.NewTool("gitwhy2_search",
		mcp.WithDescription("Search saved contexts for direct Q&A and memory retrieval. "+
			"Use this tool to answer questions about past decisions or to look up what was saved. "+
			"This is NOT the plan-stage interrupt gate — if you are about to edit files and want to "+
			"check your plan against prior decisions, call gitwhy2_tripwire instead. "+
			"Uses claim graph first, then markdown fallback unless mode is graph_only."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("domain", mcp.Description("Optional: filter by domain prefix")),
		mcp.WithString("topic", mcp.Description("Optional: filter by topic prefix")),
		mcp.WithString("mode", mcp.Description("Optional: 'auto' (default) or 'graph_only'. graph_only never falls back to markdown search.")),
	), s.handleSearch)

	s.server.AddTool(mcp.NewTool("gitwhy2_tripwire",
		mcp.WithDescription("Required pre-edit plan gate. Call this after forming a concrete plan and BEFORE editing any files. "+
			"It checks the plan against prior saved decisions and returns interrupt=true if a conflict is found. "+
			"Use this tool — not gitwhy2_search — when the user asks you to 'check against gitwhy2' or 'check memory' before coding. "+
			"gitwhy2_search is for Q&A; gitwhy2_tripwire is the interrupt gate. "+
			"If interrupt=true: stop, show the user the relevant prior claims, and ask whether to revise the plan, continue anyway, or supersede the old decision. "+
			"Never falls back to markdown search."),
		mcp.WithString("event_type", mcp.Description("Runtime event type. Defaults to agent_plan_created.")),
		mcp.WithString("project_id", mcp.Description("Optional project filter. Defaults to the current git repository name.")),
		mcp.WithString("user_request", mcp.Required(), mcp.Description("The user's current request.")),
		mcp.WithString("agent_plan", mcp.Required(), mcp.Description("The agent's proposed plan before edits begin.")),
		mcp.WithArray("files_likely_touched", mcp.Description("Files likely touched by the plan.")),
		mcp.WithArray("concepts", mcp.Description("Concepts/components in the plan.")),
		mcp.WithArray("proposed_changes", mcp.Description("Concrete changes the plan proposes.")),
		mcp.WithArray("new_dependencies", mcp.Description("New dependencies/frameworks the plan may introduce.")),
		mcp.WithArray("risk_surfaces", mcp.Description("Risk surfaces affected by the plan.")),
	), s.handleTripwire)

	s.server.AddTool(mcp.NewTool("gitwhy2_list",
		mcp.WithDescription("Browse saved contexts by domain/topic hierarchy."),
		mcp.WithString("domain", mcp.Description("Filter by domain (optional)")),
		mcp.WithString("topic", mcp.Description("Filter by topic (optional)")),
	), s.handleList)

	s.server.AddTool(mcp.NewTool("gitwhy2_sync",
		mcp.WithDescription("Upload local contexts to the cloud (requires API key). Cloud sync coming soon."),
		mcp.WithString("context_id", mcp.Description("Sync a specific context. If omitted, sync all pending.")),
	), s.handleSync)

	s.server.AddTool(mcp.NewTool("gitwhy2_publish",
		mcp.WithDescription("Share synced contexts with your team (requires Team plan). Coming soon."),
		mcp.WithString("context_id", mcp.Description("Publish a specific context. If omitted, publish all synced.")),
	), s.handlePublish)

	s.server.AddTool(mcp.NewTool("gitwhy2_post_pr",
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
		"gitwhy2_dir":         info.GitWhyDir,
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
		Prompt:                prompt,
		Reasoning:             reasoning,
		KeyDecisions:          decisions,
		RejectedAlternatives:  req.GetString("rejected_alternatives", ""),
		WhatWasDone:           req.GetString("what_was_done", ""),
		RisksAndOpenQuestions: req.GetString("risks", ""),
		Verification:          req.GetString("verification", ""),
		Domain:                req.GetString("domain", ""),
		Topic:                 req.GetString("topic", ""),
		Agent:                 req.GetString("agent", ""),
		Title:                 req.GetString("title", ""),
		Date:                  time.Now().UTC(),
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
	ctx.ID = id
	if saved, err := s.store.Get(id); err == nil {
		ctx = *saved
	}

	// Parse edge_hints: array of {context_id, edge_type} objects.
	var edgeHints []graphpkg.EdgeHint
	if raw, ok := req.GetArguments()["edge_hints"]; ok {
		if rawSlice, ok := raw.([]interface{}); ok {
			for _, item := range rawSlice {
				if m, ok := item.(map[string]interface{}); ok {
					ctxID, _ := m["context_id"].(string)
					edgeType, _ := m["edge_type"].(string)
					if ctxID != "" && edgeType != "" {
						edgeHints = append(edgeHints, graphpkg.EdgeHint{ContextID: ctxID, EdgeType: edgeType})
					}
				}
			}
		}
	}

	if s.graph != nil {
		_ = s.graph.SaveToGraph(ctx, edgeHints)
	}

	warns := contextpkg.ValidateQuality(ctx)
	warnMsgs := make([]string, 0, len(warns))
	for _, w := range warns {
		warnMsgs = append(warnMsgs, w.Field+": "+w.Message)
	}

	resp := map[string]any{
		"success":   true,
		"id":        id,
		"timestamp": ctx.Date.Format(time.RFC3339),
	}
	if len(warnMsgs) > 0 {
		resp["warnings"] = warnMsgs
		resp["hint"] = "Re-save with improved field content or see AGENTS.md for the save contract."
	}
	out, _ := json.Marshal(resp)
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
	domain := req.GetString("domain", "")
	topic := req.GetString("topic", "")
	mode := strings.ToLower(req.GetString("mode", "auto"))
	graphOnly := mode == "graph_only"

	// Try semantic graph search first; fall back to full-text store search.
	var items []map[string]any
	var cacheHit bool
	retrievalMode := "markdown_fallback"
	unavailableReason := ""

	if s.graph != nil {
		graphResults, hit, graphErr := s.graph.Search(query, domain, 10)
		if graphErr == nil && len(graphResults) > 0 {
			retrievalMode = "claim_graph"
			cacheHit = hit
			for _, r := range graphResults {
				if topic != "" && !strings.HasPrefix(r.Topic, topic) {
					continue
				}
				item := map[string]any{
					"id":        r.ID,
					"title":     r.Title,
					"domain":    r.Domain,
					"topic":     r.Topic,
					"date":      r.Date.Format(time.RFC3339),
					"prompt":    r.Prompt,
					"score":     r.Score,
					"cache_hit": cacheHit,
				}
				if r.ClaimID != "" {
					item["claim_id"] = r.ClaimID
					item["claim"] = r.ClaimText
					item["claim_type"] = r.ClaimType
					item["vector_kind"] = r.VectorKind
				}
				if r.EdgeType != "" {
					item["edge_type"] = r.EdgeType
					item["depth"] = r.Depth
				}
				items = append(items, item)
			}
		} else if graphErr != nil {
			unavailableReason = graphErr.Error()
		}
	} else {
		unavailableReason = "graph is not initialized"
	}

	if graphOnly {
		if items == nil {
			items = []map[string]any{}
		}
		out, _ := json.Marshal(map[string]any{
			"results": items,
			"total":   len(items),
			"telemetry": map[string]any{
				"retrieval_mode":        "graph_only",
				"cache_hit":             cacheHit,
				"markdown_files_opened": 0,
				"unavailable_reason":    unavailableReason,
			},
		})
		return mcp.NewToolResultText(string(out)), nil
	}

	// Full-text fallback when graph returned nothing.
	if len(items) == 0 {
		storeResults, storeErr := s.store.Search(query)
		if storeErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", storeErr)), nil
		}
		for _, r := range storeResults {
			if domain != "" && !strings.HasPrefix(r.Domain, domain) {
				continue
			}
			if topic != "" && !strings.HasPrefix(r.Topic, topic) {
				continue
			}
			items = append(items, map[string]any{
				"id":     r.ID,
				"title":  r.Title,
				"domain": r.Domain,
				"topic":  r.Topic,
				"date":   r.Date.Format(time.RFC3339),
				"prompt": r.Prompt,
			})
		}
	}

	if items == nil {
		items = []map[string]any{}
	}
	out, _ := json.Marshal(map[string]any{
		"results": items,
		"total":   len(items),
		"telemetry": map[string]any{
			"retrieval_mode": retrievalMode,
			"cache_hit":      cacheHit,
		},
	})
	return mcp.NewToolResultText(string(out)), nil
}

func (s *MCPServer) handleTripwire(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userRequest, err := req.RequireString("user_request")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	agentPlan, err := req.RequireString("agent_plan")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectID := req.GetString("project_id", "")
	if projectID == "" {
		if repo, err := contextpkg.RepoName(); err == nil {
			projectID = repo
		}
	}
	event := graphpkg.RuntimeEvent{
		EventType:          req.GetString("event_type", "agent_plan_created"),
		ProjectID:          projectID,
		UserRequest:        userRequest,
		AgentPlan:          agentPlan,
		FilesLikelyTouched: req.GetStringSlice("files_likely_touched", nil),
		Concepts:           req.GetStringSlice("concepts", nil),
		ProposedChanges:    req.GetStringSlice("proposed_changes", nil),
		NewDependencies:    req.GetStringSlice("new_dependencies", nil),
		RiskSurfaces:       req.GetStringSlice("risk_surfaces", nil),
	}
	var result *graphpkg.TripwireResult
	if s.graph == nil {
		result = &graphpkg.TripwireResult{
			Available:  false,
			Candidates: []graphpkg.TripwireCandidate{},
			Telemetry: graphpkg.RetrievalTelemetry{
				RetrievalMode:       "graph_only",
				MarkdownFilesOpened: 0,
				EventType:           event.EventType,
				UnavailableReason:   "graph is not initialized",
			},
		}
	} else {
		result, err = s.graph.CheckTripwire(event)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("tripwire failed: %v", err)), nil
		}
	}
	out, _ := json.Marshal(result)
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
