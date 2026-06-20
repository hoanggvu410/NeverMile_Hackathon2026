package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-why",
	Short: "GitWhy — agent memory layer for AI coding agents",
	Long: `GitWhy saves and queries the reasoning behind git commits.
Run as a git sub-command: git why <command>`,
}

func init() {
	rootCmd.AddCommand(newSaveCmd())
	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(newLogCmd())
	rootCmd.AddCommand(newTreeCmd())
	rootCmd.AddCommand(newSearchCmd())
	rootCmd.AddCommand(newHookCmd())
	rootCmd.AddCommand(newSetupCmd())
	rootCmd.AddCommand(newPushCmd())
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
