package context

import (
	"strings"
)

// QualityWarning describes a quality issue with a saved context field.
type QualityWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// vaguePatterns are substrings that indicate low-quality field content.
var vaguePatterns = []string{
	"fixed stuff", "updated stuff", "changed stuff",
	"fixed things", "updated things", "various changes",
	"seemed good", "looked good", "makes sense",
}

// rejectionKeywords indicate that a rejected_alternatives entry explains WHY.
var rejectionKeywords = []string{
	"because", "since", "rejected", "avoided", "too ", "would", "complex",
	"not ", "doesn't", "does not", "cannot", "can't",
}

// ValidateQuality checks a Context for common quality problems and returns
// warnings. It never blocks a save — callers decide what to do with warnings.
func ValidateQuality(ctx Context) []QualityWarning {
	var warns []QualityWarning

	// decisions: must be substantive
	d := strings.TrimSpace(ctx.KeyDecisions)
	if len(d) < 50 {
		warns = append(warns, QualityWarning{
			Field:   "decisions",
			Message: "very short — add constraint sentences (Use / Do not / Never / Prefer)",
		})
	} else {
		lower := strings.ToLower(d)
		for _, pat := range vaguePatterns {
			if strings.Contains(lower, pat) {
				warns = append(warns, QualityWarning{
					Field:   "decisions",
					Message: "contains vague language (\"" + pat + "\") — rewrite as durable constraint sentences",
				})
				break
			}
		}
	}

	// reasoning: must be more than a one-liner stub
	r := strings.TrimSpace(ctx.Reasoning)
	if len(r) < 30 {
		warns = append(warns, QualityWarning{
			Field:   "reasoning",
			Message: "very short — explain the trade-off, not just what was done",
		})
	}

	// rejected_alternatives: if present, must explain WHY
	ra := strings.TrimSpace(ctx.RejectedAlternatives)
	if ra != "" {
		lower := strings.ToLower(ra)
		hasReason := false
		for _, kw := range rejectionKeywords {
			if strings.Contains(lower, kw) {
				hasReason = true
				break
			}
		}
		if !hasReason {
			warns = append(warns, QualityWarning{
				Field:   "rejected_alternatives",
				Message: "missing explanation for WHY options were discarded — add 'because', 'rejected due to', etc.",
			})
		}
	}

	// risks: if present, must be more than a stub
	rq := strings.TrimSpace(ctx.RisksAndOpenQuestions)
	if rq != "" && len(rq) < 20 {
		warns = append(warns, QualityWarning{
			Field:   "risks",
			Message: "very short — frame as trigger → consequence: 'If X happens, Y will break'",
		})
	}

	return warns
}
