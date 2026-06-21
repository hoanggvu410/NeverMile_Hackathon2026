# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build CLI
go build ./cmd/git-why/

# Build MCP server binary
go build ./mcp/

# Build both
go build ./...

# Run CLI
go run ./cmd/git-why/ <command>

# Run tests
go test ./...

# Run tests for a specific package
go test ./internal/context/...

# Add a dependency
go get <package>
go mod tidy
```

## Architecture

Two binaries, one shared library:

```
mcp/main.go          → MCP stdio server (spawned by Claude Code / Cursor / Windsurf)
cmd/git-why/         → CLI (Cobra) — same operations, terminal interface
internal/context/    → shared storage layer both binaries use
internal/mcp/        → MCP tool definitions (wraps internal/context)
```

**Request flow (MCP path):**
AI agent calls tool → `internal/mcp/server.go` handles it → calls `internal/context/store.go` → reads/writes `.git/gitwhy/`

**Request flow (CLI path):**
`git why <cmd>` → `cmd/git-why/*.go` → calls `internal/context/store.go` → same storage

### Storage layout (inside any git repo)

```
.git/gitwhy/
  contexts/<domain>/<topic>/<id>.md   ← whyspec files (source of truth)
  pending_commits                      ← commit hashes written by post-commit hook, consumed by next gitwhy_save
  graph.db                             ← SQLite context graph (Section 2 — not yet built)
  cache/semantic.db                    ← SQLite semantic cache (Section 2 — not yet built)
```

### Whyspec format

Each context is a structured markdown file (`internal/context/whyspec.go`). `Render()` serialises a `Context` struct to markdown; `Parse()` deserialises it back. The format uses `**Key:** value` header lines followed by `## Section` blocks. `Context ID` is the only required field for parsing.

### Hook mechanism

`git why hook install` appends a shell stanza to `.git/hooks/post-commit` that writes the HEAD commit hash to `.git/gitwhy/pending_commits`. On the next `gitwhy_save` call, `store.Save()` reads and truncates that file and appends the hashes to the saved context's `Commits` field.

### Section ownership (3-person team)

| Section | Owner | Directories |
|---|---|---|
| 1 — MCP + CLI + Hook | Done ✅ | `mcp/`, `cmd/`, `internal/context/`, `internal/mcp/` |
| 2 — Graph + Cache | In progress | `internal/graph/`, `internal/cache/` |
| 3 — Cloud + Auth | Not started | `cloud/`, `db/` |

Shared types go in `internal/types.go` (not yet created). Shared interfaces go in `internal/interfaces.go` (not yet created). Do not duplicate struct definitions across packages.

**Coordinate before touching:** `go.mod`, `go.sum`, root `main.go`. Announce in chat before running `go get` to avoid go.mod conflicts — resolve by running `go mod tidy` after accepting both change sets.

## Section 2 — What needs to be built

`internal/graph/` and `internal/cache/` are empty. Section 2 must expose two functions to Section 1:

```go
Search(query string, domain string, limit int) ([]SearchResult, bool, error)
// bool = true if cache hit

SaveToGraph(ctx Context) error
```

When `SaveToGraph` is called:
1. Embed the new context (text-embedding-3-small / nomic-embed-text for offline)
2. Cosine similarity against existing embeddings in graph.db → top 3 above 0.75 threshold
3. Send new context + candidates to LLM → classify edge type (CAUSED_BY / DEPENDS_ON / INVALIDATES / CONTRADICTS / CONSTRAINED_BY)
4. Write node + edges to graph.db

SQLite schema for graph.db is in `md documents/07-data-models.md` section 2. Full edge-linking algorithm is in `md documents/04-functional-requirements.md` §GRAPH-01.

When Section 2 is ready, Section 1 needs one change: `internal/mcp/server.go:179` currently calls `s.store.Search()` (full-text only). Replace with the graph `Search()`.

## Key product docs

Full specs live in `md documents/`. Most relevant for implementation:
- `04-functional-requirements.md` — §GRAPH, §SEARCH, §HOOK
- `07-data-models.md` — SQLite schemas (graph.db, semantic.db, PostgreSQL)
- `08-api-specifications.md` — MCP tool schemas + cloud REST API
- `work.md` — section ownership, interface contracts, conflict rules
- `product.md` — product overview, use case walkthrough, business model
- `AGENTS.md` — **read first**: plan-stage tripwire gate (required before editing files) + save contract for `gitwhy2_save`
