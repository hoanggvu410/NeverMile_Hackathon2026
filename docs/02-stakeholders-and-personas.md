# 02 — Stakeholders & Personas

---

## Stakeholders

| Stakeholder | Vai trò | Mối quan tâm chính |
|-------------|---------|-------------------|
| **GitWhy Team** | Vận hành platform | Adoption, uptime, developer experience |
| **Individual Developer** | User chính | Không muốn thêm friction vào workflow |
| **Team Lead / Tech Lead** | Decision maker mua Team plan | ROI: giảm review time, tăng knowledge sharing |
| **PR Reviewer** | Đọc context trên PR | Hiểu được *why* không chỉ *what* |
| **AI Coding Agent** | Client chính của MCP server | Reliable tool schema, fast response |
| **GitHub** | Integration partner | PR bot permissions, app webhook |

---

## Personas

### Persona 1 — Individual Developer (Dev đơn lẻ)

> Freelancer hoặc dev đang dùng AI coding agent hàng ngày.

**Thông tin:**
- Dùng Claude Code / Cursor / Windsurf để code hàng ngày
- Không muốn context quan trọng mất khi đóng chat
- Quan tâm đến chi phí token

**Nhu cầu:**
- Save context nhanh, không cần context switch
- Tìm lại quyết định cũ khi cần
- Xem lịch sử reasoning theo project / domain

**Workflows chính:**
1. AI agent tự gọi `gitwhy_save` sau khi commit
2. Hỏi `gitwhy_search "tại sao đổi AWS"` → nhận chain of decisions
3. `gitwhy_sync` để backup lên cloud

**Pain points hiện tại:**
- Phải manually trigger save → hay quên
- Không tìm lại được context của 2 tuần trước
- Token cost tăng khi phải re-explain context cho AI

---

### Persona 2 — Team Lead / Tech Lead

> Quản lý team 3-10 dev, dùng AI agents heavily.

**Thông tin:**
- Chịu trách nhiệm về code quality và architecture decisions
- Đau đầu vì PR review mất thời gian — không biết context
- Muốn mới onboard nhanh, không cần hỏi lại

**Nhu cầu:**
- Team chia sẻ context với nhau automatically
- PR reviewer thấy reasoning không cần hỏi dev
- Search được "tại sao team chọn approach X" qua dashboard

**Workflows chính:**
1. Setup gitwhy cho toàn team, cấu hình post-commit hook
2. Review PR → xem gitwhy-bot comment với full reasoning
3. Onboard dev mới → chỉ họ search context cũ

**Trigger mua Team plan:** Khi muốn `gitwhy_publish` để team cùng thấy context.

---

### Persona 3 — PR Reviewer

> Dev khác trong team review pull request.

**Thông tin:**
- Nhận PR notification, vào GitHub review
- Hiện tại chỉ thấy diff, không biết *tại sao* thay đổi
- Hay phải comment hỏi thêm → delay merge

**Nhu cầu:**
- Thấy ngay: prompt gốc, decisions, alternatives bị bỏ
- Không cần rời GitHub để tìm context
- Hiểu được risk areas trước khi approve

**Workflows chính:**
1. Mở PR → thấy gitwhy-bot comment với context summary
2. Đọc decisions + rejected_alternatives → hiểu trade-off
3. Review thông minh hơn, ít câu hỏi hơn

---

### Persona 4 — AI Coding Agent

> Claude Code, Cursor, Windsurf, Cline... — MCP client chính.

**"Thông tin":**
- Kết nối với gitwhy-mcp qua stdio MCP protocol
- Gọi tools để save / search / retrieve context
- Cần tool schema rõ ràng, response nhanh, error handling tốt

**Nhu cầu:**
- `gitwhy_save` với đầy đủ fields (prompt, reasoning, decisions, rejected_alternatives)
- `gitwhy_search` trả về relevant context để inject vào prompt tiếp theo
- `gitwhy_status` để biết cần setup gì trước khi dùng

**Workflows chính:**
1. Trước khi bắt đầu task: `gitwhy_search` để load context liên quan
2. Sau khi hoàn thành + commit: `gitwhy_save` với full reasoning
3. Trước khi push: `gitwhy_sync` + `gitwhy_post_pr`

---

### Persona 5 — System Admin (GitWhy Team)

> Người vận hành GitWhy cloud platform.

**Thông tin:**
- Quản lý cloud backend, PostgreSQL, GitHub App
- Monitor uptime, context sync, PR bot

**Nhu cầu:**
- Dashboard monitoring sync jobs, error rates
- Manage team subscriptions
- Debug gitwhy-bot webhook failures

---

*Cập nhật lần cuối: 2026-06-20*
