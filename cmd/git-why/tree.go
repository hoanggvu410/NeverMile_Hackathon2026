package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/spf13/cobra"
)

func newTreeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tree",
		Short: "Show domain → topic → context hierarchy",
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
			printHierarchy(os.Stdout, summaries)
			return nil
		},
	}
}

// printHierarchy groups summaries by domain then topic and prints an indented tree.
func printHierarchy(w io.Writer, summaries []contextpkg.ContextSummary) {
	// Build: domain → topic → []summary
	type domainMap = map[string]map[string][]contextpkg.ContextSummary
	tree := make(domainMap)
	for _, s := range summaries {
		if tree[s.Domain] == nil {
			tree[s.Domain] = make(map[string][]contextpkg.ContextSummary)
		}
		tree[s.Domain][s.Topic] = append(tree[s.Domain][s.Topic], s)
	}

	domains := make([]string, 0, len(tree))
	for d := range tree {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	for _, domain := range domains {
		fmt.Fprintf(w, "%s/\n", domain)
		topics := make([]string, 0, len(tree[domain]))
		for t := range tree[domain] {
			topics = append(topics, t)
		}
		sort.Strings(topics)
		for _, topic := range topics {
			fmt.Fprintf(w, "  %s/\n", topic)
			ctxs := tree[domain][topic]
			// Already newest-first from store.List
			for _, c := range ctxs {
				title := c.Title
				if title == "" {
					title = c.Prompt
				}
				if len(title) > 55 {
					title = title[:52] + "..."
				}
				date := ""
				if !c.Date.IsZero() {
					date = c.Date.Format("2006-01-02")
				}
				fmt.Fprintf(w, "    %-16s  %s  %s\n", c.ID, date, title)
			}
		}
	}
}
