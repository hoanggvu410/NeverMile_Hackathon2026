# 13 — Environment Setup

---

## 1. Prerequisites

```bash
# Required — cho MCP server (npm package)
node 18+
npm 9+
git 2.x+

# Required — cho Go CLI (alternative/additional)
# Không cần — binary được pre-compiled và download tự động
```

---

## 2. Install

### Option A: npm (recommended cho MCP client users)

```bash
npm install -g gitwhy-mcp

# Verify
gitwhy-mcp --version
# → gitwhy-mcp 0.1.x
```

### Option B: Install script (macOS/Linux — Go binary)

```bash
curl -sSL https://dl.gitwhy.dev/install.sh | bash

# Binary sẽ được cài vào ~/.local/bin/gitwhy
# Tự thêm vào PATH

# Verify
git why --version
# → gitwhy 0.1.x
```

### Option C: Homebrew (macOS/Linux)

```bash
brew install gitwhy-cli/tap/git-why

# Verify
git why --version
```

### Option D: Scoop (Windows)

```powershell
scoop bucket add gitwhy https://github.com/quanng28/gitwhy-scoop-bucket
scoop install git-why

# Verify
git why --version
```

---

## 3. Authentication Setup

### Interactive (khuyến nghị)

```bash
git why setup
# → Mở browser: https://app.gitwhy.dev/cli-auth
# → Đăng nhập GitHub / email
# → API key tự điền vào terminal
# → Config lưu tại ~/.config/gitwhy/config.json

# Verify
git why status
# ✓ Git repository detected
# ✓ API key: gw_live_xxx*** (free plan)
```

### Manual (headless/CI environments)

```bash
# Lấy API key tại: https://app.gitwhy.dev/dashboard/api-keys

# Option 1: Environment variable
export GITWHY_API_KEY=gw_live_xxxxxxxxxxxxx

# Option 2: Config file
mkdir -p ~/.config/gitwhy
cat > ~/.config/gitwhy/config.json << 'EOF'
{
  "api_key": "gw_live_xxxxxxxxxxxxx",
  "api_url": "https://api.gitwhy.dev"
}
EOF
chmod 600 ~/.config/gitwhy/config.json
```

---

## 4. MCP Server Setup

Sau khi install, thêm vào AI agent config (xem `10-mcp-integration.md` để biết config cho từng agent).

Quick setup cho Claude Code:

```bash
# Thêm vào ~/.claude/settings.json
cat >> ~/.claude/settings.json << 'EOF'
{
  "mcpServers": {
    "gitwhy": {
      "command": "gitwhy-mcp",
      "args": []
    }
  }
}
EOF
```

---

## 5. Post-commit Hook Setup (v0.2)

```bash
# Trong git repo muốn enable auto-save
git why hook install

# Tự tạo .git/hooks/post-commit với auto-save logic
# Verify
cat .git/hooks/post-commit
# → #!/bin/sh
# → gitwhy autosave --from-commit "$1"

# Disable per-repo
git config gitwhy.autosave false

# Disable globally
git config --global gitwhy.autosave false
```

---

## 6. Local Development Setup (contribute)

```bash
# Clone
git clone https://github.com/gitwhy-cli/gitwhy-mcp.git
cd gitwhy-mcp

# Install dependencies
npm install

# Run locally (test MCP server)
node run.js

# Build (khi develop Go binary locally)
go build -o gitwhy ./cmd/main.go
```

### Environment variables cho development

```env
# .env.local (KHÔNG commit file này)

GITWHY_API_KEY=gw_test_xxxxxxxxxxxxx   # Test API key
GITWHY_API_URL=http://localhost:8080   # Local cloud API (nếu develop cloud)
GITWHY_DEBUG=true                      # Verbose logging
GITWHY_AUTOSAVE=false                  # Tắt auto-save trong dev
```

---

## 7. Cloud Backend Local Setup (team development)

```bash
# Prerequisites
go 1.22+
postgresql 15+
redis 7+
docker + docker compose

# Clone cloud backend (internal repo)
git clone https://github.com/gitwhy-cli/gitwhy-cloud.git
cd gitwhy-cloud

# Environment
cp .env.example .env
# → Điền DATABASE_URL, REDIS_URL, GITHUB_APP_ID, GITHUB_APP_PRIVATE_KEY

# Start dependencies
docker compose up -d postgres redis

# Run migrations
go run ./cmd/migrate up

# Start API server
go run ./cmd/api/main.go
# → API available at http://localhost:8080
```

### Environment variables (cloud backend)

```env
# App
ENVIRONMENT=development
PORT=8080

# Database
DATABASE_URL=postgresql://gitwhy:gitwhy@localhost:5432/gitwhy_dev

# Redis
REDIS_URL=redis://localhost:6379/0

# GitHub App
GITHUB_APP_ID=123456
GITHUB_APP_PRIVATE_KEY_PATH=./github-app.pem
GITHUB_WEBHOOK_SECRET=your-webhook-secret

# JWT (cho web dashboard auth)
JWT_SECRET=your-secret-key-here

# Semantic search (v0.2)
OPENAI_API_KEY=sk-xxx   # hoặc Anthropic key cho embeddings
```

---

## 8. Web Dashboard Local Setup

```bash
# Clone (internal repo)
git clone https://github.com/gitwhy-cli/gitwhy-dashboard.git
cd gitwhy-dashboard

# Install
npm install

# Environment
cp .env.example .env.local
# NEXT_PUBLIC_API_URL=http://localhost:8080
# NEXT_PUBLIC_GITHUB_CLIENT_ID=xxx

# Run
npm run dev
# → http://localhost:3000
```

---

*Cập nhật lần cuối: 2026-06-20*
