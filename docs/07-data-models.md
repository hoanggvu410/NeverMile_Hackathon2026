# 07 — Data Models

---

## 1. Context Schema (Whyspec)

Đây là format chuẩn của mỗi context file. Lưu dạng structured markdown tại `.git/gitwhy/contexts/{id}.md`.

### Frontmatter (YAML)

```yaml
---
id: cxt_20260620_abc123
prompt: "Migrate JWT từ HS256 sang RS256"
agent: claude-code
model: claude-opus-4-6
timestamp: 2026-06-20T10:30:00Z
commits:
  - a1b2c3d4e5f6
  - b2c3d4e5f6a1
files:
  - app/core/security.py
  - app/core/deps.py
  - tests/test_auth.py
domain: backend/auth
topic: jwt-migration
synced: true
synced_at: 2026-06-20T10:35:00Z
published: false
---
```

### Body sections

```markdown
## Reasoning
[string — Agent's explanation of its overall approach and why this direction was chosen]

## Decisions
[string — Key choices made, each with rationale. Format: bullet list preferred]

## Rejected Alternatives
[string — Options that were considered but discarded, with reason for rejection]

## Trade-offs
[string — Optional. Explicit trade-offs accepted in this decision]
```

### TypeScript interface (cho web dashboard)

```typescript
interface GitWhyContext {
  id: string;                    // "cxt_{timestamp}_{nanoid}"
  prompt: string;                // original user prompt to AI
  reasoning: string;             // agent's explanation
  decisions: string;             // key choices + rationale
  rejected_alternatives: string; // discarded options
  trade_offs?: string;           // optional trade-off notes
  files: string[];               // affected source files
  commits: string[];             // linked git commit hashes
  domain: string;                // hierarchical: "backend/auth"
  topic: string;                 // "jwt-migration"
  agent: string;                 // "claude-code" | "cursor" | "windsurf"
  model: string;                 // "claude-opus-4-6" | etc
  timestamp: string;             // ISO 8601
  synced: boolean;
  synced_at?: string;
  published: boolean;
}
```

---

## 2. Claim Graph — graph.db

> ⚠️ **Note:** Phần này đã được build xong. Thiết kế cuối cùng khác với bản draft trước — đọc phần này để hiểu đúng.
>
> ⚠️ **Note:** This section is fully built. The final design differs from the original draft — read this to understand what's actually implemented.

Lưu tại `.git/gitwhy/graph.db` (SQLite).

---

### Ý tưởng đơn giản / The simple idea

Thay vì lưu cả đoạn markdown dài và để AI tự đọc lại sau, GitWhy tách ra những **quyết định quan trọng** (gọi là *claims*) và lưu chúng ở dạng có thể tìm kiếm bằng vector.

Instead of saving long markdown blobs and making AI re-read them later, GitWhy extracts **key decisions** (called *claims*) and stores them in vector-searchable form.

```
Một session gitwhy_save →  tối đa 7 claims
                           mỗi claim → 3 vectors
                           vectors → dùng để tìm kiếm và cảnh báo tripwire
```

---

### Bảng chính / Main tables

#### `sessions` — Full markdown archive

```sql
CREATE TABLE sessions (
  id            TEXT PRIMARY KEY,   -- ctx_xxxxxxxx
  project_id    TEXT,               -- repo URL hoặc "local"
  domain        TEXT,               -- vd: "product/gitwhy2"
  topic         TEXT,               -- vd: "tripwire-agent-contract"
  title         TEXT,
  prompt        TEXT,
  created_at    DATETIME,
  full_markdown TEXT                -- toàn bộ whyspec markdown
);
```

Session chỉ là archive — không dùng trực tiếp để search. / Sessions are just the archive — not used directly for search.

---

#### `claims` — Quyết định được tách ra / Extracted decisions

Mỗi save → tách tối đa 7 claims từ các section: Key Decisions, Rejected Alternatives, Risks, What Was Done, Reasoning.

Each save → up to 7 claims extracted from: Key Decisions, Rejected Alternatives, Risks, What Was Done, Reasoning.

```sql
CREATE TABLE claims (
  id                        TEXT PRIMARY KEY,   -- clm_xxxxxxxxxxxxxxxx
  session_id                TEXT,               -- link về session gốc
  text                      TEXT,               -- nội dung quyết định
  type                      TEXT,               -- decision / constraint / risk / rejected_alternative / ...
  status                    TEXT DEFAULT 'active',
  importance                INTEGER,            -- 3–5, dùng để sort
  source_span               TEXT,               -- "Key Decisions:1"
  scope_json                TEXT,               -- {components, concepts, files, dependencies}
  aliases_json              TEXT,               -- tên gọi khác ngắn gọn
  retrieval_triggers_json   TEXT,               -- "khi nào thì cần claim này?"
  blast_radius_json         TEXT,               -- file/component nào bị ảnh hưởng
  interrupt_conditions_json TEXT                -- "khi nào thì interrupt agent?"
);
```

---

#### `claim_vectors` — 3 vectors cho mỗi claim / 3 vectors per claim

```sql
CREATE TABLE claim_vectors (
  id        TEXT PRIMARY KEY,
  claim_id  TEXT,
  kind      TEXT CHECK(kind IN ('claim', 'retrieval', 'interrupt')),
  provider  TEXT,   -- "openai" hoặc "nomic" (offline)
  dims      INTEGER,
  text      TEXT,   -- text đã được embed
  embedding BLOB    -- float32 array
);
```

| Vector kind | Embed cái gì | Dùng để làm gì |
|---|---|---|
| `claim` | Chính nội dung quyết định | Tìm kiếm giống semantic |
| `retrieval` | aliases + retrieval triggers | Tìm khi context xa với claim |
| `interrupt` | interrupt conditions + blast radius | Tripwire — cảnh báo plan sắp vi phạm |

Tại sao 3 vectors? Câu hỏi tương lai có thể khác hoàn toàn với cách claim được lưu. Ví dụ: claim "dùng FastAPI" nhưng câu hỏi là "websocket notification có reliable không?" — chỉ `retrieval` vector mới bắt được.

Why 3? Future questions may be semantically distant from the claim's own text. One vector per claim is too weak.

---

#### `edges` — Quan hệ giữa các claims / Relationships between claims

```sql
CREATE TABLE edges (
  id            TEXT PRIMARY KEY,
  from_claim_id TEXT,
  to_claim_id   TEXT,
  type          TEXT CHECK(type IN (
    'CONSTRAINS',         -- A ràng buộc B
    'IMPLEMENTS',         -- A là implementation của decision B
    'CAUSED_BY',          -- A xảy ra do B
    'SUPERSEDES',         -- A thay thế B (B đã lỗi thời)
    'CONFLICTS_WITH',     -- A và B mâu thuẫn nhau
    'RELATED_CANDIDATE'   -- có thể liên quan — chưa xác nhận
  )),
  confidence    REAL,
  evidence      TEXT,
  source        TEXT,   -- "same_session" / "cross_session" / "edge_hint"
  status        TEXT    -- "active" / "candidate"
);
```

**Edge được tạo tự động như thế nào? / How edges are auto-created:**

| Tình huống | Edge được tạo |
|---|---|
| Cùng session: implementation + decision | `IMPLEMENTS` (confidence 0.82) |
| Cùng session: constraint + implementation | `CONSTRAINS` (confidence 0.78) |
| Cùng session: decision + rejected alternative | `CONFLICTS_WITH` (confidence 0.74) |
| Khác session: semantic similarity cao | `RELATED_CANDIDATE` (status=candidate) |
| Agent cung cấp explicit hint | Bất kỳ type nào (confidence 0.90) |

Không cần LLM để classify edge — pattern matching theo claim type là đủ cho MVP.

No LLM needed for edge classification — claim type pattern matching is sufficient for MVP.

---

#### Bảng phụ / Supporting tables

```sql
-- Theo dõi mỗi khi tripwire được check
CREATE TABLE interrupt_events (
  id              TEXT PRIMARY KEY,
  event_type      TEXT,     -- "agent_plan_created"
  project_id      TEXT,
  event_json      TEXT,     -- plan của agent
  candidates_json TEXT,     -- claim nào được surface
  user_action     TEXT,
  created_at      DATETIME
);

-- Lưu provider + dims đang dùng để tránh incompatible vectors
CREATE TABLE embedding_config (
  key   TEXT PRIMARY KEY,
  value TEXT
);
```

---

### Tìm kiếm hoạt động thế nào / How search works

```
gitwhy_search(query)
    ↓
1. Embed query → vector
2. Cosine similarity với tất cả claim_vectors
3. Lấy top claims theo score
4. Traverse edges (2 hops) để lấy related claims
5. Trả về kết quả có score + claim text + edge path
```

---

### Tripwire hoạt động thế nào / How tripwire works

```
gitwhy_tripwire(plan event)
    ↓
1. Embed plan → vector
2. So sánh với "interrupt" vectors (weight cao nhất)
3. Nếu claim match scope VÀ (interrupt condition match HOẶC có edge)
   → interrupt = true
   → trả về claim liên quan + message
4. Không fallback về markdown — telemetry_mode = "graph_only"
```

Tripwire phải được gọi *trước khi* agent bắt đầu edit files. Xem `AGENTS.md`.

Tripwire must be called *before* the agent starts editing files. See `AGENTS.md`.

---

## 3. Sync State

Lưu tại `.git/gitwhy/sync.json`.

```json
{
  "last_sync": "2026-06-20T10:35:00Z",
  "contexts": {
    "cxt_20260620_abc123": {
      "status": "synced",
      "cloud_id": "cl_abc123",
      "synced_at": "2026-06-20T10:35:00Z"
    },
    "cxt_20260619_def456": {
      "status": "pending"
    }
  }
}
```

---

## 4. Cloud Database (PostgreSQL)

### Bảng: users

```sql
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        VARCHAR(255) UNIQUE NOT NULL,
    github_id    VARCHAR(50) UNIQUE,
    github_login VARCHAR(100),
    plan         VARCHAR(20) NOT NULL DEFAULT 'free', -- 'free' | 'team'
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Bảng: api_keys

```sql
CREATE TABLE api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash     VARCHAR(64) UNIQUE NOT NULL,  -- SHA-256 hex của key
    name         VARCHAR(100),                 -- label do user đặt
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX ix_api_keys_user_id  ON api_keys(user_id);
```

### Bảng: teams

```sql
CREATE TABLE teams (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    owner_id     UUID NOT NULL REFERENCES users(id),
    plan         VARCHAR(20) NOT NULL DEFAULT 'team',
    sync_quota   INT NOT NULL DEFAULT -1, -- -1 = unlimited (team plan)
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE team_members (
    team_id    UUID REFERENCES teams(id) ON DELETE CASCADE,
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL DEFAULT 'member', -- 'owner' | 'member'
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);
```

### Bảng: contexts

```sql
CREATE TABLE contexts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    local_id        VARCHAR(50) UNIQUE NOT NULL,   -- cxt_{timestamp}_{nanoid}
    user_id         UUID NOT NULL REFERENCES users(id),
    team_id         UUID REFERENCES teams(id),
    prompt          TEXT NOT NULL,
    reasoning       TEXT,
    decisions       TEXT,
    rejected_alternatives TEXT,
    trade_offs      TEXT,
    files           JSONB NOT NULL DEFAULT '[]',   -- string[]
    commits         JSONB NOT NULL DEFAULT '[]',   -- string[]
    domain          VARCHAR(200),
    topic           VARCHAR(200),
    agent           VARCHAR(50),
    model           VARCHAR(100),
    is_published    BOOLEAN NOT NULL DEFAULT FALSE,
    repo_name       VARCHAR(255),                  -- "owner/repo"
    embedding       VECTOR(1536),                  -- pgvector (v0.2)
    context_ts      TIMESTAMPTZ NOT NULL,          -- original timestamp
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_contexts_user_id     ON contexts(user_id);
CREATE INDEX ix_contexts_team_id     ON contexts(team_id);
CREATE INDEX ix_contexts_is_published ON contexts(is_published);
CREATE INDEX ix_contexts_domain      ON contexts(domain);
-- pgvector index (v0.2)
-- CREATE INDEX ix_contexts_embedding ON contexts USING ivfflat (embedding vector_cosine_ops);
```

### Bảng: sync_quota_usage (Free tier tracking)

```sql
CREATE TABLE sync_quota_usage (
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    month      DATE NOT NULL,                  -- first day of month
    sync_count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, month)
);
```

### Bảng: pr_comments

```sql
CREATE TABLE pr_comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id  UUID NOT NULL REFERENCES contexts(id),
    repo        VARCHAR(255) NOT NULL,
    pr_number   INT NOT NULL,
    comment_url TEXT,
    posted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 5. Config File (local)

Lưu tại `~/.config/gitwhy/config.json` (permission 600).

```json
{
  "api_key": "gw_live_xxxxxxxxxxxxx",
  "api_url": "https://api.gitwhy.dev",
  "user_id": "uuid",
  "plan": "team"
}
```

---

*Cập nhật lần cuối: 2026-06-20*
