# 10 — MCP Integration

---

## 1. Cài đặt MCP Server

### Bước 1: Install package

```bash
# npm (recommended)
npm install -g gitwhy-mcp

# macOS / Linux (install script — Go binary)
curl -sSL https://dl.gitwhy.dev/install.sh | bash

# macOS Homebrew
brew install gitwhy-cli/tap/git-why

# Windows Scoop
scoop bucket add gitwhy https://github.com/quanng28/gitwhy-scoop-bucket
scoop install git-why
```

### Bước 2: Authenticate (cloud features)

```bash
git why setup
# → Mở browser, đăng nhập app.gitwhy.dev
# → Copy API key, paste vào terminal
# → Config lưu tại ~/.config/gitwhy/config.json
```

Hoặc dùng API key trực tiếp (headless/CI):

```bash
export GITWHY_API_KEY=gw_live_xxxxxxxxxxxxx
```

---

## 2. Claude Code / Claude Desktop

Thêm vào `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "gitwhy": {
      "command": "gitwhy-mcp",
      "args": []
    }
  }
}
```

Hoặc dùng `npx` (không cần global install):

```json
{
  "mcpServers": {
    "gitwhy": {
      "command": "npx",
      "args": ["-y", "gitwhy-mcp"]
    }
  }
}
```

**Verify:**
```
> /mcp
# → Phải thấy "gitwhy" trong danh sách MCP servers
```

---

## 3. Cursor

Tạo hoặc chỉnh `.cursor/mcp.json` trong project:

```json
{
  "mcpServers": {
    "gitwhy": {
      "command": "npx",
      "args": ["-y", "gitwhy-mcp"]
    }
  }
}
```

---

## 4. Windsurf

Thêm vào Windsurf MCP settings:

```json
{
  "mcpServers": {
    "gitwhy": {
      "command": "gitwhy-mcp",
      "args": []
    }
  }
}
```

---

## 5. VS Code (Copilot)

Tạo `.vscode/mcp.json` trong workspace:

```json
{
  "servers": {
    "gitwhy": {
      "command": "npx",
      "args": ["-y", "gitwhy-mcp"]
    }
  }
}
```

---

## 6. Cline

Thêm vào Cline MCP Servers trong VS Code settings:

```json
{
  "cline.mcpServers": {
    "gitwhy": {
      "command": "gitwhy-mcp",
      "args": [],
      "env": {
        "GITWHY_API_KEY": "gw_live_xxxxxxxxxxxxx"
      }
    }
  }
}
```

---

## 7. Smithery / Glama (Remote)

Khi dùng qua remote MCP registry (Smithery, Glama):
- Không cần install local
- Paste API key vào registry settings khi prompted
- URL: `https://registry.modelcontextprotocol.io/servers/gitwhy-mcp`

---

## 8. Environment Variables

| Variable | Mô tả | Default |
|----------|-------|---------|
| `GITWHY_API_KEY` | API key cho cloud features | `~/.config/gitwhy/config.json` |
| `GITWHY_API_URL` | Override cloud API endpoint | `https://api.gitwhy.dev` |
| `GITWHY_DEBUG` | Enable verbose logging | `false` |
| `GITWHY_AUTOSAVE` | Auto-save hook enabled (v0.2) | `true` |

---

## 9. Recommended Agent Workflow

Cách tốt nhất để configure AI agent sử dụng GitWhy:

### System prompt gợi ý (thêm vào agent context)

```
You have access to gitwhy MCP tools for context management:

BEFORE starting any significant task:
1. Call gitwhy_search to find relevant past decisions
2. Incorporate found context into your approach

AFTER completing a task and committing:
1. Call gitwhy_save with: prompt (what user asked), reasoning (your approach),
   decisions (key choices), rejected_alternatives (what you considered but didn't do)
2. If this is for a PR, call gitwhy_post_pr with the PR number

This ensures reasoning is preserved for your team and future sessions.
```

### Claude Code — CLAUDE.md setup

```markdown
## GitWhy Context Management

Always use gitwhy tools during development:
- **Before coding**: `gitwhy_search` for relevant past decisions  
- **After committing**: `gitwhy_save` with full reasoning
- **Before pushing**: `gitwhy_post_pr` if opening a PR
```

---

## 10. Verify Installation

```bash
# CLI check
git why status

# Expected output:
# ✓ Git repository detected
# ✓ API key: gw_live_xxx*** (team plan)
# ✓ 3 contexts pending sync
# Last sync: 2026-06-20 10:35:00

# MCP check (sau khi thêm vào agent config)
# Trong Claude Code: /mcp → thấy "gitwhy" listed
```

---

## 11. Troubleshooting

| Vấn đề | Nguyên nhân | Giải quyết |
|--------|-------------|------------|
| `gitwhy-mcp: command not found` | Chưa install | `npm install -g gitwhy-mcp` |
| `Not a git repository` | Chạy ngoài git repo | `cd` vào project folder |
| `Invalid API key` | Key sai hoặc revoked | `git why setup` lại |
| `Quota exceeded` | Free: 20 syncs/tháng đã hết | Upgrade Team hoặc chờ reset |
| `GitHub App not installed` | gitwhy-bot chưa add vào repo | Install tại app.gitwhy.dev/github |
| MCP tools không xuất hiện | Config JSON sai | Validate JSON, restart agent |

---

*Cập nhật lần cuối: 2026-06-20*
