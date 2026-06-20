package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Sync contexts to the cloud (coming soon)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Cloud sync coming soon. Stay tuned at https://gitwhy.dev")
			return nil
		},
	}
}
