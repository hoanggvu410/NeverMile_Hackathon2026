package main

import (
	"fmt"
	"os"
	"path/filepath"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	graphpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/graph"
	mcppkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/mcp"
)

func main() {
	store, err := contextpkg.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: failed to initialise store: %v\n", err)
		os.Exit(1)
	}

	gitwhyDir := store.GitWhyDir()
	graph, err := graphpkg.NewGraph(
		filepath.Join(gitwhyDir, "graph.db"),
		filepath.Join(gitwhyDir, "cache", "semantic.db"),
	)
	if err != nil {
		// Graph is non-critical; log and continue without it.
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: graph init failed (continuing without graph): %v\n", err)
		graph = nil
	}
	if graph != nil {
		defer graph.Close()
	}

	srv, err := mcppkg.NewMCPServer(store, graph)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: failed to create MCP server: %v\n", err)
		os.Exit(1)
	}

	if err := srv.Serve(); err != nil {
		fmt.Fprintf(os.Stderr, "gitwhy-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}
