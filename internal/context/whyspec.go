package context

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const idCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// GenerateID returns "ctx_" + 8 random chars from [0-9a-zA-Z].
func GenerateID() string {
	b := make([]byte, 8)
	charLen := big.NewInt(int64(len(idCharset)))
	for i := range b {
		n, _ := rand.Int(rand.Reader, charLen)
		b[i] = idCharset[n.Int64()]
	}
	return "ctx_" + string(b)
}

// Render serialises ctx into the exact whyspec markdown format.
func Render(ctx *Context) string {
	var sb strings.Builder

	title := ctx.Title
	if title == "" {
		title = ctx.Prompt
		if len(title) > 72 {
			title = title[:72] + "..."
		}
	}

	sb.WriteString(fmt.Sprintf("# Context: %s\n\n", title))
	sb.WriteString(fmt.Sprintf("**Context ID:** %s\n", ctx.ID))
	sb.WriteString(fmt.Sprintf("**Saved by:** %s\n", ctx.SavedBy))
	sb.WriteString(fmt.Sprintf("**Agent:** %s\n", ctx.Agent))
	sb.WriteString(fmt.Sprintf("**Repository:** %s\n", ctx.Repository))
	sb.WriteString(fmt.Sprintf("**Branch:** %s\n", ctx.Branch))

	date := ctx.Date
	if date.IsZero() {
		date = time.Now().UTC()
	}
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", date.Format(time.RFC3339)))
	if ctx.Domain != "" {
		sb.WriteString(fmt.Sprintf("**Domain:** %s\n", ctx.Domain))
	}
	if ctx.Topic != "" {
		sb.WriteString(fmt.Sprintf("**Topic:** %s\n", ctx.Topic))
	}

	sb.WriteString("\n## Prompt\n\n")
	for _, line := range strings.Split(ctx.Prompt, "\n") {
		sb.WriteString("> " + line + "\n")
	}

	sb.WriteString("\n## What Was Done\n\n")
	sb.WriteString(ctx.WhatWasDone)
	if ctx.WhatWasDone != "" && !strings.HasSuffix(ctx.WhatWasDone, "\n") {
		sb.WriteByte('\n')
	}

	sb.WriteString("\n## Reasoning\n\n")
	sb.WriteString(ctx.Reasoning)
	if ctx.Reasoning != "" && !strings.HasSuffix(ctx.Reasoning, "\n") {
		sb.WriteByte('\n')
	}

	sb.WriteString("\n## Key Decisions\n\n")
	sb.WriteString(ctx.KeyDecisions)
	if ctx.KeyDecisions != "" && !strings.HasSuffix(ctx.KeyDecisions, "\n") {
		sb.WriteByte('\n')
	}

	sb.WriteString("\n## Rejected Alternatives\n\n")
	sb.WriteString(ctx.RejectedAlternatives)
	if ctx.RejectedAlternatives != "" && !strings.HasSuffix(ctx.RejectedAlternatives, "\n") {
		sb.WriteByte('\n')
	}

	sb.WriteString("\n## Risks & Open Questions\n\n")
	sb.WriteString(ctx.RisksAndOpenQuestions)
	if ctx.RisksAndOpenQuestions != "" && !strings.HasSuffix(ctx.RisksAndOpenQuestions, "\n") {
		sb.WriteByte('\n')
	}

	sb.WriteString("\n## Files\n\n")
	sb.WriteString("| File | Status | Description |\n")
	sb.WriteString("|------|--------|-------------|\n")
	for _, f := range ctx.Files {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", f.File, f.Status, f.Description))
	}

	sb.WriteString("\n## Commits\n\n")
	for _, c := range ctx.Commits {
		sb.WriteString("- " + c + "\n")
	}

	sb.WriteString("\n## Verification\n\n")
	sb.WriteString(ctx.Verification)
	if ctx.Verification != "" && !strings.HasSuffix(ctx.Verification, "\n") {
		sb.WriteByte('\n')
	}

	return sb.String()
}

// Parse deserialises a whyspec markdown string into a Context.
// Returns error if mandatory header fields (Context ID) are missing.
func Parse(src string) (*Context, error) {
	ctx := &Context{}
	lines := strings.Split(src, "\n")

	var currentSection string
	var sectionBuf strings.Builder
	inFilesTable := false

	flushSection := func() {
		text := strings.TrimSpace(sectionBuf.String())
		switch currentSection {
		case "prompt":
			// Strip leading "> " from blockquote lines
			var promptLines []string
			for _, l := range strings.Split(text, "\n") {
				l = strings.TrimPrefix(l, "> ")
				promptLines = append(promptLines, l)
			}
			ctx.Prompt = strings.TrimSpace(strings.Join(promptLines, "\n"))
		case "what was done":
			ctx.WhatWasDone = text
		case "reasoning":
			ctx.Reasoning = text
		case "key decisions":
			ctx.KeyDecisions = text
		case "rejected alternatives":
			ctx.RejectedAlternatives = text
		case "risks & open questions":
			ctx.RisksAndOpenQuestions = text
		case "verification":
			ctx.Verification = text
		case "commits":
			for _, l := range strings.Split(text, "\n") {
				l = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "-"))
				l = strings.TrimSpace(l)
				if l != "" {
					ctx.Commits = append(ctx.Commits, l)
				}
			}
		}
		sectionBuf.Reset()
	}

	for _, line := range lines {
		// H1: title (two variants)
		if strings.HasPrefix(line, "# Context:") {
			ctx.Title = strings.TrimSpace(strings.TrimPrefix(line, "# Context:"))
			continue
		}
		if strings.HasPrefix(line, "# ") && ctx.Title == "" {
			ctx.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}

		// Table-format metadata: | **Key** | value |  (written by gitwhy v0.8.x)
		if strings.HasPrefix(line, "| **") && strings.Contains(line, "** |") {
			if v, ok := parseTableMetaLine(line); ok {
				switch v[0] {
				case "Context ID":
					ctx.ID = v[1]
				case "Agent":
					ctx.Agent = v[1]
				case "Repository":
					ctx.Repository = v[1]
				case "Branch":
					ctx.Branch = v[1]
				case "Date":
					if t, err := time.Parse("2006-01-02 15:04:05 UTC", v[1]); err == nil {
						ctx.Date = t
					} else if t, err := time.Parse(time.RFC3339, v[1]); err == nil {
						ctx.Date = t
					}
				case "Domain/Topic":
					parts := strings.SplitN(v[1], " / ", 2)
					if len(parts) == 2 {
						ctx.Domain = strings.TrimSpace(parts[0])
						ctx.Topic = strings.TrimSpace(parts[1])
					}
				case "Tags":
					// ignore tags field
				}
				continue
			}
		}

		// Header metadata lines (bold-text format written by Render())
		if strings.HasPrefix(line, "**Context ID:**") {
			ctx.ID = strings.TrimSpace(strings.TrimPrefix(line, "**Context ID:**"))
			continue
		}
		if strings.HasPrefix(line, "**Saved by:**") {
			ctx.SavedBy = strings.TrimSpace(strings.TrimPrefix(line, "**Saved by:**"))
			continue
		}
		if strings.HasPrefix(line, "**Agent:**") {
			ctx.Agent = strings.TrimSpace(strings.TrimPrefix(line, "**Agent:**"))
			continue
		}
		if strings.HasPrefix(line, "**Repository:**") {
			ctx.Repository = strings.TrimSpace(strings.TrimPrefix(line, "**Repository:**"))
			continue
		}
		if strings.HasPrefix(line, "**Branch:**") {
			ctx.Branch = strings.TrimSpace(strings.TrimPrefix(line, "**Branch:**"))
			continue
		}
		if strings.HasPrefix(line, "**Date:**") {
			raw := strings.TrimSpace(strings.TrimPrefix(line, "**Date:**"))
			t, err := time.Parse(time.RFC3339, raw)
			if err == nil {
				ctx.Date = t
			}
			continue
		}
		if strings.HasPrefix(line, "**Domain:**") {
			ctx.Domain = strings.TrimSpace(strings.TrimPrefix(line, "**Domain:**"))
			continue
		}
		if strings.HasPrefix(line, "**Topic:**") {
			ctx.Topic = strings.TrimSpace(strings.TrimPrefix(line, "**Topic:**"))
			continue
		}

		// H2 section headers
		if strings.HasPrefix(line, "## ") {
			flushSection()
			inFilesTable = false
			heading := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "## ")))
			currentSection = heading
			if heading == "files" {
				inFilesTable = true
			}
			continue
		}

		// Files table rows
		if inFilesTable {
			// Skip header/separator rows
			if strings.HasPrefix(line, "|---") || strings.HasPrefix(line, "| File") {
				continue
			}
			if strings.HasPrefix(line, "|") {
				parts := strings.Split(line, "|")
				// parts[0] is empty, parts[1]=file, parts[2]=status, parts[3]=desc
				if len(parts) >= 4 {
					fe := FileEntry{
						File:        strings.TrimSpace(parts[1]),
						Status:      strings.TrimSpace(parts[2]),
						Description: strings.TrimSpace(parts[3]),
					}
					if fe.File != "" {
						ctx.Files = append(ctx.Files, fe)
					}
				}
			}
			continue
		}

		// Accumulate section body
		if currentSection != "" {
			sectionBuf.WriteString(line + "\n")
		}
	}
	flushSection()

	if ctx.ID == "" {
		return nil, errors.New("whyspec missing required field: Context ID")
	}

	return ctx, nil
}

// parseTableMetaLine parses a Markdown table row like:
// | **Context ID** | `ctx_abc123` |
// Returns ([key, value], true) on success.
func parseTableMetaLine(line string) ([2]string, bool) {
	parts := strings.Split(line, "|")
	if len(parts) < 4 {
		return [2]string{}, false
	}
	key := strings.TrimSpace(parts[1])
	key = strings.TrimPrefix(key, "**")
	key = strings.TrimSuffix(key, "**")
	key = strings.TrimSpace(key)
	val := strings.TrimSpace(parts[2])
	val = strings.Trim(val, "`")
	val = strings.TrimSpace(val)
	if key == "" || val == "" {
		return [2]string{}, false
	}
	return [2]string{key, val}, true
}
