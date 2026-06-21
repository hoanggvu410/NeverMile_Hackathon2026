# 08 — API Specifications

---

## Part A — MCP Tool Schemas

Tất cả tools đều follow MCP protocol. Transport: stdio. Client: Claude Code, Cursor, Windsurf, Cline, VS Code Copilot.

---

### Tool: gitwhy_save

```json
{
  "name": "gitwhy_save",
  "description": "Save development context (reasoning, decisions, trade-offs) for the current session. Link to git commits automatically.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "prompt": {
        "type": "string",
        "description": "The original user prompt given to the AI agent"
      },
      "reasoning": {
        "type": "string",
        "description": "Agent's explanation of its approach and methodology"
      },
      "decisions": {
        "type": "string",
        "description": "Key choices made with rationale (e.g. 'Chose RS256 because...')"
      },
      "rejected_alternatives": {
        "type": "string",
        "description": "Options that were considered but discarded, and why"
      },
      "files": {
        "type": "array",
        "items": { "type": "string" },
        "description": "List of source files affected by this change"
      },
      "commits": {
        "type": "array",
        "items": { "type": "string" },
        "description": "Git commit hashes to link. Auto-detected from HEAD if omitted."
      },
      "domain": {
        "type": "string",
        "description": "Hierarchical domain label (e.g. 'backend/auth', 'infra/database')"
      },
      "topic": {
        "type": "string",
        "description": "Topic slug (e.g. 'jwt-migration', 'kafka-removal')"
      }
    },
    "required": ["prompt", "reasoning", "decisions"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "id": "cxt_20260620_abc123",
  "timestamp": "2026-06-20T10:30:00Z",
  "linked_commits": ["a1b2c3d"],
  "local_path": ".git/gitwhy/contexts/cxt_20260620_abc123.md"
}
```

---

### Tool: gitwhy_get

```json
{
  "name": "gitwhy_get",
  "description": "Retrieve a saved context by its ID.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "id": {
        "type": "string",
        "description": "Context ID (e.g. 'cxt_20260620_abc123')"
      }
    },
    "required": ["id"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "context": {
    "id": "cxt_20260620_abc123",
    "prompt": "...",
    "reasoning": "...",
    "decisions": "...",
    "rejected_alternatives": "...",
    "files": ["app/core/security.py"],
    "commits": ["a1b2c3d"],
    "domain": "backend/auth",
    "topic": "jwt-migration",
    "agent": "claude-code",
    "model": "claude-opus-4-6",
    "timestamp": "2026-06-20T10:30:00Z",
    "synced": false,
    "published": false
  }
}
```

---

### Tool: gitwhy_search

```json
{
  "name": "gitwhy_search",
  "description": "Search saved contexts by keyword or natural language query. Searches local and (if synced) team cloud contexts.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search query (e.g. 'tại sao bỏ Kafka', 'AWS migration reason')"
      },
      "domain": {
        "type": "string",
        "description": "Optional: filter by domain prefix (e.g. 'infrastructure')"
      },
      "limit": {
        "type": "integer",
        "description": "Max results to return (default: 5, max: 20)"
      }
    },
    "required": ["query"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "cache_hit": false,
  "results": [
    {
      "id": "cxt_20260610_xyz789",
      "prompt": "Remove Kafka from messaging pipeline",
      "decisions": "Switched to SQS for cost savings...",
      "domain": "infrastructure",
      "topic": "kafka-removal",
      "score": 0.94,
      "timestamp": "2026-06-10T14:00:00Z"
    }
  ],
  "total": 1
}
```

---

### Tool: gitwhy_list

```json
{
  "name": "gitwhy_list",
  "description": "Browse saved contexts by domain/topic hierarchy.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "domain": { "type": "string", "description": "Filter by domain" },
      "topic":  { "type": "string", "description": "Filter by topic" },
      "limit":  { "type": "integer", "description": "Max results (default: 20)" }
    }
  }
}
```

---

### Tool: gitwhy_status

```json
{
  "name": "gitwhy_status",
  "description": "Check setup state: git repo detection, API key validity, pending sync count."
}
```

**Response:**
```json
{
  "is_git_repo": true,
  "has_api_key": true,
  "api_key_valid": true,
  "plan": "team",
  "pending_sync_count": 3,
  "last_sync_at": "2026-06-20T09:00:00Z",
  "local_context_count": 47,
  "monthly_sync_used": null,
  "monthly_sync_limit": null
}
```

---

### Tool: gitwhy_sync

```json
{
  "name": "gitwhy_sync",
  "description": "Upload local contexts to the cloud (private). Requires API key.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "context_id": {
        "type": "string",
        "description": "Sync a specific context. If omitted, sync all pending."
      }
    }
  }
}
```

**Response:**
```json
{
  "success": true,
  "synced_count": 3,
  "failed_count": 0,
  "quota_used": 3,
  "quota_remaining": 17
}
```

---

### Tool: gitwhy_publish

```json
{
  "name": "gitwhy_publish",
  "description": "Share synced contexts with your team. Requires Team plan.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "context_id": {
        "type": "string",
        "description": "Publish a specific context. If omitted, publish all synced."
      }
    }
  }
}
```

---

### Tool: gitwhy_post_pr

```json
{
  "name": "gitwhy_post_pr",
  "description": "Post a context summary as a GitHub PR comment via gitwhy-bot.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "context_id": {
        "type": "string",
        "description": "ID of the context to post"
      },
      "repo": {
        "type": "string",
        "description": "GitHub repo in owner/repo format"
      },
      "pr_number": {
        "type": "integer",
        "description": "Pull request number"
      }
    },
    "required": ["context_id", "repo", "pr_number"]
  }
}
```

**Response:**
```json
{
  "success": true,
  "comment_url": "https://github.com/org/repo/pull/42#issuecomment-123456",
  "pr_url": "https://github.com/org/repo/pull/42"
}
```

---

## Part B — Cloud REST API

Base URL: `https://api.gitwhy.dev`

Auth: `Authorization: Bearer {api_key}` cho tất cả endpoints.

---

### POST /v1/contexts/sync

Bulk upload contexts từ local.

```json
// Request
{
  "contexts": [
    {
      "local_id": "cxt_20260620_abc123",
      "prompt": "...",
      "reasoning": "...",
      "decisions": "...",
      "rejected_alternatives": "...",
      "files": ["..."],
      "commits": ["a1b2c3d"],
      "domain": "backend/auth",
      "topic": "jwt-migration",
      "agent": "claude-code",
      "model": "claude-opus-4-6",
      "context_ts": "2026-06-20T10:30:00Z"
    }
  ]
}

// Response 200
{
  "synced": ["cxt_20260620_abc123"],
  "failed": [],
  "quota_used": 1,
  "quota_remaining": 19
}

// Errors
401 CONTEXT_INVALID_API_KEY
429 CONTEXT_QUOTA_EXCEEDED  -- Free tier: 20 syncs/month
```

---

### GET /v1/contexts/search

```
GET /v1/contexts/search?q=tại+sao+bỏ+Kafka&domain=infrastructure&limit=5

// Response 200
{
  "results": [...],
  "total": 3
}
```

---

### POST /v1/pr/comment

Post PR comment via gitwhy-bot.

```json
// Request
{
  "context_local_id": "cxt_20260620_abc123",
  "repo": "org/repo",
  "pr_number": 42
}

// Response 201
{
  "comment_url": "https://github.com/org/repo/pull/42#issuecomment-123456"
}

// Errors
403 PR_GITHUB_APP_NOT_INSTALLED  -- gitwhy-bot chưa install vào repo
404 PR_NOT_FOUND
```

---

### POST /v1/auth/api-key

Tạo API key mới (qua web dashboard).

```json
// Response 201
{
  "api_key": "gw_live_xxxxxxxxxxxxx",  -- chỉ hiện 1 lần
  "key_id": "uuid",
  "name": "My Claude Code Key"
}
```

---

*Cập nhật lần cuối: 2026-06-20*
