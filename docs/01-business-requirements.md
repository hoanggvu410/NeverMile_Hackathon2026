# 01 — Business Requirements

---

## 1. Bối cảnh

Trong chu kỳ phát triển phần mềm hiện đại, AI coding agents đang generate ra phần lớn code. Nhưng toàn bộ reasoning đằng sau code — prompt gốc, tại sao chọn approach này, tại sao bỏ Kafka, tại sao đổi AWS — **biến mất khi đóng chat window**.

Hệ quả:
- Dev mất thời gian hỏi lại nhau về quyết định kỹ thuật đã có
- PR reviewer không hiểu được *tại sao* code thay đổi, chỉ thấy diff
- Onboarding dev mới kéo dài vì không có audit trail của reasoning
- AI agent bắt đầu session mới không có context của session trước → duplicate effort

**GitWhy** là agent memory layer đứng giữa để giải quyết bài toán này: pre-compute context để agents stop rediscovering everything từ đầu. Agent query GitWhy trước khi touch LLM — nhận lean, focused context thay vì crawl 50 files. LLM stops being a search engine và starts being a reasoner.

---

## 2. Business Model

- **Freemium SaaS** — cá nhân dùng free, team trả phí
- **Free tier**: 1 repository, 20 cloud syncs/tháng, unlimited local save
- **Team plan**: $20/tháng, unlimited repositories, unlimited sync, team publish, PR bot
- **Chiến lược**: bán được khi team muốn *chia sẻ* context — lúc đó mới cần subscription

---

## 3. Yêu cầu nghiệp vụ cốt lõi

### BR-01: Lưu AI reasoning gắn với commit

> Mỗi AI coding session phải có thể lưu structured context (prompt, decisions, trade-offs) và tự động link với git commit.

- Context phải tồn tại sau khi chat window đóng
- Link với commit hash để trace back được lịch sử
- Lưu local trước (offline-first), sync cloud khi cần

### BR-02: Tìm kiếm semantic context

> Developer phải có thể hỏi "tại sao bỏ Kafka" và nhận lại quyết định + PR + cost note liên quan.

- Tìm theo keyword, natural language, domain/topic
- Trả về trong < 3 giây (demo target)
- Tìm được qua MCP tool, CLI, và web dashboard

### BR-03: Chia sẻ context trong team

> Team có thể xem reasoning của nhau — đặc biệt khi review PR.

- Sync context lên cloud (private)
- Publish để team cùng thấy
- gitwhy-bot tự post summary lên GitHub PR comment

### BR-04: Context Graph (v0.2)

> Các context phải có thể liên kết với nhau theo chain of decisions.

- Mỗi commit/PR/msg → 1 node trong graph
- Auto-link A→B qua typed edges (CAUSED_BY, CONSTRAINED_BY, INVALIDATES, CONTRADICTS, DEPENDS_ON) trong SQLite adjacency table
- Query "tại sao đổi AWS" → ra luôn decision chain + config + PR

### BR-05: Zero-friction capture (v0.2)

> Dev không nên phải nhớ tay trigger save.

- Post-commit hook tự động trigger `gitwhy_save`
- Semantic cache: query deduplication layer. >90% cosine similarity với cached query → return cached result, 0 LLM token
- Agents hiện tốn 2:1 input-to-output token ratio cho context communication. Cache eliminates redundant LLM calls cho repeated queries — actual reduction depends on query repetition rate trong codebase cụ thể

### BR-06: Web Dashboard

> Team cần giao diện web để quản lý contexts, members, và API keys.

- List, search, view contexts tại `app.gitwhy.dev`
- Quản lý team members + publish settings
- Manage API keys

---

## 4. Constraints

| Constraint | Mô tả |
|------------|-------|
| GitHub-only PR bot | GitLab / Bitbucket chưa hỗ trợ (phase 2) |
| Local storage | Context lưu tại `.git/gitwhy/contexts/` — cần git repo |
| Free tier limits | 1 repo, 20 syncs/tháng |
| MCP transport | stdio only — không HTTP/SSE cho local server |
| No code execution | GitWhy là persistence layer, không phải AI/LLM |

---

## 5. Success Metrics (v0.1 → v0.2)

| Metric | v0.1 Target | v0.2 Target |
|--------|------------|------------|
| Demo time "tại sao bỏ Kafka" | < 5s | < 3s |
| Demo bill per query | < $0.05 | $0.01 |
| Semantic cache hit rate | N/A | > 70% cho repeated queries |
| Time-to-first-context (new user) | < 10 phút | < 5 phút |
| PR context coverage | Manual | Auto (post-commit hook) |

---

## 6. Go-to-Market

| Tier | Target | Giá | Trigger |
|------|--------|-----|---------|
| **Free** | Individual dev | $0 | Install + use locally |
| **Team** | Team 2-5+ dev | $20/tháng | Muốn chia sẻ context với nhau |
| **Enterprise** | >50 dev | TBD | Self-hosted, audit logs, SSO |

Chiến lược: dev dùng thật trước (không phải slideware), sau đó team mua khi thấy giá trị từ shared context.

---

*Cập nhật lần cuối: 2026-06-20*
