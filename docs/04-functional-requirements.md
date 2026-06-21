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
- Agents hiện tốn 2:1 input-to-output token ratio cho context. Cache eliminates redundant LLM calls — actual reduction depends on query repetition rate trong codebase cụ thể

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
- Sau mỗi commit: trigger `gitwhy_save` automatically — dev không cần nhớ

**Hook captures automatically:**
- Prompt (pulled từ agent context / last user message)
- Files changed trong commit
- HEAD commit hash

**Agent must still provide:**
- `reasoning` — tại sao approach này
- `decisions` — đã chọn gì
- `rejected_alternatives` — đã cân nhắc và bỏ gì

Hook reduces friction — không thể manufacture reasoning agent chưa produce. Agent vẫn phải generate 3 fields đó. Hook chỉ đảm bảo `gitwhy_save` fires mà không cần developer manually invoke.

- Dev có thể disable per-repo: `git config gitwhy.autosave false`

---

## §GRAPH — Context Graph (v0.2)

### GRAPH-01: Auto-link contexts

**Graph Storage:** SQLite với adjacency table tại `.git/gitwhy/graph.db`:

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

**Typed Edges:**

| Edge Type | Nghĩa |
|---|---|
| `CAUSED_BY` | Decision A trực tiếp trigger decision B |
| `CONSTRAINED_BY` | Decision B phải hoạt động trong bounds của A |
| `INVALIDATES` | Decision B làm decision A obsolete |
| `CONTRADICTS` | B conflict với A — fires alert |
| `DEPENDS_ON` | B chỉ đúng nếu assumption trong A vẫn còn valid |

**Tại sao typed edges, không phải similarity links:** RAG trả lời "cái gì similar với X?" Graph trả lời "cái gì downstream của X?" và "cái gì break nếu X thay đổi?" Similarity search không thể answer dependency queries. Graph traversal có thể.

**Edge Linking — how edges are built at save time (PENDING DECISION):**

Mỗi khi `gitwhy_save` được gọi, graph phải auto-link context mới với related contexts. Flow đề xuất:

1. Embed context mới (text-embedding-3-small)
2. Cosine similarity so với tất cả existing embeddings trong graph.db
3. Lấy top 3 matches có similarity > 0.75
4. Gửi context mới + top 3 matches lên LLM — prompt: "what is the causal relationship between these decisions?"
5. LLM trả về edge type: `CAUSED_BY` / `DEPENDS_ON` / `CONTRADICTS` / etc
6. Write edge vào context_edges

**Cost per save:** 1 embedding call + 1 small LLM classification call (~500 tokens)

**PENDING:** Edge type classification có thể được LLM tự suy luận từ content, hoặc agent cung cấp explicitly khi gọi `gitwhy_save`. Cần quyết định trước khi implement Section 2.

Behavior:
- Sau mỗi `gitwhy_save`: compute embedding cho context, insert node vào graph.db
- So sánh với existing embeddings → top 3 candidates (cosine similarity > 0.75)
- LLM classify edge type cho từng candidate pair
- Insert typed edges vào context_edges table

### GRAPH-02: Graph traversal query

Input: `{ "query": "tại sao đổi AWS", "hops": 2 }`

Behavior:
- Search context matching query → root node
- Traverse edges tối đa 2-hop
- Return: ordered chain of contexts (decision trail)
- Response format: `[{ context, relationship, timestamp }]`

---

*Cập nhật lần cuối: 2026-06-20*
