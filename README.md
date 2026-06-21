# GitWhy — NeverMile Hackathon 2026

> **Context layer cho git.** Lưu lại lý do đằng sau mỗi thay đổi code — để AI agent sau này không vi phạm quyết định cũ.
>
> The context layer for git. Saves the reasoning behind AI code changes so future agents don't break old decisions.

---

## Cái này làm gì? / What does it do?

Bình thường khi bạn dùng AI để code, AI đó không biết những quyết định đã được đưa ra trước đó. GitWhy giải quyết điều đó bằng cách:

When you use AI to code, it doesn't know about past decisions. GitWhy fixes that by:

1. **Lưu lý do** — `gitwhy_save` ghi lại tại sao bạn làm thứ gì đó / **Saving reasoning** — why you did something
2. **Tìm kiếm** — `gitwhy_search` tìm lại quyết định cũ liên quan / **Searching** — find related past decisions
3. **Tripwire** — `gitwhy_tripwire` cảnh báo AI *trước khi* nó phá vỡ thứ gì đó / **Tripwire** — warns AI *before* it breaks something

---

## Cài đặt nhanh / Quick start

### Build

```bash
go build ./...
```

Lệnh này tạo **3 binary** (mỗi cái một việc) / This produces **3 binaries** (each does one job):

| Binary | Từ / From | Vai trò / Role |
|---|---|---|
| `git-why.exe` | `cmd/git-why/` | CLI — `git why save / search / list / reindex / mcp install / hook install` |
| `gitwhy2-mcp.exe` | `mcp/` | **MCP server** (stdio) — Claude/Cursor tự spawn để agent gọi `gitwhy_*` tools |
| `serve.exe` | `cmd/serve/` | **HTTP API** (localhost:7420) — đọc `.git/gitwhy/` cho dashboard web |

> Bạn không chạy `gitwhy2-mcp.exe` bằng tay — Claude/Cursor spawn nó qua config. `git-why.exe` và `serve.exe` thì chạy trực tiếp.
>
> You don't run `gitwhy2-mcp.exe` by hand — Claude/Cursor spawn it via config. `git-why.exe` and `serve.exe` you run directly.

### Cài MCP cho một repo / Install MCP for a repo

```bash
cd /path/to/your-repo
/path/to/git-why mcp install
```

Lệnh này ghi 3 thứ vào repo của bạn:
- `.claude/settings.json` — đăng ký MCP server cho Claude Code
- `.cursor/mcp.json` — đăng ký MCP server cho Cursor
- `AGENTS.md` — contract bắt buộc để AI agent biết dùng tripwire đúng cách

Mỗi repo có config riêng — không dùng `--global`.

This writes 3 things into your repo:
- `.claude/settings.json` — registers the MCP server for Claude Code
- `.cursor/mcp.json` — registers the MCP server for Cursor
- `AGENTS.md` — required contract so AI agents know how to use the tripwire correctly

Each repo gets its own config — don't use `--global`.

### Hook (optional)

```bash
git why hook install
```

Tự động bắt commit hash sau mỗi `git commit` và đính vào context tiếp theo được lưu.
Auto-captures commit hashes after each `git commit` and attaches them to the next save.

### Reindex sau khi sửa file tay / Reindex after manual edits

```bash
git why reindex
```

Đồng bộ `graph.db` với các file markdown trong `.git/gitwhy/contexts/`.
Syncs `graph.db` with markdown files in `.git/gitwhy/contexts/`.

---

## Dashboard web / Web dashboard

Giao diện tối, đọc thẳng dữ liệu trong repo của bạn — không đăng nhập, không cloud.
A dark dashboard that reads your repo's data directly — no login, no cloud.

```bash
# 1. Build HTTP API (một lần / once)
go build -o serve.exe ./cmd/serve/

# 2. Chạy API trong repo của bạn / run the API from inside your repo
./serve.exe                       # → http://localhost:7420
# ./serve.exe -repo /path/to/repo # hoặc trỏ repo khác / or point at another repo

# 3. Chạy dashboard / run the dashboard (cửa sổ terminal khác / separate terminal)
cd web
npm install      # lần đầu / first time only
npm run dev      # → http://localhost:3000
```

Mở `http://localhost:3000` → tự vào `/dashboard`. Bạn sẽ thấy context vừa lưu, claim graph, heatmap, và search.
Open `http://localhost:3000` → redirects to `/dashboard`. You'll see your saved contexts, the claim graph, heatmap, and search.

> Dashboard **không** nói chuyện với MCP server. MCP **ghi** vào `.git/gitwhy/`, `serve` **đọc** lại — nên context bạn save (qua agent) hiện ngay sau khi reload.
>
> The dashboard does **not** talk to the MCP server. MCP **writes** to `.git/gitwhy/`, `serve` **reads** it back — so contexts you save (via your agent) appear right after a reload.

Chi tiết từng bước + troubleshooting: xem **Test Case 6** trong `test-cases.md`.
Step-by-step + troubleshooting: see **Test Case 6** in `test-cases.md`.

---

## Cấu trúc / Structure

```
cmd/git-why/         CLI (Cobra) — git why save / search / list / reindex / mcp install
cmd/serve/           HTTP API (localhost:7420) — đọc .git/gitwhy/ cho dashboard
mcp/                 MCP stdio server — được Claude/Cursor/Windsurf spawn
internal/context/    Lưu trữ local + parse whyspec markdown
internal/graph/      Claim graph + tripwire (SQLite: graph.db)
internal/cache/      Semantic cache (SQLite: semantic.db)
internal/mcp/        MCP tool definitions (gitwhy_save, gitwhy_search, gitwhy_tripwire...)
web/                 Next.js dashboard (App Router · Tailwind · React Flow · Framer Motion)
docs/                Tài liệu sản phẩm / Product documentation
```

---

## Cách hoạt động / How it works

```
Agent gọi gitwhy_save(context)
    ↓
1. Lưu markdown vào .git/gitwhy/contexts/
2. Tách ra tối đa 7 "claims" (quyết định quan trọng nhất)
3. Mỗi claim → 3 embedding vectors (claim / retrieval / interrupt)
4. Lưu vào graph.db
    ↓
Lần sau agent gọi gitwhy_tripwire(plan)
    ↓
So sánh plan với interrupt vectors
Nếu match → trả về interrupt=true + claim liên quan
```

---

## Module

```
github.com/hoanggvu410/NeverMile_Hackathon2026
```

---

## Docs

Xem `docs/` để biết thêm chi tiết về:
- `docs/04-functional-requirements.md` — GRAPH, SEARCH, HOOK specs
- `docs/07-data-models.md` — SQLite schemas
- `docs/08-api-specifications.md` — MCP tool schemas
- `AGENTS.md` — **đọc trước khi dùng MCP tools** — required contract for AI agents
- `test-cases.md` — Hướng dẫn test cho teammate / Teammate testing guide
- `web/README.md` — chi tiết dashboard (API routes, screens, stack) / dashboard details
