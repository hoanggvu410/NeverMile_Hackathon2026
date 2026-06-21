package main

import _ "embed"

// agentsMD is the AGENTS.md contract file written into target repos by `mcp install`.
// Keep cmd/git-why/assets/AGENTS.md in sync with the root AGENTS.md.
//
//go:embed assets/AGENTS.md
var agentsMD []byte
