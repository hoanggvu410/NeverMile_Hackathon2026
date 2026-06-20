# 06 — System Architecture

---

## 1. Kiến trúc tổng thể

```
┌─────────────────────────────────────────────────────────────────┐
│  AI Coding Agents (Claude Code / Cursor / Windsurf / Cline...)   │
│  Kết nối qua MCP stdio protocol                                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │ MCP stdio
               ┌───────────────▼────────────────┐
               │      gitwhy-mcp (Node.js)        │
               │      Local stdio MCP Server      │
               │      :stdio (không phải HTTP)    │
               │                                  │
               │  ┌──────────────────────────┐   │
               │  │  Local Context Storage    │   │
               │  │  .git/gitwhy/contexts/   │   │
               │  │  .git/gitwhy/graph.json  │   │
               │  └──────────────────────────┘   │
               └───────────────┬────────────────┘
                               │ HTTPS (khi có API key)
               ┌───────────────▼────────────────┐
               │   GitWhy Cloud API (Go)          │
               │   api.gitwhy.dev                 │
               │                                  │
               │  ┌──────────────────────────┐   │
               │  │  PostgreSQL (contexts)    │   │
               │  └──────────────────────────┘   │
               │  ┌──────────────────────────┐   │
               │  │  Redis (semantic cache)   │   │
               │  └──────────────────────────┘   │
               └───────────────┬────────────────┘
                               │
               ┌───────────────▼────────────────┐
               │   GitHub App (gitwhy-bot)        │
               │   Posts PR comments              │
               └─────────────────────────────────┘

               ┌─────────────────────────────────┐
               │   Web Dashboard (Next.js)         │
               │   app.gitwhy.dev                  │
               │   Search, view, manage contexts   │
               └─────────────────────────────────┘
```

**Nguyên tắc kiến trúc:**
- **Offline-first**: local features không phụ thuộc cloud
- **Cloud là optional**: sync, publish, PR bot cần API key
- **AI agent là primary client** của MCP server, không phải human

---

## 2. Request Flow — Save Context

```
AI Agent (Claude Code)
  │  Gọi gitwhy_save({ prompt, reasoning, decisions, ... })
  │  via MCP stdio
  ▼
gitwhy-mcp handler
  │
  ├─ [1] Validate input schema (required fields)
  ├─ [2] Auto-detect git commit hash: exec("git rev-parse HEAD")
  ├─ [3] Generate context ID: cxt_{timestamp}_{nanoid}
  ├─ [4] Render whyspec markdown (structured template)
  ├─ [5] Write to .git/gitwhy/contexts/{id}.md
  └─ [6] Return { id, timestamp, linked_commits }
  ▼
AI Agent nhận response, tiếp tục workflow
```

---

## 3. Request Flow — Search + Graph Traversal (v0.2)

```
AI Agent
  │  gitwhy_search({ query: "tại sao bỏ Kafka" })
  ▼
gitwhy-mcp
  │
  ├─ [1] Check semantic cache: embed query → cosine sim > 90%?
  │       → Cache HIT: return cached result (< 50ms, $0)
  │
  ├─ [2] Cache MISS: full-text search local contexts
  │
  ├─ [3] Nếu có API key + contexts synced:
  │       → Query cloud API cho team contexts
  │
  ├─ [4] Context Graph (v0.2):
  │       → Load graph.json
  │       → Find matching root node
  │       → Traverse edges (max 2-hop)
  │       → Build decision chain
  │
  ├─ [5] Rank results (local score + graph relevance)
  ├─ [6] Store query+result in semantic cache
  └─ [7] Return ranked context list
```

---

## 4. Request Flow — Post PR Comment

```
AI Agent
  │  gitwhy_post_pr({ context_id, repo: "org/repo", pr_number: 42 })
  ▼
gitwhy-mcp
  │
  ├─ [1] Load context từ local storage
  ├─ [2] Format PR comment markdown (prompt, decisions, rejected_alternatives)
  │
  ├─ [3] Call GitWhy Cloud API:
  │       POST /v1/pr/comment { context, repo, pr_number, api_key }
  │
  │       Cloud API:
  │       ├─ Verify API key + plan (Free OK for PR comments)
  │       ├─ Call GitHub API via gitwhy-bot App installation token
  │       └─ POST https://api.github.com/repos/{repo}/issues/{pr}/comments
  │
  └─ [4] Return { comment_url, pr_url }
```

---

## 5. Module Structure — gitwhy-mcp (Node.js)

```
gitwhy-mcp/
├── package.json
├── install.js          ← post-install: detect platform, download Go binary
├── run.js              ← entrypoint: spawn gitwhy CLI as MCP server
├── server.json         ← MCP server manifest (tool schemas)
├── checksums.json      ← binary checksums for verification
└── README.md

# Go binary (downloaded at install time)
~/.local/bin/gitwhy (or PATH equivalent)
```

**Lý do kiến trúc này (Node.js wrapper + Go binary):**
- npm là phân phối method đơn giản nhất cho MCP clients
- Logic thực tế chạy trong Go binary (performance, cross-platform)
- `run.js` spawn Go binary, bridge stdio

---

## 6. Module Structure — Go CLI / Cloud Backend

```
gitwhy/
├── cmd/
│   ├── root.go         ← cobra root command
│   ├── save.go         ← gitwhy save (MCP tool + CLI)
│   ├── search.go       ← gitwhy search
│   ├── sync.go         ← gitwhy sync
│   ├── publish.go      ← gitwhy publish
│   ├── post-pr.go      ← gitwhy post-pr
│   ├── status.go       ← gitwhy status
│   └── setup.go        ← gitwhy setup (auth)
├── internal/
│   ├── context/
│   │   ├── store.go    ← local file storage (.git/gitwhy/)
│   │   ├── whyspec.go  ← parse/render whyspec markdown
│   │   └── graph.go    ← context graph (v0.2)
│   ├── cloud/
│   │   ├── client.go   ← HTTP client cho cloud API
│   │   └── auth.go     ← API key management
│   ├── cache/
│   │   └── semantic.go ← semantic cache (v0.2)
│   └── mcp/
│       └── server.go   ← MCP stdio server (mark3labs/mcp-go)
└── cloud/              ← Cloud API server (separate binary)
    ├── api/
    │   ├── contexts.go
    │   ├── sync.go
    │   ├── publish.go
    │   └── pr.go
    ├── db/
    │   └── postgres.go
    └── github/
        └── bot.go      ← gitwhy-bot GitHub App integration
```

---

## 7. Local Storage Schema

```
.git/gitwhy/
├── contexts/
│   ├── cxt_20260620_abc123.md   ← individual context files (whyspec)
│   ├── cxt_20260619_def456.md
│   └── ...
├── graph.json                    ← context graph edges (v0.2)
├── cache/
│   └── semantic.db               ← local semantic cache (SQLite, v0.2)
└── sync.json                     ← sync state (which contexts uploaded)
```

---

## 8. Cloud Infrastructure

| Service | Technology | Ghi chú |
|---------|-----------|--------|
| API Server | Go (Fiber/Chi) | api.gitwhy.dev |
| Database | PostgreSQL 15 | Contexts, users, teams |
| Cache | Redis 7 | Semantic query cache |
| Object Storage | S3-compatible | Context exports (nếu cần) |
| GitHub App | gitwhy-bot | PR comment integration |
| CDN | Cloudflare | gitwhy.dev static site |
| Hosting | Fly.io / Railway | Go API server |

---

*Cập nhật lần cuối: 2026-06-20*
