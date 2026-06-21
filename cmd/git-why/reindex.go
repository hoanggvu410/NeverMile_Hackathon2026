package main

import (
	"fmt"
	"path/filepath"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	graphpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/graph"
	"github.com/spf13/cobra"
)

func newReindexCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reindex",
		Short: "Index all saved markdown contexts into the graph (run after migration or manual edits)",
		Long: `Walks every context file in .git/gitwhy/contexts/ and feeds each one through
SaveToGraph so the vector store and claim graph reflect the full markdown store.

Run this when:
  - contexts were saved by a different tool or edited on disk directly
  - the MCP server was pointing at the wrong repo during earlier saves
  - graph.db was deleted or recreated`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := contextpkg.NewStore()
			if err != nil {
				return err
			}

			graph, err := graphpkg.NewGraph(
				filepath.Join(store.GitWhyDir(), "graph.db"),
				filepath.Join(store.GitWhyDir(), "cache", "semantic.db"),
			)
			if err != nil {
				return fmt.Errorf("opening graph: %w", err)
			}
			defer graph.Close()

			summaries, err := store.List("", "")
			if err != nil {
				return fmt.Errorf("listing contexts: %w", err)
			}

			if len(summaries) == 0 {
				fmt.Println("no contexts found in markdown store")
				return nil
			}

			fmt.Printf("reindexing %d context(s)...\n", len(summaries))
			ok, skipped, failed := 0, 0, 0

			for _, s := range summaries {
				ctx, err := store.Get(s.ID)
				if err != nil {
					fmt.Printf("  skip %s: cannot read (%v)\n", s.ID, err)
					skipped++
					continue
				}
				if err := graph.SaveToGraph(*ctx, nil); err != nil {
					fmt.Printf("  fail %s: %v\n", s.ID, err)
					failed++
					continue
				}
				fmt.Printf("  ok   %s  %s\n", s.ID, s.Title)
				ok++
			}

			fmt.Printf("\ndone: %d indexed, %d skipped, %d failed\n", ok, skipped, failed)
			return nil
		},
	}
}
