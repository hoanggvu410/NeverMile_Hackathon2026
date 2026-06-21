# GitWhy hoạt động thế nào

## Đọc nhanh trong 90 giây

Ví dụ agent vừa đổi backend từ Express sang Go và save context:

```text
Decision:
- Use Go for the Cloud API. Do not add Express.

Reasoning:
- Go already powers the CLI and serve binary, so one runtime keeps deployment simpler.

Rejected alternative:
- Express rejected because it adds a second backend runtime and duplicate auth middleware.
```

GitWhy làm 6 việc:

1. Ghi nguyên context thành markdown trong `.git/gitwhy/contexts/...`.
2. Lưu full markdown đó vào bảng `sessions`.
3. Tách tối đa 7 câu quan trọng thành `claims`.
4. Với mỗi claim, tạo 3 embeddings: `claim`, `retrieval`, `interrupt`.
5. Nối claim bằng typed edges như `CAUSED_BY` và `CONFLICTS_WITH`.
6. Khi agent sau này muốn thêm Express, `gitwhy_tripwire` tìm lại claim cũ và cảnh báo trước khi sửa file.

Kết quả thực tế trong graph sẽ giống:

```text
sessions
  ctx_backend_runtime

claims
  clm_001: Use Go for the Cloud API. Do not add Express.
  clm_002: Go already powers the CLI and serve binary, so one runtime keeps deployment simpler.
  clm_003: Express rejected because it adds a second backend runtime and duplicate auth middleware.

claim_vectors
  clm_001 / claim
  clm_001 / retrieval
  clm_001 / interrupt
  clm_002 / claim
  ...

edges
  clm_001 -[CAUSED_BY]-> clm_002
  clm_001 -[CONFLICTS_WITH]-> clm_003
```

Điểm cần nhớ: markdown là archive, claim graph là index để search/tripwire.

## Tool nào gọi code nào?

| Entry point | Người dùng/agent gọi | Code path chính | Output |
|---|---|---|---|
| CLI save | `git why save` | `cmd/git-why/save.go` -> `Store.Save` -> `Graph.SaveToGraph` | Markdown context + graph rows. |
| MCP save | `gitwhy_save(...)` | `internal/mcp/server.go:handleSave` -> `Store.Save` -> `Graph.SaveToGraph` | JSON result có context id. |
| CLI search | `git why search "query"` | `cmd/git-why/search.go` -> `Store.Search` | CLI keyword search đơn giản. |
| MCP search | `gitwhy_search(query)` | `handleSearch` -> `Graph.Search` -> fallback `Store.Search` nếu được phép | Claim graph results hoặc markdown fallback. |
| MCP tripwire | `gitwhy_tripwire(agent_plan_created)` | `handleTripwire` -> `Graph.CheckTripwire` | `interrupt=true/false` + candidates + telemetry. |
| Reindex | `git why reindex` | `cmd/git-why/reindex.go` -> đọc markdown -> `SaveToGraph` | Rebuild graph từ markdown store. |

Trong hackathon demo, phần quan trọng nhất là MCP path: agent gọi `gitwhy_save`, `gitwhy_search`, và `gitwhy_tripwire`.

## Ý tưởng cốt lõi

AI coding agents thường không nhớ lý do của các quyết định cũ. GitWhy lưu phần "why" đó thành dữ liệu có cấu trúc để agent sau này có thể hỏi lại trước khi sửa code.

GitWhy không chỉ lưu một đoạn note dài. Nó lưu hai tầng:

| Tầng | Nằm ở đâu | Dùng để làm gì |
|---|---|---|
| `sessions` | `.git/gitwhy/contexts/.../*.md` và bảng `sessions` trong `graph.db` | Bản ghi đầy đủ của một lần save. Đây là archive và bằng chứng gốc. |
| `claims` | bảng `claims` trong `.git/gitwhy/graph.db` | Các câu quyết định quan trọng đã được rút ra từ session. Đây là đơn vị search/tripwire chính. |

Nói ngắn gọn:

```text
session = nguyên câu chuyện
claim = câu quyết định có thể search và cảnh báo
edge = quan hệ giữa các claim
claim_vector = embedding để so sánh semantic
```

## File và bảng chính

GitWhy lưu dữ liệu local trong repo đang dùng:

```text
.git/gitwhy/
  contexts/          markdown contexts
  graph.db           claim graph + vectors + interrupt events
  cache/semantic.db  semantic cache cho search/tripwire
```

Trong `graph.db`, các bảng quan trọng là:

| Bảng | Vai trò |
|---|---|
| `sessions` | Lưu full markdown của context. |
| `claims` | Lưu tối đa 7 claim được rút ra từ mỗi session. |
| `claim_vectors` | Lưu 3 embedding rows cho mỗi claim: `claim`, `retrieval`, `interrupt`. |
| `edges` | Lưu quan hệ giữa claim với claim. |
| `interrupt_events` | Log các lần tripwire chạy và candidate được trả về. |
| `embedding_config` | Ghi provider/dimensions để graph không bị trộn embedding không tương thích. |

Legacy tables `context_nodes` và `context_edges` vẫn tồn tại để mở được graph cũ, nhưng đường chạy mới dùng `sessions`, `claims`, `claim_vectors`, và `edges`.

## Lúc save: chuyện gì xảy ra?

Save có thể đi qua CLI hoặc MCP tool:

```text
git why save
gitwhy_save(...)
```

Đường chạy chính nằm ở:

- `internal/context/store.go`
- `internal/context/whyspec.go`
- `internal/graph/graph.go`
- `internal/graph/claims.go`

### 1. Agent gửi context có cấu trúc

Một context tốt có các field như:

- `prompt`
- `reasoning`
- `decisions`
- `rejected_alternatives`
- `risks`
- `what_was_done`
- `files`
- `domain`
- `topic`

Agent viết phần reasoning bằng token của nó. GitWhy không tự suy luận toàn bộ câu chuyện từ code diff.

### 2. Store ghi markdown

`Store.Save` ghi context thành markdown trong:

```text
.git/gitwhy/contexts/<domain>/<topic>/<id>.md
```

Đây là bản đầy đủ để con người đọc lại. Search graph không thay thế markdown; graph chỉ là index thông minh.

### 3. Graph lưu session

`Graph.SaveToGraph` render context thành markdown rồi lưu vào bảng `sessions`.

Ví dụ session row:

```text
sessions.id            = ctx_backend_runtime
sessions.project_id    = NeverMile_Hackathon2026
sessions.domain        = backend/api
sessions.topic         = runtime-choice
sessions.title         = Backend runtime decision
sessions.prompt        = Replace Express plan with Go API
sessions.full_markdown = toàn bộ whyspec markdown
```

Nếu cùng `session_id` được re-save, code xóa claim cũ của session đó rồi tạo lại claim/vector/edge mới để graph không giữ bản stale.

### 4. GitWhy tách claim

`extractClaims` đọc các section theo thứ tự ưu tiên:

| Section | Importance |
|---|---:|
| `Key Decisions` | 5 |
| `Rejected Alternatives` | 4 |
| `Risks & Open Questions` | 4 |
| `What Was Done` | 3 |
| `Reasoning` | 3 |

Mỗi dòng durable được phân loại thành claim type như:

- `decision`
- `constraint`
- `design_constraint`
- `design_decision`
- `architecture_decision`
- `rejected_alternative`
- `risk`
- `implementation`
- `rationale`

Ví dụ cách phân loại:

| Text | Section | Type GitWhy gán |
|---|---|---|
| `Use Go for Cloud API. Do not add Express.` | `Key Decisions` | `architecture_decision` hoặc `constraint`, tùy keyword. |
| `Express rejected because it adds duplicate auth middleware.` | `Rejected Alternatives` | `rejected_alternative` |
| `If another backend runtime is added, deployment scripts diverge.` | `Risks & Open Questions` | `risk` |
| `Updated cloud/api/router.go and cloud/main.go.` | `What Was Done` | `implementation` |
| `One runtime keeps deploy simpler.` | `Reasoning` | `rationale` |

Sau đó GitWhy:

1. dedupe claim gần giống nhau
2. sort theo `importance`
3. giữ tối đa `7` claim cho mỗi session

Nếu một session có 20 dòng, chỉ 7 dòng có `importance` cao nhất được đưa vào graph. Full markdown vẫn còn trong `sessions.full_markdown` và file markdown.

## Claim không chỉ là text

Mỗi claim được enrich bằng metadata để search và tripwire tốt hơn:

| Field | Ý nghĩa |
|---|---|
| `scope_json` | Component, concept, file pattern liên quan. |
| `aliases_json` | Cách gọi khác của cùng khái niệm. |
| `retrieval_triggers_json` | Tình huống tương lai nên kéo claim này ra khi search. |
| `blast_radius_json` | Vùng có thể bị ảnh hưởng nếu claim bị đổi. |
| `interrupt_conditions_json` | Tín hiệu cho thấy plan mới có thể đụng quyết định cũ. |

Các field này hiện được suy ra bằng rule/keyword logic trong `claims.go`, không phải bằng LLM.

Ví dụ claim:

```text
Use Go for the Cloud API. Do not add Express.
```

Metadata có thể được infer thành:

```json
{
  "scope_json": {
    "components": ["backend", "api"],
    "concepts": ["api", "server", "cli"],
    "files": ["cloud/**", "cmd/**"]
  },
  "aliases_json": ["go api", "cloud api", "backend server"],
  "retrieval_triggers_json": [
    "when changing backend API routes",
    "when adding a server framework"
  ],
  "interrupt_conditions_json": [
    "plan adds Express",
    "plan changes backend framework"
  ]
}
```

Exact strings phụ thuộc keyword rules trong `inferScope`, `inferAliases`, `inferRetrievalTriggers`, và `inferInterruptConditions`.

## Vì sao mỗi claim có 3 vector?

Mỗi claim tạo đúng 3 rows trong `claim_vectors`:

| `kind` | Text được embed | Dùng tốt nhất khi |
|---|---|---|
| `claim` | Chính câu claim | User hỏi thẳng về quyết định. |
| `retrieval` | `aliases` + `retrieval_triggers` | User mô tả tình huống liên quan nhưng không dùng đúng wording ban đầu. |
| `interrupt` | `interrupt_conditions` + `blast_radius` | Agent đưa plan có nguy cơ vi phạm quyết định cũ. |

Ví dụ cùng một claim tạo 3 text khác nhau để embed:

```text
kind=claim
Use Go for the Cloud API. Do not add Express.

kind=retrieval
go api
cloud api
backend server
when changing backend API routes
when adding a server framework

kind=interrupt
plan adds Express
plan changes backend framework
cloud/**
cmd/**
```

Search thường match `claim` hoặc `retrieval`. Tripwire thường match `interrupt`.

## Embedding chạy bằng gì?

GitWhy có hai provider:

| Provider | Dims | Khi nào dùng |
|---|---:|---|
| `local-hash-v1` | 384 | Default khi không có `OPENAI_API_KEY`, dùng được offline. |
| `openai:text-embedding-3-small` | 1536 | Dùng khi `OPENAI_API_KEY` có sẵn hoặc `GITWHY_EMBEDDING_PROVIDER=openai`. |

Provider decision cụ thể:

```text
if GITWHY_EMBEDDING_PROVIDER=openai:
  dùng OpenAI text-embedding-3-small, cần OPENAI_API_KEY

if GITWHY_EMBEDDING_PROVIDER=local:
  dùng local-hash-v1

if env không set và OPENAI_API_KEY tồn tại:
  dùng OpenAI

if env không set và không có OPENAI_API_KEY:
  dùng local-hash-v1
```

`local-hash-v1` tokenize text thành token lowercase, hash token vào vector 384 dimensions, thêm bigram weight, rồi normalize vector. Nó không thông minh bằng embedding model thật, nhưng đủ deterministic cho offline demo/test.

Graph ghi provider và dimensions vào `embedding_config`. Nếu graph đã có vector từ provider khác, code sẽ báo mismatch và yêu cầu `reindex`, thay vì trộn vector không cùng hệ.

## Edge: GitWhy nối các claim thế nào?

Edge là quan hệ giữa hai claim.

Các edge type chính:

| Edge | Ý nghĩa |
|---|---|
| `IMPLEMENTS` | Implementation claim thực thi một decision. |
| `CONSTRAINS` | Constraint giới hạn implementation. |
| `CAUSED_BY` | Decision có rationale đi kèm. |
| `CONFLICTS_WITH` | Decision xung đột với rejected alternative. |
| `RELATED_CANDIDATE` | Claim ở session khác có vẻ liên quan theo semantic similarity. |
| `SUPERSEDES` | Claim mới thay thế claim cũ, dùng khi edge hint chỉ rõ. |

Same-session edges được tạo bằng rule:

```text
implementation -> decision-like       = IMPLEMENTS
constraint-like -> implementation     = CONSTRAINS
decision-like -> rationale            = CAUSED_BY
decision-like -> rejected_alternative = CONFLICTS_WITH
```

Confidence/source/status hiện dùng như sau:

| Edge source | Type | Confidence | Status |
|---|---|---:|---|
| Same session implementation -> decision | `IMPLEMENTS` | `0.82` | `active` |
| Same session constraint -> implementation | `CONSTRAINS` | `0.78` | `active` |
| Same session decision -> rationale | `CAUSED_BY` | `0.74` | `active` |
| Same session decision -> rejected alternative | `CONFLICTS_WITH` | `0.74` | `active` |
| Cross session semantic similarity | `RELATED_CANDIDATE` | candidate score | `candidate` |
| Explicit MCP `edge_hints` | normalized edge type | `0.90` | `active` |

Cross-session related edge có hai threshold:

```text
retrieve candidates with minScore = 0.12
only save RELATED_CANDIDATE if final candidate.score >= 0.18
```

Nên `RELATED_CANDIDATE` nghĩa là "có vẻ liên quan", không phải quyết định chắc chắn.

Legacy edge names được normalize:

| Legacy | Current |
|---|---|
| `CONSTRAINED_BY` | `CONSTRAINS` |
| `INVALIDATES` | `CONFLICTS_WITH` |
| `CONTRADICTS` | `CONFLICTS_WITH` |
| `DEPENDS_ON` | `CAUSED_BY` |

## Search: khi user hỏi thì sao?

Search đi qua:

```text
gitwhy_search(query)
Graph.Search(query, domain, limit)
```

Đường chạy:

1. Embed query.
2. Check semantic cache trong `semantic.db`.
3. Nếu cache miss, so query vector với `claim_vectors`.
4. Rank candidate claim.
5. Từ mỗi top claim, traverse edge để kéo thêm claim liên quan.
6. Cache response để lần sau trả nhanh hơn.

Search dùng weight:

| Vector kind | Weight |
|---|---:|
| `claim` | 1.25 |
| `retrieval` | 1.00 |
| `interrupt` | 0.75 |

Search candidate retrieval dùng:

```text
limit mặc định: 5 nếu caller truyền <= 0
minScore: 0.08
activeOnly: true
domain filter: optional
cache namespace: search:<provider>:<dims>:<domain>
cache threshold: 0.995
```

Ví dụ query:

```text
gitwhy_search("why did we reject Express for backend API?")
```

Expected path:

```text
query -> embedText
query vector -> semantic cache lookup
cache miss -> claim_vectors similarity
top candidate -> clm_001 "Use Go for the Cloud API. Do not add Express."
edge traversal -> pulls clm_002 rationale and clm_003 rejected alternative
response cached in semantic.db
```

Ví dụ output rút gọn:

```json
{
  "results": [
    {
      "claim_id": "clm_001",
      "claim": "Use Go for the Cloud API. Do not add Express.",
      "vector_kind": "claim"
    },
    {
      "claim_id": "clm_002",
      "claim": "Go already powers the CLI and serve binary...",
      "edge_type": "CAUSED_BY"
    },
    {
      "claim_id": "clm_003",
      "claim": "Express rejected because it adds a second backend runtime...",
      "edge_type": "CONFLICTS_WITH"
    }
  ],
  "telemetry": {
    "retrieval_mode": "claim_graph",
    "cache_hit": false
  }
}
```

## Edge traversal trong search

Search không chỉ trả claim top match. Nó còn đi theo `edges` để trả decision chain.

Traversal hiện đi tối đa 2 bước qua `edges` có `status IN ('active', 'candidate')`. Nó dùng recursive SQL trong `traverseFromClaim`, nên một top claim có thể kéo cả lý do trực tiếp và claim liên quan gần đó.

Ví dụ 2-hop:

```text
clm_A: Use Go for Cloud API.
  -[CAUSED_BY]->
clm_B: One runtime keeps deployment simpler.
  -[RELATED_CANDIDATE]->
clm_C: Serve binary should read local .git/gitwhy data only.
```

Search có thể trả cả A, B, C nếu còn trong `limit`.

## Tripwire: cảnh báo trước khi sửa code

Tripwire chạy trước khi agent edit file:

```text
gitwhy_tripwire(agent_plan_created)
Graph.CheckTripwire(event)
```

Event text được build từ:

- `event_type`
- `user_request`
- `agent_plan`
- `files_likely_touched`
- `concepts`
- `proposed_changes`
- `new_dependencies`
- `risk_surfaces`

Tripwire không fallback sang markdown. Telemetry luôn là `graph_only`. Nếu graph chưa sẵn sàng, nó trả `available=false` thay vì giả vờ có kết quả.

Tripwire dùng vector weights khác search:

| Vector kind | Weight |
|---|---:|
| `claim` | 0.85 |
| `retrieval` | 1.05 |
| `interrupt` | 1.30 |

Tripwire retrieval dùng:

```text
candidate limit: 12
minScore: 0.08
activeOnly: true
cache namespace: tripwire:<provider>:<dims>:<project_id>
cache threshold: 0.995
```

Một candidate chỉ trở thành interrupt khi đủ tín hiệu:

1. vector match
2. claim đang `active`
3. scope hoặc blast radius khớp với event
4. interrupt condition khớp, hoặc claim có edge loại đáng cảnh báo

Các edge được coi là đáng cảnh báo:

- `CONSTRAINS`
- `CONFLICTS_WITH`
- `SUPERSEDES`
- `RELATED_CANDIDATE`

Ví dụ event agent gửi trước khi sửa:

```json
{
  "event_type": "agent_plan_created",
  "user_request": "Add auth routes quickly.",
  "agent_plan": "Add an Express server for auth endpoints.",
  "files_likely_touched": ["cloud/api/router.go", "cloud/main.go"],
  "concepts": ["backend", "auth", "Express"],
  "proposed_changes": ["add Express API"],
  "new_dependencies": ["express"],
  "risk_surfaces": ["backend runtime", "deployment"]
}
```

GitWhy build text từ toàn bộ event đó, embed text này, rồi search với `interrupt` weight mạnh nhất. Nếu claim cũ có scope `backend/api` và interrupt condition kiểu "plan adds Express", candidate sẽ pass:

```text
scopeMatch = true
interruptMatch = true
edgeMatch = maybe true
candidate.vectorKind = interrupt
interrupt = true
```

Response rút gọn:

```json
{
  "available": true,
  "interrupt": true,
  "message": "Relevant prior decision:\n- Use Go for the Cloud API. Do not add Express.\n\nWhy it matters now:\n- This plan matches the claim's scope and interrupt conditions.\n\nSuggested action:\n- Continue with this context, revise the plan, or explicitly supersede the old decision.",
  "candidates": [
    {
      "claim_id": "clm_001",
      "claim": "Use Go for the Cloud API. Do not add Express.",
      "vector_kind": "interrupt",
      "matched_signals": [
        "vector_match",
        "active_claim",
        "scope_blast_radius_match",
        "interrupt_condition_match",
        "interrupt_vector_match"
      ]
    }
  ]
}
```

Nếu `Graph` chưa init hoặc embedding lỗi, tripwire trả `available=false`. Nó không pretend là safe.

## Markdown fallback nằm ở đâu?

MCP search có fallback:

```text
gitwhy_search
  -> graph search trước
  -> nếu graph lỗi/không có kết quả và mode không phải graph_only
  -> markdown fallback qua Store.Search
```

MCP search có 3 mode thực tế:

| Tình huống | Kết quả |
|---|---|
| Graph có result | `retrieval_mode = claim_graph` |
| Graph lỗi hoặc rỗng, mode mặc định | fallback `Store.Search`, markdown substring scan |
| `mode = graph_only` | không fallback; trả graph result hoặc rỗng |

Tripwire thì không fallback. Đây là deliberate: tripwire là gate trước edit, nên nếu graph không sẵn sàng thì phải nói rõ, không âm thầm scan markdown và tạo cảm giác an toàn giả.

## Cái nào cần AI?

| Bước | Có cần AI không? |
|---|---|
| Agent viết context summary | Có, vì agent đang tóm tắt quyết định. |
| Ghi markdown | Không. |
| Tách claim | Không, rule/keyword logic. |
| Infer scope/alias/triggers/interrupt | Không, rule/keyword logic. |
| Tạo embedding | Có thể dùng OpenAI, nhưng default có local hash embedding offline. |
| Search vector + SQL | Không. |
| Edge traversal | Không. |
| Tripwire message | Không, template string. |

Vậy GitWhy không phải "AI tự nhớ mọi thứ". Nó là storage + graph + vector retrieval. AI chỉ giúp viết context tốt lúc save, và optional OpenAI embedding nếu bạn bật.

## Ví dụ thứ hai: spacing rule trong frontend

Context save:

```text
Decision:
- Use 4/8/16/24 spacing scale for dashboard UI. Do not add ad-hoc spacing values.

What was done:
- Updated web/src/components/layout/Sidebar.tsx to use gap-4 and p-6.

Risk:
- If a new dashboard card uses one-off margin values, layout consistency will drift.
```

Claims:

```text
clm_spacing_1 = Use 4/8/16/24 spacing scale for dashboard UI...
clm_spacing_2 = Updated Sidebar.tsx to use gap-4 and p-6.
clm_spacing_3 = If a new dashboard card uses one-off margin values...
```

Edges:

```text
clm_spacing_2 IMPLEMENTS clm_spacing_1
clm_spacing_3 CONSTRAINS clm_spacing_2
```

Search:

```text
gitwhy_search("why are dashboard gaps multiples of 4?")
```

Tripwire:

```text
Plan: add Card component with margin: 13px.
Files: web/src/components/dashboard/KnowledgeHeatmap.tsx
Concepts: dashboard, spacing
```

Expected tripwire: interrupt, because scope and interrupt terms overlap spacing/dashboard/layout.

## Ví dụ thứ ba: reindex sau khi sửa markdown tay

Nếu ai đó sửa file trong `.git/gitwhy/contexts/` bằng tay, markdown đã đổi nhưng `graph.db` chưa chắc đổi. Lúc đó cần:

```text
git why reindex
```

Reindex đọc lại markdown contexts rồi gọi lại `SaveToGraph` để rebuild `sessions`, `claims`, `claim_vectors`, và `edges`.

## Nhớ một câu

GitWhy lưu nguyên context làm archive, rút ra tối đa 7 claim làm memory thật, tạo 3 vector cho mỗi claim, nối claim bằng typed edges, rồi dùng search/tripwire để trả lại decision chain đúng lúc.

## Code map

| File | Vai trò |
|---|---|
| `internal/context/store.go` | Lưu và search markdown contexts. |
| `internal/context/whyspec.go` | Render/parse whyspec markdown. |
| `internal/graph/schema.go` | SQLite schema cho sessions, claims, vectors, edges, interrupt events. |
| `internal/graph/claims.go` | Claim extraction, claim classification, metadata inference, vector text. |
| `internal/graph/embed.go` | Local/OpenAI embedding providers. |
| `internal/graph/graph.go` | SaveToGraph, Search, CheckTripwire, edge creation, cache invalidation. |
| `internal/cache/cache.go` | Semantic cache. |
| `internal/mcp/server.go` | MCP tools: save, search, tripwire, status, sync, publish, post_pr. |

## Debug nhanh khi kết quả lạ

| Symptom | Kiểm tra |
|---|---|
| `gitwhy_search` không thấy context mới | Context đã được save vào markdown chưa? `git why reindex` có cần chạy không? |
| Tripwire không interrupt | Event có đủ `files_likely_touched`, `concepts`, `proposed_changes`, `risk_surfaces` chưa? Claim có scope/interrupt metadata liên quan không? |
| Graph báo provider mismatch | Có đổi `GITWHY_EMBEDDING_PROVIDER` hoặc thêm `OPENAI_API_KEY` sau khi graph đã tạo không? Chạy reindex với provider mong muốn. |
| Search trả quá ít chain | `limit` có quá thấp không? Edge có tồn tại trong `edges` không? |
| Cache trả kết quả cũ | Sau save mới cache sẽ invalidate; nếu vẫn lạ, kiểm tra `semantic.db` và namespace provider/dims. |
