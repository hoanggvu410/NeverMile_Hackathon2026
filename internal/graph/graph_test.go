package graph

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
)

func newTestGraph(t *testing.T) *Graph {
	t.Helper()
	t.Setenv("GITWHY_EMBEDDING_PROVIDER", "local")
	dir := t.TempDir()
	g, err := NewGraph(filepath.Join(dir, "graph.db"), filepath.Join(dir, "cache", "semantic.db"))
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	t.Cleanup(func() { _ = g.Close() })
	return g
}

func spacingContext(id string) contextpkg.Context {
	return contextpkg.Context{
		ID:          id,
		Title:       "Planner spacing decision",
		Repository:  "acme/planner",
		Domain:      "frontend",
		Topic:       "planner-ui",
		Date:        time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC),
		Prompt:      "Define planner UI spacing rules",
		WhatWasDone: "Implemented planner UI controls using the shared spacing rules.",
		Reasoning:   "The planner needs predictable visual rhythm so dense controls stay scannable.",
		KeyDecisions: strings.Join([]string{
			"Use 4-8-16 spacing for planner UI rhythm.",
			"Do not introduce ad hoc gaps in planner controls.",
		}, "\n"),
		RejectedAlternatives:  "Freeform one-off spacing: inconsistent and hard to maintain.",
		RisksAndOpenQuestions: "Changing the spacing scale later could make planner layouts inconsistent.",
		Files: []contextpkg.FileEntry{
			{File: "frontend/planner.css", Status: "modified", Description: "Planner styles"},
		},
		Verification: "Run UI tests.",
	}
}

func TestSaveToGraphCreatesSessionClaimsVectorsAndEdges(t *testing.T) {
	g := newTestGraph(t)

	if err := g.SaveToGraph(spacingContext("ctx_spacing01"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	if got := countRows(t, g.db, `SELECT COUNT(*) FROM sessions`); got != 1 {
		t.Fatalf("sessions count = %d, want 1", got)
	}
	claims := countRows(t, g.db, `SELECT COUNT(*) FROM claims`)
	if claims == 0 || claims > maxClaimsPerSession {
		t.Fatalf("claims count = %d, want 1..%d", claims, maxClaimsPerSession)
	}
	vectors := countRows(t, g.db, `SELECT COUNT(*) FROM claim_vectors`)
	if vectors != claims*3 {
		t.Fatalf("claim_vectors count = %d, want %d", vectors, claims*3)
	}
	compatibleVectors := countRows(t, g.db, `SELECT COUNT(*) FROM claim_vectors WHERE provider = 'local-hash-v1' AND dims = 384`)
	if compatibleVectors != vectors {
		t.Fatalf("compatible claim_vectors count = %d, want %d", compatibleVectors, vectors)
	}
	edges := countRows(t, g.db, `SELECT COUNT(*) FROM edges WHERE source = 'same_session' AND type IN ('CAUSED_BY','CONSTRAINS','IMPLEMENTS')`)
	if edges == 0 {
		t.Fatalf("expected explicit same-session edges")
	}
}

func TestSearchRetrievesSpacingClaimFromClaimVectors(t *testing.T) {
	g := newTestGraph(t)
	if err := g.SaveToGraph(spacingContext("ctx_spacing02"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	results, hit, err := g.Search("why 4-8-16 spacing?", "", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if hit {
		t.Fatalf("first search should not be a cache hit")
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one result")
	}
	if !strings.Contains(strings.ToLower(results[0].ClaimText), "4-8-16 spacing") {
		t.Fatalf("top result claim = %q, want spacing claim", results[0].ClaimText)
	}
}

func TestSaveInvalidatesCachedSearchMiss(t *testing.T) {
	g := newTestGraph(t)
	query := "why 4-8-16 spacing?"

	results, hit, err := g.Search(query, "", 5)
	if err != nil {
		t.Fatalf("initial Search: %v", err)
	}
	if hit || len(results) != 0 {
		t.Fatalf("initial search hit=%v len=%d, want cached miss setup", hit, len(results))
	}
	if err := g.SaveToGraph(spacingContext("ctx_spacing_miss01"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	results, hit, err = g.Search(query, "", 5)
	if err != nil {
		t.Fatalf("second Search: %v", err)
	}
	if hit {
		t.Fatalf("search returned cache hit after graph write; stale miss was not invalidated")
	}
	if len(results) == 0 {
		t.Fatalf("expected saved spacing claim after cache invalidation")
	}
}

func TestSaveInvalidatesCachedTripwireMiss(t *testing.T) {
	g := newTestGraph(t)
	event := RuntimeEvent{
		EventType:          "agent_plan_created",
		ProjectID:          "acme/planner",
		UserRequest:        "Update the planner UI spacing",
		AgentPlan:          "Adjust planner UI gaps and spacing scale before editing frontend CSS.",
		FilesLikelyTouched: []string{"frontend/planner.css"},
		Concepts:           []string{"planner UI", "spacing"},
		ProposedChanges:    []string{"change spacing scale", "adjust gaps"},
		RiskSurfaces:       []string{"layout rhythm"},
	}

	initial, err := g.CheckTripwire(event)
	if err != nil {
		t.Fatalf("initial CheckTripwire: %v", err)
	}
	if initial.Interrupt {
		t.Fatalf("initial tripwire interrupted before any claims were saved")
	}
	if err := g.SaveToGraph(spacingContext("ctx_spacing_miss02"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	second, err := g.CheckTripwire(event)
	if err != nil {
		t.Fatalf("second CheckTripwire: %v", err)
	}
	if second.Telemetry.CacheHit {
		t.Fatalf("tripwire returned cache hit after graph write; stale miss was not invalidated")
	}
	if !second.Interrupt {
		t.Fatalf("expected tripwire interrupt after saving relevant claim, got %+v", second)
	}
}

func TestTripwireSurfacesPlannerSpacingBeforeEdits(t *testing.T) {
	g := newTestGraph(t)
	if err := g.SaveToGraph(spacingContext("ctx_spacing03"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	event := RuntimeEvent{
		EventType:          "agent_plan_created",
		ProjectID:          "acme/planner",
		UserRequest:        "Update the planner UI spacing",
		AgentPlan:          "Adjust planner UI gaps and spacing scale before editing frontend CSS.",
		FilesLikelyTouched: []string{"frontend/planner.css"},
		Concepts:           []string{"planner UI", "spacing"},
		ProposedChanges:    []string{"change spacing scale", "adjust gaps"},
		RiskSurfaces:       []string{"layout rhythm"},
	}
	result, err := g.CheckTripwire(event)
	if err != nil {
		t.Fatalf("CheckTripwire: %v", err)
	}
	if !result.Available {
		t.Fatalf("tripwire unavailable: %s", result.Telemetry.UnavailableReason)
	}
	if !result.Interrupt {
		t.Fatalf("expected interrupt, got candidates=%+v telemetry=%+v", result.Candidates, result.Telemetry)
	}
	if result.Telemetry.RetrievalMode != "graph_only" {
		t.Fatalf("retrieval mode = %q, want graph_only", result.Telemetry.RetrievalMode)
	}
	if result.Telemetry.MarkdownFilesOpened != 0 {
		t.Fatalf("markdown files opened = %d, want 0", result.Telemetry.MarkdownFilesOpened)
	}
	if result.Telemetry.VectorsSearched == 0 {
		t.Fatalf("expected vectors searched telemetry")
	}
}

func TestTripwireCacheHit(t *testing.T) {
	g := newTestGraph(t)
	if err := g.SaveToGraph(spacingContext("ctx_spacing04"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}
	event := RuntimeEvent{
		EventType:          "agent_plan_created",
		ProjectID:          "acme/planner",
		UserRequest:        "Change planner spacing",
		AgentPlan:          "Edit planner UI spacing in frontend/planner.css.",
		FilesLikelyTouched: []string{"frontend/planner.css"},
		Concepts:           []string{"planner UI", "spacing"},
		ProposedChanges:    []string{"change spacing"},
	}
	if _, err := g.CheckTripwire(event); err != nil {
		t.Fatalf("first CheckTripwire: %v", err)
	}
	second, err := g.CheckTripwire(event)
	if err != nil {
		t.Fatalf("second CheckTripwire: %v", err)
	}
	if !second.Telemetry.CacheHit {
		t.Fatalf("expected second tripwire call to report cache hit, telemetry=%+v", second.Telemetry)
	}
}

func TestSlashSeparatedSpacingScaleRetrievesHyphenQuery(t *testing.T) {
	g := newTestGraph(t)
	ctx := spacingContext("ctx_spacing_slash01")
	ctx.KeyDecisions = "Use the 4/8/16/24/32/48/64 spacing scale for planner UI rhythm."
	if err := g.SaveToGraph(ctx, nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}

	claims := extractClaims(&ctx)
	if len(claims) == 0 {
		t.Fatalf("expected extracted claims")
	}
	aliases := strings.Join(claims[0].Aliases, "\n")
	for _, want := range []string{"4-8-16", "4/8/16", "4-8-16-24-32-48-64", "4/8/16/24/32/48/64"} {
		if !strings.Contains(aliases, want) {
			t.Fatalf("aliases missing %q in:\n%s", want, aliases)
		}
	}

	results, _, err := g.Search("why 4-8-16 spacing?", "", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 || !strings.Contains(results[0].ClaimText, "4/8/16/24/32/48/64") {
		t.Fatalf("top result = %+v, want slash-separated spacing claim", results)
	}
}

func TestGraphKeepsStoredEmbeddingProviderAcrossEnvChanges(t *testing.T) {
	dir := t.TempDir()
	graphDB := filepath.Join(dir, "graph.db")
	cacheDB := filepath.Join(dir, "cache", "semantic.db")

	t.Setenv("GITWHY_EMBEDDING_PROVIDER", "local")
	g1, err := NewGraph(graphDB, cacheDB)
	if err != nil {
		t.Fatalf("NewGraph local: %v", err)
	}
	if g1.embedding.Provider != localEmbeddingProvider {
		t.Fatalf("initial provider = %q, want %q", g1.embedding.Provider, localEmbeddingProvider)
	}
	if err := g1.SaveToGraph(spacingContext("ctx_provider01"), nil); err != nil {
		t.Fatalf("SaveToGraph: %v", err)
	}
	if err := g1.Close(); err != nil {
		t.Fatalf("close g1: %v", err)
	}

	t.Setenv("GITWHY_EMBEDDING_PROVIDER", "")
	t.Setenv("OPENAI_API_KEY", "not-a-real-key")
	g2, err := NewGraph(graphDB, cacheDB)
	if err != nil {
		t.Fatalf("NewGraph after OPENAI_API_KEY: %v", err)
	}
	if g2.embedding.Provider != localEmbeddingProvider {
		t.Fatalf("reopened provider = %q, want stable %q", g2.embedding.Provider, localEmbeddingProvider)
	}
	results, _, err := g2.Search("why 4-8-16 spacing?", "", 5)
	if err != nil {
		t.Fatalf("Search with stable local provider: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected local-indexed claim to remain retrievable after OPENAI_API_KEY appears")
	}
	if err := g2.Close(); err != nil {
		t.Fatalf("close g2: %v", err)
	}

	t.Setenv("GITWHY_EMBEDDING_PROVIDER", "openai")
	if _, err := NewGraph(graphDB, cacheDB); err == nil || !strings.Contains(err.Error(), "embedding provider mismatch") {
		t.Fatalf("NewGraph explicit provider mismatch err = %v, want mismatch error", err)
	}
}

func TestCrossSessionLinksStartAsRelatedCandidates(t *testing.T) {
	g := newTestGraph(t)
	if err := g.SaveToGraph(spacingContext("ctx_spacing05"), nil); err != nil {
		t.Fatalf("SaveToGraph first: %v", err)
	}
	ctx2 := spacingContext("ctx_spacing06")
	ctx2.KeyDecisions = "Planner layout must keep the 4-8-16 spacing scale."
	if err := g.SaveToGraph(ctx2, nil); err != nil {
		t.Fatalf("SaveToGraph second: %v", err)
	}
	candidates := countRows(t, g.db, `SELECT COUNT(*) FROM edges WHERE type = 'RELATED_CANDIDATE' AND status = 'candidate' AND source = 'cross_session'`)
	if candidates == 0 {
		t.Fatalf("expected cross-session RELATED_CANDIDATE edge")
	}
}

func countRows(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("countRows %q: %v", query, err)
	}
	return count
}
