# 03 — User Stories

Format: `US-{EPIC}-{ID}: As a {role}, I want to {action} so that {benefit}.`

---

## EPIC 1 — Save Context (SAVE)

| ID | Story | Priority |
|----|-------|----------|
| US-SAVE-01 | Là **AI agent**, tôi muốn gọi `gitwhy_save` với prompt + reasoning + decisions để lưu context của session hiện tại | Must |
| US-SAVE-02 | Là **AI agent**, tôi muốn context tự động link với git commit hash hiện tại khi save | Must |
| US-SAVE-03 | Là **dev**, tôi muốn context được lưu local ngay lập tức, không cần kết nối internet | Must |
| US-SAVE-04 | Là **AI agent**, tôi muốn lưu `rejected_alternatives` để reviewer biết những approach nào đã bị bỏ | Must |
| US-SAVE-05 | Là **dev**, tôi muốn post-commit hook tự trigger save để không phải nhớ tay (v0.2) | Should |
| US-SAVE-06 | Là **AI agent**, tôi muốn nhận context ID ngay sau khi save để reference sau | Must |

---

## EPIC 2 — Retrieve & Search (SEARCH)

| ID | Story | Priority |
|----|-------|----------|
| US-SEARCH-01 | Là **dev**, tôi muốn hỏi "tại sao bỏ Kafka" và nhận lại chain of decisions liên quan | Must |
| US-SEARCH-02 | Là **AI agent**, tôi muốn search context trước khi bắt đầu task mới để tránh duplicate reasoning | Must |
| US-SEARCH-03 | Là **dev**, tôi muốn list contexts theo domain/topic để duyệt qua project history | Must |
| US-SEARCH-04 | Là **AI agent**, tôi muốn retrieve context cụ thể bằng ID để inject vào prompt | Must |
| US-SEARCH-05 | Là **dev**, tôi muốn search qua web dashboard không cần CLI | Should |
| US-SEARCH-06 | Là **AI agent**, tôi muốn semantic cache trả lời câu hỏi lặp mà không tốn token (v0.2) | Should |

---

## EPIC 3 — Sync & Publish (SYNC)

| ID | Story | Priority |
|----|-------|----------|
| US-SYNC-01 | Là **dev**, tôi muốn upload contexts lên cloud để backup và truy cập từ máy khác | Must |
| US-SYNC-02 | Là **team lead**, tôi muốn publish contexts để toàn team có thể tìm kiếm | Must |
| US-SYNC-03 | Là **dev**, tôi muốn biết trạng thái sync (pending / synced / error) | Must |
| US-SYNC-04 | Là **dev**, tôi muốn sync chỉ contexts của repo hiện tại | Should |
| US-SYNC-05 | Là **dev**, tôi muốn contexts private by default, chỉ publish khi chủ động chọn | Must |

---

## EPIC 4 — PR Integration (PR)

| ID | Story | Priority |
|----|-------|----------|
| US-PR-01 | Là **dev**, tôi muốn post context summary lên GitHub PR comment để reviewer hiểu reasoning | Must |
| US-PR-02 | Là **PR reviewer**, tôi muốn thấy: prompt gốc, decisions chính, alternatives bị bỏ ngay trên GitHub | Must |
| US-PR-03 | Là **dev**, tôi muốn PR comment link về web dashboard để reviewer xem full context | Should |
| US-PR-04 | Là **team lead**, tôi muốn gitwhy-bot tự post comment khi push (auto-trigger từ CI) | Should |
| US-PR-05 | Là **dev**, tôi muốn chỉ định PR number + repo khi gọi `gitwhy_post_pr` | Must |

---

## EPIC 5 — Status & Setup (SETUP)

| ID | Story | Priority |
|----|-------|----------|
| US-SETUP-01 | Là **dev**, tôi muốn `gitwhy_status` cho tôi biết setup có đúng không (git repo? API key? pending syncs?) | Must |
| US-SETUP-02 | Là **dev**, tôi muốn setup authentication bằng `git why setup` trong terminal | Must |
| US-SETUP-03 | Là **dev trong CI/CD**, tôi muốn authenticate bằng API key thay vì interactive flow | Must |
| US-SETUP-04 | Là **dev**, tôi muốn tạo và quản lý API keys tại app.gitwhy.dev | Must |

---

## EPIC 6 — Context Graph (GRAPH — v0.2)

| ID | Story | Priority |
|----|-------|----------|
| US-GRAPH-01 | Là **dev**, tôi muốn query "tại sao đổi AWS" và nhận về chain: decision → config → PR theo thứ tự thời gian | Must |
| US-GRAPH-02 | Là **AI agent**, tôi muốn contexts tự động link với nhau khi có semantic similarity | Must |
| US-GRAPH-03 | Là **dev**, tôi muốn visualize context graph trên web dashboard | Should |
| US-GRAPH-04 | Là **dev**, tôi muốn traverse 2-hop: "tại sao commit A dẫn đến decision B" | Should |

---

## EPIC 7 — Web Dashboard (WEB)

| ID | Story | Priority |
|----|-------|----------|
| US-WEB-01 | Là **dev**, tôi muốn đăng nhập dashboard và xem tất cả contexts của mình | Must |
| US-WEB-02 | Là **team lead**, tôi muốn xem contexts đã publish của team | Must |
| US-WEB-03 | Là **dev**, tôi muốn search contexts qua web UI bằng keyword | Must |
| US-WEB-04 | Là **dev**, tôi muốn tạo và revoke API keys tại dashboard | Must |
| US-WEB-05 | Là **team lead**, tôi muốn mời members vào team | Should |
| US-WEB-06 | Là **dev**, tôi muốn xem full context detail (tất cả fields) | Must |

---

*Cập nhật lần cuối: 2026-06-20*
