package main

import (
	"fmt"
	"os"
	"strings"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search local contexts by keyword",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			store, err := contextpkg.NewStore()
			if err != nil {
				return err
			}
			results, err := store.Search(query)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				fmt.Printf("no contexts matched %q\n", query)
				return nil
			}
			printContextList(os.Stdout, results)
			return nil
		},
	}
}
