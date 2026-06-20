# 12 — Error Codes

Format response lỗi chuẩn (Cloud API):
```json
{
  "success": false,
  "error": {
    "code": "ERR_CODE",
    "message": "Mô tả lỗi cho end-user"
  }
}
```

Format lỗi MCP tool (trả về trong content):
```json
{
  "success": false,
  "error": "ERR_CODE",
  "message": "Human-readable error message"
}
```

---

## AUTH — Authentication errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `INVALID_API_KEY` | 401 | API key không tồn tại hoặc đã bị revoke |
| `API_KEY_EXPIRED` | 401 | API key đã hết hạn (nếu có expiry) |
| `ACCOUNT_DISABLED` | 403 | Tài khoản user bị vô hiệu hóa |
| `PLAN_REQUIRED` | 403 | Action yêu cầu Team plan nhưng user đang ở Free |

---

## CONTEXT — Context operation errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `CONTEXT_NOT_FOUND` | 404 | Context ID không tồn tại (local hoặc cloud) |
| `CONTEXT_INVALID_SCHEMA` | 400 | Context thiếu required fields (prompt, reasoning, decisions) |
| `CONTEXT_TOO_LARGE` | 400 | Context vượt quá 100KB |
| `CONTEXT_NOT_SYNCED` | 400 | Cố gắng publish context chưa sync lên cloud |
| `CONTEXT_QUOTA_EXCEEDED` | 429 | Free tier: đã dùng hết 20 syncs tháng này |
| `CONTEXT_ALREADY_EXISTS` | 409 | local_id đã tồn tại trên cloud (idempotent — safe to ignore) |

---

## GIT — Git integration errors

| Code | MCP | Mô tả |
|------|-----|-------|
| `NOT_GIT_REPO` | — | Working directory không phải git repository |
| `GIT_NOT_INSTALLED` | — | `git` binary không tìm thấy trong PATH |
| `NO_COMMITS` | — | Repo chưa có commit nào (git HEAD không tồn tại) |
| `GIT_COMMAND_FAILED` | — | Lỗi khi chạy git command |

---

## SYNC — Sync errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `SYNC_PARTIAL_FAILURE` | 207 | Một số contexts sync thành công, một số thất bại |
| `SYNC_NETWORK_ERROR` | 503 | Không kết nối được cloud API |
| `SYNC_CONFLICT` | 409 | Context đã bị modify trên cloud (resolution: local wins) |

---

## PR — Pull Request errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `PR_NOT_FOUND` | 404 | PR number không tồn tại trong repo |
| `PR_REPO_NOT_FOUND` | 404 | Repository không tồn tại hoặc không accessible |
| `PR_GITHUB_APP_NOT_INSTALLED` | 403 | gitwhy-bot chưa được install vào repository |
| `PR_RATE_LIMITED` | 429 | GitHub API rate limit hit |
| `PR_PERMISSION_DENIED` | 403 | gitwhy-bot không có quyền write PR comment |
| `PR_COMMENT_FAILED` | 502 | GitHub API trả về lỗi không mong muốn |

---

## GRAPH — Context Graph errors (v0.2)

| Code | MCP | Mô tả |
|------|-----|-------|
| `GRAPH_NOT_INITIALIZED` | — | graph.json chưa tồn tại — run `gitwhy_save` ít nhất 1 lần |
| `GRAPH_TRAVERSAL_TIMEOUT` | — | Graph traversal > 10s (graph quá lớn) |
| `GRAPH_CORRUPT` | — | graph.json bị corrupt — xóa và rebuild |

---

## CACHE — Semantic Cache errors (v0.2)

| Code | MCP | Mô tả |
|------|-----|-------|
| `CACHE_WRITE_FAILED` | — | Không ghi được vào cache DB (disk full?) |
| `CACHE_CORRUPT` | — | SQLite cache bị corrupt — tự xóa và tạo lại |

---

## TEAM — Team management errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `TEAM_NOT_FOUND` | 404 | Team không tồn tại |
| `TEAM_MEMBER_NOT_FOUND` | 404 | User không phải thành viên team |
| `TEAM_INVITE_EXPIRED` | 400 | Link invite đã hết hạn (72h) |
| `TEAM_MEMBER_ALREADY_EXISTS` | 409 | User đã là member của team |

---

## RATE — Rate Limiting

| Code | HTTP | Mô tả |
|------|------|-------|
| `RATE_LIMITED` | 429 | Quá nhiều requests. Header `Retry-After` cho biết giây chờ |

---

## SYS — System errors

| Code | HTTP | Mô tả |
|------|------|-------|
| `INTERNAL_ERROR` | 500 | Lỗi nội bộ không mong muốn |
| `SERVICE_UNAVAILABLE` | 503 | Cloud API đang maintenance hoặc quá tải |
| `VALIDATION_ERROR` | 422 | Input validation thất bại — message chứa field errors |

---

## Error Handling trong MCP Tools

MCP tools **không throw exception** — luôn return structured error:

```javascript
// Ví dụ xử lý trong agent:
const result = await callTool("gitwhy_save", { prompt, reasoning, decisions });

if (!result.success) {
  if (result.error === "NOT_GIT_REPO") {
    // Thông báo user chạy lệnh từ git project directory
  } else if (result.error === "CONTEXT_QUOTA_EXCEEDED") {
    // Thông báo free tier limit, suggest upgrade
  }
}
```

---

*Cập nhật lần cuối: 2026-06-20*
