# 15 — Roadmap

---

## v0.1 — Foundation (Đã ship)

**Mục tiêu:** MCP server hoạt động end-to-end với 8 tools. Dev dùng được thật ngay hôm nay.

### Đã hoàn thành ✅

- [x] MCP server (Node.js/stdio) với 8 tools đầy đủ
- [x] `gitwhy_save` — lưu context local với auto-link commit
- [x] `gitwhy_get` — retrieve bằng ID
- [x] `gitwhy_search` — full-text search local contexts
- [x] `gitwhy_list` — browse theo domain/topic
- [x] `gitwhy_status` — check setup state
- [x] `gitwhy_sync` — upload lên cloud (Free: 20/tháng)
- [x] `gitwhy_publish` — share với team (Team plan)
- [x] `gitwhy_post_pr` — post GitHub PR comment qua gitwhy-bot
- [x] Whyspec context format (structured markdown)
- [x] Cloud API (Go) — sync, publish, PR bot
- [x] Web Dashboard — `app.gitwhy.dev` (Next.js)
- [x] Product site — `gitwhy.dev` (Framer)
- [x] npm package: `gitwhy-mcp`
- [x] Homebrew tap: `gitwhy-cli/tap/git-why`
- [x] Scoop bucket (Windows)
- [x] Auth: GitHub OAuth + API keys
- [x] Free tier + Team plan ($20/tháng)
- [x] MCP Registry listing (Glama, Smithery)

---

## v0.2 — Context Graph + Auto-Agent (PRD approved)

**Mục tiêu:** Biến GitWhy thành agent memory layer. Dev không cần nhớ trigger save. Câu hỏi lặp không tốn token.

### Sprint 1 — Auto-save Hook (2 tuần)

- [ ] Extend post-commit hook: `gitwhy hook install`
- [ ] Auto-detect active AI agent session
- [ ] Extract context từ commit message + diff
- [ ] Trigger `gitwhy_save` tự động sau mỗi commit
- [ ] `git config gitwhy.autosave false` để disable per-repo
- [ ] Test với Claude Code, Cursor, Windsurf

### Sprint 2 — Semantic Cache (2 tuần)

- [ ] SQLite cache local tại `.git/gitwhy/cache/semantic.db`
- [ ] Embed query khi search (text-embedding-3-small cho cloud; nomic-embed-text cho local/offline)
- [ ] Cosine similarity lookup: > 90% → return cache
- [ ] Cache TTL: 24h, max 1,000 entries
- [ ] Metrics: cache hit rate, token savings
- [ ] Metrics: cache hit rate, token savings (actual reduction depends on query repetition rate trong codebase)

### Sprint 3 — Context Graph (3 tuần)

- [ ] Embedding generation cho mỗi context khi save
- [ ] Auto-link contexts với similarity > threshold
- [ ] SQLite adjacency table (graph.db) — context_nodes + context_edges với typed edges
- [ ] 2-hop traversal query
- [ ] `gitwhy_search` trả về decision chain (không chỉ list)
- [ ] Demo: "tại sao bỏ Kafka" → chuỗi decision + PR + cost note trong 3s

### Sprint 4 — Web UI Polish + Graph Visualization (1 tuần)

- [ ] Fix font: Inter 14px, line-height 1.6
- [ ] Tăng padding cards (chữ không díu dít nữa)
- [ ] Context graph visualization trên dashboard
- [ ] Cloud graph storage (pgvector indexing)
- [ ] Semantic search thay thế full-text cho cloud contexts

---

## v0.3 — Enterprise & Integrations (3 tháng)

- [ ] GitLab PR integration (gitwhy-bot GitLab edition)
- [ ] Bitbucket PR integration
- [ ] Jira integration: link context với Jira ticket
- [ ] Slack integration: `/gitwhy search` command
- [ ] Self-hosted cloud backend (Docker Compose package)
- [ ] SSO (GitHub Team / SAML)
- [ ] Audit logs (enterprise)
- [ ] Context export (JSON, CSV)
- [ ] API rate increase (enterprise)

---

## v1.0 — Platform

- [ ] Context Graph trên cloud (multi-team, cross-repo)
- [ ] AI-powered context summarization
- [ ] Context health score cho team (knowledge coverage)
- [ ] Analytics dashboard (decision patterns, common topics)
- [ ] Webhook notifications
- [ ] VS Code extension (sidebar view)
- [ ] GitHub Action: auto post-pr khi PR opened

---

## Priorities

| Priority | Item | Version | Status |
|----------|------|---------|--------|
| P0 | 8 MCP tools end-to-end | v0.1 | ✅ Done |
| P0 | Cloud sync + PR bot | v0.1 | ✅ Done |
| P0 | Web dashboard | v0.1 | ✅ Done |
| P0 | npm + Homebrew distribution | v0.1 | ✅ Done |
| P1 | Auto-save hook | v0.2 | 🔄 Sprint 1 |
| P1 | Semantic cache | v0.2 | 🔄 Sprint 2 |
| P1 | Context Graph | v0.2 | 🔄 Sprint 3 |
| P1 | Web UI polish | v0.2 | 🔄 Sprint 4 |
| P2 | GitLab integration | v0.3 | — |
| P2 | Self-hosted | v0.3 | — |
| P3 | SSO / SAML | v0.3 | — |
| P3 | Analytics | v1.0 | — |

---

## Demo Targets (v0.2)

**Hackathon demo script:**
1. Mở repo thật với nhiều commits
2. Hỏi: `"tại sao bỏ Kafka"`
3. GitWhy trả về: Context decision → PR diff → cost note
4. **Thời gian: 3 giây**
5. **Bill: $0.01**
6. Lần 2 hỏi câu tương tự → cache hit → **$0.00**

---

*Cập nhật lần cuối: 2026-06-20*
