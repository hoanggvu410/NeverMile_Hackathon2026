# 11 — Security

---

## 1. API Key Management

### Format

```
gw_live_[32 bytes base64url]    ← production
gw_test_[32 bytes base64url]    ← test/dev environment
```

### Lưu trữ

- **Client side**: `~/.config/gitwhy/config.json` với permission `0600`
- **Server side**: Chỉ lưu `SHA-256(api_key)` trong DB — không bao giờ store plaintext
- API key chỉ hiện **1 lần** khi tạo tại dashboard — sau đó không thể retrieve lại

### Validation flow

```
Client → HTTPS → Cloud API
  Header: Authorization: Bearer gw_live_xxx...

Cloud API:
  [1] Extract key từ header
  [2] Compute SHA-256(key)
  [3] SELECT FROM api_keys WHERE key_hash = ? AND revoked_at IS NULL
  [4] Check user.is_active = true
  [5] Check plan permissions cho action được yêu cầu
  [6] Proceed hoặc return 401/403
```

### Rotation & Revocation

- Dev có thể revoke bất kỳ lúc nào tại `app.gitwhy.dev/dashboard/api-keys`
- Revoke ngay lập tức — không có grace period
- Sau revoke: request với key đó → `401 INVALID_API_KEY`

---

## 2. Local Context Security

### Local storage

- Contexts lưu plain text tại `.git/gitwhy/contexts/`
- `.git/` directory nên gitignored (Git tự xử lý — `.git/` không bao giờ commit)
- **Dev chịu trách nhiệm** về bảo mật máy local — GitWhy không encrypt local contexts
- Không nên lưu secret values (passwords, tokens) trong context content

### Điều không bao giờ được làm

```
❌ Log API key ra stdout/stderr
❌ Ghi API key vào context file
❌ Ghi API key vào commit message
❌ Print API key trong error messages
❌ Include API key trong `gitwhy_status` response
```

---

## 3. Cloud Data Security

### Encryption at rest

- Contexts lưu trong PostgreSQL: encrypt tại disk level (storage provider)
- Sensitive fields (context nội dung) có thể encrypt at application level (AES-256) — Phase 2

### Encryption in transit

- TLS 1.3 bắt buộc cho tất cả requests đến `api.gitwhy.dev`
- Certificate pinning cho Go CLI client

### Data isolation

- Contexts của user A không bao giờ visible với user B (trừ khi published trong cùng team)
- Published contexts chỉ visible cho team members
- SQL queries luôn filter bởi `user_id` hoặc `team_id`

---

## 4. GitHub App Security

### Permissions

gitwhy-bot GitHub App chỉ yêu cầu:
- `pull_requests: write` — để post comments
- `metadata: read` — GitHub bắt buộc

**Không có** `contents: read` → gitwhy-bot không thể đọc code.

### Token management

- GitHub App installation token có TTL 1h
- Tự refresh trước khi expire
- Không lưu installation tokens vào DB

### Webhook verification

```go
// Verify GitHub webhook signature
mac := hmac.New(sha256.New, []byte(webhookSecret))
mac.Write(payload)
expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
if !hmac.Equal([]byte(signature), []byte(expected)) {
    return 401
}
```

---

## 5. Rate Limiting

| Endpoint | Limit |
|---------|-------|
| POST /v1/contexts/sync | 100 req/min per API key |
| POST /v1/pr/comment | 60 req/min per API key |
| GET /v1/contexts/search | 200 req/min per API key |
| POST /v1/auth/api-key | 10 req/hour per user |

Rate limit response:
```json
{ "error": { "code": "RATE_LIMITED", "retry_after": 60 } }
```

---

## 6. Input Validation

| Field | Validation |
|-------|-----------|
| `prompt` | Max 50,000 chars |
| `reasoning` | Max 50,000 chars |
| `decisions` | Max 20,000 chars |
| `rejected_alternatives` | Max 20,000 chars |
| `files` | Max 50 items, each max 500 chars |
| `commits` | Max 20 items, each must match `/^[0-9a-f]{7,40}$/` |
| `domain` | Max 200 chars, alphanumeric + `/` + `-` + `_` |
| `topic` | Max 200 chars, alphanumeric + `-` + `_` |
| `repo` | Must match `^[\w.-]+/[\w.-]+$` |
| `pr_number` | Positive integer |

Tất cả string inputs đều được sanitize: strip null bytes, normalize Unicode.

---

## 7. Privacy

- Context content (prompt, reasoning) không được dùng để train model của GitWhy
- Aggregate metadata (domain statistics, tool usage counts) có thể dùng cho product analytics
- Dev có thể xóa toàn bộ cloud data: `DELETE /v1/account` — xóa tất cả contexts + API keys
- GDPR: delete request được xử lý trong 30 ngày

---

*Cập nhật lần cuối: 2026-06-20*
