# 00 — Tổng quan hệ thống GitWhy

> **GitWhy** là agent memory layer cho AI coding agents — pre-compute context để agents không phải rediscover từ đầu mỗi session. Thay vì spin up 10 explore agents để tìm "tại sao Kafka bị xóa," agent query GitWhy trước, nhận answer trong 1 query, inject vào context. LLM nhận lean, focused prompt thay vì đọc 50 files.

---

## Vấn đề cốt lõi

Khi AI coding agent (Claude Code, Cursor, Windsurf...) generate hoặc modify code, toàn bộ reasoning — prompt gốc, quyết định thiết kế, alternatives bị loại bỏ, trade-off analysis — chỉ tồn tại trong ephemeral chat window. Session kết thúc, context biến mất vĩnh viễn.

| Stakeholder | Vấn đề |
|-------------|--------|
| **Dev** | 12–15 major context switches mỗi ngày × 23 phút recovery = **>4.5 giờ mất focus** (DEV Community, 2026) |
| **Reviewer** | Chỉ thấy *what* thay đổi, không thấy *why* |
| **Team** | 73% công ty vượt AI budget dù token price giảm 67% — vì **volume từ agent loops**, không phải unit price (Pebblous, 2026) |
| **AI Agent** | Multi-agent systems tốn **2:1 input-to-output token ratio** — phần lớn để communicate context, không phải generate output (arXiv, 2026) |

**Tagline: "GitWhy = why.log, không phải git.log"**

---

## Sản phẩm

GitWhy gồm ba thành phần chính:

| Thành phần | URL | Mô tả |
|------------|-----|-------|
| **MCP Server** | npm: `gitwhy-mcp` | Local stdio server cài vào AI agent |
| **Web Dashboard** | `app.gitwhy.dev` | Quản lý contexts, team, API keys |
| **Product Site** | `gitwhy.dev` | Info, docs, pricing |
| **Cloud Backend** | Internal | Sync, publish, PR bot |

---

## Core Features

### v0.1 (hiện tại — đã ship)

| Feature | Mô tả |
|---------|-------|
| **8 MCP Tools** | save / get / search / list / status / sync / publish / post_pr |
| **Context Schema** | Whyspec — structured format cho AI reasoning |
| **Cloud Sync** | Upload contexts lên cloud, private |
| **Team Publish** | Chia sẻ contexts trong team |
| **PR Bot** | gitwhy-bot post context summary lên GitHub PR comment |
| **Web Dashboard** | Quản lý contexts, API keys tại app.gitwhy.dev |

### v0.2 (roadmap — PRD đã approved)

| Feature | Mô tả |
|---------|-------|
| **Context Graph** | Mỗi context → 1 node. Typed edges: CAUSED_BY, CONSTRAINED_BY, INVALIDATES, CONTRADICTS, DEPENDS_ON. 2-hop traversal trên SQLite adjacency table |
| **Auto-save Hook** | Extend post-commit hook → tự trigger save, không cần manual |
| **Semantic Cache** | Câu hỏi lặp >90% similarity → trả lời ngay, 0 token |
| **Web UI Polish** | Font Inter 14px, line-height 1.6, padding tăng |

---

## Tech Stack

| Layer | Technology | Ghi chú |
|-------|-----------|--------|
| **MCP Server** | Node.js / JavaScript | stdio transport, npm package `gitwhy-mcp` |
| **CLI (gitwhy)** | Go + Cobra | Binary phân phối qua install script / Homebrew / Scoop |
| **Cloud API** | Go | REST API cho sync, publish, PR bot |
| **Database** | PostgreSQL | Cloud context storage |
| **Web Dashboard** | Next.js (App Router) | app.gitwhy.dev |
| **Product Site** | Framer | gitwhy.dev |
| **Context Storage (local)** | Structured Markdown | `.git/gitwhy/contexts/` |

---

## Kiến trúc tổng thể

```
AI Coding Agent (Claude Code / Cursor / Windsurf / ...)
  │  MCP stdio protocol
  ▼
gitwhy-mcp (Node.js local server)
  │
  ├── Local Storage (.git/gitwhy/contexts/) ← immediate, offline
  │
  └── GitWhy Cloud API (Go)
        ├── PostgreSQL (contexts, teams, users)
        ├── GitHub App (gitwhy-bot → PR comments)
        └── Web Dashboard (Next.js — app.gitwhy.dev)
```

---

## Install (nhanh)

```bash
# npm (recommended cho MCP clients)
npm install -g gitwhy-mcp

# macOS / Linux
curl -sSL https://dl.gitwhy.dev/install.sh | bash

# macOS Homebrew
brew install gitwhy-cli/tap/git-why

# Windows Scoop
scoop bucket add gitwhy https://github.com/quanng28/gitwhy-scoop-bucket
scoop install git-why
```

---

## Map: Task → Tài liệu

| Task | Docs |
|------|------|
| Business goals + model | `01-business-requirements.md` |
| Ai dùng GitWhy | `02-stakeholders-and-personas.md` |
| User stories theo epic | `03-user-stories.md` |
| Chi tiết functional | `04-functional-requirements.md` |
| Performance / security | `05-non-functional-requirements.md` |
| Kiến trúc hệ thống | `06-system-architecture.md` |
| Context schema + DB | `07-data-models.md` |
| MCP tool specs + REST API | `08-api-specifications.md` |
| Phân quyền free/team | `09-permissions-matrix.md` |
| Setup MCP client configs | `10-mcp-integration.md` |
| Security & API keys | `11-security.md` |
| Error codes | `12-error-codes.md` |
| Cài đặt local | `13-environment-setup.md` |
| Docker / CI/CD | `14-devops-infrastructure.md` |
| Roadmap v0.1 → v0.2 | `15-roadmap.md` |
| Web UI design spec | `DESIGN.md` |
| Frontend dev plan | `FRONTEND-PLAN.md` |

---

*Cập nhật lần cuối: 2026-06-20*
