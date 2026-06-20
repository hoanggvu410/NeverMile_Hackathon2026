package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive first-time setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("GitWhy setup")
			fmt.Println("------------")
			fmt.Println("Local storage is ready — no configuration needed for offline use.")
			fmt.Println()
			fmt.Println("To enable cloud sync, team sharing, and PR bot:")
			fmt.Println("  1. Sign up at https://gitwhy.dev")
			fmt.Println("  2. Create an API key in the dashboard")
			fmt.Println("  3. Run: git why push  (coming soon)")
			fmt.Println()
			fmt.Println("Setup complete.")
			return nil
		},
	}
}
