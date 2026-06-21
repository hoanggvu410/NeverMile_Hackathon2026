# 09 — Permissions Matrix

---

## 1. Tiers

| Tier | Target | Giá | Mô tả |
|------|--------|-----|-------|
| **Free** | Individual dev | $0 | Local full-featured, cloud sync giới hạn |
| **Team** | Team 2+ | $20/tháng | Unlimited sync, team publish, PR bot |
| **Enterprise** | >50 dev | TBD | Self-hosted, SSO, audit logs |

---

## 2. Feature Matrix

| Feature | Free | Team | Enterprise |
|---------|------|------|-----------|
| **Local save** (gitwhy_save) | ✅ Unlimited | ✅ Unlimited | ✅ Unlimited |
| **Local search** (gitwhy_search) | ✅ Unlimited | ✅ Unlimited | ✅ Unlimited |
| **Local list / get** | ✅ Unlimited | ✅ Unlimited | ✅ Unlimited |
| **gitwhy_status** | ✅ | ✅ | ✅ |
| **Cloud sync** (gitwhy_sync) | ✅ 20/tháng | ✅ Unlimited | ✅ Unlimited |
| **Team publish** (gitwhy_publish) | ❌ | ✅ | ✅ |
| **Team search** (cloud) | ❌ | ✅ | ✅ |
| **PR comment** (gitwhy_post_pr) | ✅ Unlimited | ✅ Unlimited | ✅ Unlimited |
| **Web Dashboard** | ✅ Own contexts | ✅ Own + team | ✅ Full |
| **API keys** | ✅ Max 3 | ✅ Unlimited | ✅ Unlimited |
| **Repositories** | 1 | Unlimited | Unlimited |
| **Context Graph** (v0.2) | ✅ Local | ✅ Cloud graph | ✅ |
| **Semantic cache** (v0.2) | ✅ Local | ✅ Cloud | ✅ |
| **Auto-save hook** (v0.2) | ✅ | ✅ | ✅ |
| **Self-hosted** | ❌ | ❌ | ✅ |
| **SSO (SAML/LDAP)** | ❌ | ❌ | ✅ |
| **Audit logs** | ❌ | ❌ | ✅ |
| **SLA** | None | 99.5% | 99.9% |

---

## 3. API Key Scope

Mỗi API key có scope gắn với user và tier:

```
Free API key:
  - gitwhy_sync: ✅ (quota: 20/tháng)
  - gitwhy_publish: ❌
  - gitwhy_post_pr: ✅
  - cloud search: ❌ (chỉ own contexts)

Team API key:
  - gitwhy_sync: ✅ (unlimited)
  - gitwhy_publish: ✅
  - gitwhy_post_pr: ✅
  - cloud search: ✅ (team contexts)
```

---

## 4. Web Dashboard Permissions

| Action | Unauthenticated | Free User | Team Member | Team Owner |
|--------|----------------|-----------|-------------|------------|
| View own contexts | ❌ | ✅ | ✅ | ✅ |
| View team contexts | ❌ | ❌ | ✅ | ✅ |
| Search own contexts | ❌ | ✅ | ✅ | ✅ |
| Search team contexts | ❌ | ❌ | ✅ | ✅ |
| Publish context | ❌ | ❌ | ✅ | ✅ |
| Unpublish context | ❌ | ❌ | Own only | ✅ |
| Create API key | ❌ | ✅ (max 3) | ✅ | ✅ |
| Revoke API key | ❌ | Own | Own | ✅ All |
| Invite team member | ❌ | ❌ | ❌ | ✅ |
| Remove team member | ❌ | ❌ | ❌ | ✅ |
| Manage billing | ❌ | ❌ | ❌ | ✅ |

---

## 5. GitHub App Permissions

gitwhy-bot GitHub App cần các permissions tối thiểu:

| Permission | Scope | Lý do |
|------------|-------|-------|
| `pull_requests: write` | Repository | Post PR comments |
| `contents: read` | Repository | Verify repo exists (optional) |
| `metadata: read` | Repository | Bắt buộc bởi GitHub |

**Lưu ý:** gitwhy-bot **không** cần đọc code. Scope tối thiểu cho privacy.

---

## 6. Free Tier Quota Enforcement

```
Mỗi lần gọi gitwhy_sync:
  → Cloud API check: SELECT sync_count FROM sync_quota_usage
                     WHERE user_id = ? AND month = date_trunc('month', NOW())
  
  → count >= 20: return 429 CONTEXT_QUOTA_EXCEEDED
  → count < 20:  INSERT/UPDATE sync_quota_usage (increment)
                 proceed with sync
```

Quota reset: ngày đầu mỗi tháng (UTC).

Idempotent: sync lại context đã sync (same `local_id`) → không tốn quota.

---

## 7. Business Rules

| Rule | Mô tả |
|------|-------|
| PR comments không tính vào quota | Free tier unlimited `gitwhy_post_pr` |
| Local features không cần API key | save / get / search local / list — hoàn toàn offline |
| Team bắt đầu từ owner | Owner tạo team, mời members |
| Context ownership | Context thuộc về user tạo ra, không transfer |
| Publish = team-readable | Unpublish chỉ owner context hoặc team owner |

---

*Cập nhật lần cuối: 2026-06-20*
