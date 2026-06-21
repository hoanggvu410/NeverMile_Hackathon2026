# 05 — Non-Functional Requirements

---

## 1. Performance

| Metric | Target | Ghi chú |
|--------|--------|--------|
| `gitwhy_save` (local) | < 100ms | Ghi file local, không gọi network |
| `gitwhy_search` (local only) | < 500ms | Full-text search trên local contexts |
| `gitwhy_search` (cloud) | < 3s | Target demo — "hỏi tại sao bỏ Kafka, trả lời trong 3s" |
| `gitwhy_sync` per context | < 2s/context | Upload 1 context lên cloud |
| `gitwhy_post_pr` | < 5s | GitHub API call + format |
| Semantic cache hit response | < 50ms | Không gọi LLM, chỉ cache lookup |
| Web Dashboard load | < 2s FCP | Next.js SSR |
| Context Graph traversal (2-hop) | < 1s | Local graph query |

---

## 2. Availability

| Metric | Target |
|--------|--------|
| Cloud API uptime | 99.5% / tháng |
| Local MCP server | 100% (không phụ thuộc cloud) |
| Graceful degradation | Nếu cloud down → local features vẫn hoạt động bình thường |
| gitwhy-bot webhook | Retry 3 lần với exponential backoff nếu fail |

**Nguyên tắc offline-first**: `gitwhy_save`, `gitwhy_get`, `gitwhy_search` (local), `gitwhy_list` (local) **không yêu cầu internet**. Chỉ `sync`, `publish`, `post_pr` cần network.

---

## 3. Security

- API key: lưu trong `~/.config/gitwhy/config.json` với permission 600
- API key không bao giờ log ra stdout/stderr hoặc ghi vào context file
- Cloud sync: TLS 1.3 bắt buộc
- Context nội dung: không encrypt at rest (local) — dev chịu trách nhiệm bảo mật máy
- Cloud contexts: encrypt at rest (AES-256) trên PostgreSQL
- GitHub App: scope tối thiểu — chỉ `pull_requests: write`, không có repo read

> Chi tiết xem `11-security.md`

---

## 4. Scalability

**MCP Server (local):**
- Stateless — mỗi tool call độc lập
- Không cần state shared giữa calls
- Memory footprint nhỏ: < 50MB

**Cloud Backend:**
- Horizontal scale API servers
- PostgreSQL connection pooling (PgBouncer)
- Semantic cache layer (Redis) để giảm LLM calls cho graph queries
- Target: 10,000 active teams

---

## 5. Compatibility

| Platform | Support |
|---------|---------|
| macOS (Apple Silicon + Intel) | ✅ |
| Linux (x86_64, ARM64) | ✅ |
| Windows 10/11 | ✅ (Scoop) |
| Claude Code | ✅ |
| Cursor | ✅ |
| Windsurf | ✅ |
| Cline | ✅ |
| VS Code Copilot | ✅ |
| Any MCP-compatible agent | ✅ |
| GitLab PR bot | ❌ Phase 2 |
| Bitbucket PR bot | ❌ Phase 2 |

**Node.js requirement:** v18+ cho `gitwhy-mcp` (npm package)

---

## 6. Maintainability

- MCP tool schema versioned: `gitwhy_save_v1`, tăng version khi breaking change
- Context format (whyspec) backward compatible — field mới thêm optional
- Local storage path: `.git/gitwhy/` — gitignored tự động
- Logging: structured JSON logs cho cloud backend
- Error messages: human-readable, actionable (không "undefined error")

---

## 7. Data & Context Limits

| Limit | Value |
|-------|-------|
| Max context size | 100KB per context |
| Max files per context | 50 files |
| Max commits per context | 20 commits |
| Reasoning field | Max 10,000 tokens |
| decisions + rejected_alternatives | Max 5,000 tokens each |
| Graph edges per context | Max 50 edges |
| Semantic cache entries (local) | Max 1,000 entries |

---

## 8. Token Cost Targets

| Scenario | Cost |
|---------|------|
| Demo query "tại sao bỏ Kafka" | $0.01 |
| Repeated query (cache hit) | $0.00 |
| Full context save | ~$0.002 |
| PR comment generation | ~$0.005 |
| Monthly power user (no cache) | < $5 |
| Monthly power user (với cache 70% hit rate) | < $1.50 |

---

*Cập nhật lần cuối: 2026-06-20*
