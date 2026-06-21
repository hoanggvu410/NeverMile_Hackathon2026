package graph

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	contextpkg "github.com/hoanggvu410/NeverMile_Hackathon2026/internal/context"
)

const maxClaimsPerSession = 7

var digitSequencePattern = regexp.MustCompile(`\d+`)

type claimDraft struct {
	ID                  string
	SessionID           string
	Text                string
	Type                string
	Status              string
	Importance          int
	SourceSpan          string
	Scope               map[string][]string
	Aliases             []string
	RetrievalTriggers   []string
	BlastRadius         map[string][]string
	InterruptConditions []string
}

func extractClaims(ctx *contextpkg.Context) []claimDraft {
	var drafts []claimDraft
	addClaims := func(section, source string, importance int) {
		for i, item := range splitDurableStatements(source) {
			typ := classifyClaim(section, item)
			if typ == "" {
				continue
			}
			draft := buildClaimDraft(ctx, item, typ, importance, fmt.Sprintf("%s:%d", section, i+1))
			drafts = append(drafts, draft)
		}
	}

	addClaims("Key Decisions", ctx.KeyDecisions, 5)
	addClaims("Rejected Alternatives", ctx.RejectedAlternatives, 4)
	addClaims("Risks & Open Questions", ctx.RisksAndOpenQuestions, 4)
	addClaims("What Was Done", ctx.WhatWasDone, 3)
	addClaims("Reasoning", ctx.Reasoning, 3)

	drafts = dedupeClaims(drafts)
	sort.SliceStable(drafts, func(i, j int) bool {
		if drafts[i].Importance == drafts[j].Importance {
			return sourcePriority(drafts[i].SourceSpan) < sourcePriority(drafts[j].SourceSpan)
		}
		return drafts[i].Importance > drafts[j].Importance
	})
	if len(drafts) > maxClaimsPerSession {
		drafts = drafts[:maxClaimsPerSession]
	}
	return drafts
}

func splitDurableStatements(source string) []string {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil
	}
	lines := strings.Split(source, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 220 {
			for _, part := range strings.Split(line, ". ") {
				part = strings.TrimSpace(part)
				if part != "" {
					out = append(out, trimTrailingSentence(part))
				}
			}
			continue
		}
		out = append(out, trimTrailingSentence(line))
	}
	return out
}

func trimTrailingSentence(text string) string {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, " \t\r\n-")
	return text
}

func classifyClaim(section, text string) string {
	lower := strings.ToLower(text)
	switch section {
	case "Rejected Alternatives":
		if isDurableText(lower) {
			return "rejected_alternative"
		}
	case "Risks & Open Questions":
		if isDurableText(lower) {
			return "risk"
		}
	case "What Was Done":
		if isDurableText(lower) {
			return "implementation"
		}
	case "Reasoning":
		if isDurableText(lower) {
			return "rationale"
		}
	}
	if containsAny(lower, "do not", "don't", "never", "must", "only", "constraint", "required", "avoid", "preserve", "keep") {
		if containsAny(lower, "spacing", "typography", "layout", "ui", "design") {
			return "design_constraint"
		}
		return "constraint"
	}
	if containsAny(lower, "spacing", "typography", "layout", "color", "palette", "ui", "dashboard", "planner") {
		return "design_decision"
	}
	if containsAny(lower, "fastapi", "express", "framework", "backend", "api", "server", "database", "sqlite", "postgres", "supabase", "cache", "mcp", "cli") {
		return "architecture_decision"
	}
	if containsAny(lower, "use ", "chosen", "choose", "decided", "decision", "prefer", "standardize", "replace", "migrate") {
		return "decision"
	}
	if isDurableText(lower) && section == "Key Decisions" {
		return "decision"
	}
	return ""
}

func isDurableText(lower string) bool {
	if len(strings.Fields(lower)) < 3 {
		return false
	}
	return containsAny(lower,
		"use", "chosen", "choose", "decided", "decision", "prefer", "must", "never", "avoid",
		"risk", "because", "reason", "constraint", "architecture", "backend", "frontend", "api",
		"spacing", "auth", "security", "database", "cache", "mcp", "cli", "graph", "tripwire",
	)
}

func buildClaimDraft(ctx *contextpkg.Context, text, typ string, importance int, sourceSpan string) claimDraft {
	scope := inferScope(ctx, text, typ)
	aliases := inferAliases(ctx, text, scope)
	retrieval := inferRetrievalTriggers(text, typ, scope, aliases)
	blast := inferBlastRadius(scope, text, typ)
	interrupts := inferInterruptConditions(text, typ, scope, aliases)

	return claimDraft{
		Text:                text,
		Type:                typ,
		Status:              "active",
		Importance:          importance,
		SourceSpan:          sourceSpan,
		Scope:               scope,
		Aliases:             uniqueStrings(aliases),
		RetrievalTriggers:   uniqueStrings(retrieval),
		BlastRadius:         blast,
		InterruptConditions: uniqueStrings(interrupts),
	}
}

func inferScope(ctx *contextpkg.Context, text, typ string) map[string][]string {
	lower := strings.ToLower(text + " " + ctx.Domain + " " + ctx.Topic)
	components := []string{}
	concepts := []string{}
	dependencies := []string{}
	files := []string{}

	if ctx.Domain != "" {
		components = append(components, ctx.Domain)
	}
	if ctx.Topic != "" {
		concepts = append(concepts, ctx.Topic)
	}
	for _, f := range ctx.Files {
		if f.File != "" {
			files = append(files, filepathSlash(f.File))
		}
	}

	if containsAny(lower, "frontend", "ui", "design", "spacing", "layout", "typography", "planner", "dashboard", "css") {
		components = append(components, "frontend", "ui", "design")
		concepts = append(concepts, "spacing", "layout", "planner ui", "design system")
		files = append(files, "frontend/**", "ui/**", "styles/**", "css/**")
	}
	if containsAny(lower, "backend", "api", "server", "route", "fastapi", "express", "websocket", "mcp") {
		components = append(components, "backend", "api", "server")
		concepts = append(concepts, "backend framework", "api routes", "server architecture", "websocket backend")
		files = append(files, "backend/**", "server/**", "internal/api/**", "app/**", "mcp/**")
	}
	if containsAny(lower, "database", "sqlite", "postgres", "supabase", "schema", "migration") {
		components = append(components, "database", "storage")
		concepts = append(concepts, "data model", "schema", "migration")
		files = append(files, "db/**", "migrations/**", "internal/storage/**")
	}
	if containsAny(lower, "cache", "semantic cache") {
		components = append(components, "cache")
		concepts = append(concepts, "semantic cache", "query cache")
		files = append(files, "internal/cache/**")
	}
	if containsAny(lower, "graph", "claim", "tripwire", "edge", "embedding") {
		components = append(components, "graph", "memory", "tripwire")
		concepts = append(concepts, "claim graph", "claim vectors", "typed edges", "interrupt conditions")
		files = append(files, "internal/graph/**")
	}
	if containsAny(lower, "cli", "cobra", "git why") {
		components = append(components, "cli")
		concepts = append(concepts, "command line")
		files = append(files, "cmd/**")
	}

	for _, dep := range []string{"fastapi", "express", "flask", "django", "sqlite", "postgres", "supabase", "openai", "mcp", "cobra"} {
		if strings.Contains(lower, dep) {
			dependencies = append(dependencies, dep)
		}
	}
	if strings.Contains(typ, "design") {
		components = append(components, "design")
	}

	return map[string][]string{
		"components":   uniqueStrings(components),
		"concepts":     uniqueStrings(concepts),
		"files":        uniqueStrings(files),
		"dependencies": uniqueStrings(dependencies),
	}
}

func inferAliases(ctx *contextpkg.Context, text string, scope map[string][]string) []string {
	// Aliases are alternate noun-phrase names for the thing being decided about.
	// Do NOT include: claim text (already stored in text), scope components/concepts
	// (already in scope_json), or generic labels. Only short variant names that
	// someone might search for to refer to the same concept.
	var aliases []string
	aliases = append(aliases, scope["dependencies"]...)

	lower := strings.ToLower(text)
	if containsAny(lower, "spacing", "gap", "padding", "margin") {
		aliases = append(aliases, spacingScaleAliases(text)...)
	}
	if strings.Contains(lower, "4-8-16") || strings.Contains(lower, "4 8 16") || strings.Contains(lower, "4/8/16") {
		aliases = append(aliases, "4-8-16", "4/8/16", "4 8 16", "spacing scale", "planner UI spacing", "gap padding margin")
	}
	if strings.Contains(lower, "fastapi") {
		aliases = append(aliases, "fast api", "python api server", "backend framework", "api layer")
	}
	if strings.Contains(lower, "mcp") {
		aliases = append(aliases, "model context protocol", "mcp server", "mcp tools")
	}
	if ctx.Topic != "" {
		aliases = append(aliases, ctx.Topic)
	}
	return aliases
}

func spacingScaleAliases(text string) []string {
	numbers := digitSequencePattern.FindAllString(text, -1)
	if len(numbers) < 3 {
		return []string{"spacing scale", "numeric spacing scale"}
	}

	aliases := []string{
		"spacing scale",
		"numeric spacing scale",
		"design spacing scale",
		"planner UI spacing",
		"gap padding margin",
	}
	addScaleVariants := func(prefix string, values []string) {
		hyphen := strings.Join(values, "-")
		slash := strings.Join(values, "/")
		spaced := strings.Join(values, " ")
		aliases = append(aliases,
			hyphen,
			slash,
			spaced,
			prefix+" "+hyphen,
			prefix+" "+slash,
			prefix+" "+spaced,
		)
	}

	addScaleVariants("spacing scale", numbers)
	addScaleVariants("full spacing scale", numbers)
	addScaleVariants("base spacing scale", numbers[:3])
	if len(numbers) >= 4 {
		addScaleVariants("compact spacing scale", numbers[:4])
	}
	return uniqueStrings(aliases)
}

func inferRetrievalTriggers(text, typ string, scope map[string][]string, aliases []string) []string {
	// Retrieval triggers are future situation descriptions that should surface this claim.
	// They describe scenarios, not the claim itself. Do NOT include: claim text or aliases
	// (already stored separately). Focus on "what would someone be doing when they need this?"
	var triggers []string
	for _, component := range scope["components"] {
		triggers = append(triggers, "work touching "+component, component+" changes")
	}
	for _, concept := range scope["concepts"] {
		triggers = append(triggers, "changes to "+concept, concept+" work")
	}

	// Use text only — not typ — to avoid type label words like "design" in
	// "design_constraint" triggering unrelated domain heuristics.
	lower := strings.ToLower(text)
	if containsAny(lower, "spacing", "layout", "ui", "design") {
		triggers = append(triggers,
			"planner UI spacing",
			"layout gap padding margin changes",
			"design system spacing scale",
			"inconsistent spacing in UI",
		)
	}
	if containsAny(lower, "fastapi", "backend", "api", "server") {
		triggers = append(triggers,
			"backend route changes",
			"third-party API integration",
			"websocket realtime backend",
			"API framework changes",
			"Express Flask Django Node backend",
		)
	}
	if containsAny(lower, "claim", "tripwire", "embedding", "edge") {
		triggers = append(triggers,
			"agent plan created",
			"memory tripwire",
			"claim retrieval",
			"typed edge traversal",
		)
	}
	return triggers
}

func inferBlastRadius(scope map[string][]string, text, typ string) map[string][]string {
	blast := map[string][]string{
		"components":   append([]string{}, scope["components"]...),
		"concepts":     append([]string{}, scope["concepts"]...),
		"files":        append([]string{}, scope["files"]...),
		"dependencies": append([]string{}, scope["dependencies"]...),
	}
	lower := strings.ToLower(text)
	if containsAny(lower, "spacing", "layout", "ui", "design") {
		blast["files"] = append(blast["files"], "frontend/**", "styles/**", "css/**")
		blast["concepts"] = append(blast["concepts"], "spacing scale", "layout rhythm")
	}
	if containsAny(lower, "backend", "api", "server", "fastapi") {
		blast["files"] = append(blast["files"], "backend/**", "server/**", "mcp/**")
		blast["concepts"] = append(blast["concepts"], "backend architecture", "api routing")
	}
	for key, values := range blast {
		blast[key] = uniqueStrings(values)
	}
	return blast
}

func inferInterruptConditions(text, typ string, scope map[string][]string, aliases []string) []string {
	conditions := []string{}
	lower := strings.ToLower(text)
	if isConstraintLike(typ) {
		conditions = append(conditions, "future plan violates this constraint: "+text)
	}
	if containsAny(lower, "spacing", "layout", "ui", "design") {
		conditions = append(conditions,
			"plan changes spacing scale",
			"plan edits planner UI spacing",
			"plan introduces inconsistent gaps padding or margins",
			"plan changes layout rhythm",
		)
	}
	if containsAny(lower, "fastapi", "backend", "api", "server") {
		conditions = append(conditions,
			"plan introduces another backend framework",
			"plan replaces the existing API server",
			"plan creates a separate Express Flask Django or Node backend",
			"plan changes backend route architecture",
		)
	}
	if containsAny(lower, "claim", "tripwire", "embedding", "edge") {
		conditions = append(conditions,
			"plan changes claim retrieval",
			"plan changes tripwire interruption",
			"plan changes typed edge semantics",
			"plan changes embedding vector layout",
		)
	}
	for _, concept := range scope["concepts"] {
		conditions = append(conditions, "plan changes "+concept)
	}
	return conditions
}

func vectorTextForClaim(claim claimDraft, kind string) string {
	switch kind {
	case "claim":
		return claim.Text
	case "retrieval":
		return strings.Join(append(append([]string{}, claim.Aliases...), claim.RetrievalTriggers...), "\n")
	case "interrupt":
		var parts []string
		parts = append(parts, claim.InterruptConditions...)
		parts = append(parts, flattenStringMap(claim.BlastRadius)...)
		return strings.Join(parts, "\n")
	default:
		return claim.Text
	}
}

func buildEventText(event RuntimeEvent) string {
	var parts []string
	parts = append(parts, "Event type: "+event.EventType)
	parts = append(parts, "User request: "+event.UserRequest)
	parts = append(parts, "Agent plan: "+event.AgentPlan)
	parts = append(parts, "Files likely touched: "+strings.Join(event.FilesLikelyTouched, ", "))
	parts = append(parts, "Concepts: "+strings.Join(event.Concepts, ", "))
	parts = append(parts, "Proposed changes: "+strings.Join(event.ProposedChanges, ", "))
	parts = append(parts, "New dependencies: "+strings.Join(event.NewDependencies, ", "))
	parts = append(parts, "Risk surfaces: "+strings.Join(event.RiskSurfaces, ", "))
	return strings.Join(parts, "\n")
}

func claimMatchesScope(event RuntimeEvent, eventText string, claim claimRow) bool {
	scopeText := strings.Join(append(flattenJSONMap(claim.scopeJSON), flattenJSONMap(claim.blastRadiusJSON)...), " ")
	if tokenOverlap(eventText, scopeText) >= 1 {
		return true
	}
	if tokenOverlap(strings.Join(event.Concepts, " "), scopeText) >= 1 {
		return true
	}
	if tokenOverlap(strings.Join(event.NewDependencies, " "), scopeText) >= 1 {
		return true
	}
	files := append(stringsFromJSONMap(claim.scopeJSON, "files"), stringsFromJSONMap(claim.blastRadiusJSON, "files")...)
	return filesOverlap(event.FilesLikelyTouched, files)
}

func claimMatchesInterrupt(eventText string, claim claimRow) bool {
	interruptText := strings.Join(append(stringsFromJSONArray(claim.interruptJSON), stringsFromJSONArray(claim.retrievalJSON)...), " ")
	return tokenOverlap(eventText, interruptText) >= 1
}

func candidateLexicalBoost(queryText string, candidate claimCandidate) float64 {
	if strings.TrimSpace(queryText) == "" {
		return 0
	}
	haystack := strings.Join([]string{
		candidate.vectorText,
		candidate.claim.text,
		candidate.claim.aliasesJSON,
		candidate.claim.retrievalJSON,
		candidate.claim.interruptJSON,
	}, " ")

	boost := 0.0
	if intersects(spacingScaleSignatures(queryText), spacingScaleSignatures(haystack)) {
		boost += 0.25
	}
	if tokenOverlap(queryText, haystack) >= 3 {
		boost += 0.03
	}
	return boost
}

func spacingScaleSignatures(text string) []string {
	numbers := digitSequencePattern.FindAllString(text, -1)
	if len(numbers) < 3 {
		return nil
	}
	signatures := []string{
		strings.Join(numbers, "-"),
		strings.Join(numbers, "/"),
		strings.Join(numbers[:3], "-"),
		strings.Join(numbers[:3], "/"),
	}
	if len(numbers) >= 4 {
		signatures = append(signatures, strings.Join(numbers[:4], "-"), strings.Join(numbers[:4], "/"))
	}
	return uniqueStrings(signatures)
}

func intersects(a, b []string) bool {
	seen := make(map[string]bool, len(a))
	for _, value := range a {
		seen[value] = true
	}
	for _, value := range b {
		if seen[value] {
			return true
		}
	}
	return false
}

func hasTripwireEdge(edgeTypes []string) bool {
	for _, edgeType := range edgeTypes {
		switch edgeType {
		case "CONSTRAINS", "CONFLICTS_WITH", "SUPERSEDES", "RELATED_CANDIDATE":
			return true
		}
	}
	return false
}

func buildTripwireMessage(candidate TripwireCandidate) string {
	return fmt.Sprintf(
		"Relevant prior decision:\n- %s\n\nWhy it matters now:\n- This plan matches the claim's scope and interrupt conditions.\n\nSuggested action:\n- Continue with this context, revise the plan, or explicitly supersede the old decision.",
		candidate.Claim,
	)
}

func tokenOverlap(a, b string) int {
	left := meaningfulTokenSet(a)
	right := meaningfulTokenSet(b)
	count := 0
	for token := range left {
		if right[token] {
			count++
		}
	}
	return count
}

func meaningfulTokenSet(text string) map[string]bool {
	out := make(map[string]bool)
	for _, token := range tokenizeText(text) {
		if len(token) < 3 || stopWords[token] {
			continue
		}
		out[token] = true
	}
	return out
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "this": true, "that": true,
	"plan": true, "future": true, "work": true, "changes": true, "change": true,
	"touching": true, "touches": true, "use": true, "uses": true, "using": true,
	"add": true, "adds": true, "agent": true, "created": true, "request": true,
}

func filesOverlap(eventFiles, patterns []string) bool {
	for _, eventFile := range eventFiles {
		eventFile = filepathSlash(eventFile)
		for _, pattern := range patterns {
			pattern = filepathSlash(pattern)
			if pattern == "" {
				continue
			}
			if strings.HasSuffix(pattern, "/**") && strings.HasPrefix(eventFile, strings.TrimSuffix(pattern, "/**")+"/") {
				return true
			}
			if ok, _ := path.Match(pattern, eventFile); ok {
				return true
			}
			if strings.Contains(eventFile, strings.Trim(pattern, "*")) {
				return true
			}
		}
	}
	return false
}

func dedupeClaims(claims []claimDraft) []claimDraft {
	seen := make(map[string]bool)
	var out []claimDraft
	for _, claim := range claims {
		key := strings.Join(tokenizeText(claim.Text), " ")
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, claim)
	}
	return out
}

func sourcePriority(sourceSpan string) int {
	switch {
	case strings.HasPrefix(sourceSpan, "Key Decisions"):
		return 0
	case strings.HasPrefix(sourceSpan, "Rejected Alternatives"):
		return 1
	case strings.HasPrefix(sourceSpan, "Risks"):
		return 2
	case strings.HasPrefix(sourceSpan, "What Was Done"):
		return 3
	default:
		return 4
	}
}

func isDecisionLike(typ string) bool {
	return typ == "decision" || typ == "architecture_decision" || typ == "design_decision" ||
		typ == "constraint" || typ == "design_constraint"
}

func isConstraintLike(typ string) bool {
	return typ == "constraint" || typ == "design_constraint"
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func stringsFromJSONArray(raw string) []string {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values
	}
	return nil
}

func stringsFromJSONMap(raw, key string) []string {
	var values map[string][]string
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values[key]
	}
	return nil
}

func flattenJSONMap(raw string) []string {
	var values map[string][]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}
	return flattenStringMap(values)
}

func flattenStringMap(values map[string][]string) []string {
	var out []string
	for _, items := range values {
		out = append(out, items...)
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func makeClaimID(sessionID string, index int, text string) string {
	return "clm_" + shortHash(fmt.Sprintf("%s:%02d:%s", sessionID, index, text))
}

func makeEdgeID(fromID, toID, edgeType string) string {
	return "edge_" + shortHash(fromID+"|"+toID+"|"+edgeType)
}

func shortHash(text string) string {
	sum := sha1.Sum([]byte(text))
	return hex.EncodeToString(sum[:])[:16]
}

func filepathSlash(value string) string {
	return strings.ReplaceAll(value, "\\", "/")
}
