# Section 2 — Context for New Session

## Who you are

You are building Section 2 of GitWhy: the Context Graph + Semantic Cache.
This is `internal/graph/` and `internal/cache/`. Both directories are currently empty.

## What GitWhy is

Agent memory layer. AI coding agents are stateless — every session starts from zero.
GitWhy saves decisions + reasoning after every git commit, then makes them queryable
via a causal graph so agents don't re-discover what already happened.

One-liner: "GitWhy lets AI agents query why decisions were made."
Tagline: "GitWhy = why.log, not git.log"

## What Section 1 built (already done)

- MCP server with 8 tools: `internal/mcp/server.go`
- Local file storage: `internal/context/store.go`
- Whyspec format (structured .md files): `internal/context/whyspec.go`
- CLI (Cobra): `cmd/git-why/`
- Post-commit hook: `cmd/git-why/hook.go`

Section 1's `Search()` is currently full-text substring scan — no graph, no embeddings.
Your job is to replace it with graph traversal + semantic cache.

## What you need to build

### internal/graph/ — SQLite graph.db

Two tables (schema in `md documents/07-data-models.md` section 2):

```sql
CREATE TABLE context_nodes (
  id TEXT PRIMARY KEY,
  context_id TEXT,
  domain TEXT,
  topic TEXT,
  embedding BLOB
);

CREATE TABLE context_edges (
  from_id TEXT REFERENCES context_nodes(id),
  to_id TEXT REFERENCES context_nodes(id),
  edge_type TEXT CHECK(edge_type IN (
    'CAUSED_BY','CONSTRAINED_BY','INVALIDATES','CONTRADICTS','DEPENDS_ON'
  )),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (from_id, to_id, edge_type)
);
```

### internal/cache/ — SQLite semantic.db

Query deduplication. Cosine similarity > 90% → cache hit → return instantly, $0.00.
TTL: 24h, max 1,000 entries.

### Two functions you expose to Section 1

```go
// Section 1 calls this inside gitwhy_search
Search(query string, domain string, limit int) ([]SearchResult, bool, error)
// bool = true if cache hit

// Section 1 calls this after writing the .md file in gitwhy_save
SaveToGraph(ctx Context) error
```

## Edge linking algorithm (the core decision — DECIDED: Option A)

When `SaveToGraph(ctx)` is called:

1. Embed new context (text-embedding-3-small for cloud, nomic-embed-text for offline)
2. Cosine similarity against all existing embeddings in graph.db
3. Take top 3 matches with similarity > 0.75
4. Send new context + top 3 to LLM with prompt:
   "Given these two decisions, what is the causal relationship?
    Choose exactly one: CAUSED_BY / CONSTRAINED_BY / INVALIDATES / CONTRADICTS / DEPENDS_ON / NONE"
5. Write edge to context_edges for each non-NONE result
6. If no candidates above 0.75 → save as isolated node, no edges (normal for first few saves)

Cost per save: ~1,500 tokens (1 embedding + up to 3 LLM classification calls)

## Traversal query (2-hop recursive CTE)

```sql
WITH RECURSIVE chain AS (
  SELECT to_id, edge_type, 1 AS depth
  FROM context_edges WHERE from_id = ?
  UNION ALL
  SELECT e.to_id, e.edge_type, c.depth + 1
  FROM context_edges e
  JOIN chain c ON e.from_id = c.to_id
  WHERE c.depth < 2
)
SELECT n.*, c.edge_type, c.depth
FROM chain c JOIN context_nodes n ON n.id = c.to_id;
```

## The demo that must work

```
Search("tại sao bỏ Kafka")
→ returns 3-node decision chain in ~3s, ~$0.01

Search("why did we remove kafka")  (same query, different wording)
→ cosine similarity 0.94 → cache hit → <50ms, $0.00
```

## Integration point with Section 1

One change needed in `internal/mcp/server.go` line 179:

```go
// Current (full-text only):
results, err := s.store.Search(query)

// Replace with:
results, cacheHit, err := graph.Search(query, domain, limit)
```

## Files to read before starting

- `md documents/07-data-models.md` — SQLite schemas (sections 2 + 3)
- `md documents/04-functional-requirements.md` — §SEARCH, §GRAPH
- `md documents/06-system-architecture.md` — search flow diagram
- `internal/context/types.go` — Context struct you'll be working with
- `internal/mcp/server.go` — where your Search() gets called
- `CLAUDE.md` — full repo architecture

## Directories you own

```
internal/graph/    ← build here
internal/cache/    ← build here
```

Do NOT create files outside these directories without asking.

## Module

```
github.com/hoanggvu410/NeverMile_Hackathon2026
```

Go 1.26.4. SQLite: use `modernc.org/sqlite` (pure Go, no CGO needed).
