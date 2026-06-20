# 04 — Functional Requirements

---

## §SAVE — gitwhy_save

### SAVE-01: Lưu context mới

**MCP tool:** `gitwhy_save`

Input schema:
```json
{
  "prompt":                "string — original prompt given to AI agent",
  "reasoning":             "string — agent's explanation of its approach",
  "decisions":             "string — key choices made with rationale",
  "rejected_alternatives": "string — options considered and discarded",
  "files":                 ["array of strings — affected source files"],
  "commits":               ["array of strings — linked git commit hashes, optional"],
  "domain":                "string — e.g. 'backend/auth'",
  "topic":                 "string — e.g. 'jwt-migration'"
}
```

Behavior:
- Assign unique ID (`cxt_{timestamp}_{random}`)
- Persist to `.git/gitwhy/contexts/{id}.md` (structured markdown per whyspec)
- Auto-detect current git commit hash nếu `commits` không được cung cấp
- Return `{ id, timestamp, linked_commits }` trên success
- Hoạt động offline — không cần API key cho local save

### SAVE-02: Format lưu trữ (whyspec)

```markdown
---
id: cxt_20260620_abc123
prompt: "Migrate từ JWT HS256 sang RS256"
agent: claude-code
model: claude-opus-4-6
timestamp: 2026-06-20T10:30:00Z
commits: ["a1b2c3d"]
files: ["app/core/security.py", "app/core/deps.py"]
domain: backend/auth
topic: jwt-migration
---

## Reasoning
Cần asymmetric key để services khác verify token mà không cần private key.

## Decisions
- Chọn RS256: public key có thể distribute an toàn
- Key size 2048-bit: đủ secure, không quá chậm

## Rejected Alternatives
- HS256: symmetric key phải share với tất cả services → security risk
- ES256: cần thêm library, team chưa quen

## Trade-offs
RS256 sign chậm hơn HS256 ~10x nhưng verify nhanh → acceptable với access token TTL 15 phút
```

---

## §GET — gitwhy_get

### GET-01: Retrieve context bằng ID

**MCP tool:** `gitwhy_get`

Input: `{ "id": "cxt_20260620_abc123" }`

Behavior:
- Tìm file trong `.git/gitwhy/contexts/{id}.md`
- Parse structured markdown → return full context object
- Nếu không tìm thấy local → query cloud (nếu có API key + synced)
- Return toàn bộ fields bao gồm reasoning, decisions, rejected_alternatives

---

## §SEARCH — gitwhy_search

### SEARCH-01: Tìm kiếm contexts

**MCP tool:** `gitwhy_search`

Input: `{ "query": "string", "domain": "optional string", "limit": "optional int (default 5)" }`

Behavior:
- Full-text search trên local contexts (prompt + reasoning + decisions + topic)
- Nếu có API key + synced → cũng search trên cloud contexts của team
- Return danh sách contexts sorted by relevance score
- Trả về trong < 3s (target demo)

### SEARCH-02: Semantic Cache (v0.2)

- Cache query embedding + result tại local
- Câu hỏi mới có cosine similarity > 90% với cached query → return cached result ngay
- Không gọi LLM → 0 token cost
- Cache TTL: 24h (configurable)
- Mục tiêu: giảm 80% token cost cho repeated queries

---

## §LIST — gitwhy_list

### LIST-01: Browse contexts theo domain/topic

**MCP tool:** `gitwhy_list`

Input: `{ "domain": "optional string", "topic": "optional string", "limit": "optional int" }`

Behavior:
- List contexts từ local storage + cloud (nếu synced)
- Filter theo domain và/hoặc topic hierarchy
- Return: `[{ id, timestamp, prompt_preview, domain, topic, linked_commits }]`

---

## §STATUS — gitwhy_status

### STATUS-01: Kiểm tra setup

**MCP tool:** `gitwhy_status`

Behavior:
- Check working directory có phải git repo không
- Check API key hợp lệ (nếu có)
- Count pending contexts chưa sync
- Return sync status: `{ is_git_repo, has_api_key, pending_sync_count, last_sync_at, plan }`

---

## §SYNC — gitwhy_sync

### SYNC-01: Upload contexts lên cloud

**MCP tool:** `gitwhy_sync`

Behavior:
- Require API key (Free tier: 20 syncs/tháng; Team: unlimited)
- Upload tất cả pending contexts (chưa sync)
- Contexts private by default — không ai trong team thấy cho đến khi publish
- Return: `{ synced_count, failed_count, errors }`
- Idempotent: sync lại context đã sync → skip (không trừ quota)

---

## §PUBLISH — gitwhy_publish

### PUBLISH-01: Share contexts với team

**MCP tool:** `gitwhy_publish`

Input: `{ "context_id": "optional — nếu null thì publish tất cả synced" }`

Behavior:
- Require Team plan API key
- Mark contexts là `published = true` trên cloud
- Từ thời điểm này, team members có thể search + view contexts này
- Return: `{ published_count, context_ids }`

---

## §POST_PR — gitwhy_post_pr

### POST_PR-01: Post comment lên GitHub PR

**MCP tool:** `gitwhy_post_pr`

Input: `{ "context_id": "string", "repo": "owner/repo", "pr_number": "int" }`

Behavior:
- Require API key + GitHub repo access (via gitwhy-bot GitHub App)
- Format context thành PR comment với sections:
  - 🧠 **Original Prompt**
  - ✅ **Key Decisions**
  - ❌ **Rejected Alternatives**
  - 📝 **Linked Commits**
  - 🔗 Link về app.gitwhy.dev/contexts/{id} để xem full context
- Post comment qua GitHub API dùng gitwhy-bot identity
- Return: `{ comment_url, pr_url }`
- Free tier: unlimited PR comments

---

## §HOOK — Auto-save Hook (v0.2)

### HOOK-01: Post-commit hook auto-trigger

Behavior:
- Extend `.git/hooks/post-commit` khi install
- Sau mỗi commit: detect nếu AI agent (Claude Code, Cursor...) đang active trong session
- Nếu có pending reasoning từ session → trigger `gitwhy_save` automatically
- Context prompt extracted từ: last user message / commit message / diff summary
- Dev có thể disable per-repo: `git config gitwhy.autosave false`

---

## §GRAPH — Context Graph (v0.2)

### GRAPH-01: Auto-link contexts

Behavior:
- Sau mỗi `gitwhy_save`: compute embedding cho context
- So sánh với existing contexts trong project
- Nếu similarity > threshold → tạo edge A→B trong local graph
- Graph lưu tại `.git/gitwhy/graph.json`

### GRAPH-02: Graph traversal query

Input: `{ "query": "tại sao đổi AWS", "hops": 2 }`

Behavior:
- Search context matching query → root node
- Traverse edges tối đa 2-hop
- Return: ordered chain of contexts (decision trail)
- Response format: `[{ context, relationship, timestamp }]`

---

*Cập nhật lần cuối: 2026-06-20*
