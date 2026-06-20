package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "List all contexts newest-first",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := contextpkg.NewStore()
			if err != nil {
				return err
			}
			summaries, err := store.List("", "")
			if err != nil {
				return err
			}
			if len(summaries) == 0 {
				fmt.Println("no contexts saved yet")
				return nil
			}
			printContextList(os.Stdout, summaries)
			return nil
		},
	}
}

// printContextList writes a tabular list of context summaries to w.
func printContextList(w io.Writer, summaries []contextpkg.ContextSummary) {
	fmt.Fprintf(w, "%-16s  %-30s  %-10s  %s\n", "ID", "Domain/Topic", "Date", "Title")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 90))
	for _, s := range summaries {
		domainTopic := s.Domain + "/" + s.Topic
		if len(domainTopic) > 30 {
			domainTopic = domainTopic[:27] + "..."
		}
		title := s.Title
		if title == "" {
			title = s.Prompt
		}
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		date := ""
		if !s.Date.IsZero() {
			date = s.Date.Format("2006-01-02")
		}
		fmt.Fprintf(w, "%-16s  %-30s  %-10s  %s\n", s.ID, domainTopic, date, title)
	}
}
