# Work Division

3 people. Each owns one section completely. No one touches another section's files without asking.

---

## Section 1 — MCP Server + CLI + Auto-save Hook

### What this is

The product that judges interact with. Everything an agent calls goes through here.

- Node.js stdio MCP server — receives tool calls from Claude Code / Cursor / Windsurf
- 8 MCP tools — validate input, call Go functions, return structured JSON
- Go CLI (`gitwhy save`, `gitwhy search`, `gitwhy status`)
- Write/read `.md` whyspec files to `.git/gitwhy/contexts/`
- Auto-detect HEAD commit hash (`git rev-parse HEAD`)
- Post-commit hook that fires `gitwhy_save` automatically

The 5 tools that must work for the demo: `save`, `search`, `get`, `list`, `status`.
`sync`, `publish`, `post_pr` come after the demo path works.

### Files you own

```
mcp/              ← Node.js MCP server
cmd/              ← Go CLI (Cobra)
internal/context/ ← whyspec read/write, local file storage
internal/mcp/     ← MCP stdio server (Go)
```

### What you expose to others

Section 2 (graph + cache) gives you two functions. You call them, you don't implement them:

```go
// call this inside the gitwhy_search tool handler
Search(query string, domain string, limit int) ([]SearchResult, bool, error)
// bool = cache hit

// call this after writing the .md file in gitwhy_save
SaveToGraph(ctx Context) error
```

Section 3 (cloud) gives you a client. You call it, you don't implement it:

```go
Sync(contexts []Context) (SyncResult, error)
PostPR(contextID, repo string, prNumber int) (string, error)
```

On day 1, agree on these function signatures with Sections 2 and 3. Then build against them independently — stub them out if the other sections aren't ready.

### Starting prompt

```
You are building Section 1 of GitWhy: the MCP server, Go CLI, and auto-save hook.

Read these files first:
- md documents/04-functional-requirements.md (sections §SAVE, §GET, §SEARCH, §LIST, §STATUS, §HOOK)
- md documents/08-api-specifications.md (Part A — MCP Tool Schemas)
- md documents/06-system-architecture.md (sections 2, 5, 6, 7)
- md documents/07-data-models.md (section 1 — whyspec format)

What you are building:
1. Node.js stdio MCP server that receives tool calls and routes them to Go
2. Go CLI with Cobra — commands: save, search, get, list, status
3. Local file storage — write whyspec .md files to .git/gitwhy/contexts/{id}.md
4. Auto-detect current git commit hash with git rev-parse HEAD
5. Post-commit hook at .git/hooks/post-commit that calls gitwhy_save automatically

What you are NOT building (stub these out):
- The graph and cache (Section 2) — stub Search() to return mock results
- The cloud API client (Section 3) — stub Sync() and PostPR() to return success

Your directory: mcp/, cmd/, internal/context/, internal/mcp/
Do not create files outside these directories.

The demo path that must work first: gitwhy_save → writes .md → gitwhy_search → returns results.
```

---

## Section 2 — Context Graph + Semantic Cache

### What this is

The intelligence behind `gitwhy_search`. What makes the answer a causal chain instead of a flat list. What makes the second query cost $0.00.

- SQLite at `.git/gitwhy/graph.db` — two tables: `context_nodes` and `context_edges`
- 5 typed edge types: CAUSED_BY, CONSTRAINED_BY, INVALIDATES, CONTRADICTS, DEPENDS_ON
- 2-hop recursive CTE traversal
- Embeddings via `text-embedding-3-small` (OpenAI) or `nomic-embed-text` (local)
- Cosine similarity lookup in `semantic.db`
- Cache TTL 24h, max 1,000 entries

This section is self-contained. It doesn't need the cloud API. It only needs to expose two clean functions to Section 1.

### Files you own

```
internal/graph/   ← SQLite graph.db, nodes, edges, traversal
internal/cache/   ← SQLite semantic.db, embedding, cosine lookup
```

### What you expose to Section 1

```go
// gitwhy_search calls this
func Search(query string, domain string, limit int) ([]SearchResult, bool, error)
// []SearchResult = ordered chain of contexts
// bool = true if cache hit

// gitwhy_save calls this after writing the .md file
func SaveToGraph(ctx Context) error
```

That is all Section 1 needs from you. Keep the interface small.

### Starting prompt

```
You are building Section 2 of GitWhy: the context graph and semantic cache.

Read these files first:
- md documents/04-functional-requirements.md (sections §SEARCH, §GRAPH)
- md documents/07-data-models.md (section 2 — context graph SQLite schema)
- md documents/06-system-architecture.md (section 3 — search flow)
- md documents/product.md (use case section — shows exactly what the graph must return)

What you are building:
1. SQLite graph.db with two tables: context_nodes and context_edges
   Schema is in 07-data-models.md section 2 — use it exactly.
2. 2-hop recursive CTE traversal on typed edges
3. SQLite semantic.db for query caching
4. Embedding via text-embedding-3-small (OpenAI API) — fall back to nomic-embed-text for offline
5. Cosine similarity comparison against cached query embeddings

What you expose (write these functions first, implement after):
  Search(query string, domain string, limit int) ([]SearchResult, bool, error)
  SaveToGraph(ctx Context) error

Your directory: internal/graph/, internal/cache/
Do not create files outside these directories.

**Key design decision you need to make on day 1 — how edges are built:**

When SaveToGraph(ctx) is called, you must auto-link the new context to related existing ones.
Proposed flow (see 07-data-models.md for full detail):
1. Embed the new context
2. Cosine similarity against all existing embeddings → top 3 above 0.75
3. Send each pair to LLM → classify edge type (CAUSED_BY / DEPENDS_ON / etc)
4. Write edges to context_edges

PENDING: decide whether edge type is always LLM-inferred (Option A) or agent can pass it explicitly at save time (Option B). Document your decision in 07-data-models.md before implementing.

Build and test Search() in isolation with seed data before integrating with Section 1.
The demo: Search("tại sao bỏ Kafka") must return a 3-node chain. Search it twice — second call must return cache hit = true.
```

---

## Section 3 — Cloud API + Auth

### What this is

Everything that requires a network connection. Section 1 and 2 work completely offline. Section 3 unlocks cloud features: syncing contexts to the cloud, team sharing, PR bot.

- Go REST API — 4 endpoints: sync, search (cloud), PR comment, create API key
- PostgreSQL — full schema in `07-data-models.md`
- GitHub OAuth flow + JWT sessions
- API key creation (store SHA-256 hash only, plaintext shown once)
- API key validation middleware on every endpoint
- gitwhy-bot GitHub App for PR comments
- Free tier quota enforcement (20 syncs/month)

### Files you own

```
cloud/            ← Go REST API server
cloud/api/        ← endpoint handlers
cloud/db/         ← PostgreSQL client
cloud/github/     ← gitwhy-bot GitHub App
db/migrations/    ← SQL migration files
```

### What you expose to Section 1

A client package Section 1 can import:

```go
// Section 1 calls these — you implement them
func Sync(contexts []Context) (SyncResult, error)
func PostPR(contextID, repo string, prNumber int) (string, error)  // returns comment URL
func ValidateAPIKey(key string) (User, error)
```

### Starting prompt

```
You are building Section 3 of GitWhy: the cloud API and auth.

Read these files first:
- md documents/08-api-specifications.md (Part B — Cloud REST API)
- md documents/07-data-models.md (section 4 — PostgreSQL schema)
- md documents/09-permissions-matrix.md (free vs team limits, quota enforcement)
- md documents/11-security.md (API key hashing, GitHub App permissions)
- md documents/12-error-codes.md (all error codes you must return)

What you are building:
1. Go REST API — endpoints: POST /v1/contexts/sync, GET /v1/contexts/search,
   POST /v1/pr/comment, POST /v1/auth/api-key
2. PostgreSQL — run migrations from 07-data-models.md schema
3. GitHub OAuth — login flow, JWT session
4. API key management — SHA-256 hash storage, shown once on creation
5. Auth middleware — validate Bearer token on every request
6. Free tier quota — 20 syncs/month tracked in sync_quota_usage table
7. gitwhy-bot GitHub App — POST PR comment via GitHub API

Priority order: auth + API keys first (Section 1 needs this to test cloud features),
then sync endpoint, then PR comment, then GitHub OAuth UI.

Register the GitHub App on day 1 — it requires manual setup and has lead time.

Your directory: cloud/, db/migrations/
Do not create files outside these directories.
```

---

## Avoiding Conflicts

All 3 people will be coding with AI in the same repo at the same time. These rules prevent stepping on each other.

### File ownership — hard rule

Each section owns its directories. If you need to create a file outside your directories, ask first.

| Section | Owns |
|---|---|
| 1 — MCP + CLI + Hook | `mcp/`, `cmd/`, `internal/context/`, `internal/mcp/` |
| 2 — Graph + Cache | `internal/graph/`, `internal/cache/` |
| 3 — Cloud + Auth | `cloud/`, `db/` |
| Shared (coordinate before touching) | `go.mod`, `go.sum`, root `main.go` |

### Interfaces first — day 1

Before anyone writes implementation, agree on the function signatures between sections:

```go
// Section 2 → Section 1
Search(query string, domain string, limit int) ([]SearchResult, bool, error)
SaveToGraph(ctx Context) error

// Section 3 → Section 1
Sync(contexts []Context) (SyncResult, error)
PostPR(contextID, repo string, prNumber int) (string, error)
ValidateAPIKey(key string) (User, error)
```

Write these as Go interfaces in a shared `internal/interfaces.go` file. Then each section implements and Section 1 stubs. This way Section 1 can build the full MCP server without waiting for Sections 2 or 3 to finish.

### Shared types

Put shared structs (`Context`, `SearchResult`, `User`, `SyncResult`) in `internal/types.go`. One person creates this file, everyone imports from it. Do not duplicate struct definitions across packages.

### go.mod conflicts

If two people add a dependency at the same time, `go.mod` and `go.sum` will conflict. Coordinate: announce in chat before running `go get`. Merge conflicts in these files are resolved by running `go mod tidy` after accepting both sets of changes.

### Branch strategy

Each section works on its own branch:
```
section/1-mcp-cli-hook
section/2-graph-cache
section/3-cloud-auth
```

Merge into `main` only when a section's demo path works end-to-end. Section 1 merges first (it's the entry point). Sections 2 and 3 merge after.

---

## Demo Checklist — What Must Work

```
[ ] gitwhy_save  → writes .md file to .git/gitwhy/contexts/
[ ] post-commit hook → fires gitwhy_save automatically
[ ] gitwhy_search → returns decision chain from graph.db
[ ] gitwhy_search (repeat) → returns cache hit, no LLM call
[ ] gitwhy_status → shows correct state
[ ] Web dashboard → shows context list + detail view
```

Everything else (sync, publish, PR bot, team features) is secondary to this path.

---

*Updated: 2026-06-20*
