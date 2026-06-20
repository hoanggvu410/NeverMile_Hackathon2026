# GitWhy — NeverMile Hackathon 2026

The context layer for git. Saves, searches, and shares the reasoning
behind AI-generated code changes — tied to commits, delivered to PRs.

See `markdown/` for full product documentation.

## Module
github.com/hoanggvu410/NeverMile_Hackathon2026

## Structure
cmd/git-why/        CLI entry point (Cobra)
internal/context/   local storage + whyspec parsing
internal/mcp/       MCP tool definitions
mcp/                MCP stdio server (spawned by npm wrapper)
markdown/           Product documentation (18 docs)
