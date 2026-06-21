// Command serve runs a local, no-auth HTTP API that wraps the GitWhy store and
// claim graph so the web dashboard (web/) can read real data from the local
// repo. It is meant for local development / demos only — there is no auth.
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	"github.com/hoanggvu410/NeverMile_Hackathon2026/internal/graph"
	_ "modernc.org/sqlite"
)

type server struct {
	store    *contextpkg.Store
	graph    *graph.Graph
	graphDB  *sql.DB // read-only handle for node/edge queries
	repoName string
	branch   string
}

func main() {
	addr := flag.String("addr", "localhost:7420", "listen address")
	repo := flag.String("repo", "", "path to the git repo to serve (defaults to cwd)")
	flag.Parse()

	if *repo != "" {
		if err := os.Chdir(*repo); err != nil {
			log.Fatalf("chdir %s: %v", *repo, err)
		}
	}

	store, err := contextpkg.NewStore()
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	graphDBPath := filepath.Join(store.GitWhyDir(), "graph.db")
	cacheDBPath := filepath.Join(store.GitWhyDir(), "cache", "semantic.db")

	g, err := graph.NewGraph(graphDBPath, cacheDBPath)
	if err != nil {
		log.Printf("warning: graph unavailable (%v); search will fall back to store", err)
		g = nil
	}

	var graphDB *sql.DB
	if db, err := sql.Open("sqlite", graphDBPath+"?mode=ro"); err == nil {
		db.SetMaxOpenConns(1)
		graphDB = db
	} else {
		log.Printf("warning: could not open graph.db for node/edge queries: %v", err)
	}

	repoName := cleanRepoName(mustRepoName())
	branch, _ := contextpkg.CurrentBranch()

	s := &server{store: store, graph: g, graphDB: graphDB, repoName: repoName, branch: branch}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/contexts", s.handleContexts)
	mux.HandleFunc("/api/contexts/", s.handleContextByID)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/graph/nodes", s.handleGraphNodes)
	mux.HandleFunc("/api/graph/edges", s.handleGraphEdges)
	mux.HandleFunc("/api/domains", s.handleDomains)

	log.Printf("GitWhy local API listening on http://%s (repo: %s, branch: %s)", *addr, repoName, branch)
	if err := http.ListenAndServe(*addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

// cors allows any localhost origin (dashboard dev server) with simple GETs.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ---- response shapes ----

type statusResponse struct {
	Repository     string   `json:"repository"`
	Branch         string   `json:"branch"`
	GitRoot        string   `json:"git_root"`
	ContextCount   int      `json:"context_count"`
	PendingCommits []string `json:"pending_commits"`
	GraphReady     bool     `json:"graph_ready"`
}

type contextSummaryDTO struct {
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Domain string    `json:"domain"`
	Topic  string    `json:"topic"`
	Date   time.Time `json:"date"`
	Prompt string    `json:"prompt"`
}

type fileEntryDTO struct {
	File        string `json:"file"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type contextDTO struct {
	ID                    string         `json:"id"`
	Title                 string         `json:"title"`
	SavedBy               string         `json:"saved_by"`
	Agent                 string         `json:"agent"`
	Repository            string         `json:"repository"`
	Branch                string         `json:"branch"`
	Date                  time.Time      `json:"date"`
	Domain                string         `json:"domain"`
	Topic                 string         `json:"topic"`
	Prompt                string         `json:"prompt"`
	WhatWasDone           string         `json:"what_was_done"`
	Reasoning             string         `json:"reasoning"`
	KeyDecisions          string         `json:"key_decisions"`
	RejectedAlternatives  string         `json:"rejected_alternatives"`
	RisksAndOpenQuestions string         `json:"risks_and_open_questions"`
	Verification          string         `json:"verification"`
	Files                 []fileEntryDTO `json:"files"`
	Commits               []string       `json:"commits"`
}

type graphNodeDTO struct {
	ID         string  `json:"id"`
	SessionID  string  `json:"session_id"`
	Domain     string  `json:"domain"`
	Topic      string  `json:"topic"`
	Title      string  `json:"title"`
	Claim      string  `json:"claim"`
	ClaimType  string  `json:"claim_type"`
	Importance int     `json:"importance"`
	EdgeCount  int     `json:"edge_count"`
}

type graphEdgeDTO struct {
	ID         string  `json:"id"`
	Source     string  `json:"source"`
	Target     string  `json:"target"`
	Type       string  `json:"type"`
	Confidence float64 `json:"confidence"`
	Status     string  `json:"status"`
}

// ---- handlers ----

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	info := s.store.Status()
	resp := statusResponse{
		Repository:     s.repoName,
		Branch:         s.branch,
		GitRoot:        info.GitRoot,
		ContextCount:   info.LocalContextCount,
		PendingCommits: info.PendingCommits,
		GraphReady:     s.graphDB != nil,
	}
	if resp.PendingCommits == nil {
		resp.PendingCommits = []string{}
	}
	writeJSON(w, resp)
}

func (s *server) handleContexts(w http.ResponseWriter, r *http.Request) {
	domainFilter := r.URL.Query().Get("domain")
	topicFilter := r.URL.Query().Get("topic")
	// List with no filter, then enrich + filter on the real header domain/topic.
	// (The store derives domain from path depth, which is lossy for nested
	// contexts, so we resolve the authoritative values per context.)
	list, err := s.store.List("", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]contextSummaryDTO, 0, len(list))
	for _, c := range list {
		domain, topic := s.resolveDomainTopic(c)
		if domainFilter != "" && !strings.HasPrefix(domain, domainFilter) {
			continue
		}
		if topicFilter != "" && !strings.HasPrefix(topic, topicFilter) {
			continue
		}
		out = append(out, contextSummaryDTO{
			ID: c.ID, Title: c.Title, Domain: domain,
			Topic: topic, Date: c.Date, Prompt: c.Prompt,
		})
	}
	writeJSON(w, out)
}

// resolveDomainTopic returns the authoritative domain/topic for a summary,
// falling back to a full parse when the path-derived values are missing.
func (s *server) resolveDomainTopic(c contextpkg.ContextSummary) (string, string) {
	if c.Domain != "" && c.Domain != "_" {
		return c.Domain, c.Topic
	}
	if full, err := s.store.Get(c.ID); err == nil {
		d, t := full.Domain, full.Topic
		if d == "" {
			d = "_"
		}
		return d, t
	}
	return c.Domain, c.Topic
}

func (s *server) handleContextByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/contexts/")
	id = strings.Trim(id, "/")
	if id == "" {
		http.Error(w, "missing context id", http.StatusBadRequest)
		return
	}
	ctx, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "context not found", http.StatusNotFound)
		return
	}
	files := make([]fileEntryDTO, 0, len(ctx.Files))
	for _, f := range ctx.Files {
		files = append(files, fileEntryDTO{File: f.File, Status: f.Status, Description: f.Description})
	}
	commits := ctx.Commits
	if commits == nil {
		commits = []string{}
	}
	writeJSON(w, contextDTO{
		ID: ctx.ID, Title: ctx.Title, SavedBy: ctx.SavedBy, Agent: ctx.Agent,
		Repository: ctx.Repository, Branch: ctx.Branch, Date: ctx.Date,
		Domain: ctx.Domain, Topic: ctx.Topic, Prompt: ctx.Prompt,
		WhatWasDone: ctx.WhatWasDone, Reasoning: ctx.Reasoning,
		KeyDecisions: ctx.KeyDecisions, RejectedAlternatives: ctx.RejectedAlternatives,
		RisksAndOpenQuestions: ctx.RisksAndOpenQuestions, Verification: ctx.Verification,
		Files: files, Commits: commits,
	})
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if strings.TrimSpace(q) == "" {
		writeJSON(w, []graph.SearchResult{})
		return
	}

	if s.graph != nil {
		results, _, err := s.graph.Search(q, "", limit)
		if err == nil {
			if results == nil {
				results = []graph.SearchResult{}
			}
			writeJSON(w, results)
			return
		}
		log.Printf("graph search failed (%v); falling back to store", err)
	}

	// Fallback: substring search over the store, shaped like SearchResult.
	summaries, err := s.store.Search(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]graph.SearchResult, 0, len(summaries))
	for i, sum := range summaries {
		if i >= limit {
			break
		}
		out = append(out, graph.SearchResult{
			ID: sum.ID, Domain: sum.Domain, Topic: sum.Topic,
			Title: sum.Title, Prompt: sum.Prompt, Date: sum.Date,
			ClaimText: sum.Prompt,
		})
	}
	writeJSON(w, out)
}

func (s *server) handleGraphNodes(w http.ResponseWriter, r *http.Request) {
	if s.graphDB == nil {
		writeJSON(w, []graphNodeDTO{})
		return
	}
	rows, err := s.graphDB.Query(`
		SELECT c.id, c.session_id, s.domain, s.topic, s.title, c.text, c.type, c.importance,
		       (SELECT COUNT(*) FROM edges e
		         WHERE (e.from_claim_id = c.id OR e.to_claim_id = c.id)
		           AND e.status IN ('active','candidate')) AS edge_count
		FROM claims c
		JOIN sessions s ON s.id = c.session_id
		WHERE c.status = 'active'
		ORDER BY c.importance DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := make([]graphNodeDTO, 0)
	for rows.Next() {
		var n graphNodeDTO
		if err := rows.Scan(&n.ID, &n.SessionID, &n.Domain, &n.Topic, &n.Title,
			&n.Claim, &n.ClaimType, &n.Importance, &n.EdgeCount); err != nil {
			continue
		}
		out = append(out, n)
	}
	writeJSON(w, out)
}

func (s *server) handleGraphEdges(w http.ResponseWriter, r *http.Request) {
	if s.graphDB == nil {
		writeJSON(w, []graphEdgeDTO{})
		return
	}
	rows, err := s.graphDB.Query(`
		SELECT id, from_claim_id, to_claim_id, type, confidence, status
		FROM edges
		WHERE status IN ('active','candidate')`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	out := make([]graphEdgeDTO, 0)
	for rows.Next() {
		var e graphEdgeDTO
		if err := rows.Scan(&e.ID, &e.Source, &e.Target, &e.Type, &e.Confidence, &e.Status); err != nil {
			continue
		}
		out = append(out, e)
	}
	writeJSON(w, out)
}

func (s *server) handleDomains(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.List("", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	seen := map[string]bool{}
	var domains []string
	for _, c := range list {
		d, _ := s.resolveDomainTopic(c)
		if d == "" || d == "_" {
			continue
		}
		if !seen[d] {
			seen[d] = true
			domains = append(domains, d)
		}
	}
	sort.Strings(domains)
	if domains == nil {
		domains = []string{}
	}
	writeJSON(w, domains)
}

func mustRepoName() string {
	n, _ := contextpkg.RepoName()
	return n
}

// cleanRepoName reduces a remote URL / path to "owner/repo".
func cleanRepoName(raw string) string {
	raw = strings.TrimSuffix(strings.TrimSpace(raw), ".git")
	raw = strings.Trim(raw, "/")
	parts := strings.Split(raw, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	if raw == "" {
		return "local/repo"
	}
	return raw
}
