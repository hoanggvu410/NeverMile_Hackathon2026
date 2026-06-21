package context

import (
	"testing"
)

func TestValidateQuality_vague(t *testing.T) {
	ctx := Context{
		Prompt:       "do some stuff",
		Reasoning:    "ok",
		KeyDecisions: "fixed stuff",
	}
	warns := ValidateQuality(ctx)
	if len(warns) == 0 {
		t.Fatal("expected warnings for vague context, got none")
	}
	fields := map[string]bool{}
	for _, w := range warns {
		fields[w.Field] = true
	}
	if !fields["decisions"] {
		t.Error("expected warning for 'decisions' field")
	}
	if !fields["reasoning"] {
		t.Error("expected warning for 'reasoning' field")
	}
}

func TestValidateQuality_good(t *testing.T) {
	ctx := Context{
		Prompt:    "set up spacing scale for planner UI",
		Reasoning: "Chose a fixed 8pt spacing scale to prevent layout drift across planner components. Freeform spacing caused inconsistency within two PRs.",
		KeyDecisions: "- Use 4/8/16/24/32/48/64 spacing scale for all planner UI elements.\n" +
			"- Do not introduce ad-hoc margin or padding values in planner controls.\n" +
			"- Prefer gap-based layout over margin on flex containers.",
		RejectedAlternatives: "Freeform spacing rejected because it caused layout inconsistency across components within two PRs.",
		RisksAndOpenQuestions: "If a new planner control bypasses the spacing scale, layout will diverge — check spacing before any planner UI edit.",
	}
	warns := ValidateQuality(ctx)
	if len(warns) != 0 {
		t.Errorf("expected no warnings for well-written context, got: %+v", warns)
	}
}

func TestValidateQuality_rejectedWithoutReason(t *testing.T) {
	ctx := Context{
		Prompt:               "migrate auth",
		Reasoning:            "This approach is better for the team and avoids the prior issues with sessions.",
		KeyDecisions:         "Use JWT tokens for authentication. Do not store session data in the database.",
		RejectedAlternatives: "tried session-based auth, considered OAuth",
	}
	warns := ValidateQuality(ctx)
	found := false
	for _, w := range warns {
		if w.Field == "rejected_alternatives" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for rejected_alternatives missing reason")
	}
}
