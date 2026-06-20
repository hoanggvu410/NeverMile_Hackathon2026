package context

import "time"

// FileEntry is a row in the ## Files table of a whyspec document.
type FileEntry struct {
	File        string
	Status      string
	Description string
}

// Context is the in-memory representation of a saved whyspec document.
type Context struct {
	ID                    string
	Title                 string
	SavedBy               string
	Agent                 string
	Repository            string
	Branch                string
	Date                  time.Time
	Prompt                string
	WhatWasDone           string
	Reasoning             string
	KeyDecisions          string
	RejectedAlternatives  string
	RisksAndOpenQuestions string
	Files                 []FileEntry
	Commits               []string
	Verification          string
	Domain                string
	Topic                 string
}

// ContextSummary is a lightweight view used by List, Search, and tree/log output.
type ContextSummary struct {
	ID     string
	Title  string
	Domain string
	Topic  string
	Date   time.Time
	Prompt string
	Path   string
}

// StatusInfo is the payload returned by Status().
type StatusInfo struct {
	IsGitRepo         bool
	GitRoot           string
	GitWhyDir         string
	LocalContextCount int
	PendingCommits    []string
}
