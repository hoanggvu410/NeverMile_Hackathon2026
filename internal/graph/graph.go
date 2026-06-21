// Package graph implements GitWhy's claim-level memory graph and tripwire retrieval.
package graph

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hoanggvu410/NeverMile_Hackathon2026/internal/cache"
	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
	_ "modernc.org/sqlite"
)

// SearchResult is one claim-level retrieval result, shaped so existing MCP
// callers can still render session metadata.
type SearchResult struct {
	ID         string    `json:"id"`
	ClaimID    string    `json:"claim_id,omitempty"`
	Domain     string    `json:"domain"`
	Topic      string    `json:"topic"`
	Title      string    `json:"title"`
	Prompt     string    `json:"prompt"`
	Date       time.Time `json:"date"`
	Score      float64   `json:"score"`
	EdgeType   string    `json:"edge_type,omitempty"`
	Depth      int       `json:"depth,omitempty"`
	ClaimText  string    `json:"claim,omitempty"`
	ClaimType  string    `json:"claim_type,omitempty"`
	VectorKind string    `json:"vector_kind,omitempty"`
}

// RuntimeEvent is the normalized event shape used by plan-time tripwires.
type RuntimeEvent struct {
	EventType          string   `json:"event_type"`
	ProjectID          string   `json:"project_id,omitempty"`
	UserRequest        string   `json:"user_request"`
	AgentPlan          string   `json:"agent_plan"`
	FilesLikelyTouched []string `json:"files_likely_touched"`
	Concepts           []string `json:"concepts"`
	ProposedChanges    []string `json:"proposed_changes"`
	NewDependencies    []string `json:"new_dependencies"`
	RiskSurfaces       []string `json:"risk_surfaces"`
}

// TripwireCandidate is a prior claim surfaced for a runtime event.
type TripwireCandidate struct {
	ClaimID        string   `json:"claim_id"`
	SessionID      string   `json:"session_id"`
	Claim          string   `json:"claim"`
	ClaimType      string   `json:"claim_type"`
	Score          float64  `json:"score"`
	VectorKind     string   `json:"vector_kind"`
	EdgeTypes      []string `json:"edge_types,omitempty"`
	MatchedSignals []string `json:"matched_signals"`
}

// RetrievalTelemetry explains how a retrieval or tripwire decision was made.
type RetrievalTelemetry struct {
	RetrievalMode        string `json:"retrieval_mode"`
	VectorsSearched      int    `json:"vectors_searched"`
	CandidatesConsidered int    `json:"candidates_considered"`
	MarkdownFilesOpened  int    `json:"markdown_files_opened"`
	CacheHit             bool   `json:"cache_hit"`
	EventType            string `json:"event_type,omitempty"`
	UnavailableReason    string `json:"unavailable_reason,omitempty"`
}

// TripwireResult is returned by CheckTripwire and the MCP tripwire tool.
type TripwireResult struct {
	Available  bool                `json:"available"`
	Interrupt  bool                `json:"interrupt"`
	Message    string              `json:"message,omitempty"`
	Candidates []TripwireCandidate `json:"candidates"`
	Telemetry  RetrievalTelemetry  `json:"telemetry"`
}

// Graph wraps graph.db and the semantic cache.
type Graph struct {
	db        *sql.DB
	cache     *cache.Cache
	embedding embeddingSpec
}

type claimRow struct {
	id              string
	sessionID       string
	projectID       string
	domain          string
	topic           string
	title           string
	prompt          string
	date            time.Time
	text            string
	claimType       string
	status          string
	importance      int
	sourceSpan      string
	scopeJSON       string
	aliasesJSON     string
	retrievalJSON   string
	blastRadiusJSON string
	interruptJSON   string
}

type claimCandidate struct {
	claim      claimRow
	vectorKind string
	vectorText string
	rawScore   float64
	score      float64
	edgeType   string
	depth      int
}

// EdgeHint specifies a pre-classified edge to write when saving to the graph.
// ContextID may be either a claim ID or a session/context ID; the save path maps
// session IDs to their first claim for backward compatibility with existing MCP calls.
type EdgeHint struct {
	ContextID string
	EdgeType  string
}

var validEdgeTypes = map[string]bool{
	"CONSTRAINS":        true,
	"IMPLEMENTS":        true,
	"CAUSED_BY":         true,
	"SUPERSEDES":        true,
	"CONFLICTS_WITH":    true,
	"RELATED_CANDIDATE": true,
	"CONSTRAINED_BY":    true, // legacy Section 2 spelling
	"INVALIDATES":       true, // legacy, mapped to CONFLICTS_WITH
	"CONTRADICTS":       true, // legacy, mapped to CONFLICTS_WITH
	"DEPENDS_ON":        true, // legacy, mapped to CAUSED_BY
	"NONE":              true,
}

// NewGraph opens (or creates) graph.db and semantic.db.
func NewGraph(graphDBPath, cacheDBPath string) (*Graph, error) {
	if err := os.MkdirAll(filepath.Dir(graphDBPath), 0755); err != nil {
		return nil, fmt.Errorf("create graph dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cacheDBPath), 0755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	db, err := sql.Open("sqlite", graphDBPath)
	if err != nil {
		return nil, fmt.Errorf("open graph.db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := ensureGraphMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	embedding, err := resolveGraphEmbeddingSpec(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	c, err := cache.NewCache(cacheDBPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("open cache: %w", err)
	}
	return &Graph{db: db, cache: c, embedding: embedding}, nil
}

// Close releases database connections.
func (g *Graph) Close() error {
	if g == nil {
		return nil
	}
	if g.cache != nil {
		_ = g.cache.Close()
	}
	return g.db.Close()
}

// SaveToGraph stores the full session markdown and up to seven durable claims.
// Each claim receives exactly three vector rows: claim, retrieval, interrupt.
func (g *Graph) SaveToGraph(ctx contextpkg.Context, hints []EdgeHint) error {
	if g == nil {
		return nil
	}
	if ctx.ID == "" {
		return fmt.Errorf("context ID is required")
	}
	if ctx.Date.IsZero() {
		ctx.Date = time.Now().UTC()
	}
	fullMarkdown := contextpkg.Render(&ctx)
	claims := extractClaims(&ctx)
	cacheDirty := false
	defer func() {
		if cacheDirty {
			_ = g.invalidateRetrievalCache()
		}
	}()

	tx, err := g.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if err := saveSessionTx(tx, ctx, fullMarkdown); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM claims WHERE session_id = ?`, ctx.ID); err != nil {
		return err
	}
	for i := range claims {
		claims[i].SessionID = ctx.ID
		claims[i].ID = makeClaimID(ctx.ID, i, claims[i].Text)
		if err := saveClaimTx(tx, claims[i]); err != nil {
			return err
		}
		for _, kind := range []string{"claim", "retrieval", "interrupt"} {
			text := vectorTextForClaim(claims[i], kind)
			embedding, err := g.embedText(text)
			if err != nil {
				return fmt.Errorf("embed %s vector for %s: %w", kind, claims[i].ID, err)
			}
			if err := saveClaimVectorTx(tx, claims[i].ID, kind, text, embedding, g.embedding); err != nil {
				return err
			}
		}
	}
	if err := saveSameSessionEdgesTx(tx, claims); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	cacheDirty = true

	if len(claims) > 0 {
		if err := g.saveCrossSessionCandidates(claims); err != nil {
			return err
		}
		if err := g.saveEdgeHints(ctx.ID, claims, hints); err != nil {
			return err
		}
	}
	return nil
}

// Search performs graph retrieval over claim_vectors and traverses typed edges.
func (g *Graph) Search(query, domain string, limit int) ([]SearchResult, bool, error) {
	if g == nil {
		return nil, false, fmt.Errorf("graph unavailable")
	}
	if limit <= 0 {
		limit = 5
	}
	embedding, err := g.embedText(query)
	if err != nil {
		return nil, false, fmt.Errorf("embed query: %w", err)
	}

	namespace := g.cacheNamespace("search", domain)
	if cachedJSON, hit, err := g.cache.GetInNamespace(namespace, embedding, 0.995); err == nil && hit {
		var results []SearchResult
		if json.Unmarshal(cachedJSON, &results) == nil {
			return results, true, nil
		}
	}

	candidates, _, err := g.retrieveClaimCandidates(embedding, query, "", domain, limit, searchKindWeights(), 0.08, true)
	if err != nil {
		return nil, false, err
	}
	results := make([]SearchResult, 0, len(candidates))
	seenClaims := make(map[string]bool)
	for _, candidate := range candidates {
		results = append(results, searchResultFromCandidate(candidate))
		seenClaims[candidate.claim.id] = true
	}
	for _, candidate := range candidates {
		chain, err := g.traverseFromClaim(candidate.claim.id, seenClaims)
		if err == nil {
			results = append(results, chain...)
		}
	}
	if len(results) > limit {
		results = results[:limit]
	}
	if b, err := json.Marshal(results); err == nil {
		_ = g.cache.SetInNamespace(namespace, embedding, b)
	}
	return results, false, nil
}

// CheckTripwire retrieves claim-level constraints for an agent_plan_created event.
// It never falls back to markdown; telemetry always reports graph_only mode.
func (g *Graph) CheckTripwire(event RuntimeEvent) (*TripwireResult, error) {
	if event.EventType == "" {
		event.EventType = "agent_plan_created"
	}
	result := &TripwireResult{
		Available:  true,
		Candidates: []TripwireCandidate{},
		Telemetry: RetrievalTelemetry{
			RetrievalMode:       "graph_only",
			MarkdownFilesOpened: 0,
			EventType:           event.EventType,
		},
	}
	if g == nil {
		result.Available = false
		result.Telemetry.UnavailableReason = "graph is not initialized"
		return result, nil
	}

	eventText := buildEventText(event)
	embedding, err := g.embedText(eventText)
	if err != nil {
		result.Available = false
		result.Telemetry.UnavailableReason = err.Error()
		return result, nil
	}

	namespace := g.cacheNamespace("tripwire", event.ProjectID)
	if cachedJSON, hit, err := g.cache.GetInNamespace(namespace, embedding, 0.995); err == nil && hit {
		if json.Unmarshal(cachedJSON, result) == nil {
			result.Telemetry.CacheHit = true
			result.Telemetry.RetrievalMode = "graph_only"
			result.Telemetry.MarkdownFilesOpened = 0
			_ = g.recordInterruptEvent(event, result)
			return result, nil
		}
	}

	candidates, vectorsSearched, err := g.retrieveClaimCandidates(embedding, eventText, event.ProjectID, "", 12, tripwireKindWeights(), 0.08, true)
	if err != nil {
		return nil, err
	}
	result.Telemetry.VectorsSearched = vectorsSearched
	result.Telemetry.CandidatesConsidered = len(candidates)
	result.Candidates = g.tripwireCandidates(event, eventText, candidates)
	if len(result.Candidates) > 0 {
		result.Interrupt = true
		result.Message = buildTripwireMessage(result.Candidates[0])
	}

	if b, err := json.Marshal(result); err == nil {
		_ = g.cache.SetInNamespace(namespace, embedding, b)
	}
	_ = g.recordInterruptEvent(event, result)
	return result, nil
}

func saveSessionTx(tx *sql.Tx, ctx contextpkg.Context, fullMarkdown string) error {
	projectID := projectIDForContext(ctx)
	title := ctx.Title
	if title == "" {
		title = ctx.Prompt
		if len(title) > 72 {
			title = title[:72] + "..."
		}
	}
	_, err := tx.Exec(`
		INSERT INTO sessions (id, project_id, domain, topic, title, prompt, created_at, full_markdown)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			project_id = excluded.project_id,
			domain = excluded.domain,
			topic = excluded.topic,
			title = excluded.title,
			prompt = excluded.prompt,
			created_at = excluded.created_at,
			full_markdown = excluded.full_markdown`,
		ctx.ID, projectID, ctx.Domain, ctx.Topic, title, ctx.Prompt,
		ctx.Date.UTC().Format(time.RFC3339), fullMarkdown,
	)
	return err
}

func saveClaimTx(tx *sql.Tx, claim claimDraft) error {
	scopeJSON, _ := json.Marshal(claim.Scope)
	aliasesJSON, _ := json.Marshal(claim.Aliases)
	retrievalJSON, _ := json.Marshal(claim.RetrievalTriggers)
	blastJSON, _ := json.Marshal(claim.BlastRadius)
	interruptJSON, _ := json.Marshal(claim.InterruptConditions)

	_, err := tx.Exec(`
		INSERT INTO claims (
			id, session_id, text, type, status, importance, source_span,
			scope_json, aliases_json, retrieval_triggers_json, blast_radius_json,
			interrupt_conditions_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			session_id = excluded.session_id,
			text = excluded.text,
			type = excluded.type,
			status = excluded.status,
			importance = excluded.importance,
			source_span = excluded.source_span,
			scope_json = excluded.scope_json,
			aliases_json = excluded.aliases_json,
			retrieval_triggers_json = excluded.retrieval_triggers_json,
			blast_radius_json = excluded.blast_radius_json,
			interrupt_conditions_json = excluded.interrupt_conditions_json`,
		claim.ID, claim.SessionID, claim.Text, claim.Type, claim.Status, claim.Importance,
		claim.SourceSpan, string(scopeJSON), string(aliasesJSON), string(retrievalJSON),
		string(blastJSON), string(interruptJSON),
	)
	return err
}

func saveClaimVectorTx(tx *sql.Tx, claimID, kind, text string, embedding []float32, spec embeddingSpec) error {
	id := claimID + ":" + kind
	_, err := tx.Exec(`
		INSERT INTO claim_vectors (id, claim_id, kind, provider, dims, text, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(claim_id, kind) DO UPDATE SET
			provider = excluded.provider,
			dims = excluded.dims,
			text = excluded.text,
			embedding = excluded.embedding`,
		id, claimID, kind, spec.Provider, spec.Dims, text, cache.EncodeEmbedding(embedding),
	)
	return err
}

func saveSameSessionEdgesTx(tx *sql.Tx, claims []claimDraft) error {
	for i := range claims {
		for j := range claims {
			if i == j {
				continue
			}
			from := claims[i]
			to := claims[j]
			switch {
			case from.Type == "implementation" && isDecisionLike(to.Type):
				if err := saveEdgeTx(tx, from.ID, to.ID, "IMPLEMENTS", 0.82, "same session implementation references a saved decision", "same_session", "active"); err != nil {
					return err
				}
			case isConstraintLike(from.Type) && to.Type == "implementation":
				if err := saveEdgeTx(tx, from.ID, to.ID, "CONSTRAINS", 0.78, "same session constraint applies to implementation work", "same_session", "active"); err != nil {
					return err
				}
			case isDecisionLike(from.Type) && to.Type == "rationale":
				if err := saveEdgeTx(tx, from.ID, to.ID, "CAUSED_BY", 0.74, "decision and rationale were saved together", "same_session", "active"); err != nil {
					return err
				}
			case isDecisionLike(from.Type) && to.Type == "rejected_alternative":
				if err := saveEdgeTx(tx, from.ID, to.ID, "CONFLICTS_WITH", 0.74, "decision explicitly rejected an alternative in the same session", "same_session", "active"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func saveEdgeTx(tx *sql.Tx, fromID, toID, edgeType string, confidence float64, evidence, source, status string) error {
	if fromID == "" || toID == "" || fromID == toID {
		return nil
	}
	edgeType = normalizeEdgeType(edgeType)
	if edgeType == "" || edgeType == "NONE" {
		return nil
	}
	id := makeEdgeID(fromID, toID, edgeType)
	_, err := tx.Exec(`
		INSERT INTO edges (id, from_claim_id, to_claim_id, type, confidence, evidence, source, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(from_claim_id, to_claim_id, type) DO UPDATE SET
			confidence = excluded.confidence,
			evidence = excluded.evidence,
			source = excluded.source,
			status = excluded.status`,
		id, fromID, toID, edgeType, confidence, evidence, source, status,
	)
	return err
}

func (g *Graph) saveCrossSessionCandidates(claims []claimDraft) error {
	for _, claim := range claims {
		embedding, err := g.embedText(vectorTextForClaim(claim, "retrieval"))
		if err != nil {
			return err
		}
		queryText := vectorTextForClaim(claim, "retrieval")
		candidates, _, err := g.retrieveClaimCandidates(embedding, queryText, "", "", 5, map[string]float64{
			"claim": 0.90, "retrieval": 1.10, "interrupt": 0.80,
		}, 0.12, true)
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			if candidate.claim.sessionID == claim.SessionID || candidate.claim.id == claim.ID {
				continue
			}
			if candidate.score < 0.18 {
				continue
			}
			if err := g.saveEdge(claim.ID, candidate.claim.id, "RELATED_CANDIDATE", candidate.score, "semantic similarity between claim retrieval triggers", "cross_session", "candidate"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Graph) saveEdgeHints(sessionID string, claims []claimDraft, hints []EdgeHint) error {
	if len(hints) == 0 || len(claims) == 0 {
		return nil
	}
	fromID := claims[0].ID
	for _, hint := range hints {
		edgeType := normalizeEdgeType(hint.EdgeType)
		if edgeType == "" || edgeType == "NONE" || hint.ContextID == "" {
			continue
		}
		toID, err := g.resolveClaimID(hint.ContextID)
		if err != nil || toID == "" || toID == sessionID {
			continue
		}
		if err := g.saveEdge(fromID, toID, edgeType, 0.90, "explicit edge_hints argument", "edge_hint", "active"); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) saveEdge(fromID, toID, edgeType string, confidence float64, evidence, source, status string) error {
	tx, err := g.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if err := saveEdgeTx(tx, fromID, toID, edgeType, confidence, evidence, source, status); err != nil {
		return err
	}
	return tx.Commit()
}

func (g *Graph) resolveClaimID(id string) (string, error) {
	var claimID string
	err := g.db.QueryRow(`SELECT id FROM claims WHERE id = ?`, id).Scan(&claimID)
	if err == nil {
		return claimID, nil
	}
	err = g.db.QueryRow(`SELECT id FROM claims WHERE session_id = ? ORDER BY importance DESC, id LIMIT 1`, id).Scan(&claimID)
	if err != nil {
		return "", err
	}
	return claimID, nil
}

func (g *Graph) retrieveClaimCandidates(embedding []float32, queryText, projectID, domain string, limit int, weights map[string]float64, minScore float64, activeOnly bool) ([]claimCandidate, int, error) {
	if limit <= 0 {
		limit = 10
	}
	var sqlText strings.Builder
	sqlText.WriteString(`
		SELECT
			cv.claim_id, cv.kind, cv.text, cv.embedding,
			c.session_id, c.text, c.type, c.status, c.importance, c.source_span,
			c.scope_json, c.aliases_json, c.retrieval_triggers_json, c.blast_radius_json,
			c.interrupt_conditions_json,
			s.project_id, s.domain, s.topic, s.title, s.prompt, s.created_at
		FROM claim_vectors cv
		JOIN claims c ON c.id = cv.claim_id
		JOIN sessions s ON s.id = c.session_id
		WHERE cv.provider = ? AND cv.dims = ?`)
	args := []any{g.embedding.Provider, g.embedding.Dims}
	if activeOnly {
		sqlText.WriteString(` AND c.status = 'active'`)
	}
	if projectID != "" {
		sqlText.WriteString(` AND s.project_id = ?`)
		args = append(args, projectID)
	}
	if domain != "" {
		sqlText.WriteString(` AND s.domain LIKE ?`)
		args = append(args, domain+"%")
	}

	rows, err := g.db.Query(sqlText.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	byClaim := make(map[string]claimCandidate)
	vectorsSearched := 0
	for rows.Next() {
		vectorsSearched++
		var candidate claimCandidate
		var dateStr string
		var blob []byte
		if err := rows.Scan(
			&candidate.claim.id, &candidate.vectorKind, &candidate.vectorText, &blob,
			&candidate.claim.sessionID, &candidate.claim.text, &candidate.claim.claimType,
			&candidate.claim.status, &candidate.claim.importance, &candidate.claim.sourceSpan,
			&candidate.claim.scopeJSON, &candidate.claim.aliasesJSON, &candidate.claim.retrievalJSON,
			&candidate.claim.blastRadiusJSON, &candidate.claim.interruptJSON,
			&candidate.claim.projectID, &candidate.claim.domain, &candidate.claim.topic,
			&candidate.claim.title, &candidate.claim.prompt, &dateStr,
		); err != nil {
			continue
		}
		stored, err := cache.DecodeEmbedding(blob)
		if err != nil {
			continue
		}
		candidate.rawScore = cache.CosineSim(embedding, stored)
		candidate.score = candidate.rawScore*weights[candidate.vectorKind] + candidateLexicalBoost(queryText, candidate)
		if candidate.score < minScore {
			continue
		}
		candidate.claim.date, _ = time.Parse(time.RFC3339, dateStr)
		if current, ok := byClaim[candidate.claim.id]; !ok || candidate.score > current.score {
			byClaim[candidate.claim.id] = candidate
		}
	}
	if err := rows.Err(); err != nil {
		return nil, vectorsSearched, err
	}

	candidates := make([]claimCandidate, 0, len(byClaim))
	for _, candidate := range byClaim {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].claim.importance > candidates[j].claim.importance
		}
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, vectorsSearched, nil
}

func (g *Graph) traverseFromClaim(rootID string, seen map[string]bool) ([]SearchResult, error) {
	rows, err := g.db.Query(`
		WITH RECURSIVE chain AS (
			SELECT to_claim_id, type, 1 AS depth
			FROM edges
			WHERE from_claim_id = ? AND status IN ('active', 'candidate')
			UNION ALL
			SELECT e.to_claim_id, e.type, c.depth + 1
			FROM edges e
			JOIN chain c ON e.from_claim_id = c.to_claim_id
			WHERE c.depth < 2 AND e.status IN ('active', 'candidate')
		)
		SELECT
			c.id, c.session_id, c.text, c.type, c.status, c.importance, c.source_span,
			c.scope_json, c.aliases_json, c.retrieval_triggers_json, c.blast_radius_json,
			c.interrupt_conditions_json,
			s.project_id, s.domain, s.topic, s.title, s.prompt, s.created_at,
			chain.type, MIN(chain.depth) AS depth
		FROM chain
		JOIN claims c ON c.id = chain.to_claim_id
		JOIN sessions s ON s.id = c.session_id
		GROUP BY c.id
		ORDER BY depth, c.importance DESC`, rootID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var row claimRow
		var edgeType, dateStr string
		var depth int
		if err := rows.Scan(
			&row.id, &row.sessionID, &row.text, &row.claimType, &row.status,
			&row.importance, &row.sourceSpan, &row.scopeJSON, &row.aliasesJSON,
			&row.retrievalJSON, &row.blastRadiusJSON, &row.interruptJSON,
			&row.projectID, &row.domain, &row.topic, &row.title, &row.prompt,
			&dateStr, &edgeType, &depth,
		); err != nil {
			continue
		}
		if seen[row.id] {
			continue
		}
		seen[row.id] = true
		row.date, _ = time.Parse(time.RFC3339, dateStr)
		results = append(results, SearchResult{
			ID:        row.sessionID,
			ClaimID:   row.id,
			Domain:    row.domain,
			Topic:     row.topic,
			Title:     row.title,
			Prompt:    row.prompt,
			Date:      row.date,
			EdgeType:  edgeType,
			Depth:     depth,
			ClaimText: row.text,
			ClaimType: row.claimType,
		})
	}
	return results, rows.Err()
}

func (g *Graph) tripwireCandidates(event RuntimeEvent, eventText string, candidates []claimCandidate) []TripwireCandidate {
	out := make([]TripwireCandidate, 0, 3)
	for _, candidate := range candidates {
		signals := []string{"vector_match", "active_claim"}
		scopeMatch := claimMatchesScope(event, eventText, candidate.claim)
		interruptMatch := claimMatchesInterrupt(eventText, candidate.claim)
		edgeTypes := g.edgeTypesForClaim(candidate.claim.id)
		edgeMatch := hasTripwireEdge(edgeTypes)

		if scopeMatch {
			signals = append(signals, "scope_blast_radius_match")
		}
		if interruptMatch {
			signals = append(signals, "interrupt_condition_match")
		}
		if edgeMatch {
			signals = append(signals, "edge_relation_match")
		}
		if candidate.vectorKind == "interrupt" && candidate.rawScore >= 0.10 {
			signals = append(signals, "interrupt_vector_match")
			interruptMatch = true
		}

		if !(scopeMatch && (interruptMatch || edgeMatch)) {
			continue
		}
		out = append(out, TripwireCandidate{
			ClaimID:        candidate.claim.id,
			SessionID:      candidate.claim.sessionID,
			Claim:          candidate.claim.text,
			ClaimType:      candidate.claim.claimType,
			Score:          candidate.score,
			VectorKind:     candidate.vectorKind,
			EdgeTypes:      edgeTypes,
			MatchedSignals: signals,
		})
		if len(out) == 3 {
			break
		}
	}
	return out
}

func (g *Graph) edgeTypesForClaim(claimID string) []string {
	rows, err := g.db.Query(`
		SELECT DISTINCT type
		FROM edges
		WHERE status IN ('active', 'candidate')
		  AND (from_claim_id = ? OR to_claim_id = ?)
		ORDER BY type`, claimID, claimID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var types []string
	for rows.Next() {
		var typ string
		if rows.Scan(&typ) == nil {
			types = append(types, typ)
		}
	}
	return types
}

func (g *Graph) recordInterruptEvent(event RuntimeEvent, result *TripwireResult) error {
	eventJSON, _ := json.Marshal(event)
	candidatesJSON, _ := json.Marshal(result.Candidates)
	id := "evt_" + shortHash(fmt.Sprintf("%s:%d", string(eventJSON), time.Now().UnixNano()))
	_, err := g.db.Exec(`
		INSERT INTO interrupt_events (id, event_type, project_id, event_json, candidates_json, user_action, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, event.EventType, event.ProjectID, string(eventJSON), string(candidatesJSON), "",
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func searchResultFromCandidate(candidate claimCandidate) SearchResult {
	return SearchResult{
		ID:         candidate.claim.sessionID,
		ClaimID:    candidate.claim.id,
		Domain:     candidate.claim.domain,
		Topic:      candidate.claim.topic,
		Title:      candidate.claim.title,
		Prompt:     candidate.claim.prompt,
		Date:       candidate.claim.date,
		Score:      candidate.score,
		EdgeType:   candidate.edgeType,
		Depth:      candidate.depth,
		ClaimText:  candidate.claim.text,
		ClaimType:  candidate.claim.claimType,
		VectorKind: candidate.vectorKind,
	}
}

func searchKindWeights() map[string]float64 {
	return map[string]float64{"claim": 1.25, "retrieval": 1.00, "interrupt": 0.75}
}

func tripwireKindWeights() map[string]float64 {
	return map[string]float64{"claim": 0.85, "retrieval": 1.05, "interrupt": 1.30}
}

func normalizeEdgeType(edgeType string) string {
	edgeType = strings.ToUpper(strings.TrimSpace(edgeType))
	if !validEdgeTypes[edgeType] {
		return ""
	}
	switch edgeType {
	case "CONSTRAINED_BY":
		return "CONSTRAINS"
	case "INVALIDATES", "CONTRADICTS":
		return "CONFLICTS_WITH"
	case "DEPENDS_ON":
		return "CAUSED_BY"
	default:
		return edgeType
	}
}

func projectIDForContext(ctx contextpkg.Context) string {
	if strings.TrimSpace(ctx.Repository) != "" {
		return strings.TrimSpace(ctx.Repository)
	}
	return "local"
}

func (g *Graph) embedText(text string) ([]float32, error) {
	return embedTextWithSpec(text, g.embedding)
}

func (g *Graph) cacheNamespace(kind, suffix string) string {
	return fmt.Sprintf("%s:%s:%d:%s", kind, g.embedding.Provider, g.embedding.Dims, suffix)
}

func (g *Graph) invalidateRetrievalCache() error {
	if g.cache == nil {
		return nil
	}
	return g.cache.ClearNamespacePrefixes("search:", "tripwire:")
}
