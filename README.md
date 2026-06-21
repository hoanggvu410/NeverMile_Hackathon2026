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

Binary sẽ xuất hiện ở `./git-why.exe` (Windows) hoặc `./git-why`.

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

## Cấu trúc / Structure

```
cmd/git-why/         CLI (Cobra) — git why save / search / list / reindex / mcp install
internal/context/    Lưu trữ local + parse whyspec markdown
internal/graph/      Claim graph + tripwire (SQLite: graph.db)
internal/cache/      Semantic cache (SQLite: semantic.db)
internal/mcp/        MCP tool definitions (gitwhy_save, gitwhy_search, gitwhy_tripwire...)
mcp/                 MCP stdio server — được Claude/Cursor/Windsurf spawn
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
