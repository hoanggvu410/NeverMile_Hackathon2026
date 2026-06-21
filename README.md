# GitWhy

GitWhy là **context layer cho git**: nó lưu lại *why* đằng sau code changes để AI agents sau này không phải đoán lại decision cũ.

Core loop:

```text
agent saves context -> GitWhy extracts claims -> graph indexes decisions -> next agent searches/tripwires before editing
```

## Làm gì?

- `gitwhy_save`: lưu reasoning, decisions, rejected alternatives, risks vào `.git/gitwhy/contexts/`.
- `gitwhy_search`: tìm lại decision chain bằng claim graph + semantic cache.
- `gitwhy_tripwire`: check plan trước khi sửa code; nếu đụng decision cũ thì trả `interrupt=true`.
- Web dashboard: đọc local `.git/gitwhy/` để xem contexts, claim graph, heatmap, search.

GitWhy không phải chatbot. Nó là local memory + graph + MCP tools cho coding agents.

## Run nhanh

Build 3 binaries:

```powershell
go build -o git-why.exe ./cmd/git-why
go build -o gitwhy2-mcp.exe ./mcp
go build -o serve.exe ./cmd/serve
```

Run dashboard backend:

```powershell
.\serve.exe
```

API chạy ở:

```text
http://localhost:7420
```

Run web dashboard:

```powershell
cd web
npm install
npm run dev
```

Dashboard:

```text
http://localhost:3000/dashboard
```

## Install GitWhy vào một repo khác

Trong repo muốn dùng memory:

```powershell
C:\path\to\git-why.exe mcp install
```

Lệnh này tạo repo-local config:

- `.claude/settings.json`
- `.cursor/mcp.json`
- `AGENTS.md`

Quan trọng: MCP config phải có `cwd` là root của repo đó. Không dùng global config cho nhiều repo.

## Dùng CLI

```powershell
.\git-why.exe save
.\git-why.exe search "why did we remove kafka"
.\git-why.exe reindex
.\git-why.exe mcp install
.\git-why.exe hook install
```

`reindex` dùng khi markdown context trong `.git/gitwhy/contexts/` bị sửa tay và cần sync lại `graph.db`.

## Các file `.exe` và `.bat` làm gì?

| File | Là gì | Chạy khi nào |
|---|---|---|
| `git-why.exe` | CLI chính. Commands: `save`, `search`, `reindex`, `mcp install`, `hook install`. | Chạy trực tiếp trong terminal. |
| `gitwhy2-mcp.exe` | MCP stdio server. Exposes tools như `gitwhy_save`, `gitwhy_search`, `gitwhy_tripwire`. | Claude/Cursor spawn tự động qua MCP config; thường không chạy tay. |
| `serve.exe` | Local HTTP API cho dashboard. Đọc `.git/gitwhy/` của repo. | Chạy tay trước khi mở web dashboard. |
| `gitwhy2-mcp-wrapper.bat` | Windows wrapper: `cd` vào repo này rồi chạy `gitwhy2-mcp.exe`. | Dùng bởi MCP config khi client cần một command ổn định trên Windows. |

Tên binary `gitwhy2-mcp.exe` là tên lịch sử. Product name/user-facing name là **GitWhy**.

## Directory map

```text
cmd/git-why/       CLI source
cmd/serve/         local HTTP API for dashboard
mcp/               MCP server entrypoint
internal/context/  whyspec markdown save/load/search
internal/graph/    claim graph, embeddings, tripwire
internal/cache/    semantic cache
internal/mcp/      MCP tool definitions
cloud/             cloud API prototype
db/                cloud DB models + migrations
web/               Next.js dashboard
docs/              product/spec/algorithm docs
tests/             manual/test scripts
```

## Docs cần đọc

| Doc | Khi nào đọc |
|---|---|
| `docs/algo.md` | Quan trọng nhất: save/search/tripwire algorithm, tables, weights, examples. |
| `docs/04-functional-requirements.md` | Feature requirements cho graph/search/hook/MCP. |
| `docs/07-data-models.md` | SQLite/Postgres schema details. |
| `docs/08-api-specifications.md` | MCP tools + REST API shapes. |
| `docs/13-environment-setup.md` | Local setup details. |
| `docs/15-roadmap.md` | Hackathon scope + next work. |
| `test-cases.md` | Manual testing guide. |
| `web/README.md` | Dashboard-specific notes. |
| `AGENTS.md` | Agent contract: call tripwire before non-trivial edits. |

## Demo mental model

```text
1. Agent finishes work.
2. Agent calls gitwhy_save with decisions/reasoning.
3. GitWhy writes markdown + graph rows.
4. Later agent creates a plan.
5. Agent calls gitwhy_tripwire before edits.
6. If plan conflicts with saved decisions, GitWhy interrupts.
```

That is the product: make old reasoning queryable before new code gets changed.
