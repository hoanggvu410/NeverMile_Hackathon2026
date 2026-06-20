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

## 2. Context Graph (v0.2)

Lưu tại `.git/gitwhy/graph.db` (SQLite). Flat JSON file không thể do graph traversal — SQLite adjacency table cho phép query typed edges.

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
    'CAUSED_BY',
    'CONSTRAINED_BY',
    'INVALIDATES',
    'CONTRADICTS',
    'DEPENDS_ON'
  )),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (from_id, to_id, edge_type)
);
```

### Typed Edges

| Edge Type | Nghĩa |
|---|---|
| `CAUSED_BY` | Decision A trực tiếp trigger decision B |
| `CONSTRAINED_BY` | Decision B phải hoạt động trong bounds của A |
| `INVALIDATES` | Decision B làm decision A obsolete |
| `CONTRADICTS` | B conflict với A — fires alert |
| `DEPENDS_ON` | B chỉ đúng nếu assumption trong A vẫn còn valid |

**Tại sao typed edges, không phải similarity links:** RAG trả lời "cái gì similar với X?" Graph trả lời "cái gì downstream của X?" và "cái gì break nếu X thay đổi?" Similarity search không thể answer dependency queries. Graph traversal có thể.

### Traversal Query Pattern (2-hop)

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
FROM chain c
JOIN context_nodes n ON n.id = c.to_id;
```

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
