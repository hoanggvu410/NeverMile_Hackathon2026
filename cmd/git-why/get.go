package main

import (
	"fmt"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Print a saved context by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := contextpkg.NewStore()
			if err != nil {
				return err
			}
			ctx, err := store.Get(args[0])
			if err != nil {
				return fmt.Errorf("%s: %w", args[0], err)
			}
			fmt.Print(contextpkg.Render(ctx))
			return nil
		},
	}
}
