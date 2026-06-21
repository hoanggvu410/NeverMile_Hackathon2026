# GitWhy — Hướng dẫn test cho teammate / Teammate Testing Guide

> Đọc cái này nếu bạn muốn test GitWhy trong repo của mình.
> Read this if you want to test GitWhy in your own repo.

---

## Cần gì trước khi bắt đầu / Prerequisites

- Go 1.21+ (`go version`)
- Git repo bất kỳ / Any git repo
- Claude Code hoặc Cursor (để dùng MCP tools)
- OpenAI API key (tuỳ chọn — không có thì tự động dùng local embedding)

> **Offline by default.** GitWhy runs fully local with no API key. Set `OPENAI_API_KEY` only if you want higher-quality semantic embeddings.

---

## Bước 1 — Build binary / Step 1 — Build the binary

```bash
cd path/to/hackathon
go build ./...
```

Sau đó bạn có file `git-why.exe` (Windows) hoặc `git-why` (Mac/Linux).

---

## Bước 2 — Cài vào repo của bạn / Step 2 — Install into your repo

```bash
cd path/to/YOUR-repo
path/to/git-why mcp install
```

Lệnh này làm 3 thứ:
- Tạo `.claude/settings.json` (cho Claude Code)
- Tạo `.cursor/mcp.json` (cho Cursor)
- Ghi `AGENTS.md` vào root repo — chứa contract bắt buộc cho AI agent

Cả hai đều trỏ đúng vào repo của bạn. **Mỗi repo phải install riêng** — không dùng `--global`.

This command does 3 things:
- Creates `.claude/settings.json` for Claude Code
- Creates `.cursor/mcp.json` for Cursor
- Writes `AGENTS.md` into the repo root — the required AI agent contract

Both point at your repo's root. **Install separately per repo** — don't use `--global`.

---

## Bước 3 — (Tuỳ chọn) Set OpenAI key / Step 3 — (Optional) Set OpenAI key

GitWhy chạy local hoàn toàn không cần API key. Bỏ qua bước này nếu muốn test offline.

GitWhy runs fully local with no API key. Skip this step to test offline.

```bash
# Windows (optional)
set OPENAI_API_KEY=sk-...

# Mac/Linux (optional)
export OPENAI_API_KEY=sk-...
```

---

## Bước 4 — Mở Claude Code trong repo của bạn / Step 4 — Open Claude Code in your repo

Mở Claude Code (hoặc Cursor) với working directory là repo của bạn. MCP server sẽ tự khởi động.

Open Claude Code (or Cursor) with your repo as the working directory. The MCP server starts automatically.

Kiểm tra MCP hoạt động bằng cách nhờ Claude:
Verify MCP works by asking Claude:

> "List the gitwhy tools available"

Bạn sẽ thấy: `gitwhy_save`, `gitwhy_search`, `gitwhy_tripwire`, `gitwhy_list`, `gitwhy_get`.

---

## Test Case 1 — Lưu một context / Save a context

**Mục tiêu:** Lưu lý do của một quyết định gần đây trong repo.
**Goal:** Save the reasoning behind a recent decision in your repo.

Nói với Claude:

```
Please save a gitwhy context for the decision to use [technology X] in this project.
- domain: backend/architecture
- topic: framework-choice
Include: what we decided, why, what we rejected, and what could go wrong.
```

**Kết quả mong đợi / Expected result:**
- Claude gọi `gitwhy_save`
- File markdown xuất hiện trong `.git/gitwhy/contexts/`
- Claude báo cáo context ID (dạng `ctx_xxxxxxxx`)
- Graph.db được update (có thể check bằng `git why reindex`)

**Lỗi thường gặp / Common issues:**
- "graph unavailable" → OpenAI key chưa set
- "context ID required" → Claude thiếu trường ID trong save — thử lại với prompt rõ hơn

---

## Test Case 2 — Tìm kiếm context / Search contexts

**Mục tiêu:** Tìm lại quyết định vừa lưu bằng câu hỏi tự nhiên.
**Goal:** Find the saved decision using a natural language question.

Sau khi đã save ít nhất 1-2 contexts, nói với Claude:

```
Search gitwhy: why did we choose our backend framework?
```

**Kết quả mong đợi / Expected result:**
- Claude gọi `gitwhy_search`
- Trả về claim text + score + domain/topic
- Score > 0.1 nghĩa là có kết quả liên quan

**Thử câu hỏi xa hơn / Try a semantically distant question:**
```
Search gitwhy: what should I know before changing the API server?
```
Nếu retrieval vectors hoạt động, câu hỏi này vẫn nên tìm ra context về framework choice.

---

## Test Case 3 — Tripwire (quan trọng nhất / most important)

**Mục tiêu:** Kiểm tra xem tripwire có cảnh báo đúng không khi agent plan sắp vi phạm quyết định cũ.
**Goal:** Check that tripwire correctly warns when an agent plan would violate a past decision.

**Bước 1:** Lưu một constraint rõ ràng:

```
Save a gitwhy context with domain: backend, topic: framework-lock.
Key decision: "We must use FastAPI for all backend routes. Do not introduce Express or Node.js."
Reasoning: team standardized on Python stack for consistency.
```

**Bước 2:** Giả vờ tạo một plan vi phạm nó. Nói với Claude:

```
Check gitwhy tripwire before I start this task:
- I'm planning to add a Node.js Express service for the notification endpoint
- Files I'll touch: backend/notifications/, package.json
- Concepts: websocket, express, node backend
```

**Kết quả mong đợi / Expected result:**

```json
{
  "interrupt": true,
  "message": "Relevant prior decision: We must use FastAPI for all backend routes...",
  "candidates": [{ "claim": "...", "score": 0.xx }]
}
```

**Nếu `interrupt: false`:** Claim có thể chưa được index vào graph. Chạy:

```bash
git why reindex
```

Rồi thử lại.

---

## Test Case 4 — List và Get / List and Get

**Xem tất cả contexts đã lưu / See all saved contexts:**

```
List my gitwhy contexts
```

Claude gọi `gitwhy_list` → trả về domain/topic tree.

**Đọc một context cụ thể / Read a specific context:**

```
Get gitwhy context ctx_xxxxxxxx
```

Trả về full markdown với reasoning, decisions, rejected alternatives.

---

## Test Case 5 — Hook (optional)

Cài hook để tự động bắt commit hash:

```bash
git why hook install
```

Sau đó:
1. Làm một thay đổi và commit
2. Lưu gitwhy context
3. Context đó sẽ tự động có commit hash đính kèm

---

## Khi có lỗi / Troubleshooting

| Lỗi | Nguyên nhân | Fix |
|---|---|---|
| `graph unavailable` | graph.db không khởi tạo được | Kiểm tra quyền ghi vào `.git/gitwhy/` |
| MCP tools không hiện | MCP chưa install đúng | Chạy lại `git why mcp install` trong repo của bạn |
| `interrupt: false` sau khi save | graph.db chưa sync | `git why reindex` |
| Context không tìm thấy | Sai cwd khi search | Đảm bảo Claude Code mở đúng repo |
| Vague save (warnings xuất hiện) | Nội dung save quá ngắn/chung chung | Xem `AGENTS.md` để biết format đúng |

---

## Lưu ý quan trọng / Important notes

1. **Mỗi repo install riêng** — `git why mcp install` phải chạy trong từng repo
2. **`gitwhy_tripwire` ≠ `gitwhy_search`** — tripwire dùng *trước khi* edit, search dùng để tìm kiếm bình thường
3. **Đọc `AGENTS.md`** trước khi save — nó có BAD/GOOD examples về cách save đúng
4. **Cross-session edges** là `RELATED_CANDIDATE` — chưa xác nhận, chỉ là gợi ý
5. Nếu rebuild binary và binary nằm ở chỗ khác → chạy lại `git why mcp install`

---

## Flow đầy đủ / Full flow

```
1. git why mcp install          (một lần mỗi repo — cũng ghi AGENTS.md tự động)
2. export OPENAI_API_KEY=...    (mỗi terminal session)
3. Mở Claude Code trong repo
4. AGENTS.md đã có sẵn — Claude đọc tự động khi mở repo
5. Làm code change
6. gitwhy_tripwire(plan)        ← TRƯỚC KHI edit
7. Commit
8. gitwhy_save(context)         ← SAU KHI xong
9. git why reindex              ← nếu cần sync
```
