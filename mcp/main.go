package main

import (
	"fmt"
	"os"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	mcppkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/mcp"
)

func main() {
	store, err := contextpkg.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: failed to initialise store: %v\n", err)
		os.Exit(1)
	}

	srv, err := mcppkg.NewMCPServer(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: failed to create MCP server: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}
