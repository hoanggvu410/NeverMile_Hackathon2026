package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	graphpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/graph"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save",
		Short: "Read whyspec markdown from stdin and save locally",
		Long: `Reads a whyspec-formatted markdown document from stdin and saves it
to .git/gitwhy/contexts/<domain>/<topic>/<id>.md.

Example:
  git why save < my-context.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			src, err := readStdin()
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			ctx, err := contextpkg.Parse(src)
			if err != nil {
				return fmt.Errorf("parsing whyspec: %w", err)
			}

			store, err := contextpkg.NewStore()
			if err != nil {
				return err
			}

			id, err := store.Save(*ctx)
			if err != nil {
				return fmt.Errorf("saving context: %w", err)
			}

			saved, err := store.Get(id)
			if err != nil {
				fmt.Printf("saved: %s\n", id)
				return nil
			}
			if graph, err := graphpkg.NewGraph(
				filepath.Join(store.GitWhyDir(), "graph.db"),
				filepath.Join(store.GitWhyDir(), "cache", "semantic.db"),
			); err == nil {
				if err := graph.SaveToGraph(*saved, nil); err != nil {
					fmt.Fprintf(os.Stderr, "warning: graph index failed: %v\n", err)
				}
				_ = graph.Close()
			}

			domain := saved.Domain
			if domain == "" {
				domain = "_"
			}
			topic := saved.Topic
			if topic == "" {
				topic = "_"
			}
			fmt.Printf("saved: %s\n", id)
			fmt.Printf("path:  .git/gitwhy/contexts/%s/%s/%s.md\n", domain, topic, id)
			return nil
		},
	}
}

// readStdin reads all content from os.Stdin; returns error if empty.
func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("stdin is empty — pipe a whyspec markdown document")
	}
	return string(data), nil
}
